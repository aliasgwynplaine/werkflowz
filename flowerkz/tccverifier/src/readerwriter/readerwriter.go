package main

import (
	"ccmeshclient/pkg/ccmesh"
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type readerwriterHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &readerwriterHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you :)")
}

func (h *readerwriterHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("readerwriter")
	client := ccmesh.NewMeshGoClient("readerwriter")
	RunId := string(input)
	//client.InitTxn()
	value := client.Read("x")
	fmt.Println(RunId, "x = ", value)
	new_value := strconv.Itoa(rand.Intn(9999) + 1)
	client.Write("y", new_value)
	fmt.Println(RunId, "y <- ", value)
	client.CommitTxn()
	fmt.Println("-----------------")

	response, err := h.env.InvokeFunc(ctx, "finalreader", input)

	if err != nil {
		panic("erro: InvokeFunc")
	}

	return []byte(response), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
