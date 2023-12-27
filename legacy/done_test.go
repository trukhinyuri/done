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
	expected := VERSION + "\n"

	*versionPtr = true
	args := []string{"--version"}

	r, w, _ := os.Pipe()
	os.Stdout = w

	submain(args)

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()

	actual := <-outC

	if actual != expected {
		t.Error("Incorrect --version flag behavior. Result: " + actual)
	}
}

func BenchmarkVersionParameterPassed(b *testing.B) {
	*versionPtr = true
	args := []string{"--version"}

	for i := 0; i < b.N; i++ {
		submain(args)

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
