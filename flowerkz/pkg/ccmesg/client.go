package ccmesg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	. "ccmeshclient/pkg/common"

	"math/rand/v2"
	"os"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var NIGHTCORE_GW_ADDR = os.Getenv("NIGHTCORE_GW_ADDR")
var SNITCH_PORT = "46655"

var RPCC *MeshClient = nil

type Envelope struct {
	Tid       string                   `json:"tid"`
	Fname     string                   `json:"fname"`
	Payload   string                   `json:"payload"`
	Deps      map[string]VC            `json:"deps"`
	Writes    map[string]string        `json:"writes"`
	Workload  []map[string]interface{} `json:"workload"`
	ReturnAdr string                   `json:"return"`
	Abort     bool                     `json:"abort"`
}

type MeshGoClient struct {
	Rpcc *MeshClient
	/* messages */
	mu        sync.Mutex
	server    *http.Server
	Tid       string `json:"tid"`
	Fname     string `json:"fname"`
	Payload   string `json:"payload"`
	ReturnAdr string `json:"return"`
	received  map[string]bool
	interChan chan struct{}
	/* causalmesh tcc */
	Workload []map[string]interface{} `json:"workload"`
	Writes   map[string]string        `json:"writes"`
	Deps     map[string]VC            `json:"deps"`
	Abort    bool                     `json:"abort"`
}

func (c *MeshGoClient) SendMessageSync(to string, v string, direct bool) error {
	//fmt.Println(c.Tid, "-", c.Fname, " asking for ", to)
	addr, err := getAddr(c.Tid, to, direct)
	CHECK(err)

	invokeurl := "http://" + addr + "/function/" + to
	message := Envelope{
		Tid:      c.Tid,
		Fname:    c.Fname,
		Payload:  v,
		Deps:     c.Deps,
		Writes:   c.Writes,
		Workload: c.Workload,
	}
	data, err := json.Marshal(message)
	CHECK(err)
	output := bytes.NewBuffer(data)

	response, err := http.Post(invokeurl, "*/*", output) // invokation

	if err != nil {
		//fmt.Println(c.Tid, "-", c.Fname, "- returning uu ", err)
		return err
	}

	//fmt.Println(c.Tid, "-", c.Fname, "- response: ", response)
	response.Body.Close()

	//fmt.Println(c.Tid, "-", c.Fname, "- Message sent to ", to, "through ", invokeurl, " - payload:", v)

	return nil
}

func (c *MeshGoClient) SendMessage(to string, v string, direct bool) error {
	//fmt.Println(c.Tid, "-", c.Fname, " asking for ", to)
	addr, err := getAddr(c.Tid, to, direct)
	CHECK(err)

	invokeurl := "http://" + addr + "/function/" + to
	message := Envelope{
		Tid:       c.Tid,
		Fname:     c.Fname,
		Payload:   v,
		Deps:      c.Deps,
		Writes:    c.Writes,
		Workload:  c.Workload,
		ReturnAdr: c.ReturnAdr,
	}
	data, err := json.Marshal(message)
	CHECK(err)
	output := bytes.NewBuffer(data)
	//fmt.Println(c.Tid, "-", c.Fname, "- Sending request to: ", invokeurl)
	go func() {
		//http.DefaultClient.Timeout = 1 * time.Second

		response, err := http.Post(invokeurl, "*/*", output) // invokation

		if err != nil {
			//fmt.Println(c.Tid, "-", c.Fname, "- returning uu ", err)
			return
		}

		//fmt.Println(c.Tid, "-", c.Fname, "- response: ", response)
		response.Body.Close()
	}()

	//fmt.Println(c.Tid, "-", c.Fname, "- Message sent to ", to, "through ", invokeurl, " - payload:", v)

	return nil
}

