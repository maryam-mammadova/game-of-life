package stubs

var GolAllTurns = "GoLOperations.GolAllTurns"

type GoLDataInP struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
	InputSlice  [][]uint8
}

//type GoldDataInC struct {
//	events:     chan,
//	ioCommand:  chan ioCommand
//	ioIdle:     ioIdle,
//	ioFilename: ioFilename,
//	ioOutput:   ioOutput,
//	ioInput:    ioInput,
//huhu
//}

type GoLDataOut struct {
	OutputSlice [][]uint8
}
