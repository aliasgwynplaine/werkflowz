package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	. "ccmeshclient/pkg/common"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type infoRep struct {
	Origin    string `json:"origin"`
	Rounds    int    `json:"rounds"`
	FuncId    int    `json:"funcid"`
	Initiator bool   `json:"initiator"`
}

type initHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &initHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *initHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("Pingpong init... ", string(input))

	rounds := 5

	fmt.Printf("Sending %d rounds!\n", rounds)

	ips, err := GetLocalIPv4()
	fmt.Println("ips: ", ips)

	CHECK(err)

	invokeurl := "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/"
	fmt.Println("Seding request. GW: ", os.Getenv("NIGHTCORE_GW_ADDR"))

	fmt.Println("Openning connection...")
	listener, err := net.Listen("tcp", "0.0.0.0:0")

	CHECK(err)

	defer listener.Close()

	fmt.Println("ips: ", ips)
	fmt.Println("preparing payload...")
	addr := ips[0].String() + ":" + strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	pingobject := infoRep{addr, rounds, 0, true}  // TODO: fix hardcoded
	pongobject := infoRep{addr, rounds, 1, false} // TODO: fix hardcoded
	pingdata, err := json.Marshal(pingobject)
	pongdata, err := json.Marshal(pongobject)

	CHECK(err)

	fmt.Println("pingdata: ", pingdata)
	fmt.Println("pongdata: ", pongdata)

	pingpayload := bytes.NewBuffer(pingdata)
	pongpayload := bytes.NewBuffer(pongdata)

	fmt.Println("invoking the pong...")

	go http.Post(invokeurl+"pong", "*/*", pongpayload)

	fmt.Println("invoking the ping...")

	go http.Post(invokeurl+"ping", "*/*", pingpayload)

	fmt.Println("invokation is done. Waiting response")

	c, err := listener.Accept()

	CHECK(err)

	defer c.Close()

	fmt.Println("Connection accepted: ", c.RemoteAddr())

	buf := make([]byte, 1024)

	n, err := c.Read(buf)

	CHECK(err)

	fmt.Println("Recv: ", buf[:n], " -> ", string(buf[:n]))

	return []byte(buf[:n]), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
