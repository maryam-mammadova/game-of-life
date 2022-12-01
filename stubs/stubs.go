package stubs

var GolAllTurns = "GoLOperations.GolAllTurns"

type GoLDataInP struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
	InputSlice  [][]uint8
}

type GoldDataInC struct {
	events:     events,
	ioCommand:  ioCommand,
	ioIdle:     ioIdle,
	ioFilename: ioFilename,
	ioOutput:   ioOutput,
	ioInput:    ioInput,
}

type GoLDataOut struct {
	OutputSlice [][]uint8
}
