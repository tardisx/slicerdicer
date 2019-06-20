package main;

import "image"
import "image/png"
import "github.com/disintegration/imaging"
import "runtime"
import "flag"


import "fmt"
import "os"

func main() {

  filenamePtr := flag.String("filename", "screenshot.png", "filename to open")
  tile_size   := flag.Int   ("tile-size", 512, "tile size, in pixels")

  flag.Parse()

  fmt.Println("opening file:", *filenamePtr)
  src, err := imaging.Open(*filenamePtr)
  if err != nil {
    fmt.Println("could not open file:", err)
    return;
  }

  size := src.Bounds().Max

  tile_size_x := *tile_size
  tile_size_y := *tile_size

  fmt.Println("starting tiling")

  z := 0

  // outer loop for zoom
  for {
    if (z == 0) {
      // do nothing
    } else {
      // halve image size
      src = imaging.Resize(src, size.X/2, 0, imaging.Linear)
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
 
        output_filename := fmt.Sprintf("tile-%d-%d-%d.png", z, x, y)
        cropped := imaging.Crop(src, image.Rect(tile_size_x*x, tile_size_y*y, tile_size_x*x+tile_size_x, tile_size_y*y+tile_size_y));

        fmt.Print("writing to: ", output_filename, "        ");
        fmt.Print("\r")

        writer, _ := os.Create(output_filename)
        err = png.Encode(writer, cropped)
        writer.Close()
        runtime.GC()
        if err != nil {
          fmt.Println(err)
        }
      }
    }

    fmt.Print("\r                                   \r")
    z++
  }

  fmt.Println("done")

}
