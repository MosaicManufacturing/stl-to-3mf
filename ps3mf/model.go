package ps3mf

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"github.com/hpinc/go3mf"
	"github.com/hpinc/go3mf/importer/stl"
	"github.com/hpinc/go3mf/spec"
	"io"
	"io/ioutil"
	"mosaicmfg.com/stl-to-3mf/util"
	"os"
)

type ModelOpts struct {
	Name           string
	ColorsPath     string
	SupportsPath   string
	MeshPath       string
	Transforms     string // serialized util.Matrix4
	Extruder       string // 1-indexed
	WipeIntoInfill bool
	WipeIntoModel  bool
}

type Model struct {
	Name           string
	Model          *go3mf.Model
	Transforms     util.Matrix4
	Colors         *util.RLE
	Supports       *util.RLE
	Extruder       string
	WipeIntoInfill bool
	WipeIntoModel  bool
}

type xmlns struct {
	Value string
}

func (n xmlns) Marshal3MFAttr(spec.Encoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{
			Name: xml.Name{
				Space: "xmlns",
				Local: "slic3rpe",
			},
			Value: n.Value,
		},
	}, nil
}

const slic3rPENamespace = "http://schemas.slic3r.org/3mf/2017/06"

// minimum size of a closed mesh in binary STL is 284 bytes,
// but go3mf incorrectly notes that the minimum is 384 bytes,
// and requires at least 300 bytes in each STL file
const go3mfMinHeaderSize = 300

func getSlicerPENamespace() spec.MarshalerAttr {
	return xmlns{slic3rPENamespace}
}

func getMetadataElement(name, value string) go3mf.Metadata {
	return go3mf.Metadata{
		Name:  xml.Name{Local: name},
		Value: value,
	}
}

func addDefaultMetadata(model *go3mf.Model) {
	model.Language = "en-US"
	model.AnyAttr = append(model.AnyAttr, getSlicerPENamespace())
	model.Metadata = append(model.Metadata, getMetadataElement("slic3rpe:Version3mf", "1"))
	model.Metadata = append(model.Metadata, getMetadataElement("Application", "Canvas"))
}

func STLtoModel(opts ModelOpts, filamentIds map[byte]byte) (Model, error) {
	model := Model{
		Name:           opts.Name,
		Model:          new(go3mf.Model),
		Transforms:     util.Matrix4{},
		Colors:         nil,
		Supports:       nil,
		Extruder:       opts.Extruder,
		WipeIntoInfill: opts.WipeIntoInfill,
		WipeIntoModel:  opts.WipeIntoModel,
	}

	// load the STL file using 3MF conversion
	file, err := os.Open(opts.MeshPath)
	if err != nil {
		return model, err
	}

	var reader io.Reader

	stat, err := file.Stat()
	if err != nil {
		return model, err
	}
	fileSize := stat.Size() // bytes
	if fileSize < go3mfMinHeaderSize {
		// handle STL files too small for go3mf to process (even if they contain
		// enough triangles to be valid) by padding the file with 0s to achieve
		// the minimum length go3mf requires
		content, err := ioutil.ReadAll(file)
		if err != nil {
			return model, err
		}
		content = append(content, make([]byte, go3mfMinHeaderSize-len(content))...)
		reader = bytes.NewReader(content)
	} else {
		reader = bufio.NewReader(file)
	}

	decoder := stl.NewDecoder(reader)
	if err = decoder.Decode(model.Model); err != nil {
		return model, err
	}

	// add stock PS metadata
	addDefaultMetadata(model.Model)

	// decode transforms matrix
	matrix, err := util.UnserializeMatrix4(opts.Transforms)
	if err != nil {
		return model, err
	}
	model.Transforms = matrix

	// load RLE data
	if opts.ColorsPath != "" {
		colors, err := util.LoadRLE(opts.ColorsPath, filamentIds)
		if err != nil {
			return model, err
		}
		model.Colors = colors
	}
	if opts.SupportsPath != "" {
		supports, err := util.LoadRLE(opts.SupportsPath, nil)
		if err != nil {
			return model, err
		}
		model.Supports = supports
	}

	if err = file.Close(); err != nil {
		return model, err
	}
	return model, nil
}

func (m *Model) GetTransformedBbox() util.BoundingBox {
	bbox := util.NewBoundingBox()
	for _, vertex := range m.Model.Resources.Objects[0].Mesh.Vertices {
		point := util.NewVector3(float64(vertex[0]), float64(vertex[1]), float64(vertex[2])).Transform(m.Transforms)
		bbox.ExpandByPoint(point)
	}
	return bbox
}
