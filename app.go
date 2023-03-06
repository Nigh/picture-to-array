// pic2array project main.go
package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"localhost/picarray"

	_ "github.com/hotei/bmp"

	_ "image/jpeg"
	_ "image/png"

	"os"
)

var Version string = "v1.41"

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
	colorMode  string
)

func getParrayColorMode() func(string) picarray.ColorMode {
	colormode := map[string]picarray.ColorMode{
		"alpha":  picarray.Alpha,
		"black":  picarray.Black,
		"white":  picarray.White,
		"rgb565": picarray.RGB565,
		"rgb888": picarray.RGB888,
	}
	return func(key string) picarray.ColorMode {
		if _, ok := colormode[key]; ok {
			return colormode[key]
		}
		return picarray.Alpha
	}
}

func colormodeExplain() string {
	ret := "alpha|black|white|rgb565|rgb888"
	ret += "\nalpha - alpha pixel as 0"
	ret += "\nblack - black pixel as 1"
	ret += "\nwhite - white pixel as 1"
	ret += "\nrgb565 - 16bit R5 G6 B5"
	ret += "\nrgb888 - 24bit R8 G8 B8"
	return ret
}
func init() {
	flag.BoolVar(&hp, "h", false, "help")
	flag.StringVar(&inputPath, "in", "", "the picture file or dir for convert to c language array")
	flag.StringVar(&outputPath, "out", "", "the c format array output filename")
	flag.StringVar(&colorMode, "c", "alpha", colormodeExplain())
}

type picUnit struct {
	name string
	cbuf bytes.Buffer
	hbuf bytes.Buffer
}

// 片段buffer，用于排序后输出至最终文件
var picUnits []picUnit

// 最终输出文件buffer
var finalCFileBuffer bytes.Buffer
var finalHFileBuffer bytes.Buffer
var w, h int

func get_byte_size(w, h int) int {
	if picarray.GetMode() < picarray.MonoColor {
		if h/8*8 < h {
			return w * (1 + (h / 8))
		}
		return w * (h / 8)
	} else {
		if picarray.GetMode() == picarray.RGB565 {
			return w * h * 2
		}
		if picarray.GetMode() == picarray.RGB888 {
			return w * h * 3
		}
	}
	return 0
}

var totalByteSize int = 0
var totalFileCount int = 0

type arrayElement struct {
	varName  string
	fileName string
}

type picarrayElement struct {
	dirName  string
	varName  string
	fileName string
}

var picArrayMap map[string][]arrayElement
var picArraySlice [][]picarrayElement

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
			picArrayMap[dirVarName] = append(picArrayMap[dirVarName],
				arrayElement{
					varName:  varName,
					fileName: f.Name(),
				})
		}
		picUnits = append(picUnits, picUnit{})
		picUnits[len(picUnits)-1].name = varName
		totalByteSize += pic2c(
			realPath,
			picUnits[len(picUnits)-1].name,
			&picUnits[len(picUnits)-1].cbuf,
			&picUnits[len(picUnits)-1].hbuf)
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

func picslice2c(cBuffer *bytes.Buffer, hBuffer *bytes.Buffer) {
	for _, v := range picArraySlice {
		if len(v) > 0 {
			cBuffer.WriteString(fmt.Sprintf("\nconst sBITMAP* %s_array[%d] = {\n", v[0].dirName, len(v)))
			for _, e := range v {
				cBuffer.WriteString("\t&" + e.varName + "_bmp,")
				cBuffer.WriteString(" // " + e.fileName + "\n")
			}
			cBuffer.WriteString("};\n")
			hBuffer.WriteString(fmt.Sprintf("extern const sBITMAP* %s_array[%d];\n", v[0].dirName, len(v)))
		}
	}
}
func picmap2slice() {
	for name, v := range picArrayMap {
		if len(v) > 0 {
			picArraySlice = append(picArraySlice, []picarrayElement{})
			for _, e := range v {
				picArraySlice[len(picArraySlice)-1] = append(picArraySlice[len(picArraySlice)-1],
					picarrayElement{dirName: name, varName: e.varName, fileName: e.fileName})
			}
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

	picUnits = make([]picUnit, 0)
	picArraySlice = make([][]picarrayElement, 0)
	picArrayMap = make(map[string][]arrayElement)

	getC := getParrayColorMode()

	picarray.SetMode(getC(colorMode))

	filepath.Walk(inputPath, walker)

	// DONE: buffer排序后写入final buffer
	sort.SliceStable(picUnits, func(i, j int) bool {
		return strings.Compare(picUnits[i].name, picUnits[j].name) < 0
	})
	for _, v := range picUnits {
		finalCFileBuffer.Write(v.cbuf.Bytes())
		finalHFileBuffer.Write(v.hbuf.Bytes())
	}
	// DONE: picArray排序后再写入final buffer
	// first: picArrayMap to picArraySlice
	picmap2slice()
	// then: sort picArraySlice
	sort.SliceStable(picArraySlice, func(i, j int) bool {
		return strings.Compare(picArraySlice[i][0].dirName, picArraySlice[j][0].dirName) < 0
	})
	// finally: write picArraySlice to buffer
	picslice2c(&finalCFileBuffer, &finalHFileBuffer)

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
	outputCFile.WriteString(finalCFileBuffer.String())

	hash := sha1.New()
	hash.Write(finalCFileBuffer.Bytes())
	hashStr := hex.EncodeToString(hash.Sum(nil))

	outputHFile.WriteString("#ifndef _" + string(hashStr) + "_\n")
	outputHFile.WriteString("#define _" + string(hashStr) + "_\n")
	outputHFile.WriteString(`#include "bitmap.h"` + "\n\n")
	outputHFile.WriteString(finalHFileBuffer.String())
	outputHFile.WriteString("#endif\n")

	fmt.Println("Total " + strconv.Itoa(totalFileCount) + " Files")
	fmt.Println("Total " + strconv.Itoa(totalByteSize) + " Bytes")
	fmt.Println("Hash = " + hashStr)
	fmt.Println("Convert Complete!")
}
