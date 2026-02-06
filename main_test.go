package main

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"mosaicmfg.com/stl-to-3mf/ps3mf"
)

type ContentTypes struct {
	XMLName xml.Name      `xml:"Types"`
	Xmlns   string        `xml:"xmlns,attr"`
	Default []DefaultType `xml:"Default"`
}

type DefaultType struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

func setupDateMock(t testing.TB) func() {
	originalFunc := ps3mf.CurrentDate
	ps3mf.CurrentDate = "2025-05-17"
	return func() {
		ps3mf.CurrentDate = originalFunc
	}
}

// TestPaintedCube tests the conversion of a single STL file to a 3MF file
func TestPaintedCube(t *testing.T) {
	// Setup date mock
	cleanup := setupDateMock(t)
	defer cleanup()

	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }() // Restore after test

	// Get absolute paths for test files
	testDir := filepath.Join("test")
	outPath := filepath.Join(testDir, "single.3mf")
	configPath := filepath.Join(testDir, "Slic3r_PE.config")
	colorsPath := filepath.Join(testDir, "colors.rle")
	stlPath := filepath.Join(testDir, "cube.stl")
	transform := "1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|177.500000,177.500000,8.000000,1.000000"

	// Simulate CLI arguments
	os.Args = []string{
		"cmd",                           // os.Args[0] is the program name
		outPath,                         // OutPath - save to test/cube.3mf
		configPath,                      // ConfigPath
		`{"filamentIds":[[0,0],[1,1]]}`, // FilamentIDs as string
		"--colors", colorsPath,
		"|cube", // Model Name
		transform,
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		stlPath,
	}

	// Remove the output file if it exists
	os.Remove(outPath)

	// Call Run function to execute the test
	run()

	// Snapshot the generated 3MF contents
	snapshotZipContents(t, outPath)
}

func TestModelWithCustomSupports(t *testing.T) {
	// Setup date mock
	cleanup := setupDateMock(t)
	defer cleanup()

	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }() // Restore after test

	// Get absolute paths for test files
	testDir := filepath.Join("test")
	outPath := filepath.Join(testDir, "customSupport.3mf")
	configPath := filepath.Join(testDir, "Slic3r_PE.config")
	supportsPath := filepath.Join(testDir, "supports.rle")
	infillDensity := 15
	stlPath := filepath.Join(testDir, "supportTest.stl")
	transform := "1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|53.117386,28.125000,0.000000,1.000000"

	// Simulate CLI arguments
	os.Args = []string{
		"cmd",                     // os.Args[0] is the program name
		outPath,                   // OutPath - save to test/cube.3mf
		configPath,                // ConfigPath
		`{"filamentIds":[[0,0]]}`, // FilamentIDs as string
		"--supports", supportsPath,
		"--infill", strconv.Itoa(infillDensity),
		"|cube", // Model Name
		transform,
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		stlPath,
	}

	// Remove the output file if it exists
	os.Remove(outPath)

	// Call Run function to execute the test
	run()

	// Snapshot the generated 3MF contents
	snapshotZipContents(t, outPath)
}

func TestGroupedModels(t *testing.T) {
	// Setup date mock
	cleanup := setupDateMock(t)
	defer cleanup()

	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }() // Restore after test

	// Test directory and output path
	testDir := filepath.Join("test")
	outPath := filepath.Join(testDir, "group.3mf")
	configPath := filepath.Join(testDir, "Slic3r_PE.config")

	// Get paths to STL test files
	cubePath := filepath.Join(testDir, "cube.stl")
	part1Path := filepath.Join(testDir, "Part Studio 1 - Part 1.stl")
	part2Path := filepath.Join(testDir, "Part Studio 1 - Part 2.stl")
	part3Path := filepath.Join(testDir, "Part Studio 1 - Part 3.stl")
	part4Path := filepath.Join(testDir, "Part Studio 1 - Part 4.stl")
	part5Path := filepath.Join(testDir, "Part Studio 1 - Part 5.stl")

	// Remove the output file if it exists
	os.Remove(outPath)

	// Set the arguments directly
	os.Args = []string{
		"cmd",                     // Program name
		outPath,                   // Output path
		configPath,                // Config path
		`{"filamentIds":[[0,0]]}`, // Filament IDs JSON

		// Group 2|cube2
		"Group (2)|cube2",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|177.000000,206.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		cubePath,

		// Group|Part Studio 1 - Part 4
		"Group|Part Studio 1 - Part 4",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part4Path,

		// Single cube3
		"|cube 3",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|199.000000,206.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		cubePath,

		// Group 2|cube
		"Group (2)|cube",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|157.000000,206.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		cubePath,

		// Group|Part Studio 1 - Part 1
		"Group|Part Studio 1 - Part 1",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part1Path,

		// Group|Part Studio 1 - Part 2
		"Group|Part Studio 1 - Part 2",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part2Path,

		// Group|Part Studio 1 - Part 3
		"Group|Part Studio 1 - Part 3",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part3Path,

		// Group|Part Studio 1 - Part 5
		"Group|Part Studio 1 - Part 5",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part5Path}

	// Call the Run function
	run()

	// Snapshot the generated 3MF contents
	snapshotZipContents(t, outPath)
}

// Read ZIP file and creates individual snapshots for each file
func snapshotZipContents(t *testing.T, zipPath string) {
	// Verify the output file was created
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", zipPath)
		return
	}

	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open 3MF file: %v", err)
	}
	defer reader.Close()

	// Create individual snapshots
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		srcFile, err := file.Open()
		if err != nil {
			t.Logf("Error opening %s: %v", file.Name, err)
			continue
		}

		content, err := io.ReadAll(srcFile)
		srcFile.Close()
		if err != nil {
			t.Logf("Error reading %s: %v", file.Name, err)
			continue
		}

		// For XML files, sort them to ensure deterministic output
		if file.Name == "[Content_Types].xml" {
			content = sortContentTypesXML(content)
		}

		// Create a snapshot organized by test name and file path
		snapshotPath := filepath.Join(t.Name(), file.Name)
		snaps.WithConfig(snaps.Filename(snapshotPath)).MatchSnapshot(t, string(content))
	}

	// Clean up the output file
	os.Remove(zipPath)
}

// Sorts XML content to ensure deterministic output
func sortContentTypesXML(content []byte) []byte {
	var contentTypes ContentTypes
	if err := xml.Unmarshal(content, &contentTypes); err == nil {
		// Sort Default elements
		sort.Slice(contentTypes.Default, func(i, j int) bool {
			return contentTypes.Default[i].Extension < contentTypes.Default[j].Extension
		})

		// Marshal back to XML
		if sortedData, err := xml.MarshalIndent(contentTypes, "", "    "); err == nil {
			return append([]byte(xml.Header), sortedData...)
		}
	}
	return content
}
