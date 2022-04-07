// pic2array project main.go
package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"path/filepath"
	"strconv"
	"strings"

	"localhost/picarray"

	_ "github.com/hotei/bmp"

	_ "image/jpeg"
	_ "image/png"

	"os"

	"github.com/rubenfonseca/fastimage"
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
	hp         bool
	inputPath  string
	outputPath string
	alpha      bool
	black      bool
	white      bool
)

func init() {
	flag.BoolVar(&hp, "h", false, "help")
	flag.StringVar(&inputPath, "in", "", "the picture file or dir for convert to c language array")
	flag.StringVar(&outputPath, "out", "", "the c format array output filename")
	flag.BoolVar(&alpha, "a", false, "alpha mode - alpha pixel as 0")
	flag.BoolVar(&black, "b", false, "black mode - black pixel as 1")
	flag.BoolVar(&white, "w", false, "white mode - white pixel as 1")
}

var dot_c_buffer bytes.Buffer
var dot_h_buffer bytes.Buffer
var w, h int

var colorMode picarray.Mode = picarray.Alpha

func get_byte_size(w, h int) int {
	if h/8*8 < h {
		return w * (1 + (h / 8))
	}
	return w * (h / 8)
}

func walker(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		fmt.Println("Found Dir " + f.Name())
		filepath.Walk(path, walker)
	} else {
		fmt.Println("Found File " + f.Name())
	}
	return nil
}

func pic2array(path string, varName string, cBuffer *bytes.Buffer, hBuffer *bytes.Buffer) (byteSize int) {
	f1, err := os.Open(path)
	check(err)
	defer f1.Close()

	img, _, err := image.Decode(f1)
	check(err)

	w = img.Bounds().Size().X
	h = img.Bounds().Size().Y
	fmt.Println("Width:", w, "Height:", h)
	byteSize = get_byte_size(w, h)
	cBuffer.WriteString(fmt.Sprintf("const uint8_t %s[%d] = {\n", varName, byteSize))
	picarray.Image2buffer(img, w, h, cBuffer)
	cBuffer.WriteString("\n};")
	cBuffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_bmp = {%d, %d, %s};\n", varName, w, h, varName))
	hBuffer.WriteString(fmt.Sprintf("extern const sBITMAP %s_bmp;\n", varName))
	return
}

func main() {
	flag.Parse()
	if !hp && len(flag.Args()) == 1 {
		inputPath = flag.Arg(0)
	}
	if hp || len(inputPath) == 0 {
		flag.Usage()
		return
	}
	if len(outputPath) == 0 {
		flag.Usage()
		return
	}

	picarray.SetMode(picarray.Alpha)
	if black {
		picarray.SetMode(picarray.Black)
	} else if white {
		picarray.SetMode(picarray.White)
	}

	dot_c_buffer.WriteString(`#include "bitmap.h"` + "\n\n")
	dot_h_buffer.WriteString(`#include "bitmap.h"` + "\n\n")

	var totalByteSize int = 0
	var totalFileCount int = 0
	filepath.Walk(inputPath,
		func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				fmt.Println("[DIR]", f.Name(), path)
			} else {
				relPath, _ := filepath.Rel(inputPath, path)
				fmt.Println("\t[FILE]", "「"+f.Name()+"」@『"+relPath+"』", "\t|\t", strconv.FormatInt(f.Size(), 10)+" Bytes", path)
				relPath = filepath.FromSlash(relPath)
				relPath = strings.Replace(relPath, string(filepath.Separator), "_", -1)
				varName := strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(f.Name()))
				totalByteSize += pic2array(path, varName, &dot_c_buffer, &dot_h_buffer)
				totalFileCount += 1
			}
			return nil
		})

	if checkFileIsExist(outputPath + ".c") {
		check(os.Remove(outputPath + ".c"))
	}
	if checkFileIsExist(outputPath + ".h") {
		check(os.Remove(outputPath + ".h"))
	}

	outputCFile, err := os.Create(outputPath + ".c")
	check(err)
	defer outputCFile.Close()
	outputCFile.WriteString(dot_c_buffer.String())

	dot_h_buffer.WriteString("#endif\n")
	outputHFile, err := os.Create(outputPath + ".h")
	check(err)
	defer outputHFile.Close()

	hash := sha1.New()
	hash.Write(dot_h_buffer.Bytes())
	hashStr := hex.EncodeToString(hash.Sum(nil))

	outputHFile.WriteString("#ifndef _" + string(hashStr) + "_\n")
	outputHFile.WriteString("#define _" + string(hashStr) + "_\n")
	outputHFile.WriteString(dot_h_buffer.String())

	fmt.Println("Total " + strconv.Itoa(totalFileCount) + " Files")
	fmt.Println("Total " + strconv.Itoa(totalByteSize) + " Bytes")
	fmt.Println("Hash = " + hashStr)
	fmt.Println("Convert Complete!")
	return
}

func attrImage(f1 *os.File) (img_type string, frameLen int) {
	f1.Seek(0, 0)
	imagetype, _, _ := fastimage.DetectImageTypeFromReader(f1)
	f1.Seek(0, 0)
	switch imagetype {
	case fastimage.GIF:
		img_type = "GIF"
		fmt.Println("GIF desu")
	case fastimage.PNG:
		img_type = "PNG"
		fmt.Println("PNG desu")
	case fastimage.JPEG:
		img_type = "JPEG"
		fmt.Println("JPEG desu")
	case fastimage.BMP:
		img_type = "BMP"
		fmt.Println("BMP desu")
	default:
		img_type = "OTHER FORMAT"
		fmt.Println("Onknown format")
	}
	frameLen = 1
	if imagetype == fastimage.GIF {
		g, _ := gif.DecodeAll(f1)
		frameLen = len(g.Image)
	}
	return img_type, frameLen
}
