// gif2array project main.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"path/filepath"
	"strings"
	//"image/color"
	//"image/color/palette"
	"image/gif"

	_ "github.com/hotei/bmp"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	//	"golang.org/x/image/bmp"

	"github.com/rubenfonseca/fastimage"
	// "net/http"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

var (
	hp          bool
	input_name  string
	output_name string
	alpha       bool
	black       bool
	white       bool
	dir         bool
)

func init() {
	flag.BoolVar(&hp, "h", false, "help")
	flag.StringVar(&input_name, "in", "", "the picture file for convert to c format array")
	flag.StringVar(&output_name, "out", "output", "the c format array output filename")
	flag.BoolVar(&alpha, "a", false, "alpha mode - alpha pixel as 0")
	flag.BoolVar(&black, "b", false, "black mode - black pixel as 1")
	flag.BoolVar(&white, "w", false, "white mode - white pixel as 1")
	flag.BoolVar(&dir, "d", false, "dir mode - input as directory, output as bitmap array")
}

var output_buffer bytes.Buffer
var w, h int

func get_byte_size(w, h int) int {
	if h/8*8 < h {
		return w * (1 + (h / 8))
	}
	return w * (h / 8)
}
func main() {
	flag.Parse()
	if !hp && len(flag.Args()) == 1 {
		input_name = flag.Arg(0)
	}
	if hp || len(input_name) == 0 {
		flag.Usage()
		return
	}
	if !alpha && !black && !white {
		alpha = true
	}

	f, _ := os.Stat(input_name)
	dir = f.IsDir()
	if dir {
		var file_count int = 0
		var image_size int = 0
		output_name = filepath.Base(input_name)
		filepath.Walk(input_name, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				return nil
			}
			file_count += 1

			f0, err := os.Open(path)
			check(err)

			ic, _, err := image.DecodeConfig(f0)
			check(err)
			w = ic.Width
			h = ic.Height
			f0.Close()
			if get_byte_size(w, h) > image_size {
				image_size = get_byte_size(w, h)
			}

			return nil
		})

		output_buffer.WriteString(fmt.Sprintf("const unsigned char %s[%d][%d]={", output_name, file_count, image_size))

		filepath.Walk(input_name, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				return nil
			}
			println(path)

			f1, err := os.Open(path)
			check(err)
			output_buffer.WriteString(fmt.Sprintf("\n{"))
			img, _, err := image.Decode(f1)
			image2buffer(img, w, h, &output_buffer)
			output_buffer.WriteString(fmt.Sprintf("\t//%s\n},", f.Name()))
			f1.Close()
			return nil
		})
		output_buffer.WriteString("\n};")

		output_buffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_array[]={", output_name))
		for i := 0; i < file_count; i++ {
			output_buffer.WriteString(fmt.Sprintf("\n\t{%d,%d,(unsigned char*)&(%s[%d][0])},", w, h, output_name, i))
		}
		output_buffer.WriteString("\n};\n")

	} else {
		output_name = strings.TrimSuffix(filepath.Base(input_name), filepath.Ext(input_name))
		f1, err := os.Open(input_name)
		check(err)
		defer f1.Close()

		img, _, err := image.Decode(f1)
		check(err)

		w = img.Bounds().Size().X
		h = img.Bounds().Size().Y
		fmt.Println("Width:", w, "Height:", h)

		img_type, frame := attrImage(f1)
		f1.Seek(0, 0)
		switch img_type {
		case "GIF":
			if frame > 1 {
				output_buffer.WriteString(fmt.Sprintf("const unsigned char %s[%d][%d]=\n{", output_name, frame, get_byte_size(w, h)))
				g, err := gif.DecodeAll(f1)
				check(err)
				gif2buffer(g, w, h, &output_buffer)
				output_buffer.WriteString("\n};")

				output_buffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_anime[]={", output_name))
				for i := 0; i < frame; i++ {
					output_buffer.WriteString(fmt.Sprintf("\n\t{%d,%d,(unsigned char*)&(%s[%d][0])},", w, h, output_name, i))
				}
				output_buffer.WriteString("\n};\n")
			} else {
				output_buffer.WriteString(fmt.Sprintf("const unsigned char %s[%d]=\n{", output_name, get_byte_size(w, h)))
				g, err := gif.DecodeAll(f1)
				check(err)
				gif2buffer(g, w, h, &output_buffer)
				output_buffer.WriteString("\n};")

				output_buffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_icon={%d,%d,%s[0]};\n", output_name, w, h, output_name))
			}
		default:
			output_buffer.WriteString(fmt.Sprintf("const unsigned char %s[%d]=\n{", output_name, get_byte_size(w, h)))
			image2buffer(img, w, h, &output_buffer)
			output_buffer.WriteString("\n};")
			output_buffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_icon={%d,%d,%s};\n", output_name, w, h, output_name))
		}
	}

	if checkFileIsExist(output_name + ".c") {
		check(os.Remove(output_name + ".c"))
	}
	output, err := os.Create(output_name + ".c")
	check(err)
	defer output.Close()

	output.WriteString(output_buffer.String())
	// fmt.Println(output_buffer.String())
	fmt.Println("Convert Complete!")
}

