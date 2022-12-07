package gol

import (
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

// import
type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPress   <-chan rune
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	world := makeMatrix(p)

	turn := 0

	c.ioCommand <- ioInput
	newFileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) //declare file name
	c.ioFilename <- newFileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput //take in input

		}

	} //put slice into distributor channel

	// Ticker goroutine
	ticker := time.NewTicker(2 * time.Second)

	chanW := make([]chan [][]uint8, p.Threads) //make global channel of channels
	for i := range chanW {
		chanW[i] = make(chan [][]uint8)
	}

	// TODO: Execute all turns of the Game of Life.
	if p.Threads == 1 { //single threaded implementation
		for turn < p.Turns {
			world = calculateNextState(p, world, c, turn, 0, p.ImageHeight)
			turn++
		}
	} else {
		maxHeight := p.ImageHeight
		threads := p.Threads

		maxExtra := maxHeight % threads //bit leftover

		for turn < p.Turns {

			select {
			case <-ticker.C:
				numAlive := len(calculateAliveCells(p, world))
				c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: numAlive} //report alive cell count

				world = executeTurns(p, c, maxHeight, maxExtra, threads, turn, chanW, world)
			default:
				world = executeTurns(p, c, maxHeight, maxExtra, threads, turn, chanW, world)
			}
			turn++
		}

	}

	output(p, c, turn, world)

	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{CompletedTurns: turn,
		Alive: calculateAliveCells(p, world)}
	//done <- true

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}

//func keyPressFunctionality(p Params, c distributorChannels, slice [][]uint8, turn int, newFileName string) {
//	keyPress := string(<-c.keyPress)
//	switch keyPress {
//	case string('s'):
//		output(p, slice, c, turn, newFileName)
//	case string('q'):
//		output(p, slice, c, turn, newFileName)
//		sdl.Quit()
//		os.Exit(1)
//
//	case string('p'):
//		for {
//			time.Sleep(time.Second * 1)
//			switch keyPress {
//			case string('p'):
//				fmt.Println("Continuing")
//			}
//		}
//	default:
//	}
//}

func output(p Params, c distributorChannels, turn int, world [][]byte) {

	c.ioCommand <- ioOutput
	FileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turn)
	c.ioFilename <- FileName //send filename down channel
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col] //output each pixel
		}
	} //put slice into distributor channel
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {

	max := p.ImageHeight
	var cells []util.Cell

	for row := 0; row < max; row++ {
		for col := 0; col < max; col++ {

			if world[row][col] == 255 { //if cell is alive
				c := util.Cell{col, row}
				cells = append(cells, c) //append cells together
			}
		}
	}

	return cells
}

func calculateNextState(p Params, world [][]uint8, c distributorChannels, turn int, startY int, endY int) [][]byte {

	world2 := make([][]byte, endY-startY)
	for i := range world2 {
		world2[i] = make([]byte, p.ImageWidth) //make copy of slice

	}
	rowT := 0
	for row := startY; row < endY; row++ {
		for col := 0; col < p.ImageWidth; col++ { //for each cell

			aliveNeighbours := 0

			for dy := -1; dy <= 1; dy++ { //go around neighbouring cells
				for dx := -1; dx <= 1; dx++ {

					nRow := (row + dx + p.ImageHeight) % p.ImageHeight //is it alive calculations
					nCol := (col + dy + p.ImageWidth) % p.ImageWidth

					if world[nRow][nCol] == 255 { //if the cell is alive
						aliveNeighbours++ //increment counter

					}
				}
			}

			if world[row][col] == 255 { //ensure we do not count the cell itself
				aliveNeighbours--
			}
			//any live cell with fewer than two live neighbours dies
			//any live cell with two or three live neighbours is unaffected
			//any live cell with more than three live neighbours dies
			//any dead cell with exactly three live neighbours becomes alive

			if world[row][col] == 0 && aliveNeighbours == 3 {
				world2[rowT][col] = 255
				c.events <- CellFlipped{ //report cells flipped
					CompletedTurns: turn,
					Cell:           util.Cell{col, row},
				}
			} else if world[row][col] == 255 { //report cells flipped
				if aliveNeighbours < 2 {
					world2[rowT][col] = 0
					c.events <- CellFlipped{
						CompletedTurns: turn,
						Cell:           util.Cell{col, row}}
				} else if aliveNeighbours > 3 {
					world2[rowT][col] = 0
					c.events <- CellFlipped{ //report cells flipped
						CompletedTurns: turn,
						Cell:           util.Cell{col, row}}
				} else {
					world2[rowT][col] = 255 //unaffected cell
				}
			}
		}
		rowT++
	}

	return world2 //return new slice
}

func executeTurns(p Params, c distributorChannels, maxHeight int, maxExtra int, threads int, turn int, chanW []chan [][]uint8, world [][]byte) [][]byte {

	for n := 0; n < threads; n++ { //for each thread
		startY := n * (p.ImageHeight / p.Threads)         //calculate the starting y
		maxHeight = (n + 1) * (p.ImageHeight / p.Threads) //calculate the ending y
		if n == threads-1 {                               //if it's the final piece
			maxHeight = maxHeight + maxExtra //add the missing bit that may be left off
		}
		go worker(p, world, c, turn, startY, maxHeight, chanW[n]) //pass to the worker function

	}
	newPixelData := makeMatrixS(0, 0) //make a new world

	for n := 0; n < threads; n++ {
		newPixelData = append(newPixelData, <-chanW[n]...) //append the world to be the process slices appended together
	}
	world = newPixelData

	return world //return result of appending
}

func makeMatrix(p Params) [][]uint8 {
	matrix := make([][]uint8, p.ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < p.ImageHeight; i++ {     //maybe have to pass height and width into channels
		matrix[i] = make([]uint8, p.ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return matrix
}

func makeMatrixS(ImageHeight, ImageWidth int) [][]uint8 {
	matrix := make([][]uint8, ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < ImageHeight; i++ {     //maybe have to pass height and width into channels
		matrix[i] = make([]uint8, ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return matrix
}

func worker(p Params, world [][]uint8, c distributorChannels, turn int, startY int, endY int, chanW chan [][]uint8) {

	world2 := calculateNextState(p, world, c, turn, startY, endY)

	chanW <- world2 //pass results into channel
}
