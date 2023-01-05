package ps3mf

import (
	"archive/zip"
	"encoding/xml"
	"github.com/hpinc/go3mf"
	"io"
	"io/ioutil"
	"os"
	"time"
)

const (
	version3mf                 = "1"
	mmPaintingVersion          = "1"
	fdmSupportsPaintingVersion = "1"
)

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
			output, marshalErr := model.Marshal()
			if marshalErr != nil {
				err = marshalErr
				return
			}

			if _, writeErr := fileWriter.Write(output); writeErr != nil {
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
		if _, writeErr := fileWriter.Write(b.Config); writeErr != nil {
			err = writeErr
			return
		}
	}

	// generate and write in Metadata/Slic3r_PE_model.config
	modelConfig := b.GetModelConfig(&model, idPairs, path)
	output, marshalErr := modelConfig.Marshal()
	if marshalErr != nil {
		err = marshalErr
		return
	}
	fileWriter, writerErr := writer.Create("Metadata/Slic3r_PE_model.config")
	if writerErr != nil {
		err = writerErr
		return
	}
	if _, writeErr := io.WriteString(fileWriter, xml.Header); writeErr != nil {
		err = writeErr
		return
	}
	if _, writeErr := fileWriter.Write(output); writeErr != nil {
		err = writeErr
		return
	}

	closeErr := writer.Close()
	if closeErr != nil {
		err = closeErr
	}
	return
}
