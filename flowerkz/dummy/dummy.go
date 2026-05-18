package main

import (
	"context"
	"fmt"
	"net"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type incrementorRep struct {
	Number int      `json:"number"`
	Step   int      `json:"step"`
	Origin []string `json:"origin"` // [ip, port]
}

type dummyHandler struct {
	env types.Environment
}

type funcHandlerFactory struct{}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &dummyHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. fuck you")
}

func (h *dummyHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	data := incrementorRep{}

	interfc, err := net.InterfaceAddrs()

	check(err)

	fmt.Println("input: ", string(input))
	fmt.Println("InterfaceAddrs: ", interfc)
	fmt.Println("Context: ", ctx)

	//fmt.Println("Opening socket and waiting for response...")

	if err != nil {
		panic(err)
	}

	fmt.Println("new data: ", data)

	return []byte("fuck you\n"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
