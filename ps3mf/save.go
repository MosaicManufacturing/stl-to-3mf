package ps3mf

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/hpinc/go3mf"
	"io"
	"io/ioutil"
	"mosaicmfg.com/stl-to-3mf/util"
	"os"
	"time"
)

const (
	version3mf = "1"
	mmPaintingVersion = "1"
	fdmSupportsPaintingVersion = "1"
)

type ModelXML struct {
	XMLName xml.Name `xml:"model"`
	Units string `xml:"unit,attr"`
	Language string `xml:"xml:lang,attr"`
	Namespace string `xml:"xmlns,attr"`
	Slic3rNamespace string `xml:"xmlns:slic3rpe,attr"`
	Metadata []Meta `xml:"metadata"`
	Resources []Resource `xml:"resources>object"`
	Build []BuildItem `xml:"build>item"`
}

type Meta struct {
	XMLName xml.Name `xml:"metadata"`
	Name string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}

type Resource struct {
	XMLName xml.Name `xml:"object"`
	Id string `xml:"id,attr"`
	Type string `xml:"type,attr,omitempty"`
	Mesh Mesh `xml:"mesh"`
}

type Mesh struct {
	XMLName xml.Name `xml:"mesh"`
	Vertices []Vertex `xml:"vertices>vertex"`
	Triangles []Triangle `xml:"triangles>triangle"`
}

type Vertex struct {
	XMLName xml.Name `xml:"vertex"`
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
	Z float64 `xml:"z,attr"`
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
	XMLName xml.Name `xml:"triangle"`
	V1 int `xml:"v1,attr"`
	V2 int `xml:"v2,attr"`
	V3 int `xml:"v3,attr"`
	Segmentation string `xml:"slic3rpe:mmu_segmentation,attr,omitempty"`
	CustomSupports string `xml:"slic3rpe:custom_supports,attr,omitempty"`
}

type BuildItem struct {
	XMLName xml.Name `xml:"item"`
	ObjectId string `xml:"objectid,attr"`
	Transform string `xml:"transform,attr,omitempty"`
	Printable string `xml:"printable,attr,omitempty"`
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
			m.Triangles[triIdx].Segmentation = fmt.Sprintf("%xC", currentColor - 2)
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
		}
		currentRunLength--
	}
}

func GetMeta(name, value string) Meta {
	return Meta{
		Name: name,
		Value: value,
	}
}

