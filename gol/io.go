package gol

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"uk.ac.bris.cs/gameoflife/util"
)

type ioChannels struct {
	command <-chan ioCommand
	idle    chan<- bool

	filename <-chan string
	output   <-chan uint8
	input    chan<- uint8
}

// ioState is the internal ioState of the io goroutine.
type ioState struct {
	params   Params
	channels ioChannels
}

// ioCommand allows requesting behaviour from the io (pgm) goroutine.
type ioCommand uint8

// This is a way of creating enums in Go.
// It will evaluate to:
//
//	ioOutput 	= 0
//	ioInput 	= 1
//	ioCheckIdle = 2
const (
	ioOutput    ioCommand = iota
	ioInput               //TODO (MINE) assign value
	ioCheckIdle           //TODO (MINE) assign value
)

// writePgmImage receives an array of bytes and writes it to a pgm file.
func (io *ioState) writePgmImage() {
	_ = os.Mkdir("out", os.ModePerm)

	// Request a filename from the distributor.
	filename := <-io.channels.filename

	file, ioError := os.Create("out/" + filename + ".pgm")
	util.Check(ioError)
	defer file.Close()

	_, _ = file.WriteString("P5\n")
	//_, _ = file.WriteString("# PGM file writer by pnmmodules (https://github.com/owainkenwayucl/pnmmodules).\n") //TODO (mine) why is this cancelled? Does it need to be uncancelled at a later date?
	_, _ = file.WriteString(strconv.Itoa(io.params.ImageWidth)) //TODO (MINE) will I need these params in another file?
	_, _ = file.WriteString(" ")
	_, _ = file.WriteString(strconv.Itoa(io.params.ImageHeight)) //TODO (MINE) will I need these params in another file?
	_, _ = file.WriteString("\n")
	_, _ = file.WriteString(strconv.Itoa(255)) //8 bit?
	_, _ = file.WriteString("\n")

	world := make([][]byte, io.params.ImageHeight) //makes world
	for i := range world {
		world[i] = make([]byte, io.params.ImageWidth) //make each pixel and work accross?
	}

	for y := 0; y < io.params.ImageHeight; y++ {
		for x := 0; x < io.params.ImageWidth; x++ { //loop through pixels/cells
			val := <-io.channels.output //output into channel
			//TODO (MINE) Why is this block of code cancelled? Does it need to be uncancelled later?
			//if val != 0 {//would run for too long?
			//	fmt.Println(x, y)
			//}
			world[y][x] = val //displays?
		}
	}

	for y := 0; y < io.params.ImageHeight; y++ {
		for x := 0; x < io.params.ImageWidth; x++ { //loop through pixels/cells
			_, ioError = file.Write([]byte{world[y][x]}) //write to file
			util.Check(ioError)                          //check for error
		}
	}

	ioError = file.Sync()
	util.Check(ioError)

	fmt.Println("File", filename, "output done!")
}

// readPgmImage opens a pgm file and sends its data as an array of bytes.
func (io *ioState) readPgmImage() {

	// Request a filename from the distributor.
	filename := <-io.channels.filename

	data, ioError := ioutil.ReadFile("images/" + filename + ".pgm") //TODO (MINE) He said we'd need this later on. See recording
	util.Check(ioError)

	fields := strings.Fields(string(data)) //table of images?

	//validation of file conditions
	if fields[0] != "P5" {
		panic("Not a pgm file")
	}

	width, _ := strconv.Atoi(fields[1])
	if width != io.params.ImageWidth {
		panic("Incorrect width")
	}

	height, _ := strconv.Atoi(fields[2])
	if height != io.params.ImageHeight {
		panic("Incorrect height")
	}

	maxval, _ := strconv.Atoi(fields[3])
	if maxval != 255 {
		panic("Incorrect maxval/bit depth")
	}

	image := []byte(fields[4])

	for _, b := range image {
		io.channels.input <- b //put each pixel into the channel
	}

	fmt.Println("File", filename, "input done!")
}

// startIo should be the entrypoint of the io goroutine.//TODO MINE IMPORTANT NOTE
func startIo(p Params, c ioChannels) {
	io := ioState{ //state of game of life
		params:   p,
		channels: c,
	}

	for {
		select {
		// Block and wait for requests from the distributor
		case command := <-io.channels.command: //TODO MINE Use this later when we need to do the pause and quit function
			switch command {
			case ioInput:
				io.readPgmImage()
			case ioOutput:
				io.writePgmImage()
			case ioCheckIdle:
				io.channels.idle <- true
			}
		}
	}
}
