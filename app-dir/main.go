package main

import "github.com/Go-Golang-Training/toolkit"

func main() {
	var tools toolkit.Tools

	tools.CreateDirIfNotExist(("./test-dir"))
}
