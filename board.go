package main

import (
	"errors"
	"fmt"
	"strings"
)

// Orientation is a type that indicates in which direction the piece is moved.
type Orientation int

// Location is a type that represents a 2 dimensional coordinate.
type Location struct {
	x int
	y int
}

// Direction constants.
const (
	Up Orientation = iota
	Right
	Down
	Left
)

// Goal - The "win" situation of the board.
type Goal struct {
	Piece *Piece
	X, Y  int
}

// Board is a structure that represents the puzzle board.
type Board struct {
	// Width contains the width of the puzzle board.
	Width int

	// Height contains the height of the puzzle board.
	Height int

	Goal *Goal

	slots [][]*Piece

	pieces map[*Piece]Location
}

// NewBoard constructs a new empty puzzle board.
func NewBoard(width, height int) (*Board, error) {
	if width < 1 {
		return nil, errors.New("Cannot create a board with a width that is less than 1.")
	}

	if height < 1 {
		return nil, errors.New("Cannot create a board with a height that is less than 1.")
	}

	slotsValue := make([][]*Piece, width)
	for i := range slotsValue {
		slotsValue[i] = make([]*Piece, height)
	}

	return &Board{
		Width:  width,
		Height: height,
		slots:  slotsValue,
		pieces: make(map[*Piece]Location),
	}, nil
}

func NewWellKnownBoard() (board *Board, err error) {
	/*
		Start:
			A B B C
			A B B C
			D E E F
			D G H F
			  I J
		Goal:
		   X X X X
		   X X X X
		   X X X X
		   X B B X
		   X B B X
	*/
	board, _ = NewBoard(4, 5)
	for _, toAdd := range []struct {
		id, width, height, x, y int
	}{
		{1, 1, 2, 0, 0},  // A
		{2, 2, 2, 1, 0},  // B
		{3, 1, 2, 3, 0},  // C
		{4, 1, 2, 0, 2},  // D
		{5, 2, 1, 1, 2},  // E
		{6, 1, 2, 3, 2},  // F
		{7, 1, 1, 1, 3},  // G
		{8, 1, 1, 2, 3},  // H
		{9, 1, 1, 1, 4},  // I
		{10, 1, 1, 2, 4}, // J
	} {
		piece, _ := NewPiece(toAdd.id, toAdd.width, toAdd.height)
		err = board.AddPiece(piece, toAdd.x, toAdd.y)
		if err != nil {
			return
		}
	}

	sun, err := board.GetPieceAt(1, 0)
	if err != nil || sun == nil {
		panic("couldn't set goal")
	}
	board.SetGoal(sun, 1, 3)

	return
}

