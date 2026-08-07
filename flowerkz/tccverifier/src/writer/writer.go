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

type writerHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &writerHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you :)")
}

func (h *writerHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("writer: ", input, string(input))
	client := ccmesh.NewMeshGoClient("writer")
	value := strconv.Itoa(1 + rand.Intn(9999))
	RunId := string(input)

	//client.InitTxn()
	client.Write("x", value)
	client.CommitTxn()
	fmt.Println(RunId, "x <-", value)
	fmt.Println("--------------------")

	response, err := h.env.InvokeFunc(ctx, "readerwriter", input)

	if err != nil {
		panic("error: InvokeFunc")
	}

	return response, nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
