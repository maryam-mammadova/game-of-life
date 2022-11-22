package gol

import (
	"fmt"
	"strconv"
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
var chanW = make(chan [][]byte) //chan of a whole world
var chanS = make(chan [][]byte) //chan of a world slice

// TODO: Execute all turns of the Game of Life.

func distributor(p Params, c distributorChannels) {

	//fmt.Println("distributor")

	// TODO: Create a 2D slice to store the world.
	slice := makeMatrix(p)
	turn := 0

	c.ioCommand <- ioInput
	newFileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- newFileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			slice[row][col] = <-c.ioInput

		}
	} //put slice into distributor channel
	//println("first for")

	// TODO: Execute all turns of the Game of Life.

	// ERRROOOOOOOOOOOOOOOOOOOORRRRRRRRRRRRRRRRRRRR

	//fmt.Println(p.Turns)
	for turn < p.Turns {
		go calculateNextState(p, slice)
		slice = <-chanW
		turn++
		//println(turn)
	}

	//slice = <-chanW
	//println("second for")

	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{CompletedTurns: turn,
		Alive: calculateAliveCells(p, c, slice)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func calculateAliveCells(p Params, c distributorChannels, world [][]byte) []util.Cell {

	//fmt.Println("calc")

	maxHeight := p.ImageHeight
	maxWidth := p.ImageWidth
	var cells []util.Cell
	//threads := p.Threads

	//ticker := time.NewTicker(2 * time.Second)
	//done := make(chan bool)

	for row := 0; row < maxWidth; row++ {
		//fmt.Println("calc issue1")
		for col := 0; col < maxHeight; col++ {
			//fmt.Println("calc issue2")
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

func calculateNextState(p Params, world [][]byte) {

	go workerWorldCreate(world)

	maxHeight := p.ImageHeight
	maxWidth := p.ImageWidth
	rowStart := 0
	colStart := 0

	if p.Threads == 1 {

		for rowStart < maxWidth {
			for colStart < maxHeight {
				go workerWorldChange(rowStart, colStart, maxHeight, world, p)
				colStart++
			}
			rowStart++
		}

	} else {

		threads := p.Threads
		maxExtra := maxHeight / threads

		var slice [][]byte

		for n := 0; n < threads; n++ {
			slice[n] = make([]byte, 10000000)

			for rowStart < maxExtra {
				for colStart < maxExtra {
					go workerWorldChange(rowStart, colStart, maxHeight, slice, p)
					colStart++
				}
				rowStart++
			}
			maxHeight = maxHeight + maxExtra
			maxWidth = maxWidth + maxWidth
		}

		for n := 0; n < threads; n++ {

		}

	}

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

func workerWorldChange(row int, col int, max int, world [][]byte, p Params) {
	element := world[row][col]
	counter := 0
	//what i need to do is get rid of chanW and instead have the ability to just change a slice and send it back
	//then in the state func i can send into chanW after assembling all the slices into a world
	//median filter vibe

	if p.Threads == 1 {

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

	} else {

		fmt.Println("multiple threads")
	}

}
