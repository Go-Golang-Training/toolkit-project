package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestTools_PushJSONToRemote(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test Request Parameters
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
			Header:     make(http.Header),
		}
	})

	var testTools Tools
	var foo struct {
		Bar string `json:"bar"`
	}
	foo.Bar = "bar"

	_, _, err := testTools.PushJSONToRemote("http://example.com/some/path", foo, client)
	if err != nil {
		t.Error("failed to call remote url:", err)
	}
}

func TestTools_RandomString(t *testing.T) {

	var tools Tools

	s := tools.RandomString(10)
	if len(s) != 10 {
		t.Errorf("Expected string length of 10, but got %d", len(s))
	}

	s2 := tools.RandomString(10)
	if s == s2 {
		t.Errorf("Expected different random strings, but got the same: %s", s)
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed with rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed file type", allowedTypes: []string{"image/png"}, renameFile: false, errorExpected: true},
	{name: "no allowed types", allowedTypes: []string{}, renameFile: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// Setup a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()

			/// create a form data field 'file'
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Errorf("%s: Error creating form file: %v", e.name, err)
				return
			}

			f, err := os.Open("./testdata/img.png")

			if err != nil {
				t.Errorf("%s: Error opening file: %v", e.name, err)
				return
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Errorf("%s: Error decoding image: %v", e.name, err)
				return
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Errorf("%s: Error encoding image to PNG: %v", e.name, err)
				return
			}
		}()

		// read from the pipe which receives data from the goroutine
		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: Unexpected error: %v", e.name, err)
			return
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: File was not uploaded: %s", e.name, err.Error())
			}
			// clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: No expected error but did get one: %v", e.name, err)
			return
		}

		wg.Wait()

	}
}

func TestTools_UploadOneFile(t *testing.T) {

	// Setup a pipe to avoid buffering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()

		/// create a form data field 'file'
		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Errorf("Error creating form file: %v", err)
			return
		}

		f, err := os.Open("./testdata/img.png")

		if err != nil {
			t.Errorf("Error opening file: %v", err)
			return
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			t.Errorf("Error decoding image: %v", err)
			return
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Errorf("Error encoding image to PNG: %v", err)
			return
		}
	}()

	// read from the pipe which receives data from the goroutine
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools

	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("File was not uploaded: %s", err.Error())
	}
	// clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))

}

func TestTools_CreateDirIfNotExist(t *testing.T) {

	var testTool Tools

	//Try to create a directory when it does not exist
	err := testTool.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	//Try to create a directory when it already exists
	err = testTool.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	//clean up
	err = os.Remove("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

}

// Table test for Slugify
var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "now is the time", expected: "now-is-the-time", errorExpected: false},
	{name: "empty string", s: "", expected: "", errorExpected: true},
	{name: "complex string", s: "Now??!! is the time 5:45PM ++++***SQL Injection.SQL", expected: "now-is-the-time-5-45pm-sql-injection-sql", errorExpected: false},
	{name: "arabic string", s: "دابا هو الوقت للمرح", expected: "", errorExpected: true},
	{name: "japanese string", s: "今こそ楽しむ時です", expected: "", errorExpected: true},
	{name: "japanese, arabic and roman characters", s: "今こそ楽しむ時です دابا هو الوقت للمرح now is the time for fun", expected: "now-is-the-time-for-fun", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {

	var testTool Tools

	// Try to create a slug from a weird string
	_, err := testTool.Slugify("Hello World from Go Slugify test function!!!")
	if err != nil {
		t.Error(err)
	}

	for _, e := range slugTests {
		slug, err := testTool.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error received when none expected: %s", e.name, err.Error())
		}
		if !e.errorExpected && slug != e.expected {
			t.Errorf("%s: wrong slug returned; expected %s, but got %s", e.name, e.expected, slug)
		}
		if err == nil && e.errorExpected {
			t.Errorf("%s: no error received when expected", e.name)
		}

	}

}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var testTool Tools

	testTool.DownloadStaticFile(rr, req, "./testdata", "pic.jpg", "puppy.jpg")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "98827" {

		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachment; filename=\"puppy.jpg\"" {
		t.Error("wrong content disposition", res.Header["Content-Disposition"][0])
	}

	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
}{
	{name: "good json", json: `{"foo": "bar"}`, errorExpected: false, maxSize: 1024, allowUnknown: false},
	{name: "badly formatted json", json: `{"foo":}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type", json: `{"foo": 1}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "two json files", json: `{"foo": "1"}{"alpha": "beta"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "empty body", json: ``, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "syntax error in json", json: `{"foo": 1"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "unknown field in json", json: `{"fooo": "1"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "allow unknown fields in json", json: `{"fooo": "1"}`, errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "missing field name", json: `{jack: "1"}`, errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "file too large", json: `{"foo": "bar"}`, errorExpected: true, maxSize: 5, allowUnknown: true},
	{name: "not json", json: `Hello, world!`, errorExpected: true, maxSize: 1024, allowUnknown: true},
}

func TestTools_ReadJSON(t *testing.T) {

	var testTool Tools

	for _, e := range jsonTests {
		// set the max file size
		testTool.MaxJSONSize = e.maxSize

		// allow/disallow unknown fields
		testTool.AllowUnknownFields = e.allowUnknown

		// declare a variable to read the decoded json into
		var decodedJSON struct {
			Foo string `json:"foo"`
		}

		// create a request with the body
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Log("Error:", err)
		}

		// create a recorder
		rr := httptest.NewRecorder()

		err = testTool.ReadJSON(rr, req, &decodedJSON)

		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected, but none received", e.name)
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error not expected, but one received: %s", e.name, err.Error())
		}
	}
}

func TestTools_WriteJSON(t *testing.T) {
	var testTools Tools

	rr := httptest.NewRecorder()
	payload := JSONResponse{
		Error:   false,
		Message: "foo",
	}

	headers := make(http.Header)
	headers.Add("FOO", "BAR")

	err := testTools.WriteJSON(rr, http.StatusOK, payload, headers)
	if err != nil {
		t.Errorf("failed to write JSON: %v", err)
	}
}

func TestTools_ErrorJSON(t *testing.T) {
	var testTools Tools

	rr := httptest.NewRecorder()
	err := testTools.ErrorJSON(rr, errors.New("some error"), http.StatusServiceUnavailable)
	if err != nil {
		t.Error(err)
	}

	var payload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		t.Error("received error when decoding JSON", err)
	}

	if !payload.Error {
		t.Error("error set to false in JSON, and it should be true")
	}

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("wrong status code returned; expected 503, but got %d", rr.Code)
	}
}
