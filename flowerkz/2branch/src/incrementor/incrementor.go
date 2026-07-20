package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	. "ccmeshclient/pkg/ccmesh"
	. "ccmeshclient/pkg/common"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

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
	fmt.Println("incrementor...", string(input))
	var envelope Envelope
	err := json.Unmarshal(input, &envelope)
	CHECK(err)
	client := NewMeshGoClient(envelope.Payload)
	client.OpenEnvelope(envelope)

	InitRPCClient(client) // TODO: CHECK THIS
	//client.InitMailBoxService()

	if envelope.Payload == "incrementor0" {
		//x := client.Read("x")
		//v_x, _ := strconv.Atoi(x)
		client.Write("x", strconv.Itoa(12345))
		client.Write("a", "123")
	} else {
		//y := client.Read("y")
		//v_y, _ := strconv.Atoi(y)
		client.Write("y", strconv.Itoa(98765))
		client.Write("b", "5193")
	}

	client.SendMessage("fanin", "fanin")

	//time.Sleep(3 * time.Second)

	fmt.Println("Returning...")
	return []byte("Ok"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
