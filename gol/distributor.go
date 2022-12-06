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
	//done := make(chan bool)
	// TODO: Create a 2D slice to store the world.
	world := makeMatrix(p)
	//	fmt.Println("InitLength-", len(world))

	turn := 0
	//mut := sync.Mutex{}

	c.ioCommand <- ioInput
	newFileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- newFileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput
			//fmt.Printf("%3d", world[row][col])

		}
		//fmt.Println()
	} //put slice into distributor channel
	//go reportAliveCellCount(p, world, c, &mut, &turn)

	// Ticker goroutine
	ticker := time.NewTicker(2 * time.Second)

	chanW := make([]chan [][]uint8, p.Threads)
	for i := range chanW {
		chanW[i] = make(chan [][]uint8)
	}

	// TODO: Execute all turns of the Game of Life.
	if p.Threads == 1 {
		for turn < p.Turns {
			world = calculateNextState(p, world, c, turn, 0, p.ImageHeight)
			turn++
		}
	} else {
		maxHeight := p.ImageHeight
		threads := p.Threads
		//startY := 0
		//fmt.Println("THREADS-", threads)

		maxExtra := maxHeight % threads //threads can be an odd number

		for turn < p.Turns {

			select {
			case <-ticker.C:
				numAlive := len(calculateAliveCells(p, world))
				c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: numAlive}

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
	c.ioFilename <- FileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col]
		}
	} //put slice into distributor channel
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {

	max := p.ImageHeight
	var cells []util.Cell

	for row := 0; row < max; row++ {
		for col := 0; col < max; col++ {
			//fmt.Println("row", row)
			//fmt.Println("column", col)
			if world[row][col] == 255 {
				c := util.Cell{col, row}
				cells = append(cells, c)
			}
		}
	}

	return cells
}

func calculateNextState(p Params, world [][]uint8, c distributorChannels, turn int, startY int, endY int) [][]byte {

	world2 := make([][]byte, endY-startY)
	//fmt.Println("len:", len(world2))
	//fmt.Println("CNS", endY, "-", startY, "=", endY-startY, "length ", len(world2))
	for i := range world2 {
		world2[i] = make([]byte, p.ImageWidth)
		//copy(world2[i], world[i])
	}
	rowT := 0
	for row := startY; row < endY; row++ {
		for col := 0; col < p.ImageWidth; col++ {

			counter := 0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {

					nRow := (row + dx + p.ImageHeight) % p.ImageHeight
					nCol := (col + dy + p.ImageWidth) % p.ImageWidth

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

func executeTurns(p Params, c distributorChannels, maxHeight int, maxExtra int, threads int, turn int, chanW []chan [][]uint8, world [][]byte) [][]byte {

	for n := 0; n < threads; n++ {
		startY := n * (p.ImageHeight / p.Threads)
		maxHeight = (n + 1) * (p.ImageHeight / p.Threads)
		if n == threads-1 {
			maxHeight = maxHeight + maxExtra
		}
		go worker(p, world, c, turn, startY, maxHeight, chanW[n])

	}
	newPixelData := makeMatrixS(0, 0)

	for n := 0; n < threads; n++ {
		newPixelData = append(newPixelData, <-chanW[n]...)
	}
	world = newPixelData

	return world
}

func makeMatrix(p Params) [][]uint8 {
	slice := make([][]uint8, p.ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < p.ImageHeight; i++ {    //maybe have to pass height and width into channels
		slice[i] = make([]uint8, p.ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return slice
}

func makeMatrixS(ImageHeight, ImageWidth int) [][]uint8 {
	slice := make([][]uint8, ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < ImageHeight; i++ {    //maybe have to pass height and width into channels
		slice[i] = make([]uint8, ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return slice
}

func worker(p Params, world [][]uint8, c distributorChannels, turn int, startY int, endY int, chanW chan [][]uint8) {

	world2 := calculateNextState(p, world, c, turn, startY, endY)

	//	fmt.Println("WORKER ", startY, "-", endY, " = ", len(world2))
	chanW <- world2
}
