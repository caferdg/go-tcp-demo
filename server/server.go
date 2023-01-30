package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var HOST string = "localhost"
var stopWorker user = user{nil, -1, -1, nil, nil} // Special user to stop workers

const chanSize int = 100
const nbWorkersPerUser int = 10

type job struct {
	x       int
	y       int
	ligne   *[]float64
	colonne *[]float64
}

type res struct {
	value float64
	x     int
	y     int
}

type user struct {
	connection net.Conn
	id         int
	sizeMat    int
	matA       [][]float64
	matB       [][]float64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func initMat(N int) (matA [][]float64, matB [][]float64, matC [][]float64) {
	A := make([][]float64, N)
	B := make([][]float64, N)
	C := make([][]float64, N)
	for i := 0; i < N; i++ {
		A[i] = make([]float64, N)
		B[i] = make([]float64, N)
		C[i] = make([]float64, N)
	}
	return A, B, C
}

func inputTextToMat(text string) (matA [][]float64, matB [][]float64, matC [][]float64) {
	mat := strings.Split(text, "-")
	mattA := strings.Split(mat[0], "\n")
	mattB := strings.Split(mat[1], "\n")[1:]
	size := len(mattB)
	A, B, C := initMat(size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			A[i][j], _ = strconv.ParseFloat(strings.Split(mattA[i], " ")[j], 3)
			B[i][j], _ = strconv.ParseFloat(strings.Split(mattB[i], " ")[j], 3)
		}
	}
	return A, B, C
}

func matToString(mat [][]float64) string {
	size := len(mat)
	res := ""
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			res += strconv.FormatFloat(mat[i][j], 'f', 1, 64) + " "
		}
		res += "\n"
	}
	return res
}

func calcCoef(line *[]float64, column *[]float64) float64 {
	var result float64
	size := len(*line)
	for k := 0; k < size; k++ {
		result += (*line)[k] * (*column)[k]
	}
	return result
}

func worker(jobCh chan job, resCh chan res) {
	for {
		job := <-jobCh
		if job.x == -1 && job.y == -1 {
			break
		}
		var result res
		result.x = job.x
		result.y = job.y
		result.value = calcCoef(job.ligne, job.colonne)
		resCh <- result
	}
}

func handleUser(newUser user) {
	jobChannel := make(chan job, chanSize)
	resChannel := make(chan res, chanSize)
	for i := 0; i < nbWorkersPerUser; i++ {
		go worker(jobChannel, resChannel)
	}

	defer newUser.connection.Close()
	fmt.Println("New connection, id :", newUser.id)

	reader := bufio.NewReader(newUser.connection)
	data, err := reader.ReadString('$')
	check(err)
	data = strings.TrimSuffix(data, "$")
	var C [][]float64

	start := time.Now()

	newUser.matA, newUser.matB, C = inputTextToMat(data)
	newUser.sizeMat = len(newUser.matA)
	go func() {
		for i := 0; i < newUser.sizeMat; i++ {
			for j := 0; j < newUser.sizeMat; j++ {
				line := newUser.matA[i]
				column := newUser.matB[:][j]
				jobChannel <- job{i, j, &line, &column}
			}
		}
	}()

	for i := 0; i < newUser.sizeMat*newUser.sizeMat; i++ {
		result := <-resChannel
		C[result.x][result.y] = result.value
	}

	resMessage := matToString(C)

	elapsed := time.Since(start)
	println("------- Time elapsed : ", elapsed.String())

	io.WriteString(newUser.connection, resMessage+"$")
	newUser.connection.Close()
	fmt.Println("Connection", newUser.id, "closed")
	// Killing workers
	for i := 0; i < nbWorkersPerUser; i++ {
		jobChannel <- job{-1, -1, nil, nil}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage : go run server.go <port>")
		os.Exit(1)
	}

	PORT := os.Args[1]
	socket := HOST + ":" + PORT
	listen, err := net.Listen("tcp", socket)
	check(err)

	fmt.Println("Listening to socket : ", socket)

	nbUsers := 0

	for {
		conn, err := listen.Accept()
		defer conn.Close()
		check(err)

		var newUser user
		newUser.id = nbUsers
		newUser.connection = conn

		go handleUser(newUser)
		nbUsers++
	}

}
