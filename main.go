package main

import (
	"./ps3mf"
	"./util"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func help() {
	fmt.Println("stl-to-3mf <models> outpath.3mf")
	fmt.Println()
	fmt.Println("  <models>: [--colors, colors.rle] [--supports supports.rle] matrix model1.stl [...]")
}

// stl-to-3mf outpath.3mf inpath.config <models>
//
//   <models>: <model> [<model> [...]]
//   <model>:  [--colors colors.rle] [--supports supports.rle] transforms model1.stl [...]

type ModelOpts struct {
	ColorsPath string
	SupportsPath string
	MeshPath string
	Transforms util.Matrix4
}

type Opts struct {
	Models []ModelOpts
	OutPath string
	ConfigPath string
}

func getOpts() Opts {
	argv := os.Args[1:]
	argc := len(argv)

	opts := Opts{}

	opts.OutPath = argv[0]
	opts.ConfigPath = argv[1]
	for i := 2; i < argc; {
		modelOpts := ModelOpts{}
		if argv[i] == "--colors" {
			i++
			modelOpts.ColorsPath = argv[i]
			i++
		}
		if argv[i] == "--supports" {
			i++
			modelOpts.SupportsPath = argv[i]
			i++
		}
		mat, err := util.UnserializeMatrix4(argv[i])
		if err != nil {
			log.Fatalln(err)
		}
		modelOpts.Transforms = mat
		i++
		modelOpts.MeshPath = argv[i]
		i++
		opts.Models = append(opts.Models, modelOpts)
	}
	return opts
}

func main() {
	//opts := getOpts()

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(dir, "test", "Slic3r_PE.config")
	outPath := filepath.Join(dir, "test", "output3mf.zip")
	stlPath := filepath.Join(dir, "test", "cube.stl")
	colorsPath := filepath.Join(dir, "test", "colors.rle")
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

	if err := bundle.Save(outPath); err != nil {
		log.Fatalln(err)
	}

	//writer, err := go3mf.CreateWriter(outPath)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//if err := writer.Encode(bundle.Model); err != nil {
	//	log.Fatalln(err)
	//}
	//if err := writer.Close(); err != nil {
	//	log.Fatalln(err)
	//}
}
