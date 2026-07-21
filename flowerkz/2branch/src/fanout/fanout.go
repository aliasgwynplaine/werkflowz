package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	. "ccmeshclient/pkg/ccmesh"
	//. "ccmeshclient/pkg/common"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type fanoutHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &fanoutHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *fanoutHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("2branch fanout... ", string(input))
	fmt.Println("GW: ", os.Getenv("NIGHTCORE_GW_ADDR"))

	client := NewMeshGoClient("fanout") // TODO: fix hardcoded !
	client.InitTxn()

	//var_x := "x"
	//val_x := "1235"
	//var_y := "y"
	//val_y := "0"

	//client.Write(var_x, val_x)
	//client.Write(var_y, val_y)

	var wg sync.WaitGroup

	wg.Go(func() {
		err := client.SendMessage("incrementor0", "incrementor0")
		if err != nil {
			panic(err)
		}
	})

	wg.Go(func() {
		err := client.SendMessage("incrementor1", "incrementor1")
		if err != nil {
			panic(err)
		}
	})

	//time.Sleep(3 * time.Second)

	wg.Wait()

	return []byte("Ok"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
