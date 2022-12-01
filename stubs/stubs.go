package stubs

var GolAllTurns = "GoLOperations.GolAllTurns"

type Response struct {
	Message string
}

type Request struct {
	Message string
}

type GoLDataIn struct {
	Turns       int
	ImageWidth  int
	ImageHeight int
	InputSlice  [][]uint8
}

type GoLDataOut struct {
	OutputSlice [][]uint8
}
