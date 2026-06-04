package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"ccmeshclient/pkg/ccmesh"

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

type writerHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &writerHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *writerHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("writer...")
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(10)

	data := infoRep{}

	err := json.Unmarshal(input, &data)

	check(err)

	fmt.Println("context: ", ctx)
	fmt.Println("data: ", data)
	fmt.Println("writing var 001 -> ", randNum)

	cclient := ccmesh.NewMeshGoClient()
	if cclient.Rpcc == nil {
		fmt.Println("Rpcc is null in the function")
	}
	cclient.Write("001", strconv.Itoa(randNum))

	fmt.Println("write op was made...")

	nxt := "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/reader"

	cclient.Rpcc = nil
	clientStr, err := json.Marshal(cclient)
	fmt.Println("clientStr: ", clientStr)

	payload := bytes.NewBuffer(clientStr)

	fmt.Println("Sending req to the reader: ", nxt)

	go http.Post(nxt, "*/*", payload)

	fmt.Println("request sent. Sleeping...")
	time.Sleep(87 * time.Second)

	fmt.Println("Going out!")

	return []byte("OK"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
