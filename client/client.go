package main

import (
	"io"
	"net"
	"os"
)

var HOST string = "localhost"
var PORT string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	if len(os.Args) != 5 {
		print("Usage: go run client.go <port> <inputfile> <outputFile> <showTime (y/n)>\n")
		os.Exit(1)
	}

	PORT = os.Args[1]
	input := os.Args[2]
	output := os.Args[3]
	socket := HOST + ":" + PORT

	conn, err := net.Dial("tcp", socket)
	check(err)
	defer conn.Close()

	fi, err := os.Open(input)
	check(err)
	defer fi.Close()
	_, err = io.Copy(conn, fi)
	check(err)
	print("File sent\n")

	fo, err := os.Create(output)
	check(err)
	defer fo.Close()
	_, err = io.Copy(fo, conn)
	check(err)
	print("File received\n")

}
