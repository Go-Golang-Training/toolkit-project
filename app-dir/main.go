package main

import "github.com/Go-Golang-Training/toolkit-project/toolkit"

func main() {
	var tools toolkit.Tools

	tools.CreateDirIfNotExist(("./test-dir"))
}
