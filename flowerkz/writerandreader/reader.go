package main

import (
	"ccmeshclient/pkg/ccmesh"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type infoRep struct {
	Origin string `json:"origin"`
}

type readerHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &readerHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *readerHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("reader...")
	fmt.Println("context: ", ctx)
	fmt.Println("input: ", input)

	cclient := ccmesh.NewMeshGoClient()

	err := json.Unmarshal(input, &cclient)

	check(err)

	fmt.Println("cclient: ", cclient)

	v := cclient.Read("001")

	fmt.Println("just read: ", v)

	conn, err := net.Dial("tcp", cclient.Origin)

	check(err)

	defer conn.Close()

	conn.Write([]byte("Ok"))

	time.Sleep(87 * time.Second)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Going out!")

	return []byte("Ok"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
