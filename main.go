package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
)

const currentVersion = "0.01"

func main() {

	filenamePtr := flag.String("filename", "", "filename to open")
	tileSizePtr := flag.Int("tile-size", 256, "tile size, in pixels")
	concurrencyPtr := flag.Int("concurrency", 5, "how many tiles to generate concurrently (threads)")
	baseName := flag.String("basename", "tile", "base of the output files")

	flag.Parse()

	if *filenamePtr == "" {
		fmt.Println("Error: You must specify a filename with --filename")
		return
	}

	fmt.Println("opening file:", *filenamePtr)
	src, err := imaging.Open(*filenamePtr)
	if err != nil {
		fmt.Println("Error: Could not open file:", err)
		return
	}

	size := src.Bounds().Max

	tile_size_x := *tileSizePtr
	tile_size_y := *tileSizePtr

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
	fmt.Println("maximum zoom level is", max_zoom)

	concurrency := *concurrencyPtr
	sem := make(chan bool, concurrency)

	fmt.Println("starting tiling with concurrency of", concurrency)

	// outer loop for zoom
	for {
		if z == max_zoom {
			// do nothing
		} else {
			// halve image size
			src = imaging.Resize(src, size.X/2, 0, imaging.NearestNeighbor)
			// recalculate size
			size = src.Bounds().Max
		}

		fmt.Print(fmt.Sprintf("zoom level: %d (%d x %d)\n", z, size.X, size.Y))

		for y := 0; y < (size.Y / tile_size_y); y++ {
			for x := 0; x < (size.X / tile_size_x); x++ {
				sem <- true
				go tile(*baseName, src, z, x, y, tile_size_x, tile_size_y, sem)
			}

		}

		z--
		if z < 0 {
			break
		}
	}

	// drain at the end of each zoom level
	// since we are about to modify the source image in memory
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	fmt.Println("done")
}

func tile(basename string, src image.Image, z, x, y int, tile_size_x, tile_size_y int, sem chan bool) {
	defer func() { <-sem }()
	output_filename := fmt.Sprintf("%s-%d-%d-%d.png", basename, z, x, y)
	cropped := imaging.Crop(src, image.Rect(tile_size_x*x, tile_size_y*y, tile_size_x*x+tile_size_x, tile_size_y*y+tile_size_y))

	writer, _ := os.Create(output_filename)
	err := png.Encode(writer, cropped)
	if err != nil {
		fmt.Println(err)
	}
	writer.Close()

	return
}