func attrImage(f1 *os.File) (img_type string, frame int) {
	f1.Seek(0, 0)
	imagetype, _, _ := fastimage.DetectImageTypeFromReader(f1)
	f1.Seek(0, 0)
	switch imagetype {
	case fastimage.GIF:
		img_type = "GIF"
		fmt.Println("GIF desu")
	default:
		img_type = "OTHER FORMAT"
		fmt.Println("GIF dewanai")
	}
	frame = 1
	if imagetype == fastimage.GIF {
		g, _ := gif.DecodeAll(f1)
		frame = len(g.Image)
	}
	return img_type, frame
}

func image2buffer(img image.Image, w int, h int, buffer *bytes.Buffer) {
	var x, y, line int
	for line = 0; line < 1+(h-1)/8; line++ {
		buffer.WriteString("\n\t")
		for x = 0; x < w; x++ {
			var c_byte uint8 = 0
			for y = line * 8; y < line*8+8; y++ {
				c_byte >>= 1
				if y < h {
					r, g, b, a := img.At(x, y).RGBA()
					// fmt.Println(r, g, b, a, (r>>8)+(g>>8)+(b>>8))
					if alpha && a > 0 {
						c_byte |= 0x80
					} else if white && (r>>8)+(g>>8)+(b>>8) > 200 {
						c_byte |= 0x80
					} else if black && (r>>8)+(g>>8)+(b>>8) < 100 {
						c_byte |= 0x80
					}
				}
			}
			buffer.WriteString(fmt.Sprintf("0x%02X,", c_byte))
		}
	}
}

func gif2buffer(gif *gif.GIF, w int, h int, buffer *bytes.Buffer) {
	for _, img := range gif.Image {
		var x, y, line, cnt int
		cnt = 0
		if len(gif.Image) > 1 {
			buffer.WriteString("\n\t{")
		} else {
			buffer.WriteString("\n")
		}
		for line = 0; line < 1+(gif.Config.Height-1)/8; line++ {
			if len(gif.Image) > 1 {
				buffer.WriteString("\n\t\t")
			} else {
				buffer.WriteString("\n\t")
			}
			for x = 0; x < gif.Config.Width; x++ {
				var c_byte uint8 = 0
				for y = line * 8; y < line*8+8; y++ {
					c_byte >>= 1
					if y < gif.Config.Height {
						r, g, b, a := img.At(x, y).RGBA()
						//fmt.Println(x, y, r, g, b, a)
						if alpha && a > 0 {
							c_byte |= 0x80
						} else if white && r+b+g > 200 {
							c_byte |= 0x80
						} else if black && r+b+g < 100 {
							c_byte |= 0x80
						}
					}
				}
				buffer.WriteString(fmt.Sprintf("0x%02X,", c_byte))
				cnt += 1
			}
		}
		if len(gif.Image) > 1 {
			buffer.WriteString("\n\t},")
		} else {
			buffer.WriteString("\n")
		}
	}

}
