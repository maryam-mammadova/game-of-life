package gol

import (
	"fmt"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

var counterChannel = make(chan<- int)
var chanW = make(chan [][]byte, 10000000)

// TODO: Execute all turns of the Game of Life.
//get slice
//should use goroutines

func calculateAliveCells(p Params, c distributorChannels, world [][]byte) []util.Cell {

	fmt.Println("calc")

	maxHeight := p.ImageHeight
	maxWidth := p.ImageWidth
	var cells []util.Cell

	//ticker := time.NewTicker(2 * time.Second)
	//done := make(chan bool)

	for row := 0; row < maxWidth; row++ {
		for col := 0; col < maxHeight; col++ {
			if world[row][col] == 255 {
				cell := util.Cell{col, row}
				cells = append(cells, cell)
				//c.events <- AliveCellsCount{CompletedTurns: 0, CellsCount: 0}
				//fmt.Println(cells)

			}
		}
	}

	//for _ = range ticker.C {
	//	c.events <- AliveCellsCount{CompletedTurns: 0, CellsCount: 0}
	//}

	return cells
}

func calculateNextState(p Params, world [][]byte) [][]byte {
	max := p.ImageHeight

	go workerWorldCreate(world)

	for row := 0; row < max; row++ {
		for col := 0; col < max; col++ {

			go workerWorldChange(row, col, max, world)

		}
	}

	world2 := <-chanW

	return world2
}

func makeMatrix(p Params) [][]uint8 {
	slice := make([][]uint8, p.ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < p.ImageHeight; i++ {    //maybe have to pass height and width into channels
		slice[i] = make([]uint8, p.ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return slice
}

func workerWorldCreate(world [][]byte) {

	world2 := make([][]byte, len(world))
	for i := range world {
		world2[i] = make([]byte, len(world[i]))
		copy(world2[i], world[i])
	}
	chanW <- world2
}

func workerWorldChange(row int, col int, max int, world [][]byte) {
	element := world[row][col]
	counter := 0

	world2 := <-chanW

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nRow := (row + dx + max) % max
			nCol := (col + dy + max) % max

			if world[nRow][nCol] == 255 {
				counter++
			}
		}
	}

	if element == 255 {
		counter--
	}
	if element == 0 {
		if counter == 3 {
			world2[row][col] = 255
		}
	} else {
		if counter < 2 {
			world2[row][col] = 0
		} else if counter > 3 {
			world2[row][col] = 0
		}
	}

	chanW <- world2
}

//func rowsCol(p Params func func, ) {
//
//	maxHeight := p.ImageHeight
//	maxWidth := p.ImageWidth
//
//	for row := 0; row < maxWidth; row++ {
//		for col := 0; col < maxHeight; col++ {
//
//			world2 = creation(row, col, max, world, world2)
//
//		}
//	}
//
//}

//func worker(startY, endY, startX, endX int, data func(y, x int) uint8, out chan<- [][]uint8) {
//
//	slice := medianFilter(startY, endY, startX, endX, data)
//	out <- slice
//}
