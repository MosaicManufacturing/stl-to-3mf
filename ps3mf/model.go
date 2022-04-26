package ps3mf

import (
	"bufio"
	"encoding/xml"
	"github.com/hpinc/go3mf"
	"github.com/hpinc/go3mf/importer/stl"
	"github.com/hpinc/go3mf/spec"
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

func STLtoModel(opts ModelOpts) (model Model, err error) {
	model = Model{
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
	file, openErr := os.Open(opts.MeshPath)
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
	if decodeErr := decoder.Decode(model.Model); decodeErr != nil {
		err = decodeErr
		return
	}

	// add stock PS metadata
	addDefaultMetadata(model.Model)

	// decode transforms matrix
	matrix, matrixErr := util.UnserializeMatrix4(opts.Transforms)
	if matrixErr != nil {
		err = matrixErr
		return
	}
	model.Transforms = matrix

	// load RLE data
	if opts.ColorsPath != "" {
		colors, colorsErr := util.LoadRLE(opts.ColorsPath)
		if colorsErr != nil {
			err = colorsErr
			return
		}
		model.Colors = colors
	}
	if opts.SupportsPath != "" {
		supports, supportsErr := util.LoadRLE(opts.SupportsPath)
		if supportsErr != nil {
			err = supportsErr
			return
		}
		model.Supports = supports
	}

	return
}

func (m *Model) GetTransformedBbox() util.BoundingBox {
	bbox := util.NewBoundingBox()
	for _, vertex := range m.Model.Resources.Objects[0].Mesh.Vertices {
		point := util.NewVector3(float64(vertex[0]), float64(vertex[1]), float64(vertex[2])).Transform(m.Transforms)
		bbox.ExpandByPoint(point)
	}
	return bbox
}
