package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestIntegration(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Join(filepath.Dir(filename), "test")
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		matched, err := regexp.MatchString(`\.js$`, f.Name())
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			t.Run(f.Name(), func(t *testing.T) {
				testSingleFile(t, filepath.Join(testDir, f.Name()))
			})
		}
	}
}

func readFile(t *testing.T, filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	return string(bytes)
}

func testSingleFile(t *testing.T, testFilePath string) {
	// Read input file
	inputBytes := readFile(t, testFilePath)

	// Process it
	output := jsToScad(inputBytes)

	// Optional: Write output file
	if os.Getenv("REGENERATE_OUTPUT") != "" {
		err := ioutil.WriteFile(testFilePath+".scad", []byte(output), 0644)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}

	// Read expected output
	expectedOutput := readFile(t, testFilePath+".scad")

	// Compare
	if output != expectedOutput {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(output, expectedOutput, false)
		t.Error("output doesn't match " + filepath.Base(testFilePath) + ":\n" +
			"\x1b[31m- actual\x1b[0m \x1b[32m+ expected\x1b[0m\n" +
			dmp.DiffPrettyText(diffs))
	}
}
