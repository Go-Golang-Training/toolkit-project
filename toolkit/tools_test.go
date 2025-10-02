package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

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
