package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var N int
var A [][]float64
var B [][]float64
var C [][]float64

var nbWorkers int
var jobChannel chan job
var resChannel chan res

type job struct {
	x int
	y int
}

type res struct {
	value float64
	x     int
	y     int
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func write(text string, file *os.File) {
	_, err := file.WriteString(text)
	check(err)
}

func read(filename string) string {
	data, err := ioutil.ReadFile(filename)
	check(err)
	return string(data)
}

func calcCoef(line int, column int) float64 {
	var result float64
	for k := 0; k < N; k++ {
		result += A[line][k] * B[column][k]
	}
	return result
}

func worker(jobCh chan job, resCh chan res) {
	for {
		var job job
		job = <-jobCh
		var result res
		result.x = job.x
		result.y = job.y
		result.value = calcCoef(result.x, result.y)
		resCh <- result
	}
}

func main() {
	inputFile, err := os.OpenFile("input.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	defer inputFile.Close()
	check(err)

	data := read(inputFile.Name())
	slidedData := strings.Split(data, "-")

	matA := strings.Split(slidedData[0], "\n")
	matB := strings.Split(slidedData[1], "\n")[1:]
	N = len(matB)
	A = make([][]float64, N)
	B = make([][]float64, N)
	C = make([][]float64, N)

	for i := 0; i < N; i++ {
		A[i] = make([]float64, N)
		B[i] = make([]float64, N)
		C[i] = make([]float64, N)
		for j := 0; j < N; j++ {
			A[i][j], _ = strconv.ParseFloat(strings.Split(matA[i], " ")[j], 8)
			B[i][j], _ = strconv.ParseFloat(strings.Split(matB[i], " ")[j], 8)
		}
	}

	var jobChannel = make(chan job, N)
	var resChannel = make(chan res, N)

	nbWorkers = 4
	for i := 0; i < nbWorkers; i++ {
		go worker(jobChannel, resChannel)
	}

	go func() {
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				jobChannel <- job{i, j}
			}
		}
	}()

	for i := 0; i < N*N; i++ {
		result := <-resChannel
		C[result.x][result.y] = result.value
	}

	outputFile, err := os.OpenFile("output.txt", os.O_CREATE|os.O_WRONLY, 0600)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			write(strconv.FormatFloat(C[i][j], 'f', 1, 64)+" ", outputFile)
		}
		write("\n", outputFile)
	}
	defer outputFile.Close()
}
