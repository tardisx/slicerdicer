package main;

import "image"
import "image/png"
import "github.com/disintegration/imaging"
import "runtime"
import "flag"
import "fmt"
import "os"

func main() {

  filenamePtr  := flag.String("filename", "screenshot.png", "filename to open")
  tileSizePtr := flag.Int   ("tile-size", 512, "tile size, in pixels")
  concurrencyPtr := flag.Int ("concurrency", 5, "how many tiles to generate concurrently (threads)")

  flag.Parse()

  fmt.Println("opening file:", *filenamePtr)
  src, err := imaging.Open(*filenamePtr)
  if err != nil {
    fmt.Println("could not open file:", err)
    return;
  }

  size := src.Bounds().Max

  tile_size_x := *tileSizePtr
  tile_size_y := *tileSizePtr

  z := 0

  concurrency := *concurrencyPtr
  sem := make(chan bool, concurrency)

  fmt.Println("starting tiling with concurrency of", concurrency)

  // outer loop for zoom
  for {
    if (z == 0) {
      // do nothing
    } else {
      // halve image size
      src = imaging.Resize(src, size.X/2, 0, imaging.NearestNeighbor)
      runtime.GC()
      // recalculate size
      size = src.Bounds().Max
      // we are done if we are now smaller then the tile
      if (size.X < tile_size_x || size.Y < tile_size_y) {
        break;
      }
    }

    fmt.Print(fmt.Sprintf("zoom level: %d (%d x %d)\n", z, size.X, size.Y))

    for y := 0 ; y <= (size.Y / tile_size_y) ; y++ {
      for x := 0 ; x <= (size.X / tile_size_x) ; x++ {
        sem <- true
        go tile(src, z, x, y, tile_size_x, tile_size_y, sem)
      }

    }

    z++
  }

  // drain at the end of each zoom level
  // since we are about to modify the source image
  for i := 0; i < cap(sem); i++ {
    sem <- true
  }

  fmt.Println("done")
}

func tile (src image.Image, z, x, y int, tile_size_x, tile_size_y int, sem chan bool) {
  defer func() { <-sem }()
  output_filename := fmt.Sprintf("tile-%d-%d-%d.png", z, x, y)
  cropped := imaging.Crop(src, image.Rect(tile_size_x*x, tile_size_y*y, tile_size_x*x+tile_size_x, tile_size_y*y+tile_size_y));

  writer, _ := os.Create(output_filename)
  err := png.Encode(writer, cropped)
  if err != nil {
    fmt.Println(err)
  }
  writer.Close()

  runtime.GC()
  return;
}
