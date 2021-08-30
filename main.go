package main

import (
	"./ps3mf"
	"fmt"
	"log"
	"os"
)

// stl-to-3mf outpath.3mf inpath.config <models>
//
//   <models>: <model> [<model> [...]]
//   <model>:  [--colors colors.rle] [--supports supports.rle] transforms model1.stl [...]

type ModelOpts struct {
	ColorsPath string
	SupportsPath string
	MeshPath string
	Transforms string // serialized util.Matrix4
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
		modelOpts.Transforms = argv[i]
		i++
		modelOpts.MeshPath = argv[i]
		i++
		opts.Models = append(opts.Models, modelOpts)
	}
	return opts
}

// todo: remove
//func demo() {
//	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	configPath := filepath.Join(dir, "test", "Slic3r_PE.config")
//	outPath := filepath.Join(dir, "test", "output3mf.zip")
//	stlPath := filepath.Join(dir, "test", "cube.stl")
//	colorsPath := filepath.Join(dir, "test", "colors.rle")
//	transforms1 := "2,0,0,0|0,2,0,0|0,0,2,0|0,0,0,1"
//	transforms2 := "1,0,0,0|0,1,0,0|0,0,1,0|50,60,70,1"
//
//	bundle := ps3mf.NewBundle()
//	if err := bundle.LoadConfig(configPath); err != nil {
//		log.Fatalln(err)
//	}
//
//	model1, err := ps3mf.STLtoModel(stlPath, transforms1, colorsPath, "")
//	if err != nil {
//		log.Fatalln(err)
//	}
//	model2, err := ps3mf.STLtoModel(stlPath, transforms2, colorsPath, "")
//	if err != nil {
//		log.Fatalln(err)
//	}
//	bundle.AddModel(&model1)
//	bundle.AddModel(&model2)
//	fmt.Println(bundle.BoundingBox.Serialize())
//
//	if err := bundle.Save(outPath); err != nil {
//		log.Fatalln(err)
//	}
//}

func main() {
	//demo()
	opts := getOpts()

	bundle := ps3mf.NewBundle()
	if err := bundle.LoadConfig(opts.ConfigPath); err != nil {
		log.Fatalln(err)
	}

	for _, modelOpts := range opts.Models {
		model, err := ps3mf.STLtoModel(modelOpts.MeshPath, modelOpts.Transforms, modelOpts.ColorsPath, modelOpts.SupportsPath)
		if err != nil {
			log.Fatalln(err)
		}
		bundle.AddModel(&model)
	}

	if err := bundle.Save(opts.OutPath); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(bundle.BoundingBox.Serialize())
}
