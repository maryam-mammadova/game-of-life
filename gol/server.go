package gol

import "uk.ac.bris.cs/gameoflife/util"

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

var counterChannel = make(chan<- int)

// TODO: Execute all turns of the Game of Life.
//get slice
//should use goroutines

func calculateAliveCells(p Params, world [][]byte) []util.Cell {

	max := p.ImageHeight
	var cells []util.Cell

	for row := 0; row < max; row++ {
		for col := 0; col < max; col++ {
			if world[row][col] == 255 {
				c := util.Cell{col, row}
				cells = append(cells, c)
			}
		}
	}

	return cells
}

func calculateNextState(p Params, world [][]byte) [][]byte {
	max := p.ImageHeight

	world2 := make([][]byte, len(world))
	for i := range world {
		world2[i] = make([]byte, len(world[i]))
		copy(world2[i], world[i])
	}

	for row := 0; row < max; row++ {
		for col := 0; col < max; col++ {
			element := world[row][col]
			counter := 0

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
		}
	}

	return world2
}

func makeMatrix(p Params) [][]uint8 {
	slice := make([][]uint8, p.ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < p.ImageHeight; i++ {    //maybe have to pass height and width into channels
		slice[i] = make([]uint8, p.ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return slice
}
