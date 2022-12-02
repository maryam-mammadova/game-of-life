package stubs

var GolAllTurns = "GoLOperations.GolAllTurns"
var AliveCellsOp = "GolOperations.AliveCells"

type GoLDataInP struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
	InputSlice  [][]uint8
}

type GoLDataOut struct {
	OutputSlice    [][]uint8
	AliveCells     int
	CompletedTurns int
}
