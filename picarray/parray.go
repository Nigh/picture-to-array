package picarray

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type ColorMode int

const (
	Alpha ColorMode = iota
	White
	Black
	MonoColor
	RGB565
	RGB888
)

// TODO: bit order && byte order

var colorMode ColorMode = Alpha

func SetMode(mode ColorMode) {
	colorMode = mode
}

func GetMode() ColorMode {
	return colorMode
}

func Image2buffer(img image.Image, w int, h int, buffer *bytes.Buffer) {
	if colorMode < MonoColor {
		MonoImage2buffer(img, w, h, buffer)
	} else {
		RGBImage2buffer(img, w, h, buffer)
	}
}

func MonoImage2buffer(img image.Image, w int, h int, buffer *bytes.Buffer) {
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
					switch colorMode {
					case Alpha:
						if a > 0 {
							c_byte |= 0x80
						}
					case White:
						if (r>>8)+(g>>8)+(b>>8) > 192 {
							c_byte |= 0x80
						}
					case Black:
						if (r>>8)+(g>>8)+(b>>8) < 64 {
							c_byte |= 0x80
						}
					}
				}
			}
			if x > 0 {
				buffer.WriteString(" ")
			}
			buffer.WriteString(fmt.Sprintf("0x%02X,", c_byte))
		}
	}
}

func RGBImage2buffer(img image.Image, w int, h int, buffer *bytes.Buffer) {
	var x, y int
	for y = 0; y < h; y++ {
		buffer.WriteString("\n\t")
		for x = 0; x < w; x++ {
			var c_byte []uint8
			r, g, b, _ := img.At(x, y).RGBA()
			switch colorMode {
			case RGB565:
				c_byte = append(c_byte, (uint8)((r&0xF8)|((g&0xFF)>>5)))
				c_byte = append(c_byte, (uint8)(((g<<3)&0xE0)|((b&0xFF)>>3)))
			case RGB888:
				c_byte = append(c_byte, (uint8)(r))
				c_byte = append(c_byte, (uint8)(g))
				c_byte = append(c_byte, (uint8)(b))
			}
			if x > 0 {
				buffer.WriteString(" ")
			}
			for i := 0; i < len(c_byte); i++ {
				buffer.WriteString(fmt.Sprintf("0x%02X,", c_byte[i]))
			}
		}
	}
}

func Gif2buffer(gif *gif.GIF, w int, h int, buffer *bytes.Buffer) {
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
						switch colorMode {
						case Alpha:
							if a > 0 {
								c_byte |= 0x80
							}
						case White:
							if (r>>8)+(g>>8)+(b>>8) > 200 {
								c_byte |= 0x80
							}
						case Black:
							if (r>>8)+(g>>8)+(b>>8) < 100 {
								c_byte |= 0x80
							}
						}
					}
				}
				if x > 0 {
					buffer.WriteString(" ")
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
