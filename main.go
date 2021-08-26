package main

import (
	"bufio"
	"github.com/hpinc/go3mf"
	"github.com/hpinc/go3mf/importer/stl"
	"log"
	"os"
)

func loadSTL(path string) (model *go3mf.Model, err error) {
	model = new(go3mf.Model)
	file, openErr := os.Open(path)
	if openErr != nil {
		err = openErr
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	reader := bufio.NewReader(file)
	decoder := stl.NewDecoder(reader)
	if decodeErr := decoder.Decode(model); decodeErr != nil {
		err = decodeErr
		return
	}
	return
}

func main() {
	model, err := loadSTL("/Users/brandonbloch/Downloads/cube.stl")
	if err != nil {
		log.Fatalln(err)
	}
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
