package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type incrementorRep struct {
	Number int    `json:"number"`
	Step   int    `json:"step"`
	Origin string `json:"origin"`
}

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
	fmt.Println("Incrementor...")
	//decode json
	data := incrementorRep{}

	json.Unmarshal(input, &data)

	fmt.Println("context: ", ctx)
	fmt.Println("data: ", data)

	var nxt string

	if data.Step == 1 {
		// call finish
		nxt = "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/incrementorFinish"
	} else {
		data.Step = data.Step - 1

		nxt = "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/incrementor"
	}

	data.Number = data.Number + 1

	resp, err := json.Marshal(data)

	if err != nil {
		panic(err)
	}

	payload := bytes.NewBuffer(resp)

	go http.Post(nxt, "application/json", payload)

	return nil, nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
