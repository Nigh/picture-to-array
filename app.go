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
	"regexp"
	"strconv"
	"strings"

	"localhost/picarray"

	_ "github.com/hotei/bmp"

	_ "image/jpeg"
	_ "image/png"

	"os"

	"github.com/rubenfonseca/fastimage"
)

var Version string = "v1.29"

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

var totalByteSize int = 0
var totalFileCount int = 0

type arrayElement struct {
	varName  string
	fileName string
}

var picArray map[string][]arrayElement

func walker(realPath string, f os.FileInfo, err error) error {
	// 忽略 . 与 _ 开头的文件和目录
	if f.IsDir() {
		if strings.HasPrefix(f.Name(), ".") || strings.HasPrefix(f.Name(), "_") {
			fmt.Println("[DIR] " + f.Name() + " !!!IGNORED!!!")
			return filepath.SkipDir
		}

		fmt.Println("[DIR] " + f.Name() + " " + realPath)
	} else {
		// TODO: maintain a ignore list in which type of files which would be genarated by system automatically, such as Thumb.db
		if strings.HasPrefix(f.Name(), ".") || strings.HasPrefix(f.Name(), "_") || strings.HasSuffix(f.Name(), "db") {
			fmt.Println("\t[FILE]\t「" + f.Name() + "」 !!!IGNORED!!!")
			return nil
		}
		relPath, _ := filepath.Rel(inputPath, realPath)
		fmt.Println("\t[FILE]\t「" + f.Name() + "」@『" + relPath + "』\t| " + strconv.FormatInt(f.Size(), 10) + " Bytes")
		relPath = strings.Replace(filepath.FromSlash(relPath), string(filepath.Separator), "_", -1)
		r := regexp.MustCompile(`\[.+?\]`)
		// DONE: 去除varName中的[.+?]
		varName := r.ReplaceAllString(strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(f.Name())), "")
		// DONE: 如果文件父目录含有[array]，则将文件加入picArray
		// varName, fileName
		if strings.Contains(filepath.Base(filepath.Dir(realPath)), "[array]") {
			dirVarName, _ := filepath.Rel(inputPath, realPath)
			dirVarName = filepath.Dir(dirVarName)
			dirVarName = strings.Replace(filepath.FromSlash(dirVarName), string(filepath.Separator), "_", -1)
			dirVarName = r.ReplaceAllString(dirVarName, "")
			picArray[dirVarName] = append(picArray[dirVarName],
				arrayElement{
					varName:  varName,
					fileName: f.Name(),
				})
		}
		totalByteSize += pic2c(realPath, varName, &dot_c_buffer, &dot_h_buffer)
		totalFileCount += 1
	}
	return nil
}

func pic2c(path string, varName string, cBuffer *bytes.Buffer, hBuffer *bytes.Buffer) (byteSize int) {
	f1, err := os.Open(path)
	check(err)
	defer f1.Close()

	img, _, err := image.Decode(f1)
	check(err)

	w = img.Bounds().Size().X
	h = img.Bounds().Size().Y
	byteSize = get_byte_size(w, h)
	strconv.Itoa(w)
	fmt.Println("\t\tSize:[" + strconv.Itoa(w) + "x" + strconv.Itoa(h) + "] " + strconv.Itoa(byteSize) + " bytes")
	cBuffer.WriteString(fmt.Sprintf("const uint8_t %s[%d] = {", varName, byteSize))
	picarray.Image2buffer(img, w, h, cBuffer)
	cBuffer.WriteString("\n};")
	cBuffer.WriteString(fmt.Sprintf("\nconst sBITMAP %s_bmp = {%d, %d, %s};\n", varName, w, h, varName))
	hBuffer.WriteString(fmt.Sprintf("extern const sBITMAP %s_bmp;\n", varName))
	return
}
func array2c(cBuffer *bytes.Buffer, hBuffer *bytes.Buffer) {
	for k, v := range picArray {
		if len(v) > 0 {
			cBuffer.WriteString(fmt.Sprintf("\nconst sBITMAP* %s_array[%d] = {\n", k, len(v)))
			for _, e := range v {
				cBuffer.WriteString("\t&" + e.varName + "_bmp,")
				cBuffer.WriteString(" // " + e.fileName + "\n")
			}
			cBuffer.WriteString("};\n")
			hBuffer.WriteString(fmt.Sprintf("extern const sBITMAP* %s_array[%d];\n", k, len(v)))
		}
	}
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

	picArray = make(map[string][]arrayElement)
	filepath.Walk(inputPath, walker)
	array2c(&dot_c_buffer, &dot_h_buffer)

	if checkFileIsExist(outputPath + ".c") {
		check(os.Remove(outputPath + ".c"))
	}
	if checkFileIsExist(outputPath + ".h") {
		check(os.Remove(outputPath + ".h"))
	}

	outputCFile, err := os.Create(outputPath + ".c")
	check(err)
	defer outputCFile.Close()
	outputHFile, err := os.Create(outputPath + ".h")
	check(err)
	defer outputHFile.Close()

	versionStr := "// clang-format off\n// *INDENT-OFF*\n// Generated by https://github.com/Nigh/picture-to-array\n// Version:" + Version + "\n"
	outputCFile.WriteString(versionStr)
	outputHFile.WriteString(versionStr)

	outputCFile.WriteString(`#include "bitmap.h"` + "\n\n")
	outputCFile.WriteString(dot_c_buffer.String())

	hash := sha1.New()
	hash.Write(dot_h_buffer.Bytes())
	hashStr := hex.EncodeToString(hash.Sum(nil))

	outputHFile.WriteString("#ifndef _" + string(hashStr) + "_\n")
	outputHFile.WriteString("#define _" + string(hashStr) + "_\n")
	outputHFile.WriteString(`#include "bitmap.h"` + "\n\n")
	outputHFile.WriteString(dot_h_buffer.String())
	outputHFile.WriteString("#endif\n")

	fmt.Println("Total " + strconv.Itoa(totalFileCount) + " Files")
	fmt.Println("Total " + strconv.Itoa(totalByteSize) + " Bytes")
	fmt.Println("Hash = " + hashStr)
	fmt.Println("Convert Complete!")
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
