package main

import (
	"ccmeshclient/pkg/ccmesh"
	"context"
	"fmt"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type finalreaderHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &finalreaderHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you :)")
}

func (h *finalreaderHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("finalreader")
	client := ccmesh.NewMeshGoClient("finalreader")
	RunId := string(input)

	//client.InitTxn()
	value_x := client.Read("x")
	value_y := client.Read("y")
	fmt.Println(RunId, "x = ", value_x)
	fmt.Println(RunId, "y = ", value_y)
	//client.CommitTxn()

	return []byte("Ok"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
