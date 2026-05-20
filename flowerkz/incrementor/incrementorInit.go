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
	"time"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

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

type incrementorRep struct {
	Number int    `json:"number"`
	Step   int    `json:"step"`
	Origin string `json:"origin"`
}

type incrementorInitHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &incrementorInitHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func (h *incrementorInitHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("Init... ", string(input))
	ips, err := getLocalIPv4()
	fmt.Println("ips: ", ips)

	if err != nil {
		panic(err)
	}

	num, err := strconv.Atoi(string(input))

	if err != nil {
		panic(err)
	}

	fmt.Println("ips: ", ips)
	fmt.Println("preparing payload...")
	object := incrementorRep{num, 99, ips[0].String() + ":12345"}

	fmt.Println(object)

	data, err := json.Marshal(object)

	if err != nil {
		panic(err)
	}

	fmt.Println("data: ", data)

	payload := bytes.NewBuffer(data)

	invokeurl := "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/incrementor"
	go http.Post(invokeurl, "application/json", payload) // todo: check if this is the "good" way to do it

	listener, err := net.Listen("tcp", "0.0.0.0:12345") // todo: randomize port

	if err != nil {
		panic(err)
	}

	defer listener.Close()

	c, err := listener.Accept()

	if err != nil {
		panic(err)
	}

	fmt.Println("Connection accepted: ", c.RemoteAddr())

	buf := make([]byte, 1024)
	err = c.SetDeadline(time.Now().Add(1 * time.Minute))
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
