package main

import (
	"log"

	"github.com/Go-Golang-Training/toolkit-project/toolkit"
)

func main() {
	var toSlug = "NOW??!! is the tImE 123"

	var tools toolkit.Tools

	slugified, err := tools.Slugify(toSlug)
	if err != nil {
		log.Println(err)

	}

	log.Println(slugified)
}
