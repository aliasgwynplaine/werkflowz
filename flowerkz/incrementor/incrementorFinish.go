package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type incrementorRep struct {
	Number int    `json:"number"`
	Step   int    `json:"step"`
	Origin string `json:"origin"`
}

type incrementorHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &incrementorHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *incrementorHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("IncrementorFinish...")
	//decode json
	data := incrementorRep{}

	json.Unmarshal(input, &data)

	fmt.Println("context: ", ctx)
	fmt.Println("data: ", data)

	conn, err := net.Dial("tcp", data.Origin)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	conn.Write([]byte(strconv.Itoa(data.Number)))

	return nil, nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
