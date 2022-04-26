package ps3mf

import (
	"encoding/xml"
	"github.com/hpinc/go3mf"
	"github.com/hpinc/go3mf/spec"
	"io/ioutil"
	"mosaicmfg.com/stl-to-3mf/util"
)

// structure of PrusaSlicer 3MF file:
//   /_rels/
//     (empty dir)
//   /[Content_Types].xml
//     <?xml version="1.0" encoding="UTF-8"?>
//     <Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
//      <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml" />
//      <Default Extension="model" ContentType="application/vnd.ms-package.3dmanufacturing-3dmodel+xml" />
//      <Default Extension="png" ContentType="image/png" />
//     </Types>
//   /3D
//     /3dmodel.model
//       (contains information about all meshes, including transforms and color/support data)
//   /Metadata
//     /Slic3r_PE_model.config
//     /Slic3r_PE.config
//     /thumbnail.png

type Bundle struct {
	Names          []string
	Model          *go3mf.Model
	Matrices       []util.Matrix4
	Colors         []*util.RLE // nil for objects with no data
	Supports       []*util.RLE // nil for objects with no data
	Extruders      []string    // 1-indexed ints
	WipeIntoInfill []bool
	WipeIntoModel  []bool
	BoundingBox    util.BoundingBox

	Config []byte
}

func NewBundle() Bundle {
	return Bundle{
		Names:          make([]string, 0),
		Model:          new(go3mf.Model),
		Matrices:       make([]util.Matrix4, 0),
		Colors:         make([]*util.RLE, 0),
		Supports:       make([]*util.RLE, 0),
		Extruders:      make([]string, 0),
		WipeIntoInfill: make([]bool, 0),
		WipeIntoModel:  make([]bool, 0),
		BoundingBox:    util.NewBoundingBox(),
	}
}

func (b *Bundle) LoadConfig(path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	b.Config = bytes
	return nil
}

type printableAttr struct {
	Value bool
}

func (att printableAttr) Marshal3MFAttr(spec.Encoder) ([]xml.Attr, error) {
	val := "0"
	if att.Value {
		val = "1"
	}
	return []xml.Attr{
		{
			Name: xml.Name{
				Local: "printable",
			},
			Value: val,
		},
	}, nil
}

func getPrintableAttr(printable bool) spec.MarshalerAttr {
	return printableAttr{printable}
}

func (b *Bundle) AddModel(model *Model) {
	objectId := uint32(len(b.Model.Resources.Objects) + 1)
	model.Model.Resources.Objects[0].ID = objectId
	model.Model.Build.Items[0].ObjectID = objectId
	model.Model.Build.Items[0].Transform = model.Transforms.To3MF()
	model.Model.Build.Items[0].AnyAttr = append(model.Model.Build.Items[0].AnyAttr, getPrintableAttr(true))

	b.Model.Resources.Objects = append(b.Model.Resources.Objects, model.Model.Resources.Objects[0])
	b.Model.Build.Items = append(b.Model.Build.Items, model.Model.Build.Items[0])
	b.Names = append(b.Names, model.Name)
	b.Matrices = append(b.Matrices, model.Transforms)
	b.Colors = append(b.Colors, model.Colors)
	b.Supports = append(b.Supports, model.Supports)
	b.Extruders = append(b.Extruders, model.Extruder)
	b.WipeIntoInfill = append(b.WipeIntoInfill, model.WipeIntoInfill)
	b.WipeIntoModel = append(b.WipeIntoModel, model.WipeIntoModel)

	b.BoundingBox.ExpandByBox(model.GetTransformedBbox())
}
