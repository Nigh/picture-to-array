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

func Image2buffer(img image.Image, varName string, buffer *bytes.Buffer) {
	w := img.Bounds().Size().X
	h := img.Bounds().Size().Y
	buffer.WriteString(fmt.Sprintf("const uint8_t %s[] = {", varName))
	if colorMode < MonoColor {
		MonoImage2buffer(img, w, h, buffer)
	} else {
		RGBImage2buffer(img, w, h, buffer)
	}
	buffer.WriteString("\n};")
	buffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_bmp = {%d, %d, %s};\n", varName, w, h, varName))
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

func Gif2buffer(gifx *gif.GIF, varName string, buffer *bytes.Buffer) (byteSize int) {
	byteSize = 0

	paletteLength := len(gifx.Image[0].Palette)
	if paletteLength != 16 && paletteLength != 256 {
		panic("GIF parser support 16 or 256 depth palette!!!")
	}
	// generate palette array (global)
	var c_byte []uint8
	buffer.WriteString(fmt.Sprintf("const uint8_t %s_palette[] = {", varName))
	for _, v := range gifx.Image[0].Palette {
		r, g, b, _ := v.RGBA()
		c_byte = append(c_byte, (uint8)((r&0xF8)|((g&0xFF)>>5)))
		c_byte = append(c_byte, (uint8)(((g<<3)&0xE0)|((b&0xFF)>>3)))
	}
	for i := 0; i < len(c_byte); i++ {
		if i&0xF == 0 {
			buffer.WriteString("\n\t")
		}
		buffer.WriteString(fmt.Sprintf("0x%02X,", c_byte[i]))
	}
	buffer.WriteString("\n};\n")
	byteSize += len(c_byte)

	for i, p := range gifx.Image {
		var postfix string
		if len(gifx.Image) > 1 {
			postfix = fmt.Sprintf("_%03d", i)
		} else {
			postfix = ""
		}
		// generate pix map
		buffer.WriteString(fmt.Sprintf("const uint8_t %s_pixel%s[] = {", varName, postfix))
		for i2, v := range p.Pix {
			if paletteLength == 16 {
				if i2&0x1F == 0 {
					buffer.WriteString("\n\t")
				}
			} else {
				if i2&0xF == 0 {
					buffer.WriteString("\n\t")
				}
			}
			if paletteLength == 16 {
				if i2%2 == 0 {
					buffer.WriteString(fmt.Sprintf("0x%X", v))
				} else {
					buffer.WriteString(fmt.Sprintf("%X,", v))
				}
			} else {
				buffer.WriteString(fmt.Sprintf("0x%02X,", v))
			}
		}
		if paletteLength == 16 && len(p.Pix)%2 == 1 {
			buffer.WriteString("0,")
		}
		if paletteLength == 16 {
			byteSize += len(p.Pix) / 2
		} else {
			byteSize += len(p.Pix)
		}
		buffer.WriteString("\n};\n")

		// generate frame
		buffer.WriteString(fmt.Sprintf("const sFRAME %s_frame%s = {", varName, postfix))
		if paletteLength == 16 {
			buffer.WriteString("\n\t.paletteDepth = 16,")
		} else {
			buffer.WriteString("\n\t.paletteDepth = 0,")
		}
		buffer.WriteString(fmt.Sprintf("\n\t.palette = %s_palette,", varName))
		rx := p.Rect.Min.X
		ry := p.Rect.Min.Y
		rw := p.Rect.Max.X - rx
		rh := p.Rect.Max.Y - ry
		buffer.WriteString(fmt.Sprintf("\n\t.drawArea = {%d,%d,%d,%d},", rx, ry, rw, rh))

		buffer.WriteString(fmt.Sprintf("\n\t.pixel = %s_pixel%s,", varName, postfix))
		buffer.WriteString("\n};\n")
		byteSize += 1 + 4 + 8 + 4
	}

	if len(gifx.Image) > 1 {
		// generate frame pointer array
		buffer.WriteString(fmt.Sprintf("const sFRAME* %s_frames[] = {", varName))
		for i := range gifx.Image {
			buffer.WriteString(fmt.Sprintf("\n\t&%s_frame_%03d,", varName, i))
		}
		buffer.WriteString("\n};\n")
		byteSize += len(gifx.Image) * 4

		// generate gif struct
		buffer.WriteString(fmt.Sprintf("const sGIF %s_gif = {", varName))
		buffer.WriteString(fmt.Sprintf("\n\t.frameCount = %d,", len(gifx.Image)))
		buffer.WriteString(fmt.Sprintf("\n\t.frames = %s_frames,", varName))
		buffer.WriteString("\n};\n")
		byteSize += 1 + 4
	}
	return
}
