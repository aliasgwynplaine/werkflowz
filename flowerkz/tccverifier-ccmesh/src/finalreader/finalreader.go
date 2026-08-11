package main

import (
	"ccmeshclient/pkg/ccmesg"
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
	client := ccmesg.NewMeshGoClient("finalreader")
	RunId := string(input)

	//client.InitTxn()
	value_y := client.Read("2")
	value_x := client.Read("1")
	fmt.Println(RunId, "2 = ", value_y)
	fmt.Println(RunId, "1 = ", value_x)
	//client.CommitTxn()

	return []byte("Ok"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
