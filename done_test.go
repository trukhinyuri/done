package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestVersionParameterPassed(t *testing.T) {
	// Skip test if BuildVersion is empty
	if BuildVersion == "" {
		t.Skip("BuildVersion is empty, skipping version test")
	}

	expected := BuildVersion + "\n"

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create buffer to capture output
	var buf bytes.Buffer

	// Create a pipe and redirect stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call printServiceVersion
	printServiceVersion()

	// Close the writer to signal we're done
	w.Close()

	// Read all output
	io.Copy(&buf, r)
	actual := buf.String()

	if actual != expected {
		t.Errorf("Incorrect --version flag behavior. Expected: %q, Got: %q", expected, actual)
	}
}

func BenchmarkVersionParameterPassed(b *testing.B) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create null writer to discard output
	nullFile, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer nullFile.Close()
	os.Stdout = nullFile

	*versionPtr = true

	for i := 0; i < b.N; i++ {
		printServiceVersion()
	}
}

//func BenchmarkMain (b *testing.B) {
//	main()
//}

//func TestMain (t *testing.T) {
//	if 1 != 0 {
//		t.Skip("One not equal zero", "ok")
//	}
//}
//
