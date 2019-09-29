package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
)

func main() {
	flag.Usage = func() {
		txt := `Usage: %s [flags] infile outfile

Converts png images to pacman rom files.
Input image should be a paletted png file.

Flags:
`
		fmt.Fprintf(os.Stderr, txt, os.Args[0])
		flag.PrintDefaults()
	}

	var palflag = flag.Bool("p", false, "Just output palette (32 colour entries)")
	var spriteFlag = flag.Bool("s", false, "Convert as 16x16 sprites (default is 8x8 characters)")

	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(2)
	}

	var infilename = flag.Arg(0)
	var outfilename = flag.Arg(1)

	// Open the infile.
	infile, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	// Decode the image.
	m, err := png.Decode(infile)
	if err != nil {
		log.Fatal(err)
	}
	p, ok := m.(image.PalettedImage)
	if !ok {
		log.Fatal("Image is not paletted")
	}

	outfile, err := os.Create(outfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	if *palflag {
		err = dumpPalette(p, outfile)
	} else if *spriteFlag {
		err = dumpSprites(outfile, p)
	} else {
		err = dumpChars(outfile, p)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func dumpPalette(m image.PalettedImage, out io.Writer) error {
	var pal color.Palette
	pal = m.ColorModel().(color.Palette)

	if len(pal) != 32 {
		return fmt.Errorf("Expected 32 colours, got %d", len(pal))
	}
	buf := make([]uint8, 32)
	for i := 0; i < len(pal); i++ {
		c := pal[i].(color.RGBA)
		// output format is
		// bbgggrrr

		buf[i] = (c.B & 0xc0) | (c.G&0xe0)>>2 | (c.R&0xe0)>>5

		//	fmt.Printf("%d,%d,%d => 0x%02x\n", c.R, c.G, c.B, buf[i])
	}
	_, err := out.Write(buf)
	return err
}

// Dump out 8x8 characters
func dumpChars(outfile *os.File, m image.PalettedImage) error {

	chars_w := m.Bounds().Dx() / 8
	chars_h := m.Bounds().Dy() / 8
	if chars_w*chars_h != 256 {
		return fmt.Errorf("bad size - expected 256 8x8 characters, got %d", chars_w*chars_h)
	}
	var n_chars int = 0
	buf := []byte{}
	for cy := 0; cy < chars_h && n_chars < 256; cy++ {
		for cx := 0; cx < chars_w && n_chars < 256; cx++ {
			n_chars++
			// lower 4 rows first
			buf = append(buf, encode_8x4(m, (cx*8)+7, (cy*8)+7)...)
			// then upper 4 rows
			buf = append(buf, encode_8x4(m, (cx*8)+7, (cy*8)+3)...)
		}
	}
	_, err := outfile.Write(buf)
	return err
}

// Dump out 16x16 sprites
func dumpSprites(outfile *os.File, m image.PalettedImage) error {

	chars_w := m.Bounds().Dx() / 16
	chars_h := m.Bounds().Dy() / 16
	if chars_w*chars_h != 64 {
		return fmt.Errorf("bad size - expected 64 16x16 sprites, got %d", chars_w*chars_h)
	}
	var n_chars int = 0
	buf := []byte{}

	for cy := 0; cy < chars_h && n_chars < 64; cy++ {
		for cx := 0; cx < chars_w && n_chars < 64; cx++ {
			n_chars++
			x := cx * 16
			y := cy * 16
			// ordering of 8x4 chunks:
			// (so bottom-right chunk is output first)
			// 5 1
			// 6 2
			// 7 3
			// 4 0
			buf = append(buf, encode_8x4(m, x+15, y+15)...)
			buf = append(buf, encode_8x4(m, x+15, y+3)...)
			buf = append(buf, encode_8x4(m, x+15, y+7)...)
			buf = append(buf, encode_8x4(m, x+15, y+11)...)

			buf = append(buf, encode_8x4(m, x+7, y+15)...)
			buf = append(buf, encode_8x4(m, x+7, y+3)...)
			buf = append(buf, encode_8x4(m, x+7, y+7)...)
			buf = append(buf, encode_8x4(m, x+7, y+11)...)
		}
	}
	_, err := outfile.Write(buf)
	return err
}

func encode_8x4(m image.PalettedImage, x int, y int) []byte {
	out := []byte{}
	for col := 0; col < 8; col++ {
		var b byte = 0
		for row := 0; row < 4; row++ {
			pix := m.ColorIndexAt(x-col, y-row)
			b |= (pix & 0x01) << uint(row)
			b |= ((pix & 0x02) >> 1) << uint(4+row)
		}
		out = append(out, b)
	}
	return out
}
