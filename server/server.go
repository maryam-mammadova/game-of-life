package server

import (
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type GoLOperations struct{}

func (s *GoLOperations) GolAllTurns(req stubs.GoLDataInP) {

	turn := req.Turns
	world := req.InputSlice

	// Ticker goroutine
	ticker := time.NewTicker(2 * time.Second)

	chanW := make([]chan [][]uint8, req.Threads)
	for i := range chanW {
		chanW[i] = make(chan [][]uint8)
	}

	// TODO: Execute all turns of the Game of Life.
	if req.Threads == 1 {
		for turn < req.Turns {
			world = calculateNextState(req, world, c, turn, 0, req.ImageHeight)
			turn++
		}
	} else {
		maxHeight := req.ImageHeight
		threads := req.Threads

		maxExtra := maxHeight % threads //threads can be an odd number

		for turn < req.Turns {

			select {
			case <-ticker.C:
				numAlive := len(calculateAliveCells(req, world))
				c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: numAlive}
				world = executeTurns(req, c, maxHeight, maxExtra, threads, turn, chanW, world)
			default:
				world = executeTurns(req, c, maxHeight, maxExtra, threads, turn, chanW, world)
			}
			turn++
		}

	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{CompletedTurns: turn,
		Alive: calculateAliveCells(req, world)}

}

func executeTurns(req stubs.GoLDataInP, c distributorChannels, maxHeight int, maxExtra int, threads int, turn int, chanW []chan [][]uint8, world [][]byte) [][]byte {

	for n := 0; n < threads; n++ {
		startY := n * (req.ImageHeight / req.Threads)
		maxHeight = (n + 1) * (req.ImageHeight / req.Threads)
		if n == threads-1 {
			maxHeight = maxHeight + maxExtra
		}
		go worker(req, world, c, turn, startY, maxHeight, chanW[n])

	}
	newPixelData := makeMatrixS(0, 0)

	for n := 0; n < threads; n++ {
		newPixelData = append(newPixelData, <-chanW[n]...)
	}
	world = newPixelData

	return world
}

func calculateAliveCells(req stubs.GoLDataInP, world [][]byte) []util.Cell {

	max := req.ImageHeight
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

func calculateNextState(req stubs.GoLDataInP, world [][]uint8, c distributorChannels, turn int, startY int, endY int) [][]byte {

	world2 := make([][]byte, endY-startY)
	for i := range world2 {
		world2[i] = make([]byte, req.ImageWidth)
	}
	rowT := 0
	for row := startY; row < endY; row++ {
		for col := 0; col < req.ImageWidth; col++ {

			counter := 0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {

					nRow := (row + dx + req.ImageHeight) % req.ImageHeight
					nCol := (col + dy + req.ImageWidth) % req.ImageWidth

					if world[nRow][nCol] == 255 {
						counter++

					}
				}
			}

			if world[row][col] == 255 {
				counter--
			}
			//any live cell with fewer than two live neighbours dies
			//any live cell with two or three live neighbours is unaffected
			//any live cell with more than three live neighbours dies
			//any dead cell with exactly three live neighbours becomes alive

			if world[row][col] == 0 && counter == 3 {
				world2[rowT][col] = 255
				c.events <- CellFlipped{
					CompletedTurns: turn,
					Cell:           util.Cell{col, row},
				}
			} else if world[row][col] == 255 {
				if counter < 2 {
					world2[rowT][col] = 0
					c.events <- CellFlipped{
						CompletedTurns: turn,
						Cell:           util.Cell{col, row}}
				} else if counter > 3 {
					world2[rowT][col] = 0
					c.events <- CellFlipped{
						CompletedTurns: turn,
						Cell:           util.Cell{col, row}}
				} else {
					world2[rowT][col] = 255
				}
			}
		}
		rowT++
	}

	return world2
}

func worker(req stubs.GoLDataInP, world [][]uint8, c distributorChannels, turn int, startY int, endY int, chanW chan [][]uint8) {

	world2 := calculateNextState(req, world, c, turn, startY, endY)

	chanW <- world2
}

func makeMatrixS(ImageHeight, ImageWidth int) [][]uint8 {
	slice := make([][]uint8, ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < ImageHeight; i++ {    //maybe have to pass height and width into channels
		slice[i] = make([]uint8, ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return slice
}

func main() {
	rpc.Register(&GoLOperations{})
	listener, _ := net.Listen("tcp", ":"+"8030")
	defer listener.Close()
	rpc.Accept(listener)
}
