package ps3mf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"mosaicmfg.com/stl-to-3mf/util"
)

type ModelXML struct {
	XMLName         xml.Name    `xml:"model"`
	Units           string      `xml:"unit,attr"`
	Language        string      `xml:"xml:lang,attr"`
	Namespace       string      `xml:"xmlns,attr"`
	Slic3rNamespace string      `xml:"xmlns:slic3rpe,attr"`
	Metadata        []Meta      `xml:"metadata"`
	Resources       []Resource  `xml:"resources>object"`
	Build           []BuildItem `xml:"build>item"`
}

type Meta struct {
	XMLName xml.Name `xml:"metadata"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

type Resource struct {
	XMLName xml.Name `xml:"object"`
	Id      string   `xml:"id,attr"`
	Type    string   `xml:"type,attr,omitempty"`
	Mesh    Mesh     `xml:"mesh"`
}

type Mesh struct {
	XMLName   xml.Name   `xml:"mesh"`
	Vertices  []Vertex   `xml:"vertices>vertex"`
	Triangles []Triangle `xml:"triangles>triangle"`
}

type Vertex struct {
	XMLName xml.Name `xml:"vertex"`
	X       float64  `xml:"x,attr"`
	Y       float64  `xml:"y,attr"`
	Z       float64  `xml:"z,attr"`
}

func (v Vertex) Transform(m util.Matrix4) Vertex {
	vec := util.NewVector3(v.X, v.Y, v.Z)
	vec.TransformInPlace(m)
	return Vertex{
		XMLName: v.XMLName,
		X:       vec.Vector[0],
		Y:       vec.Vector[1],
		Z:       vec.Vector[2],
	}
}

type Triangle struct {
	XMLName        xml.Name `xml:"triangle"`
	V1             int      `xml:"v1,attr"`
	V2             int      `xml:"v2,attr"`
	V3             int      `xml:"v3,attr"`
	Segmentation   string   `xml:"slic3rpe:mmu_segmentation,attr,omitempty"`
	CustomSupports string   `xml:"slic3rpe:custom_supports,attr,omitempty"`
}

type BuildItem struct {
	XMLName   xml.Name `xml:"item"`
	ObjectId  string   `xml:"objectid,attr"`
	Transform string   `xml:"transform,attr,omitempty"`
	Printable string   `xml:"printable,attr,omitempty"`
}

func (m *ModelXML) MergeMeshes(matrices []util.Matrix4) []IdPair {
	idPairs := make([]IdPair, 0, len(m.Resources))
	idPairs = append(idPairs, IdPair{
		FirstId: 0,
		LastId:  len(m.Resources[0].Mesh.Triangles) - 1,
	})

	currentVertCount := len(m.Resources[0].Mesh.Vertices)
	currentTriCount := len(m.Resources[0].Mesh.Triangles)

	// apply transformations to first mesh's vertices
	for vertIdx, vert := range m.Resources[0].Mesh.Vertices {
		transformedVert := vert.Transform(matrices[0])
		m.Resources[0].Mesh.Vertices[vertIdx] = transformedVert
	}

	for i := 1; i < len(m.Resources); i++ {
		mesh := m.Resources[i].Mesh
		// apply transformations to this mesh's vertices
		for _, vert := range mesh.Vertices {
			transformedVert := vert.Transform(matrices[i])
			m.Resources[0].Mesh.Vertices = append(m.Resources[0].Mesh.Vertices, transformedVert)
		}
		for _, tri := range mesh.Triangles {
			tri.V1 += currentVertCount
			tri.V2 += currentVertCount
			tri.V3 += currentVertCount
			m.Resources[0].Mesh.Triangles = append(m.Resources[0].Mesh.Triangles, tri)
		}
		idPairs = append(idPairs, IdPair{
			FirstId: currentTriCount,
			LastId:  currentTriCount + len(mesh.Triangles) - 1,
		})
		currentVertCount += len(mesh.Vertices)
		currentTriCount += len(mesh.Triangles)
	}

	m.Resources = m.Resources[:1]
	m.Resources[0].Type = "model"

	m.Build = m.Build[:1]
	// use identity matrix since vertices are already transformed
	m.Build[0].Transform = "1 0 0 0 0 1 0 0 0 0 1 0"

	return idPairs
}

func (m *Mesh) AddColors(rle *util.RLE) {
	color := *rle
	currentRunIndex := -1
	currentRunLength := 0
	currentColor := uint8(0)

	for triIdx := range m.Triangles {
		if currentRunLength <= 0 {
			if currentRunIndex < len(color.Runs) {
				currentRunIndex++
				currentRunLength = int(color.Runs[currentRunIndex].Length)
				currentColor = color.Runs[currentRunIndex].Value
			}
		}
		// 1 ->  8 -> 0000 1000
		// 2 -> 0C -> 0000 1100
		// 3 -> 1C -> 0001 1100
		// 4 -> 2C -> 0010 1100
		// 5 -> 3C -> 0011 1100
		// 6 -> 4C -> 0100 1100
		// 7 -> 5C -> 0101 1100
		// 8 -> 6C -> 0110 1100
		// ...
		if currentColor == 1 {
			m.Triangles[triIdx].Segmentation = "8"
		} else if currentColor > 1 {
			m.Triangles[triIdx].Segmentation = fmt.Sprintf("%xC", currentColor-2)
		}
		currentRunLength--
	}
}

func (m *Mesh) AddCustomSupports(rle *util.RLE) {
	color := *rle
	currentRunIndex := -1
	currentRunLength := 0
	currentSupported := 0

	for triIdx := range m.Triangles {
		if currentRunLength <= 0 {
			if currentRunIndex < len(color.Runs) {
				currentRunIndex++
				currentRunLength = int(color.Runs[currentRunIndex].Length)
				currentSupported = int(color.Runs[currentRunIndex].Value)
			}
		}
		// enforce support (currentSupported == 1): "4"
		// block support (currentSupported == 0): "8"
		if currentSupported > 0 {
			m.Triangles[triIdx].CustomSupports = "4"
		} else {
			m.Triangles[triIdx].CustomSupports = "8"
		}
		currentRunLength--
	}
}

func GetMeta(name, value string) Meta {
	return Meta{
		Name:  name,
		Value: value,
	}
}

func (m *ModelXML) Marshal() ([]byte, error) {
	output, marshalErr := xml.MarshalIndent(m, "", " ")
	if marshalErr != nil {
		return nil, marshalErr
	}
	// prepend header
	output = append([]byte(xml.Header), output...)
	// replace self-closing tags
	output = bytes.ReplaceAll(output, []byte("></vertex>"), []byte("/>"))
	output = bytes.ReplaceAll(output, []byte("></triangle>"), []byte("/>"))
	output = bytes.ReplaceAll(output, []byte("></item>"), []byte("/>"))
	// add trailing newline
	output = append(output, '\n')
	return output, nil
}
