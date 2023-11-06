package binaryafforder

var Libary map[FuncKey]func() any = map[FuncKey]func() any{
	"some leaf node": func() any { return 0 },
}

type FuncKey string
