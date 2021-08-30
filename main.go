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

func run() {
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

func main() {
	run()
}
