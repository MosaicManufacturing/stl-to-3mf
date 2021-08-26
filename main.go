package main

import (
	"github.com/hpinc/go3mf"
	"log"
)

func main() {
	model := new(go3mf.Model)
	writer, err := go3mf.CreateWriter("/Users/brandonbloch/Desktop/3mf.zip")
	if err != nil {
		log.Fatalln(err)
	}
	if err := writer.Encode(model); err != nil {
		log.Fatalln(err)
	}
	if err := writer.Close(); err != nil {
		log.Fatalln(err)
	}
}
