package ps3mf

import (
	"../util"
	"archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/hpinc/go3mf"
	"io"
	"io/ioutil"
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
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
	Z string `xml:"z,attr"`
}

type Triangle struct {
	XMLName xml.Name `xml:"triangle"`
	V1 string `xml:"v1,attr"`
	V2 string `xml:"v2,attr"`
	V3 string `xml:"v3,attr"`
	Segmentation string `xml:"slic3rpe:mmu_segmentation,attr,omitempty"`
	CustomSupports string `xml:"slic3rpe:custom_supports,attr,omitempty"`
}

type BuildItem struct {
	XMLName xml.Name `xml:"item"`
	ObjectId string `xml:"objectid,attr"`
	Transform string `xml:"transform,attr,omitempty"`
	Printable string `xml:"printable,attr,omitempty"`
}

func (m *Mesh) AddColors(rle *util.RLE) {
	color := *rle
	currentRunIndex := -1
	currentRunLength := 0
	currentColor := 0

	for triIdx := range m.Triangles {
		if currentRunLength <= 0 {
			if currentRunIndex < len(color.Runs) {
				currentRunIndex++
				currentRunLength = int(color.Runs[currentRunIndex].Length)
				currentColor = int(color.Runs[currentRunIndex].Value)
			}
		}
		if currentColor == 1 {
			m.Triangles[triIdx].Segmentation = "8"
		} else if currentColor > 1 {
			m.Triangles[triIdx].Segmentation = fmt.Sprintf("%dC", currentColor - 2)
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

func (m *Bundle) Save(path string) (err error) {
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
	if err = tempWriter.Encode(m.Model); err != nil {
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

	for _, file := range reader.File {
		fileWriter, writerErr := writer.Create(file.Name)
		if writerErr != nil {
			err = writerErr
			return
		}
		if file.Name == "3D/3dmodel.model" {
			var model ModelXML
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
				if m.Colors[idx] != nil {
					model.Resources[idx].Mesh.AddColors(m.Colors[idx])
					hasCustomColors = true
				}
				if m.Supports[idx] != nil {
					model.Resources[idx].Mesh.AddCustomSupports(m.Supports[idx])
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
	if len(m.Config) > 0 {
		fileWriter, writerErr := writer.Create("Metadata/Slic3r_PE.config")
		if writerErr != nil {
			err = writerErr
			return
		}
		if _, writeErr := io.WriteString(fileWriter, m.Config); writeErr != nil {
			err = writeErr
			return
		}
	}

	// generate and write in Metadata/Slic3r_PE_model.config
	modelConfig := m.GetModelConfig()
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
