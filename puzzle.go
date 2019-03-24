package main

import "fmt"

type Test struct {
	W     int
	H     int
	slots [][]int
}

func main() {
	board, err := NewWellKnownBoard()
	if err != nil {
		panic(err)
	}

	fmt.Println(board)
}
