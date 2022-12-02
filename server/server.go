package main

import (
	"errors"
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type GoLOperations struct{}

func (s *GoLOperations) GolAllTurns(req stubs.GoLDataInP, res *stubs.GoLDataOut) (err error) {

	if req.InputSlice == nil {
		err = errors.New("Invalid input")
		return
	}

	turn := 0
	world := req.InputSlice
	res.OutputSlice = req.InputSlice

	// ticker should be in distr
	//ticker := time.NewTicker(2 * time.Second)

	for turn < req.Turns {
		res.OutputSlice = calculateNextState(req, world)
		turn++
	}

	return
}

func (s *GoLOperations) AliveCells(req stubs.GoLDataInP, res *stubs.GoLDataOut) (err error) {
	res.AliveCells = calculateAliveCells(req, res)
	return
}

func calculateAliveCells(req stubs.GoLDataInP, res *stubs.GoLDataOut) []util.Cell {

	max := req.ImageHeight
	var cells []util.Cell
	world := res.OutputSlice

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

func calculateNextState(req stubs.GoLDataInP, world [][]byte) [][]byte {
	max := req.ImageHeight

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

func main() {
	pAddr := flag.String("port", ":8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GoLOperations{})

	listener, err := net.Listen("tcp", *pAddr)

	if err != nil {
		panic(err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			panic(err)
		}
	}(listener)
	rpc.Accept(listener)

}

//func executeTurns(req stubs.GoLDataInP, c distributorChannels, maxHeight int, maxExtra int, threads int, turn int, chanW []chan [][]uint8, world [][]byte) [][]byte {
//
//	for n := 0; n < threads; n++ {
//		startY := n * (req.ImageHeight / req.Threads)
//		maxHeight = (n + 1) * (req.ImageHeight / req.Threads)
//		if n == threads-1 {
//			maxHeight = maxHeight + maxExtra
//		}
//		go worker(req, world, c, turn, startY, maxHeight, chanW[n])
//
//	}
//	newPixelData := makeMatrixS(0, 0)
//
//	for n := 0; n < threads; n++ {
//		newPixelData = append(newPixelData, <-chanW[n]...)
//	}
//	world = newPixelData
//
//	return world
//}
