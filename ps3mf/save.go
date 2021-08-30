package ps3mf

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/hpinc/go3mf"
	"io/ioutil"
	"log"
	"os"
)

type ModelXML struct {
	XMLName xml.Name `xml:"model"`
	Units string `xml:"unit,attr"`
	Language string `xml:"xml:lang,attr"`
	Namespace string `xml:"xmlns,attr"`
	Slic3rNamespace string `xml:"xmlns:slic3rpe,attr"`
	//Version struct{
	//	XMLName xml.Name `xml:"metadata"`
	//	Name string `xml:"name,attr"`
	//} `xml:"metadata"`
	//Title struct{
	//	XMLName xml.Name `xml:"metadata"`
	//	Name string `xml:"title,attr"`
	//} `xml:"metadata,any"`
	Resources []Resource `xml:"resources>object"`
	Build []BuildItem `xml:"build>item"`
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
}

type BuildItem struct {
	XMLName xml.Name `xml:"item"`
	ObjectId string `xml:"objectid,attr"`
	Transform string `xml:"transform,attr,omitempty"`
	Printable string `xml:"printable,attr,omitempty"`
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
	tmpFile, err := ioutil.TempFile(os.TempDir(), "stl-to-3mf-")
	if err != nil {
		return
	}
	// clean up the temp file after we're done with it
	defer func() {
		err = os.Remove(tmpFile.Name())
	}()

	// close the file now so go3mf can create its own reference
	if err = tmpFile.Close(); err != nil {
		return
	}

	// write "vanilla" 3MF data to temp file
	writer, err := go3mf.CreateWriter(tmpFile.Name())
	if err != nil {
		log.Fatalln(err)
	}
	if err := writer.Encode(m.Model); err != nil {
		log.Fatalln(err)
	}
	if err := writer.Close(); err != nil {
		log.Fatalln(err)
	}

	// TODO: open a final zip for writing at the output path

	// read 3MF data as a zip
	reader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return
	}
	defer func() {
		err = reader.Close()
	}()

	for _, file := range reader.File {
		if file.Name == "3D/3dmodel.model" {
			fmt.Printf("File: %s\n", file.Name)
			var model ModelXML
			// read file and parse XML into struct
			fmt.Println("open")
			readCloser, openErr := file.Open()
			if openErr != nil {
				err = openErr
				return
			}
			fmt.Println("read")
			fileBytes, readErr := ioutil.ReadAll(readCloser)
			if readErr != nil {
				err = readErr
				return
			}
			fmt.Println("unmarshal")
			unmarshalErr := xml.Unmarshal(fileBytes, &model)
			if unmarshalErr != nil {
				fmt.Println("not nil?")
				fmt.Println(err)
				err = unmarshalErr
				return
			}

			model.Language = "en-US"
			model.Slic3rNamespace = "http://schemas.slic3r.org/3mf/2017/06"

			fmt.Println("marshal")
			output, marshalErr := xml.Marshal(model)
			if marshalErr != nil {
				err = marshalErr
				return
			}

			fmt.Println("print")
			fmt.Println(string(output))

			// TODO:
			//  1. parse file content as XML
			//  2. add in custom data
			//  3. convert XML to string
			//  4. write to new zip
		} else {
			// TODO: copy file to new zip at same path
		}

		// TODO: test code, remove
		//readCloser, openErr := file.Open()
		//if openErr != nil {
		//	err = openErr
		//	return
		//}
		//buf := new(bytes.Buffer)
		//_, readErr := buf.ReadFrom(readCloser)
		//if readErr != nil {
		//	err = readErr
		//	return
		//}
		//fmt.Println(buf.String())
		//fmt.Println()
	}

	// TODO: copy in Metadata/Slic3r_PE.config

	// TODO: generate and write Metadata/Slic3r_PE_model.config

	return
}
