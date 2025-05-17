package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestPaintedCube tests the conversion of a single STL file to a 3MF file
func TestPaintedCube(t *testing.T) {
	// Backup original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }() // Restore after test

	// Get absolute paths for test files
	testDir := filepath.Join("test")
	outPath := filepath.Join(testDir, "single.3mf")
	configPath := filepath.Join(testDir, "Slic3r_PE.config")
	colorsPath := filepath.Join(testDir, "colors.rle")
	stlPath := filepath.Join(testDir, "cube.stl")

	// Get transform from transforms.txt
	transform := "1,0,0,0|0,1,0,0|0,0,1,0|0,0,0,1" // From test/transforms.txt

	// Simulate CLI arguments
	os.Args = []string{
		"cmd",      // os.Args[0] is the program name
		outPath,    // OutPath - save to test/cube.3mf
		configPath, // ConfigPath
		"nil",      // FilamentIDs as string
		"--colors", colorsPath,
		"|cube",   // Model Name
		transform, // Transforms from transforms.txt
		"0",       // Extruder
		"1",       // WipeIntoInfill
		"0",       // WipeIntoModel
		stlPath,   // MeshPath - use test/cube.stl
	}

	// Remove the output file if it exists
	os.Remove(outPath)

	// Call Run function to execute the test
	Run()

	// Verify and extract the output file
	verifyAndExtract(t, outPath, filepath.Join(testDir, "coloredCube_unzip"), false)
}

func TestGroupedModels(t *testing.T) {
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

		// single
		"|cube3",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|199.000000,206.000000,8.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		cubePath,

		// group 2
		"Group (2)|cube",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|157.000000,206.000000,8.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		cubePath,

		"Group (2)|cube2",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|177.000000,206.000000,8.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		cubePath,

		// group 1

		"Group|Part Studio 1 - Part 1",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part1Path,

		"Group|Part Studio 1 - Part 2",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part2Path,

		"Group|Part Studio 1 - Part 3",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part3Path,

		"Group|Part Studio 1 - Part 4",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part4Path,

		"Group|Part Studio 1 - Part 5",
		"1.000000,0.000000,0.000000,0.000000|0.000000,1.000000,0.000000,0.000000|0.000000,0.000000,1.000000,0.000000|174.000000,167.000000,0.000000,1.000000",
		"1", // Extruder
		"0", // WipeIntoInfill
		"0", // WipeIntoModel
		part5Path,
	}

	// Call the Run function
	Run()

	// Verify and extract the output file
	verifyAndExtract(t, outPath, filepath.Join(testDir, "groupedModels_unzip"), false)
}

// verifyAndExtract verifies that the output file was created and extracts its contents
// if removeOriginal is true, the original 3MF file will be removed after extraction
func verifyAndExtract(t *testing.T, outPath, extractDir string, removeOriginal bool) {
	// Verify the output file was created
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outPath)
		return
	}

	t.Logf("Successfully created output file at %s", outPath)

	// Unzip the output file
	reader, err := zip.OpenReader(outPath)
	if err != nil {
		t.Fatalf("Failed to open 3MF file: %v", err)
	}
	defer reader.Close()

	// Create extract directory if it doesn't exist
	os.MkdirAll(extractDir, 0755)

	// Extract all files
	for _, file := range reader.File {
		filePath := filepath.Join(extractDir, file.Name)

		// Create directory for file if needed
		os.MkdirAll(filepath.Dir(filePath), 0755)

		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			t.Logf("Error opening %s: %v", file.Name, err)
			continue
		}

		dstFile, err := os.Create(filePath)
		if err != nil {
			srcFile.Close()
			t.Logf("Error creating %s: %v", filePath, err)
			continue
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			t.Logf("Error extracting %s: %v", file.Name, err)
		}
	}

	t.Logf("Extracted 3MF to %s", extractDir)

	// Optionally remove the original 3MF file
	if removeOriginal {
		os.Remove(outPath)
		t.Logf("Removed original 3MF file: %s", outPath)
	}
}
