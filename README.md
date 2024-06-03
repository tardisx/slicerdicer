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

    slicerdicer -help

    slicerdicer -filename large_image.png -tile-size 256 -concurrency 5

## Output filenames

The destination for the tiles can be changed with the `-basename` and
`-path-template` options. The path template must contain 5 placeholders,
in the following order:

* `%s` basename, as per `-basename` (default `tile`)
* `%d` zoom level
* `%d` 'x' coordinate
* `%d` 'y' coordinate
* `%d` file format (jpg or png)

The default template is `%s-%d-%d-%d.%s` which results in a flat structure
with all files in the current directory.

For example, using `-basename map` and
`-path-template '%s/zoom-%d/%d-%d.%s'` will result in a file structure like:

    map
    ├── 0
    │   └── 0-0.png
    ├── 1
    │   ├── 0-0.png
    │   └── 1-0.png
    └── 2
        ├── 0-0.png
        ├── 0-1.png
        ├── 1-0.png
        ├── 1-1.png
        ├── 2-0.png
        ├── 2-1.png
        ├── 3-0.png
        └── 3-1.png

All tiles in a directory called `map`, with a second level directory for zoom
level, each file named `x-y.png` within that.

## Notes

### Memory

In my tests on an 32641 x 16471, 8-bit/color RGB PNG, memory usage peaks at
around 2.7GB.

### Speed

On that same test image, the run takes around 63 seconds to create the 11179
tiles, on my fairly underwhelming MacBookPro12,1 (dual core i5).
