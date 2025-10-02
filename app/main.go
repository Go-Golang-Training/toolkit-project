package main

import (
	"fmt"

	"github.com/Go-Golang-Training/toolkit-project/toolkit"
)

func main() {
	var tools toolkit.Tools

	s := tools.RandomString(10)
	fmt.Println("Random string:", s)
}
