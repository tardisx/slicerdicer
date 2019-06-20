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

  for y := 0 ; y <= (size.Y / tile_size_y) ; y++ {

    for x := 0 ; x <= (size.X / tile_size_x) ; x++ {
 
      output_filename := fmt.Sprintf("tile-0-%d-%d.png", x, y)
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

  fmt.Println()
  fmt.Println("done")

}
