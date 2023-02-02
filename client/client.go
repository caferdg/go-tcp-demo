package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

var HOST string = "localhost"
var PORT string

func write(text string, file *os.File) {
	_, err := file.WriteString(text)
	check(err)
}

func read(filename string) string {
	data, err := ioutil.ReadFile(filename)
	check(err)
	return string(data)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	if len(os.Args) != 4 {
		println("Usage: go run client.go <port> <inputPath> <outputPath>")
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	PORT = os.Args[1]
	inputPath := os.Args[2]
	outputPath := os.Args[3]
	socket := HOST + ":" + PORT

	if _, err := os.Stat(inputPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Input file not found in " + cwd + "/" + inputPath)
		os.Exit(1)
	}
	inputFile, err := os.OpenFile(inputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	defer inputFile.Close()
	check(err)

	conn, err := net.Dial("tcp", socket)
	defer conn.Close()
	check(err)
	data := read(inputFile.Name()) + "$"
	io.WriteString(conn, data)
	inputFile.Close()

	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('$')
	check(err)
	message = strings.TrimSuffix(message, "$")

	outputFile, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer outputFile.Close()
	check(err)
	write(message, outputFile)
	outputFile.Close()

	conn.Close()
	fmt.Println("File saved to " + cwd + "/" + outputPath)
}
