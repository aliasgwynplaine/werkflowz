package main

import (
	"context"
	"encoding/json"
	"fmt"

	. "ccmeshclient/pkg/ccmesh"
	. "ccmeshclient/pkg/common"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type faninHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &faninHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *faninHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("2branch fanin... ", string(input))
	var envelope Envelope
	err := json.Unmarshal(input, &envelope)
	CHECK(err)
	client := NewMeshGoClient("fanin")

	InitRPCClient(client)

	fmt.Println("Opening envelope...")
	client.OpenEnvelope(envelope)
	fmt.Println("starting mailboxservice")
	client.InitMailBoxService()
	from := []string{"incrementor0", "incrementor1"}
	fmt.Println("Waiting for messages... ", from)
	client.WaitForMessages(from)

	fmt.Println("Committing...")

	client.CommitTxn()

	fmt.Println("Committed!")

	return []byte("Ok"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
