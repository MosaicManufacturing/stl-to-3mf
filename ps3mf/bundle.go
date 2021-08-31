package ps3mf

import (
	"../util"
	"encoding/xml"
	"github.com/hpinc/go3mf"
	"github.com/hpinc/go3mf/spec"
	"io/ioutil"
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
	Model *go3mf.Model
	Colors []*util.RLE // nil for objects with no data
	Supports []*util.RLE // nil for objects with no data
	BoundingBox util.BoundingBox

	Config string
}

func NewBundle() Bundle {
	return Bundle{
		Model:       new(go3mf.Model),
		Colors:      make([]*util.RLE, 0),
		Supports:    make([]*util.RLE, 0),
		Config:      "",
	}
}

func (m *Bundle) LoadConfig(path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	m.Config = string(bytes)
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
			Name:  xml.Name{
				Local: "printable",
			},
			Value: val,
		},
	}, nil
}

func getPrintableAttr(printable bool) spec.MarshalerAttr {
	return printableAttr{printable}
}

func (m *Bundle) AddModel(model *Model) {
	objectId := uint32(len(m.Model.Resources.Objects) + 1)
	model.Model.Resources.Objects[0].ID = objectId
	model.Model.Build.Items[0].ObjectID = objectId
	model.Model.Build.Items[0].Transform = model.Transforms.To3MF()
	model.Model.Build.Items[0].AnyAttr = append(model.Model.Build.Items[0].AnyAttr, getPrintableAttr(true))

	m.Model.Resources.Objects = append(m.Model.Resources.Objects, model.Model.Resources.Objects[0])
	m.Model.Build.Items = append(m.Model.Build.Items, model.Model.Build.Items[0])
	m.Colors = append(m.Colors, model.Colors)
	m.Supports = append(m.Supports, model.Supports)

	m.BoundingBox.ExpandByBox(model.GetTransformedBbox())
}
