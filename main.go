package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"mosaicmfg.com/stl-to-3mf/ps3mf"
)

// stl-to-3mf outpath.3mf inpath.config filamentIdsJson <models>
//
//   <models>: <model> [<model> [...]]
//   <model>:  [--colors colors.rle] [--supports supports.rle] name transforms extruder wipeIntoInfill wipeIntoModel model1.stl [...]

type Opts struct {
	Models      []ps3mf.ModelOpts
	OutPath     string
	ConfigPath  string
	FilamentIDs ps3mf.FilamentIDMap
}

func getOpts() (Opts, error) {
	argv := os.Args[1:]
	argc := len(argv)

	opts := Opts{}

	opts.OutPath = argv[0]
	opts.ConfigPath = argv[1]
	if argv[2] != "nil" {
		idsMap, err := ps3mf.UnmarshalFilamentIds(argv[2])
		if err != nil {
			return opts, err
		}
		opts.FilamentIDs = idsMap
	}
	for i := 3; i < argc; {
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
	return opts, nil
}

func run() {
	opts, err := getOpts()
	if err != nil {
		log.Fatalln(err)
	}

	bundle := ps3mf.NewBundle()
	if err = bundle.LoadConfig(opts.ConfigPath); err != nil {
		log.Fatalln(err)
	}

	modelsAdded := 0
	for _, modelOpts := range opts.Models {
		model, err := ps3mf.STLtoModel(modelOpts, opts.FilamentIDs)
		if errors.Is(err, ps3mf.ErrModelTooSmall) {
			continue
		}
		if err != nil {
			log.Fatalln(err)
		}
		bundle.AddModel(&model)
		modelsAdded++
	}

	if modelsAdded == 0 {
		_, err := fmt.Fprintf(os.Stderr, "%s", ps3mf.ErrAllModelsTooSmall)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(1)
	}

	if err = bundle.Save(opts.OutPath); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(bundle.BoundingBox.Serialize())
}

func main() {
	run()
}
