# png2pac

Converts png files into character/sprite roms for the pacman arcade hardware.

# Installation

Written in Go. To build and install:

    $ go get github.com/bcampbell/png2pac
    $ go install github.com/bcampbell/png2pac

# Invocation

```
Usage: png2pac [flags] infile outfile

Converts png images to pacman rom files.
Input image should be a paletted png file.

Flags:
  -p	Just output palette (32 colour entries)
  -s	Convert as 16x16 sprites (default is 8x8 characters)
```
