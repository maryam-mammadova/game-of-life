package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

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
	ticker := time.NewTicker(2 * time.Second)

	c.ioCommand <- ioInput
	newFileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- newFileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput
		}
	} //put slice into distributor channel

	req := stubs.GoLDataInP{
		Turns:       p.Turns,
		Threads:     p.Threads,
		ImageWidth:  p.ImageWidth,
		ImageHeight: p.ImageHeight,
		InputSlice:  world,
	}

	resp := new(stubs.GoLDataOut)
	client, err := rpc.Dial("tcp", ":8030")
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	//defer client.Close()
	a := client.Go(stubs.GolAllTurns, req, resp, nil)
	resp.OutputSlice = world
	sendEvent(c, p, req.InputSlice, resp.OutputSlice, turn)

	select {
	case <-ticker.C:
		err := client.Call(stubs.AliveCellsOp, req, resp)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Println("running")
		c.events <- AliveCellsCount{
			CompletedTurns: resp.CompletedTurns,
			CellsCount:     resp.AliveCells,
		}
	case <-a.Done:
		break
	}
	c.events <- FinalTurnComplete{CompletedTurns: req.Turns,
		Alive: resp.AliveCells}

	//output(p, c, turn, newFileName, res.OutputSlice)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}

func sendEvent(c distributorChannels, p Params, world [][]uint8, world2 [][]uint8, turn int) {
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 && world2[y][x] == 0 {
				c.events <- CellFlipped{turn, util.Cell{x, y}}
			} else {
				//print("x", x)
				//	print("y", y)
				//	print(len(world))
				//	print(len(world2))
				if world[y][x] == 0 && world2[y][x] == 255 {
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				}
			}
		}
	}
}

func output(p Params, c distributorChannels, turn int, newFileName string, world [][]byte) {

	c.ioCommand <- ioOutput
	FileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turn)
	c.ioFilename <- FileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col]
		}
	} //put slice into distributor channel
}

func makeMatrix(p Params) [][]uint8 {
	slice := make([][]uint8, p.ImageHeight) // initialize a slice of dy slices//will declare slice in gol
	for i := 0; i < p.ImageHeight; i++ {    //maybe have to pass height and width into channels
		slice[i] = make([]uint8, p.ImageWidth) // initialize a slice of dx unit8 in each of dy slices
	}
	return slice
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
