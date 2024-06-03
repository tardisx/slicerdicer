package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

type resultPrinter struct {
	batchTotal   int
	count        int
	lastFilename string
	ch           chan string
}

func (rp *resultPrinter) Reset(batchSize int) {
	rp.batchTotal = batchSize
	rp.count = 0
	rp.lastFilename = ""
	fmt.Print("\033[s")
}

func (rp *resultPrinter) Run() {
	for last := range rp.ch {
		rp.count++
		rp.lastFilename = last
		fmt.Print("\033[u\033[K") // restore the cursor position and clear the line
		fmt.Printf("processing %5d/%5d - %s", rp.count, rp.batchTotal, rp.lastFilename)
	}
}

const defaultPathTemplate = "%s-%d-%d-%d.%s"

func main() {
	t0 := time.Now()
	filenamePtr := flag.String("filename", "", "filename to open")
	tileSizePtr := flag.Int("tile-size", 256, "tile size, in pixels")
	concurrencyPtr := flag.Int("concurrency", 5, "how many tiles to generate concurrently (threads)")
	baseName := flag.String("basename", "tile", "base of the output files")
	outFormat := flag.String("format", "png", "output format (jpg/png)")
	pathTemplate := flag.String("path-template", defaultPathTemplate, "template for output files - base, zoom, x, y, format")

	flag.Parse()

	if *filenamePtr == "" {
		fmt.Println("Error: You must specify a filename with --filename")
		return
	}

	if *outFormat != "jpg" && *outFormat != "png" {
		fmt.Println("Error: -format must be jpg or png")
		return
	}

	log.Println("opening file:", *filenamePtr)
	src, err := os.Open(*filenamePtr)
	if err != nil {
		fmt.Println("Error: Could not open file:", err)
		return
	}

	processImage(src, *baseName, *pathTemplate, *outFormat, *tileSizePtr, *concurrencyPtr, diskOutput{})
	log.Printf("done in %.2f", time.Since(t0).Seconds())
}

func processImage(input io.Reader, basename string, pathTemplate string, format string, tileSize int, concurrency int, output outputter) {

	src, err := imaging.Decode(input)
	if err != nil {
		panic(err)
	}

	size := src.Bounds().Max

	tile_size_x := tileSize
	tile_size_y := tileSize

	// work out maximum zoom
	var max_zoom int
	zoom_test_size_x := size.X
	zoom_test_size_y := size.Y
	for max_zoom = 0; ; max_zoom++ {
		if zoom_test_size_x <= tile_size_x && zoom_test_size_y <= tile_size_y {
			break
		}
		zoom_test_size_x = zoom_test_size_x >> 1
		zoom_test_size_y = zoom_test_size_y >> 1
	}

	z := max_zoom
	log.Println("maximum zoom level is", max_zoom)
	log.Println("starting tiling with concurrency of", concurrency)

	results := make(chan string)
	rp := resultPrinter{
		batchTotal:   0,
		count:        0,
		lastFilename: "",
		ch:           results,
	}

	// start the tileWorkers
	jobs := make(chan tileJob)
	for i := 0; i < concurrency; i++ {
		go tileWorker(jobs, results)
	}

	go func() {
		rp.Run()
	}()

	// outer loop for zoom
	for {

		if z == max_zoom {
			// do nothing
		} else {
			// halve image size
			log.Print("resizing for next zoom level")
			src = imaging.Resize(src, size.X/2, 0, imaging.NearestNeighbor)
			// recalculate size
			size = src.Bounds().Max
		}

		log.Printf("zoom level: %d (%d x %d)\n", z, size.X, size.Y)

		// size 257 => 2
		// size 256 => 1
		// size 255 => 1
		yTiles := int(math.Ceil(float64(size.Y) / float64(tile_size_y)))
		xTiles := int(math.Ceil(float64(size.X) / float64(tile_size_x)))

		if z == 0 {
			xTiles = 1
			yTiles = 1
		}
		tilesToRender := xTiles * yTiles

		rp.Reset(tilesToRender)

		wg := sync.WaitGroup{}
		wg.Add(tilesToRender)

		log.Printf("rendering %d tiles", tilesToRender)

		for y := 0; y < yTiles; y++ {
			for x := 0; x < xTiles; x++ {
				jobs <- tileJob{
					baseName:     basename,
					pathTemplate: pathTemplate,
					format:       format,
					src:          src,
					zoom:         z,
					x:            x,
					y:            y,
					tileSizeX:    tile_size_x,
					tileSizeY:    tile_size_y,
					wg:           &wg,
					output:       output,
				}
			}

		}
		wg.Wait() // wait for all tiles to be generated for this zoom level
		z--
		if z < 0 {
			break
		}

		// let the last progress be printed out
		// yes I know this is ugly :-)
		time.Sleep(time.Millisecond * 10)
		fmt.Print("\033[u\033[K") // restore the cursor position and clear the line
	}
	close(results)
	log.Printf("all tiles complete")
}

type outputter interface {
	CreatePathAndFile(fn string) (io.WriteCloser, error)
}

type diskOutput struct {
}

func (do diskOutput) CreatePathAndFile(fn string) (io.WriteCloser, error) {
	dir, _ := filepath.Split(fn)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}

	writer, err := os.Create(fn)
	return writer, err
}

type tileJob struct {
	baseName     string
	pathTemplate string
	format       string
	src          image.Image
	zoom         int
	x, y         int
	tileSizeX    int
	tileSizeY    int
	output       outputter
	wg           *sync.WaitGroup
}

func tileWorker(jobs <-chan tileJob, results chan<- string) {
	for j := range jobs {
		output_filename := fmt.Sprintf(j.pathTemplate, j.baseName, j.zoom, j.x, j.y, j.format)
		cropped := imaging.Crop(j.src, image.Rect(j.tileSizeX*j.x, j.tileSizeY*j.y, j.tileSizeX*j.x+j.tileSizeX, j.tileSizeY*j.y+j.tileSizeY))

		// log.Printf("writing to %s", output_filename)
		writer, err := j.output.CreatePathAndFile(output_filename)
		if err != nil {
			panic(err)
		}
		if j.format == "png" {
			err = png.Encode(writer, cropped)
		} else if j.format == "jpg" {
			err = jpeg.Encode(writer, cropped, &jpeg.Options{
				Quality: 40,
			})
		} else {
			panic("bad format")
		}
		if err != nil {
			panic(err)
		}
		writer.Close()
		results <- output_filename
		j.wg.Done()
	}
}
