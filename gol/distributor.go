package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"uk.ac.bris.cs/gameoflife/stubs"
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
	req2 := stubs.GoldDataInC{
		events:     c,
		ioCommand:  ioCommand,
		ioIdle:     ioIdle,
		ioFilename: ioFilename,
		ioOutput:   ioOutput,
		ioInput:    ioInput,
	}

	resp := new(stubs.GoLDataOut)
	client, err := rpc.Dial("tcp", "127.0.0.1:8040")
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	err = client.Call("GoLOperations.GoLAllTurns", req, resp)
	if err != nil {
		log.Fatal("arith error:", err)
	}

	output(p, c, turn, newFileName, world)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

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
