package main

import (
	"bytes"
	"ccmeshclient/pkg/ccmail"
	. "ccmeshclient/pkg/common"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type reg struct {
	t  string
	o  string
	vc VC
}

var mu sync.Mutex
var memPool = make([]reg, 0) // to store "ping"s and "pongs"

type infoRep struct {
	Origin    string `json:"origin"`
	Rounds    int    `json:"rounds"`
	FuncId    int    `json:"funcid"`
	Initiator bool   `json:"initiator"`
}

type pingpongHandler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	return &pingpongHandler{env: env}, nil
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented. Fuck you")
}

func pingpongdeliver(m ccmail.Message) {
	mu.Lock()
	fmt.Println("writing message in mem: ", m)
	memPool = append(memPool, reg{m.Payload, m.Origin, m.Vc})
	mu.Unlock()
}

func (f *pingpongHandler) Call(ctx context.Context, input []byte) ([]byte, error) {
	inputStr := string(input)
	fmt.Println("pingpong -> ", inputStr)
	var inpayload infoRep

	err := json.Unmarshal(input, &inpayload)

	CHECK(err)

	var mail *ccmail.MailGoClient

	rounds := inpayload.Rounds

	if inpayload.Initiator {
		fmt.Println("Initiation here. Sending PING")

		mail = ccmail.CreateNewMailClient(
			inpayload.FuncId,
			ccmail.PINGPORT,
			0,
			pingpongdeliver,
		)

		mail.SendMessage("127.0.0.1:1284", "PING")
	} else {
		mail = ccmail.CreateNewMailClient(
			inpayload.FuncId,
			ccmail.PONGPORT,
			0,
			pingpongdeliver,
		)
	}

	for times := 0; times < rounds; {
		if len(memPool) != 0 {
			mu.Lock()
			peek := memPool[0]
			memPool = memPool[1:]
			mu.Unlock()

			if peek.t == "PING" {
				fmt.Println("PING recv'd VC: ", peek.vc)
				fmt.Println("sending PONG")
				err := mail.SendMessage(peek.o, "PONG")

				CHECK(err)

				times++
			}

			if peek.t == "PONG" {
				fmt.Println("PONG recv'd VC: ", peek.vc)
				fmt.Println("sending PING")
				err := mail.SendMessage(peek.o, "PING")

				CHECK(err)

				times++
			}
		}
	}

	invokeurl := "http://" + os.Getenv("NIGHTCORE_GW_ADDR") + ":8080/function/finish"
	fmt.Println("Seding request. GW: ", os.Getenv("NIGHTCORE_GW_ADDR"))

	outdata, err := json.Marshal(inpayload)
	payload := bytes.NewBuffer(outdata)

	go http.Post(invokeurl, "*/*", payload)

	time.Sleep(time.Duration(83) * time.Second)

	return []byte("OK"), nil
}

func main() {
	faas.Serve(&funcHandlerFactory{})
}
