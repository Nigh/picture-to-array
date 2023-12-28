package main

import (
	"fmt"
	"image/gif"
	"os"
	"testing"
)

var testPng string = "./test_pic/pic0-png.png"
var testStaticGif16 string = "./test_pic/pic0-16-gif.gif"
var testStaticGif256 string = "./test_pic/pic0-256-gif.gif"

var testAnimateGif16Array []string = []string{"./test_pic/ani-16-48frames.gif", "./test_pic/ani-16-58frames.gif", "./test_pic/ani-16-108frames.gif", "./test_pic/ani-16-90frames.gif"}
var testAnimateGif256Array []string = []string{"./test_pic/ani-256-48frames.gif", "./test_pic/ani-256-58frames.gif"}

func testGif(path string) {
	f1, err := os.Open(path)
	check(err)
	defer f1.Close()
	fmt.Println("FILE", path)

	gifx, _ := gif.DecodeAll(f1)
	var avrFrameLength int = 0
	bits := len(gifx.Image[0].Palette)
	bitLen := 0
	for bits > 0 {
		bitLen += 1
		bits /= 2
	}
	for _, p := range gifx.Image {
		avrFrameLength += len(p.Pix) * bitLen
	}
	avrFrameLength /= len(gifx.Image)
	fmt.Printf("avrFrameLength = %d bytes in %d frames\n", avrFrameLength/8, len(gifx.Image))
}

func TestPictureTest(t *testing.T) {
	testGif(testStaticGif16)
	testGif(testStaticGif256)
	for _, v := range testAnimateGif16Array {
		testGif(v)
	}
	for _, v := range testAnimateGif256Array {
		testGif(v)
	}
}

func TestMain(t *testing.T) {

}
