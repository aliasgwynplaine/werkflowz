package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type infoRep struct {
	Origin string `json:"origin"`
}

func getLocalIPv4() ([]net.IP, error) {
	var ips []net.IP
	interfc, err := net.InterfaceAddrs()

	if err != nil {
		panic(err)
	}

	for _, addr := range interfc {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}

	return ips, nil
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
	fmt.Println("Init... ", string(input))

	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(10)
	fmt.Println("Random Num: ", randomNum)

	ips, err := getLocalIPv4()
	fmt.Println("ips: ", ips)

	if err != nil {
		panic(err)
	}

	/*
		TODO: write here ?????
	*/

	invokeurl := "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/writer"
	fmt.Println("Seding request. GW: ", os.Getenv("NIGHTCORE_GW_ADDR"))

	fmt.Println("Openning connection...")
	listener, err := net.Listen("tcp", "0.0.0.0:0")

	if err != nil {
		panic(err)
	}

	fmt.Println("ips: ", ips)
	fmt.Println("preparing payload...")
	addr := ips[0].String() + ":" + strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	object := infoRep{addr}
	data, err := json.Marshal(object)

	if err != nil {
		panic(err)
	}

	fmt.Println("data: ", data)

	payload := bytes.NewBuffer(data)

	fmt.Println("invoking the init...")

	_, err = http.Post(invokeurl, "*/*", payload) // todo: check if this is the "good" way to do it

	if err != nil {
		fmt.Println("error in post")
		panic(err)
	}

	fmt.Println("invokation is done. Waiting response")

	defer listener.Close()

	c, err := listener.Accept()

	if err != nil {
		fmt.Println("error...")
		panic(err)
	}

	defer c.Close()

	fmt.Println("Connection accepted: ", c.RemoteAddr())

	buf := make([]byte, 1024)

	n, err := c.Read(buf)

	if err != nil {
		panic(err)
	}

	fmt.Println("Recv: ", buf[:n], " -> ", string(buf[:n]))

	return []byte(buf[:n]), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
