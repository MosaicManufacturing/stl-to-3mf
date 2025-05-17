package ps3mf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

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

type MergedVolumesInfo struct {
	VolumeIdPairs  []MeshTriangleRange
	VolumeNames    []string
	Extruders      []string // 1-indexed ints
	WipeIntoInfill []bool
	WipeIntoModel  []bool
	BoundingBox    util.BoundingBox
}

type Group struct {
	resources        Resource
	build            BuildItem
	volumeIdPairs    []MeshTriangleRange
	volumeNames      []string
	currentVertCount int
	currentTriCount  int
	Extruders        []string // 1-indexed ints
	WipeIntoInfill   []bool
	WipeIntoModel    []bool
	BoundingBox      util.BoundingBox
}

// creates and initializes a new Group
func newGroup(
	resource Resource,
	volumeName string,
	buildItem BuildItem,
	extruder string,
	wipeIntoInfill bool,
	wipeIntoModel bool,
	boundingBox util.BoundingBox,
) Group {
	group := Group{
		resources:   resource,
		volumeNames: []string{volumeName},
		volumeIdPairs: []MeshTriangleRange{
			{
				FirstId: 0,
				LastId:  len(resource.Mesh.Triangles) - 1,
			},
		},
		currentVertCount: len(resource.Mesh.Vertices),
		currentTriCount:  len(resource.Mesh.Triangles),
		Extruders:        []string{extruder},
		WipeIntoInfill:   []bool{wipeIntoInfill},
		WipeIntoModel:    []bool{wipeIntoModel},
		BoundingBox:      boundingBox,
	}

	// set common properties
	group.resources.Type = "model"
	group.build = buildItem
	// use identity matrix since vertices are already transformed
	group.build.Transform = "1 0 0 0 0 1 0 0 0 0 1 0"
	return group
}

// updateGroupWithMesh adds a new mesh to an existing group,
// updating all indices appropriately
func updateGroupWithMesh(
	group *Group,
	resource Resource,
	volumeName string,
	extruder string,
	wipeIntoInfill bool,
	wipeIntoModel bool,
) {
	// add vertices (with correct offsets for triangles)
	group.resources.Mesh.Vertices = append(group.resources.Mesh.Vertices, resource.Mesh.Vertices...)

	// add triangles with updated vertex indices
	for _, tri := range resource.Mesh.Triangles {
		modifiedTri := Triangle{
			XMLName:        tri.XMLName,
			V1:             tri.V1 + group.currentVertCount,
			V2:             tri.V2 + group.currentVertCount,
			V3:             tri.V3 + group.currentVertCount,
			Segmentation:   tri.Segmentation,
			CustomSupports: tri.CustomSupports,
		}
		group.resources.Mesh.Triangles = append(group.resources.Mesh.Triangles, modifiedTri)
	}

	// update group metadata
	group.volumeIdPairs = append(group.volumeIdPairs, MeshTriangleRange{
		FirstId: group.currentTriCount,
		LastId:  group.currentTriCount + len(resource.Mesh.Triangles) - 1,
	})

	// Update counters
	group.currentVertCount += len(resource.Mesh.Vertices)
	group.currentTriCount += len(resource.Mesh.Triangles)

	// Update metadata arrays
	group.volumeNames = append(group.volumeNames, volumeName)
	group.Extruders = append(group.Extruders, extruder)
	group.WipeIntoInfill = append(group.WipeIntoInfill, wipeIntoInfill)
	group.WipeIntoModel = append(group.WipeIntoModel, wipeIntoModel)
}

func (m *ModelXML) MergeGroupMeshes(bundle *Bundle) ([]MergedVolumesInfo, error) {
	// input validation
	if len(m.Resources) == 0 {
		return []MergedVolumesInfo{}, nil
	}

	// transform all vertices
	for i, resource := range m.Resources {
		for vertIdx, vert := range resource.Mesh.Vertices {
			transformedVert := vert.Transform(bundle.Matrices[i])
			resource.Mesh.Vertices[vertIdx] = transformedVert
		}
	}

	groups := make(map[string]Group)
	for i, currResource := range m.Resources {
		// group them based on path name
		splitNames := strings.Split(bundle.Paths[i], "|")

		if len(splitNames) != 2 {
			return nil, fmt.Errorf("invalid path format: %s (expected 'groupName|volumeName')", bundle.Paths[i])
		}

		groupName := splitNames[0]
		volumeName := splitNames[1]

		if groupName == "" {
			groups[fmt.Sprintf("%d", i)] = newGroup(
				currResource,
				volumeName,
				m.Build[i],
				bundle.Extruders[i],
				bundle.WipeIntoInfill[i],
				bundle.WipeIntoModel[i],
				bundle.BoundingBox,
			)
		} else {
			group, groupAlreadyCreated := groups[groupName]
			if !groupAlreadyCreated {
				groups[groupName] = newGroup(
					currResource,
					volumeName,
					m.Build[i],
					bundle.Extruders[i],
					bundle.WipeIntoInfill[i],
					bundle.WipeIntoModel[i],
					bundle.BoundingBox,
				)
			} else {
				// add the new mesh to the existing group
				updateGroupWithMesh(
					&group,
					currResource,
					volumeName,
					bundle.Extruders[i],
					bundle.WipeIntoInfill[i],
					bundle.WipeIntoModel[i],
				)
				// store the modified group back in the map
				groups[groupName] = group
			}
		}
	}

	m.Resources = m.Resources[:len(groups)]
	m.Build = m.Build[:len(groups)]
	index := 0
	groupVolumeInfo := []MergedVolumesInfo{}
	for _, group := range groups {
		m.Resources[index] = group.resources
		m.Build[index] = group.build
		groupVolumeInfo = append(groupVolumeInfo, MergedVolumesInfo{
			VolumeIdPairs:  group.volumeIdPairs,
			VolumeNames:    group.volumeNames,
			Extruders:      group.Extruders,
			WipeIntoInfill: group.WipeIntoInfill,
			WipeIntoModel:  group.WipeIntoModel,
			BoundingBox:    group.BoundingBox,
		})
		index++
	}

	return groupVolumeInfo, nil
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
