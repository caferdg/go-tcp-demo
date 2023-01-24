package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net"
	"os"
)

var HOST string = "localhost"
var PORT string

var chanUsers chan user
var jobChannel chan job
var resChannel chan res

const nbWorkers int = 10

type user struct {
	connection net.Conn
	id         int
}

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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func manageUser(input chan user) {
	for {
		user := <-input
		//reader := bufio.NewReader(user.connection)
		//message, err := reader.ReadString('$')

		//io.WriteString(user.connection, fmt.Sprintf("Received time : %s$", time.Now()))
		inputFile, err := os.Create("input")
		check(err)

		_, err = io.Copy(inputFile, user.connection)
		check(err)
		print("File received\n")

		img, format := openImage("input")
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

		out, err := os.Create("output")
		check(err)
		if format == "jpeg" {
			check(err)
			jpeg.Encode(out, finalImage, nil)
		}
		if format == "png" {
			png.Encode(out, finalImage)
		}

		outputFile, err := os.Open("output")
		check(err)

		_, err = io.Copy(user.connection, outputFile)
		check(err)
		print("File sent\n")

		outputFile.Close()
		user.connection.Close()
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
	// Takes an image path and returns a matrix of colors and the image format
	file, err := os.Open(path)
	check(err)
	defer file.Close()
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
	if len(os.Args) != 2 {
		fmt.Printf("Usage : go run server.go <port>")
		os.Exit(1)
	}

	PORT = os.Args[1]

	socket := HOST + ":" + PORT
	listen, err := net.Listen("tcp", socket)
	check(err)

	fmt.Println("Listening to socket : ", socket)

	chanUsers = make(chan user, 10)

	go manageUser(chanUsers)

	nbUsers := 0

	for i := 0; i < nbWorkers; i++ {
		go worker(jobChannel, resChannel)
	}

	for {
		conn, err := listen.Accept()
		defer conn.Close()
		check(err)

		var newUser user
		newUser.id = nbUsers
		newUser.connection = conn

		chanUsers <- newUser
		nbUsers++
	}

}