func (p *Board) String() string {
	var sb strings.Builder
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			val := 0
			piece, _ := p.GetPieceAt(x, y)
			if piece != nil {
				val = piece.ID
			}
			sb.WriteString(fmt.Sprintf("%02d  ", val))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// AddPiece adds a Piece instance at the specified location.  If the piece
// would overlap with an existing piece, it is not placed and an error is
// returned.
func (p *Board) AddPiece(piece *Piece, x, y int) error {
	// Validate location inputs
	if x < 0 || x > p.Width-piece.Width || y < 0 || y > p.Height-piece.Height {
		return fmt.Errorf("Piece %d cannot be added at %d, %d because it would not fit within the board.", piece.ID, x, y)
	}

	for i := 0; i < piece.Width; i++ {
		for j := 0; j < piece.Height; j++ {
			slotPiece := p.slots[x+i][y+j]
			if slotPiece != nil {
				return fmt.Errorf("Piece %d overlaps piece %d at %d, %d\n", piece.ID, slotPiece, x+i, y+j)
			}
		}
	}

	p.pieces[piece] = Location{
		x: x,
		y: y,
	}

	for i := 0; i < piece.Width; i++ {
		for j := 0; j < piece.Height; j++ {
			p.slots[x+i][y+j] = piece
		}
	}

	return nil
}

// SetGoal sets the goal layout of the puzzle, which piece needs to be in which position for the puzzle to be solved.
func (p *Board) SetGoal(piece *Piece, x, y int) error {
	if piece == nil {
		return errors.New("goal cannot have nil piece")
	}

	if _, ok := p.pieces[piece]; !ok {
		return fmt.Errorf("invalid piece for goal, piece %d is not present in the puzzle", piece.ID)
	}

	if p.PieceFitsOnBoardAtPosition(piece, x, y) {
		return fmt.Errorf("Invalid position for goal (%d, %d), x must be positive and smaller than %d, y must be positive and smaller than %d", x, y, p.Width, p.Height)
	}

	p.Goal = &Goal{
		X: x, Y: y, Piece: piece,
	}

	return nil
}

// MovePiece moves the specified piece by 1 square in the given direction.
func (p *Board) MovePiece(piece *Piece, orientation Orientation) error {
	// Check if the piece can be moved.
	l := p.pieces[piece]
	x := l.x
	y := l.y

	switch orientation {
	case Up:
		y--
		break
	case Right:
		x++
		break
	case Down:
		y++
		break
	case Left:
		x--
		break
	}

	if p.PieceFitsOnBoardAtPosition(piece, x, y) {
		return fmt.Errorf("Moving piece %d (%d, %d - %d, %d) would put it outside the bounds of the puzzle board (0, 0 - %d, %d)", piece.ID, x, y, x+piece.Width, y+piece.Height, p.Width, p.Height)
	}

	for i := 0; i < piece.Width; i++ {
		for j := 0; j < piece.Height; j++ {
			slotPiece := p.slots[x+i][y+j]
			if slotPiece != nil && slotPiece != piece {
				return fmt.Errorf("Moving piece %d would cause it to overlap piece %d", piece.ID, slotPiece.ID)
			}
		}
	}

	p.RemovePiece(piece)
	p.AddPiece(piece, x, y)

	return nil
}

func (p *Board) Undo(piece *Piece, orientation Orientation) error {
	var undo Orientation
	switch orientation {
	case Up:
		undo = Down
	case Down:
		undo = Up
	case Left:
		undo = Right
	case Right:
		undo = Left
	default:
		return fmt.Errorf("Invalid direction")
	}

	return p.MovePiece(piece, undo)
}

// PieceFitsOnBoardAtPosition checks whether a piece would fit on the board at the position
func (p *Board) PieceFitsOnBoardAtPosition(piece *Piece, x, y int) bool {
	return x < 0 || x+piece.Width > p.Width || y < 0 || y+piece.Height > p.Height
}

// RemovePiece removes a Piece instance from the puzzle board.
func (p *Board) RemovePiece(piece *Piece) {
	if l, ok := p.pieces[piece]; ok {
		for x := 0; x < piece.Width; x++ {
			for y := 0; y < piece.Height; y++ {
				p.slots[x+l.x][y+l.y] = nil
			}
		}

		delete(p.pieces, piece)
	}
}

// GetPieceAt returns the Piece instance that occupies the specified location.
func (p *Board) GetPieceAt(x, y int) (*Piece, error) {
	if x < 0 || x >= p.Width {
		return nil, fmt.Errorf("The specified x coordinate %d is invalid.", x)
	}

	if y < 0 || y >= p.Height {
		return nil, fmt.Errorf("The specified y coordinate %d is invalid.", y)
	}

	return p.slots[x][y], nil
}

// IsSolved verifies whether the puzzle is in the solved state base on the goal set by SetGoal
func (p *Board) IsSolved() bool {
	// a puzzle with no goal is always solved! ;)
	if p.Goal == nil {
		return true
	}

	pieceInPos, err := p.GetPieceAt(p.Goal.X, p.Goal.Y)
	if err != nil {
		panic(fmt.Sprintf("Got an error while checking if solved: %v", err))
	}

	if pieceInPos == nil {
		// no piece occupying the goal
		return false
	}

	location, ok := p.pieces[pieceInPos]
	if !ok {
		panic(fmt.Sprintf("Got an error while checking if solved: %v", err))
	}

	if pieceInPos.ID == p.Goal.Piece.ID && location.x == p.Goal.X && location.y == p.Goal.Y {
		return true
	}

	return false
}

func Solve(board *Board) (bool, []string) {
	seen := make(map[string]bool)
	path := []string{board.String()}

	return innerSolve(board, seen, path)
}

func innerSolve(board *Board, seen map[string]bool, path []string) (bool, []string) {
	state := board.String()

	if seen[state] {
		return false, nil
	}
	seen[state] = true

	if board.IsSolved() {
		return true, path
	}

	for piece, _ := range board.pieces {
		for _, orientation := range []Orientation{Up, Down, Left, Right} {
			err := board.MovePiece(piece, orientation)
			if err != nil {
				continue
			}

			path = append(path, board.String())
			solved, solutionPath := innerSolve(board, seen, path)
			if solved {
				return solved, solutionPath
			}

			path = path[:len(path)-1]
			err = board.Undo(piece, orientation)
			if err != nil {
				panic("could not undo a move that was made")
			}
		}

	}

	return false, nil
}