func (b *Bundle) Save(path string) (err error) {
	// general workflow:
	// 1. use 3MF lib to write "vanilla" 3MF to temp file
	// 2. use zip lib to open and modify 3MF contents
	//    - add metadata to model (Title = project name)
	//    - add custom support/color data to model objects
	//    - create Metadata dir
	//    - generate Metadata/Slic3r_PE_model.config
	//    - copy in Metadata/Slic3r_PE.config
	//    - copy in Metadata/thumbnail.png (optional)
	// 3. save modified zip back to real path

	// get a temporary file path
	tmpFile, tmpErr := ioutil.TempFile(os.TempDir(), "stl-to-3mf-")
	if tmpErr != nil {
		err = tmpErr
		return
	}
	// clean up the temp file after we're done with it
	defer func() {
		removeErr := os.Remove(tmpFile.Name())
		if removeErr != nil {
			err = removeErr
		}
	}()

	// close the file now so go3mf can create its own reference
	if err = tmpFile.Close(); err != nil {
		return
	}

	// write "vanilla" 3MF data to temp file
	tempWriter, writerErr := go3mf.CreateWriter(tmpFile.Name())
	if writerErr != nil {
		err = writerErr
		return
	}
	if err = tempWriter.Encode(b.Model); err != nil {
		return
	}
	if err = tempWriter.Close(); err != nil {
		return
	}

	// read 3MF data as a zip
	reader, readerErr := zip.OpenReader(tmpFile.Name())
	if readerErr != nil {
		err = readerErr
		return
	}
	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	// open a zip writer for writing at the output path
	zipFile, createErr := os.Create(path)
	if createErr != nil {
		err = createErr
		return
	}
	defer func() {
		closeErr := zipFile.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()
	writer := zip.NewWriter(zipFile)

	var model ModelXML
	var idPairs []IdPair

	for _, file := range reader.File {
		fileWriter, writerErr := writer.Create(file.Name)
		if writerErr != nil {
			err = writerErr
			return
		}
		if file.Name == "3D/3dmodel.model" {
			// read file and parse XML into struct
			readCloser, openErr := file.Open()
			if openErr != nil {
				err = openErr
				return
			}
			defer func() {
				closeErr := readCloser.Close()
				if closeErr != nil {
					err = closeErr
				}
			}()
			fileBytes, readErr := ioutil.ReadAll(readCloser)
			if readErr != nil {
				err = readErr
				return
			}
			unmarshalErr := xml.Unmarshal(fileBytes, &model)
			if unmarshalErr != nil {
				err = unmarshalErr
				return
			}

			// add custom color and/or support data, if available
			hasCustomColors := false
			hasCustomSupports := false
			for idx := range model.Resources {
				if b.Colors[idx] != nil {
					model.Resources[idx].Mesh.AddColors(b.Colors[idx])
					hasCustomColors = true
				}
				if b.Supports[idx] != nil {
					model.Resources[idx].Mesh.AddCustomSupports(b.Supports[idx])
					hasCustomSupports = true
				}
			}

			// add missing metadata
			model.Language = "en-US"
			model.Slic3rNamespace = slic3rPENamespace
			currentDate := time.Now().Format("2006-01-02") // YYYY-MM-DD
			model.Metadata = append(
				model.Metadata,
				GetMeta("slic3rpe:Version3mf", version3mf),
			)
			if hasCustomColors {
				// only include if any models are painted
				model.Metadata = append(
					model.Metadata,
					GetMeta("slic3rpe:MmPaintingVersion", mmPaintingVersion),
				)
			}
			if hasCustomSupports {
				// only include if any models use custom supports
				model.Metadata = append(
					model.Metadata,
					GetMeta("slic3rpe:FdmSupportsPaintingVersion", fdmSupportsPaintingVersion),
				)
			}
			model.Metadata = append(
				model.Metadata,
				GetMeta("Title", "Project"),
				GetMeta("Designer", ""),
				GetMeta("Description", ""),
				GetMeta("Copyright", ""),
				GetMeta("LicenseTerms", ""),
				GetMeta("Rating", ""),
				GetMeta("CreationDate", currentDate),
				GetMeta("ModificationDate", currentDate),
				GetMeta("Application", "Canvas"),
			)

			// combine all meshes into a single mesh
			idPairs = model.MergeMeshes(b.Matrices)

			// write modified content into final zip
			output, marshalErr := xml.Marshal(model)
			if marshalErr != nil {
				err = marshalErr
				return
			}

			if _, writeErr := io.WriteString(fileWriter, string(output)); writeErr != nil {
				err = writeErr
				return
			}
		} else {
			// copy file to new zip at same path
			openedFile, openErr := file.Open()
			if openErr != nil {
				err = openErr
				return
			}
			if _, copyErr := io.Copy(fileWriter, openedFile); copyErr != nil {
				err = copyErr
				return
			}
		}
	}

	// copy in Metadata/Slic3r_PE.config
	if len(b.Config) > 0 {
		fileWriter, writerErr := writer.Create("Metadata/Slic3r_PE.config")
		if writerErr != nil {
			err = writerErr
			return
		}
		if _, writeErr := io.WriteString(fileWriter, b.Config); writeErr != nil {
			err = writeErr
			return
		}
	}

	// generate and write in Metadata/Slic3r_PE_model.config
	modelConfig := b.GetModelConfig(&model, idPairs)
	output, marshalErr := xml.Marshal(modelConfig)
	if marshalErr != nil {
		err = marshalErr
		return
	}
	fileWriter, writerErr := writer.Create("Metadata/Slic3r_PE_model.config")
	if writerErr != nil {
		err = writerErr
		return
	}
	if _, writeErr := io.WriteString(fileWriter, string(output)); writeErr != nil {
		err = writeErr
		return
	}

	closeErr := writer.Close()
	if closeErr != nil {
		err = closeErr
	}
	return
}
