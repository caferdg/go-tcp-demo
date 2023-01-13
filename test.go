package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var N int
var A [][]float64
var B [][]float64
var C [][]float64

var jobChannel chan job
var resChannel chan res

type job struct {
	line   []float64
	column []float64
	x      int
	y      int
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
	if _, err := file.WriteString(text); err != nil {
		panic(err)
	}
}

func read(filename string) string {
	data, err := ioutil.ReadFile(filename)
	check(err)
	return string(data)
}

func calcCoef(line []float64, column []float64) float64 {
	var result float64
	for k := 0; k < N; k++ {
		result += line[k] * column[k]
	}
	return result
}

func worker(jobCh chan job, resCh chan res) {
	var job job
	job = <-jobCh
	var result res
	result.x = job.x
	result.y = job.y
	result.value = calcCoef(job.line, job.column)
	resCh <- result
}

func main() {
	file, err := os.OpenFile("input.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	defer file.Close()
	check(err)

	data := read(file.Name())
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

	for i := 0; i < 10; i++ {
		go worker(jobChannel, resChannel)
	}

	go func() {
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				jobChannel <- job{A[i], B[j], i, j}
			}
		}
	}()

	for i := 0; i < N*N; i++ {
		result := <-resChannel
		fmt.Print("Result received :", result.x, " ", result.y, " ", result.value, "\n")
		C[result.x][result.y] = result.value
	}

	fmt.Print(C)
}
