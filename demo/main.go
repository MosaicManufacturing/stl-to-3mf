package main

import (
	"../ps3mf"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func demo() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(dir, "..", "test", "Slic3r_PE.config")
	outPath := filepath.Join(dir, "..", "test", "output3mf.zip")
	stlPath := filepath.Join(dir, "..", "test", "cube.stl")
	colorsPath := filepath.Join(dir, "..", "test", "colors.rle")
	transforms1 := "2,0,0,0|0,2,0,0|0,0,2,0|0,0,0,1"
	transforms2 := "1,0,0,0|0,1,0,0|0,0,1,0|50,60,70,1"

	bundle := ps3mf.NewBundle()
	if err := bundle.LoadConfig(configPath); err != nil {
		log.Fatalln(err)
	}

	model1, err := ps3mf.STLtoModel(stlPath, transforms1, colorsPath, "")
	if err != nil {
		log.Fatalln(err)
	}
	model2, err := ps3mf.STLtoModel(stlPath, transforms2, colorsPath, "")
	if err != nil {
		log.Fatalln(err)
	}
	bundle.AddModel(&model1)
	bundle.AddModel(&model2)
	fmt.Println(bundle.BoundingBox.Serialize())

	if err := bundle.Save(outPath); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	demo()
}