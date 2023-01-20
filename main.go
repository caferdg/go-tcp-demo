package main

import (
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	"os"
)

type job struct {
	pixel color.Color
	x     int
	y     int
}

type res struct {
	pixel color.Color
	x     int
	y     int
}

var jobChannel chan job
var resChannel chan res

const nbWorkers int = 10

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func worker(jobCh chan job, resCh chan res) {
	for {
		job := <-jobCh
		var result res
		result.x = job.x
		result.y = job.y
		result.pixel = editPixel(job.pixel)
		resCh <- result
	}
}

func editPixel(pixel color.Color) color.Color {
	realColor, _ := color.RGBAModel.Convert(pixel).(color.RGBA)
	grey := uint8(float64(realColor.R)*0.1 + float64(realColor.G)*0.9 + float64(realColor.B)*0.1)
	newColor := color.RGBA{
		grey,
		grey,
		grey,
		realColor.A,
	}
	return newColor
}

func openImage(path string) ([][]color.Color, string) {
	// Takes an image path and returns a matrix of color and the image format
	file, err := os.Open(path)
	check(err)
	img, format, err := image.Decode(file)
	check(err)
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	mat := make([][]color.Color, width)
	for x := 0; x < width; x++ {
		mat[x] = make([]color.Color, height)
		for y := 0; y < height; y++ {
			mat[x][y] = img.At(x, y)
		}
	}
	return mat, format
}

func main() {
	filePath := os.Args[1]
	outPath := os.Args[2]
	img, format := openImage(filePath)
	jobChannel = make(chan job, 1000)
	resChannel = make(chan res, 1000)

	for i := 0; i < nbWorkers; i++ {
		go worker(jobChannel, resChannel)
	}

	go func() {
		for x, column := range img {
			for y, pixel := range column {
				jobChannel <- job{pixel, x, y}
			}
		}
	}()

	rect := image.Rect(0, 0, len(img), len(img[0]))
	finalImage := image.NewRGBA(rect)

	for i := 0; i < len(img)*len(img[0]); i++ {
		result := <-resChannel
		finalImage.Set(result.x, result.y, result.pixel)
	}

	out, err := os.Create(outPath)
	check(err)
	if format == "jpeg" {
		check(err)
		jpeg.Encode(out, finalImage, nil)
	}
	if format == "png" {
		png.Encode(out, finalImage)
	}

}
