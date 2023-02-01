# Go client-server parallel matrix multiplication solver

## About
Simple demonstration of client-server TCP connection used to calcul a matrix multiplication.
The main purpose is to properly use Go routines and parallel programming to execute a task efficiently.

## Execution
Server side : `go run server/server.go <port>`\
Client side : `go run client/client.go <port> <inputPath> <outputPath>`

## Rules
Input file must follow the exact same syntax as the `example.txt` file.

## Tests
The `matrixExamples` folder contains examples of matrix to test the program.
`n.txt` is the input file for 2 matrix of size **n*n**.
