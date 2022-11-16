package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"uk.ac.bris.cs/gameoflife/golop"
)

func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	slice := makeMatrix(p)
	turn := 0

	c.ioCommand <- ioInput
	newFileName := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- newFileName
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			slice[row][col] = <-c.ioInput
			//fmt.Printf("%3d", slice[row][col])

		}
		//fmt.Println()
	} //put slice into distributor channel

	// TODO: Execute all turns of the Game of Life.
	for turn < p.Turns {
		slice = calculateNextState(p, slice)
		turn++
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{CompletedTurns: turn,
		Alive: calculateAliveCells(p, slice)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func makeCall(client *rpc.Client, message string) {
	request := golop.Request{Message: message}
	response := new(golop.Response)

	client.Call(golop.ReverseHandler, request, response)
	fmt.Println("Responded: " + response.Message)
}

func server() {
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	//file, _ := os.Open("wordlist")
	//scanner := bufio.NewScanner(file)
	//for scanner.Scan() {
	//	t := scanner.Text()
	//	fmt.Println("Called: " + t)
	//	makeCall(client, t)
	//}
}
