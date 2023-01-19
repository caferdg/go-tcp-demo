package main

import (
	"image"
	"image/color"
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
	r, g, b, a := pixel.RGBA()
	r = r * 2
	g = g * 2
	b = b * 2
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func openImage(path string) [][]color.Color {
	file, err := os.Open(path)
	check(err)
	img, format, err := image.Decode(file)
	check(err)
	if format != "jpeg" {
		print("Not jpeg ...")
		return nil
	}
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	mat := make([][]color.Color, height)
	for y := 0; y < height; y++ {
		mat[y] = make([]color.Color, width)
		for x := 0; x < width; x++ {
			mat[y][x] = img.At(x, y)
		}
	}
	return mat
}

func

func main() {
	img := openImage("test.jpg")

	jobChannel = make(chan job, 100)
	resChannel = make(chan res, 100)

	for i := 0; i < nbWorkers; i++ {
		go worker(jobChannel, resChannel)
	}
	
	for y, line := range img {
		for x, pixel := range line {
			jobChannel <- job{pixel, x, y}
		}
	}

	for {
		result := <-resChannel
		img[result.y][result.x] = result.pixel
	}
}