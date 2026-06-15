package main

import (
	. "ccmeshclient/pkg/common"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type reg struct {
	t string
	o string
}

type infoRep struct {
	Origin    string `json:"origin"`
	Rounds    string `json:"rounds"`
	FuncId    int    `json:"funcid"`
	Initiator bool   `json:"initiator"`
}

type finishHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &finishHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (f *finishHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	inputStr := string(input)
	fmt.Println("finish -> ", inputStr)
	var inpayload infoRep

	err := json.Unmarshal(input, &inpayload)

	conn, err := net.Dial("tcp", inpayload.Origin)

	CHECK(err)

	defer conn.Close()

	conn.Write([]byte("Ok!"))

	return []byte("OK"), nil

}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
