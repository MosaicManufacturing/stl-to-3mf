package main

import (
	"fmt"
	"log"
	"mosaicmfg.com/stl-to-3mf/ps3mf"
	"os"
)

// stl-to-3mf outpath.3mf inpath.config <models>
//
//   <models>: <model> [<model> [...]]
//   <model>:  [--colors colors.rle] [--supports supports.rle] name transforms extruder wipeIntoInfill wipeIntoModel model1.stl [...]

type Opts struct {
	Models []ps3mf.ModelOpts
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
		modelOpts := ps3mf.ModelOpts{}
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
		modelOpts.Name = argv[i]
		i++
		modelOpts.Transforms = argv[i]
		i++
		modelOpts.Extruder = argv[i]
		i++
		modelOpts.WipeIntoInfill = argv[i] == "1"
		i++
		modelOpts.WipeIntoModel = argv[i] == "1"
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
		model, err := ps3mf.STLtoModel(modelOpts)
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
