package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var HOST string = "localhost"
var stopWorker user = user{nil, -1, -1, nil, nil} // Special user to stop workers

const chanSize int = 100
const nbWorkersPerUser int = 1

type job struct {
	x      int
	y      int
	client user
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

func calcCoef(line int, column int, client user) float64 {
	var result float64
	size := len(client.matA)
	for k := 0; k < size; k++ {
		result += client.matA[line][k] * client.matB[k][column]
	}
	return result
}

func worker(jobCh chan job, resCh chan res) {
	for {
		job := <-jobCh
		if job.client.id == -1 {
			break
		}
		var result res
		result.x = job.x
		result.y = job.y
		result.value = calcCoef(job.x, job.y, job.client)
		resCh <- result
	}
}

func handleUser(user user) {
	jobChannel := make(chan job, chanSize)
	resChannel := make(chan res, chanSize)
	for i := 0; i < nbWorkersPerUser; i++ {
		go worker(jobChannel, resChannel)
	}

	defer user.connection.Close()
	fmt.Println("New connection, id :", user.id)

	reader := bufio.NewReader(user.connection)
	data, err := reader.ReadString('$')
	check(err)
	data = strings.TrimSuffix(data, "$")
	var C [][]float64

	user.matA, user.matB, C = inputTextToMat(data)

	go func() {
		for i := 0; i < user.sizeMat; i++ {
			for j := 0; j < user.sizeMat; j++ {
				jobChannel <- job{i, j, user}
			}
		}
	}()

	for i := 0; i < user.sizeMat*user.sizeMat; i++ {
		result := <-resChannel
		C[result.x][result.y] = result.value
	}

	resMessage := matToString(C)

	io.WriteString(user.connection, resMessage+"$")
	user.connection.Close()
	fmt.Println("Connection", user.id, "closed")
	// Killing workers
	for i := 0; i < nbWorkersPerUser; i++ {
		jobChannel <- job{0, 0, stopWorker}
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
