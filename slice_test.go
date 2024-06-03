package main

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testWriteCloser struct {
	written int
	closed  bool
}

func (twc *testWriteCloser) Close() error {
	twc.closed = true
	return nil
}

func (twc *testWriteCloser) Write(b []byte) (int, error) {
	twc.written += len(b)
	return len(b), nil
}

type testOutputter struct {
	closers map[string]*testWriteCloser
}

func (to *testOutputter) CreatePathAndFile(fn string) (io.WriteCloser, error) {
	to.closers[fn] = &testWriteCloser{}
	return to.closers[fn], nil
}

func (to *testOutputter) Dump() string {
	out := ""
	for k, v := range to.closers {
		out += fmt.Sprintf("%20s: %6d/%t\n", k, v.written, v.closed)
	}
	return out
}

func NewTestOutputter() *testOutputter {
	to := testOutputter{}
	to.closers = make(map[string]*testWriteCloser)
	return &to
}

func TestXxx(t *testing.T) {

	// testdata/gold.jpg: JPEG image data, JFIF standard 1.01, 794x447
	f, err := os.Open("testdata/gold.jpg")
	if err != nil {
		panic(err)
	}

	// with a tile size larger than the image itself, we should
	// get a single image at zoom level 0
	testOutputter := NewTestOutputter()
	processImage(f, "indy", defaultPathTemplate, "jpg", 1000, 1, testOutputter)
	t.Log(testOutputter.Dump())
	assert.Equal(t, 1, len(testOutputter.closers))
	assert.Equal(t, 28082, testOutputter.closers["indy-0-0-0.jpg"].written)

	f, _ = os.Open("testdata/gold.jpg")

	// with a tilesize smaller than *one* of the dimensions, 3 tiles
	// 1@zoom 0, 2@zoom 1
	testOutputter = NewTestOutputter()
	processImage(f, "indy", defaultPathTemplate, "jpg", 500, 1, testOutputter)
	t.Log(testOutputter.Dump())
	assert.Equal(t, 3, len(testOutputter.closers))
	assert.Equal(t, 10304, testOutputter.closers["indy-0-0-0.jpg"].written)
	assert.Equal(t, 18014, testOutputter.closers["indy-1-0-0.jpg"].written)
	assert.Equal(t, 10749, testOutputter.closers["indy-1-1-0.jpg"].written)

	f, _ = os.Open("testdata/gold.jpg")

	// with a tilesize smaller than *both* of the dimensions, 5 tiles
	// zoom 0, zoom 1 x 4
	testOutputter = NewTestOutputter()
	processImage(f, "indy", defaultPathTemplate, "jpg", 400, 1, testOutputter)
	t.Log(testOutputter.Dump())
	assert.Equal(t, 5, len(testOutputter.closers))
	assert.Equal(t, 10304, testOutputter.closers["indy-0-0-0.jpg"].written)
	assert.Equal(t, 12607, testOutputter.closers["indy-1-0-0.jpg"].written)
	assert.Equal(t, 12050, testOutputter.closers["indy-1-1-0.jpg"].written)
	assert.Equal(t, 2722, testOutputter.closers["indy-1-0-1.jpg"].written)
	assert.Equal(t, 2477, testOutputter.closers["indy-1-1-1.jpg"].written)

}