func getAddr(tid string, to string, direct bool) (string, error) {
	if tid == "" || direct {
		return NIGHTCORE_GW_ADDR + ":8080", nil
	}
	conn, err := net.Dial("tcp", NIGHTCORE_GW_ADDR+":"+SNITCH_PORT)

	if err != nil {
		return "", err
	}

	defer conn.Close()
	//fmt.Println("Sending request ", tid, "for ", to)
	conn.Write([]byte("GET " + tid + " " + to + "\n"))
	buf := make([]byte, 256) // todo: hardcoded size
	n, err := conn.Read(buf)

	if err != nil {
		return "", err
	}

	addr := string(buf[:n])
	//fmt.Println("recvd Addr: ", buf[:n], " -> ", addr)

	return addr, nil
}

func (c *MeshGoClient) subscribe(addr string) {
	conn, err := net.Dial("tcp", NIGHTCORE_GW_ADDR+":"+SNITCH_PORT)
	CHECK(err)

	defer conn.Close()

	payload := "PUT " + c.Tid + " " + c.Fname + " " + addr + "\n"
	_, err = conn.Write([]byte(payload))
	CHECK(err)
	//fmt.Println("sent ", n, " bytes to the snitch...")
	buf := make([]byte, 32) // todo hardcoded
	_, err = conn.Read(buf)
	//fmt.Println("recv'd ", buf[:n], " as response")
}

func (c *MeshGoClient) WaitForMessages(fromlist []string) {
	if c.Tid == "" {
		panic("No TID!")
	}

	ln, err := net.Listen("tcp", "0.0.0.0:0")

	if err != nil {
		fmt.Errorf("err:", err)
		c.Abort = true
		return
	}

	lip, err := GetLocalIPv4()
	CHECK(err)
	caddr := lip[0].String() + ":" + strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)

	for _, s := range fromlist {
		if _, ok := c.received[s]; !ok {
			c.received[s] = false
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/function/"+c.Fname, http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)
			defer r.Body.Close()

			var envelope Envelope
			err := decoder.Decode(&envelope)
			CHECK(err)

			c.recv(envelope)
		},
	))

	c.server = &http.Server{
		Handler: mux,
	}

	servErrCh := make(chan error, 1)

	go func() {
		servErrCh <- c.server.Serve(ln)
	}()

	c.subscribe(caddr)

	<-c.interChan
	//fmt.Println(c.Tid, " completed!")

	c.server.Close()
	<-servErrCh
}

func (c *MeshGoClient) recv(envelope Envelope) {
	//c.mu.Lock()
	//fmt.Println("recv: ", envelope)

	/*
		if c.isDelivrable(envelope) {
			c.mu.Unlock()
			c.deliver(envelope)
			c.mu.Lock()
			c.deliverCausalMessages()
		} else {
			c.Buffer = append(c.Buffer, envelope)
		}
	*/

	c.deliver(envelope)

	//c.mu.Unlock()
}

func (c *MeshGoClient) deliver(envelope Envelope) {
	//fmt.Println("Delivering ", envelope)
	c.mu.Lock()

	for k, vc := range envelope.Deps {
		InsertOrMergeVC(&c.Deps, k, &vc)
	}

	for k, v := range envelope.Writes {
		//if vv, ok := c.Writes[k]; ok {
		//	if v != vv {
		//		fmt.Println(c.Tid, "- Operation not permited: concurrent write in txn.")
		//		c.Abort = true
		//		break
		//	}
		//}

		//fmt.Println("merging ", k, ": ", v)
		c.Writes[k] = v
	}

	c.received[envelope.Fname] = true

	completed := true

	for _, r := range c.received {
		completed = r && completed
	}

	if completed {
		close(c.interChan)
	}

	c.mu.Unlock()
}

func (client *MeshGoClient) Read(k string) string {
	//fmt.Printf("read: %s ----\n", k)

	if v, ok := client.Writes[k]; ok {
		//fmt.Printf("read-after-write: %s ----\n", v)
		return v
	}

	depsStr, err := json.Marshal(client.Deps)
	//fmt.Println("deps:", depsStr)
	CHECK(err)
	res, err := (*client.Rpcc).ClientRead(context.Background(), &ClientReadRequest{Key: k, Deps: string(depsStr)})
	CHECK(err)
	//fmt.Println("res:", res)
	var vc VC
	err = json.Unmarshal([]byte(res.Vc), &vc)
	CHECK(err)

	if res.Value != "None" {
		InsertOrMergeVC(&client.Deps, k, &vc)
	}

	//fmt.Printf("end_read: %s -> %s----\n", k, res.Value)
	return res.Value
}

