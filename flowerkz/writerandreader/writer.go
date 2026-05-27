package main

import (
	"bytes"
	"context"
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

	fmt.Println("context: ", ctx)
	fmt.Println("updating var -> ", randNum)

	cclient := ccmesh.NewMeshGoClient()
	cclient.Write("k", strconv.Itoa(randNum))

	nxt := "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/reader"

	payload := bytes.NewBuffer(input)

	fmt.Println("Sending req to the next one: ", nxt)

	go http.Post(nxt, "*/*", payload)

	time.Sleep(87 * time.Second)

	fmt.Println("Going out!")

	return []byte("OK"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
