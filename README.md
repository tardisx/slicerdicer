# slicerdicer

Slice and dice an image, turning it into many equal sized tiles. Useful
for things like leaflet.js, with the Leaflet.Zoomify plugin.

The image is sliced up into equal sized tiles, based on the command line 
option `--tile-size` (default 512 pixels). 

Once the tiling is finished, the original is resized to half its current
dimensions (the orignal file on disk is not touched) and the process repeats.
Each halving is a new "zoom level".

Each file is named something like:

    tile-z-x-y.png

Where 'z' is the zoom level, x and y are the coordinates, with 0,0 being
the top left tile.

## Usage

    slicerdicer --help

    slicerdicer --filename foo.png --tile-size 256