func (client *MeshGoClient) CommitTxn() {
	//fmt.Println(client.Tid, "- Commit Txn ----", client.Deps, client.Writes)
	depsStr, err := json.Marshal(client.Deps)
	CHECK(err)
	writesStr, err := json.Marshal(client.Writes)
	CHECK(err)
	_, err = (*client.Rpcc).ClientCommitTxn(context.Background(), &ClientCommitTxnRequest{Deps: string(depsStr), Writes: string(writesStr)})

	if err != nil {
		client.Abort = true
	}

	if client.Tid == "" {
		//fmt.Println("This is a solo buddy... returning")
		return
	}

	conn, err := net.Dial("tcp", NIGHTCORE_GW_ADDR+":"+SNITCH_PORT)
	CHECK(err)
	defer conn.Close()
	_, err = conn.Write([]byte("COMMIT " + client.Tid + "\n"))
	CHECK(err)
}

func (client *MeshGoClient) Write(k string, v string) {
	//fmt.Printf("Write: %s -> %s ----\n", k, v)
	client.Writes[k] = v

	//if client.Tid == "" {
	//	fmt.Println("WARNING: Txn context not initialized")
	//}
}

func CreateClient() *MeshClient {
	conn, err := grpc.Dial(ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	CHECK(err)
	rpcc := NewMeshClient(conn)
	return &rpcc
}

func InitClient() {
	if RPCC == nil {
		RPCC = CreateClient()
	}
}

func InitRPCClient(client *MeshGoClient) {
	if client == nil {
		//fmt.Println("fuck you!")
		return
	}

	if RPCC == nil {
		RPCC = CreateClient()
	}

	//if client.Rpcc == nil {
	//	fmt.Println("InitRPCCClient: new rpcc")
	//	client.Rpcc = RPCC
	//} else {
	//	fmt.Println("InitRPCCClient: uu")
	//}
}

func (c *MeshGoClient) InitTxn() {
	c.Tid = uuid.New().String()
}

func NewMeshGoClient(fname string) *MeshGoClient {
	var client MeshGoClient
	InitClient()

	if client.Rpcc == nil {
		client.Rpcc = RPCC
	} else {
	}

	client.Writes = make(map[string]string, 0)
	client.Deps = make(map[string]VC, 0)
	client.Fname = fname
	client.received = make(map[string]bool)
	client.interChan = make(chan struct{})

	return &client
}

func (c *MeshGoClient) OpenEnvelope(envelope Envelope) {
	//fmt.Println("Opening envelope with message from ", envelope.Fname)
	c.mu.Lock()
	c.Tid = envelope.Tid
	c.Deps = envelope.Deps
	//fmt.Println("Writes: ", envelope.Writes)
	c.Writes = envelope.Writes
	c.Workload = envelope.Workload

	c.received[envelope.Fname] = true

	c.ReturnAdr = envelope.ReturnAdr
	c.mu.Unlock()
}

func Run(input []byte) []byte {
	var envelope Envelope
	err := json.Unmarshal(input, &envelope)
	CHECK(err)

	client := NewMeshGoClient(envelope.Payload)
	InitRPCClient(client)

	client.OpenEnvelope(envelope)
	//fmt.Println(client.Tid, "- Init Execution - inputstr: ", string(input))
	client.Execute()
	//if client.Abort {
	//	fmt.Println(client.Tid, "-", client.Fname, "- Execution aborted! - ", string(input))
	//} else {
	//	fmt.Println(client.Tid, "-", client.Fname, "- Execution completed! - ", string(input))
	//}

	res := Envelope{
		Abort: client.Abort,
	}

	resStr, err := json.Marshal(res)
	CHECK(err)

	return resStr
}

func (client *MeshGoClient) Execute() []byte {
	if client.Writes == nil || client.Deps == nil {
		panic("client not init")
	}
	//fmt.Println(client.Tid, " executing workload -> ", client.Workload)
	workload := client.Workload
	var ln net.Listener
outer:
	for i := 0; i < len(workload); i++ {
		op := workload[i]

		if len(op) != 1 {
			panic("op is not 1")
		}

		for k, v := range op {
			switch k {
			case "T":
				b := int(v.(float64))
				if b > 0 {
					client.InitTxn()
				}
				var err error
				ln, err = net.Listen("tcp", ":0")
				CHECK(err)
				lip, err := GetLocalIPv4()
				CHECK(err)
				caddr := lip[0].String() + ":" + strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
				client.ReturnAdr = caddr
			case "R":
				//fmt.Println("READ")
				res := client.Read(v.(string))
				if res == "None" {
					client.Abort = true
					break outer
				}
			case "W":
				//fmt.Println("WRITE")
				vs := v.([]interface{})
				client.Write(vs[0].(string), vs[1].(string))
			case "O":
				//fmt.Println(client.Tid, "- FANOUT")
				// handle fanout
				grado := int(v.(float64))
				payload := make([][]map[string]interface{}, grado)
				i++

				for nb := 0; nb < grado; i++ {
					nextop := workload[i]

					if len(nextop) != 1 {
						panic("len is not 1 uu")
					}

					ready := false

					for kk, _ := range nextop { // this is not a loop
						payload[nb] = append(payload[nb], nextop)

						if kk == "J" {
							ready = true
						}
					}

					if ready {
						nb++
					}
				}
				//fmt.Println(client.Tid, "- WoRKLOADDDDD:::: -> ", workload[i:])

				for nb := 0; nb < grado; nb++ {
					payload[nb] = append(payload[nb], workload[i:]...)
					client.Workload = payload[nb]
					if rand.IntN(2) == 1 {
						client.SendMessage(fmt.Sprintf("Entry1?t=%s&p=%s", client.Tid, fmt.Sprintf("fux_%d", nb)), fmt.Sprintf("fux_%d", nb), true)
					} else {
						client.SendMessage(fmt.Sprintf("Entry2?t=%s&p=%s", client.Tid, fmt.Sprintf("fux_%d", nb)), fmt.Sprintf("fux_%d", nb), true)
					}
				}

				break outer
			case "J":
				//fmt.Println(client.Tid, "-", client.Fname, "- JOIN")
				// prepare workload
				i++
				client.Workload = client.Workload[i:]
				client.SendMessage("Sink", "Sink", false)
				return nil
			case "I":
				//fmt.Println(client.Tid, "- FANIN")
				// handle fanin
				var fromlist []string

				for j := 0; j < int(v.(float64)); j++ {
					fromlist = append(fromlist, fmt.Sprintf("fux_%d", j))
				}

				client.WaitForMessages(fromlist)

				if client.Abort {
					conn, err := net.Dial("tcp", client.ReturnAdr)
					CHECK(err)
					defer conn.Close()
					return nil
				}
			case "M":
				i++
				client.Workload = client.Workload[i:]

				if rand.IntN(2) == 1 {
					client.SendMessage(fmt.Sprintf("Entry1?t=%s", client.Tid), "Migration", true)
				} else {
					client.SendMessage(fmt.Sprintf("Entry2?t=%s", client.Tid), "Migration", true)
				}

				return nil
			case "C":
				client.CommitTxn()
				conn, err := net.Dial("tcp", client.ReturnAdr)
				CHECK(err)
				defer conn.Close()
				conn.Write([]byte("Ok"))
				return nil
			}

		}
	}

	if ln != nil {
		//fmt.Println(client.Tid, "- Wait for result.")
		c, err := ln.Accept()

		if err != nil {
			fmt.Errorf("error. ", err)
		}

		defer c.Close()
		buf := make([]byte, 32)
		_, err = c.Read(buf)
		CHECK(err)
		ln.Close()
	}

	return nil
}

func Test() string {
	conn, err := grpc.Dial(ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	CHECK(err)
	defer conn.Close()
	c := NewMeshClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.HealthCheck(ctx, &HealthCheckRequest{})
	CHECK(err)
	return r.Status
}
