# slicerdicer

Slice and dice an image, turning it into many equal sized tiles. Useful
for tools like leaflet.js, to create interactive "slippy" maps.

The image is sliced up into equal sized tiles, based on the command line 
option `--tile-size` (default 256 pixels). 

Once the tiling is finished, the original is resized to half its current
dimensions (the orignal file on disk is not touched) and the process repeats.
Each halving is a new "zoom level".

Each file is named something like:

    tile-z-x-y.png

Where 'z' is the zoom level, x and y are the coordinates, with 0,0 being
the top left tile.

## Usage

    slicerdicer --help

    slicerdicer --filename foo.png --tile-size 256 --concurrency 5

## Notes

### Memory

In my tests on an 32641 x 16471, 8-bit/color RGB PNG, memory usage peaks at
around 2.7GB.

### Speed

On that same test image, the run takes around 63 seconds to create the 11179
tiles, on my fairly underwhelming MacBookPro12,1 (dual core i5).
