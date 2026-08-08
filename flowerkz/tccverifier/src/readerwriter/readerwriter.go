package main

import (
	"bytes"
	"ccmeshclient/pkg/ccmesh"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
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

	/* response, err := h.env.InvokeFunc(ctx, "finalreader", input)

	if err != nil {
		panic("erro: InvokeFunc")
	}
	*/

	invokeurl := "http://" + ccmesh.NIGHTCORE_GW_ADDR + ":8080/function/finalreader"
	output := bytes.NewBuffer(input)
	response, err := http.Post(invokeurl, "*/*", output)

	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	realresponse, err := io.ReadAll(response.Body)

	if err != nil {
		panic(err)
	}

	return []byte(realresponse), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
