package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"gonum.org/v1/gonum/stat/distuv"
)

var dir = flag.String("path", "images", "path to the images dir")

const size = 256

func main() {
	flag.Parse()

	a := analyzer{
		colors:  make(map[byte]int),
		changes: make(map[int]int),
	}

	files, _ := ioutil.ReadDir(*dir)
	for _, f := range files {
		file, err := os.Open(*dir + "/" + f.Name())
		if err != nil {
			log.Panicln("File open error", f.Name(), err)
		}
		img, err := png.Decode(file)
		if err != nil {
			log.Panicln("Image decode error", f.Name(), err)
		}
		a.process(img.(*image.NRGBA))
		log.Printf("Ready: " + f.Name())
	}

	max := 0
	for _, val := range a.changes {
		if max < val {
			max = val
		}
	}
	log.Println("Max changes: ", max)
	max = max / 2

	sample := make([]float64, size*size)
	for i, val := range a.changes {
		sample[i] = float64(val) / float64(max)
	}

	changes := image.NewNRGBA(image.Rect(0, 0, size, size))
	unchanged := image.NewNRGBA(image.Rect(0, 0, size, size))

	normal := distuv.Normal{
		Mu:    0.5,
		Sigma: 3,
	}
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			curr := a.changes[y*size+x]
			popularity := float64(curr) / float64(max)
			prob := 1 - normal.LogProb(popularity)
			slide := int(255 * 4 * prob * popularity)
			if slide > 255*4 {
				slide = 255 * 4
			}

			red := slide - 255*2
			green := 0
			if slide < 255*2 {
				green = slide
			} else {
				green = 4*255 - slide
			}
			blue := 2*255 - slide

			changes.Set(x, y, color.NRGBA{A: 255, R: uint8(clamp(red)), G: uint8(clamp(green)), B: uint8(clamp(blue))})
			if curr == 0 {
				unchanged.Set(x, y, color.NRGBA{A: 255, R: 255, G: 255, B: 255})
			} else {
				unchanged.Set(x, y, color.NRGBA{A: 255, R: 0, G: 0, B: 0})
			}
		}
	}

	b := bytes.Buffer{}
	_ = png.Encode(&b, changes)
	_ = ioutil.WriteFile("changes.png", b.Bytes(), 0644)
	b.Reset()
	_ = png.Encode(&b, unchanged)
	_ = ioutil.WriteFile("unchanged.png", b.Bytes(), 0644)
}

type analyzer struct {
	last    *image.NRGBA
	colors  map[byte]int
	changes map[int]int
	changed int
	frame   int
}

func (a *analyzer) process(img *image.NRGBA) {
	a.frame++
	if a.last == nil {
		a.last = img
		return
	}
	last := a.last
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			i := last.PixOffset(x, y)
			prev := last.Pix[i : i+4 : i+4]
			curr := img.Pix[i : i+4 : i+4]

			for i := 0; i < 4; i++ {
				if prev[i] != curr[i] {
					idx := getColorIndex(img.NRGBAAt(x, y))
					if idx == 16 {
						panic("Unknown color")
					}
					a.colors[idx] = a.colors[idx] + 1
					oldChanges := a.changes[y*size+x]
					a.changes[y*size+x] = oldChanges + 1
					if oldChanges == 0 {
						a.changed++
						if a.changed == size*size {
							panic(a.frame)
						}
					}
					break
				}
			}
		}
	}
	a.last = img
}

func getColorIndex(c color.NRGBA) byte {
	argb := (uint32(c.A) << 24) | (uint32(c.R) << 16) | (uint32(c.G) << 8) | uint32(c.B)
	switch argb {
	case 0xFFe4e4e4:
		return 0
	case 0xFFea7e35:
		return 1
	case 0xFFbe49c9:
		return 2
	case 0xFF6387d2:
		return 3
	case 0xFFc2b51c:
		return 4
	case 0xFF39ba2e:
		return 5
	case 0xFFd98199:
		return 6
	case 0xFF414141:
		return 7
	case 0xFFa0a7a7:
		return 8
	case 0xFF267191:
		return 9
	case 0xFF7e34bf:
		return 10
	case 0xFF253193:
		return 11
	case 0xFF56331c:
		return 12
	case 0xFF364b18:
		return 13
	case 0xFF9e2b27:
		return 14
	case 0xFF181414:
		return 15
	}
	return 16
}

func clamp(c int) int {
	if c > 255 {
		return 255
	}
	if c < 0 {
		return 0
	}
	return c
}
