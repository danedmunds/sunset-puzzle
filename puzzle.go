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

	solved, solution := Solve(board)
	if solved {
		fmt.Println("Solution found!")
		fmt.Println("Steps:")
		for _, step := range solution {
			fmt.Println(step)
			fmt.Println()
		}
	} else {
		fmt.Println("No solution found")
	}
}
