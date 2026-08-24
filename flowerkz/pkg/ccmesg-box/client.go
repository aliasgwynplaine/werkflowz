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

var magicBox = wrapper{
	box: make(map[string]*caja),
}

type wrapper struct {
	mu  sync.Mutex
	box map[string]*caja
}

type caja struct {
	nb  int
	mu  sync.Mutex
	buf []map[string]string
}

func newCaja(nb int) *caja {
	return &caja{
		nb:  nb,
		buf: make([]map[string]string, 0, nb),
	}
}

func (c *caja) append(wr map[string]string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buf = append(c.buf, wr)

	if len(c.buf) == c.nb {
		return true
	}

	return false
}

func (c *caja) getWrites() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	writes := make(map[string]string)

	for _, e := range c.buf {
		for k, v := range e {
			writes[k] = v // TODO: conflict check
		}
	}

	return writes
}

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
	interChan chan struct{}
	/* causalmesh tcc */
	Workload []map[string]interface{} `json:"workload"`
	Writes   map[string]string        `json:"writes"`
	Deps     map[string]VC            `json:"deps"`
	Abort    bool                     `json:"abort"`
}

func (c *MeshGoClient) SendMessageHashed(to string, v string) error {
	//fmt.Println(c.Tid, "-", c.Fname, " asking for ", to)
	host := GetHost(c.Tid, to) + ":8080"

	invokeurl := "http://" + host + "/function/" + to
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

func (c *MeshGoClient) SendMessageSync(to string, v string, direct bool) error {
	//fmt.Println(c.Tid, "-", c.Fname, " asking for ", to)
	addr := getAddr() + ":8080"

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

	//http.DefaultClient.Timeout = 1 * time.Second

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

func (c *MeshGoClient) SendMessage(to string, v string) error {
	//fmt.Println(c.Tid, "-", c.Fname, " asking for ", to)
	addr := getAddr() + ":8080"

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

func getAddr() string {
	idx := rand.IntN(T)
	addr := PEERS[idx]

	return addr
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
	client.interChan = make(chan struct{})

	return &client
}

func (c *MeshGoClient) OpenEnvelope(envelope Envelope) {
	c.mu.Lock()

	c.Tid = envelope.Tid
	c.Deps = envelope.Deps
	c.Writes = envelope.Writes
	c.Workload = envelope.Workload
	c.ReturnAdr = envelope.ReturnAdr

	c.mu.Unlock()
}

func Run(input []byte) []byte {
	var envelope Envelope
	err := json.Unmarshal(input, &envelope)
	CHECK(err)
	//conn, err := grpc.Dial(ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//CHECK(err)
	//
	//rpcc := NewMeshClient(conn)
	//var client MeshGoClient
	//InitClient()
	//client.Rpcc = RPCC
	client := NewMeshGoClient(envelope.Payload)
	InitRPCClient(client)

	client.OpenEnvelope(envelope)
	//fmt.Println(client.Tid, "- Init Execution - inputstr: ", string(input))
	client.Execute()
	// todo erase workload to avoid unnecessary serialization
	envelope.Abort = client.Abort
	envelopeStr, err := json.Marshal(envelope)
	CHECK(err)
	//if client.Abort {
	//	fmt.Println(client.Tid, "-", client.Fname, "- Execution aborted! - ", string(input))
	//} else {
	//	fmt.Println(client.Tid, "-", client.Fname, "- Execution completed! - ", string(input))
	//}
	return envelopeStr
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
				client.InitTxn()
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
						client.SendMessage(fmt.Sprintf("Entry1?t=%s&p=%s", client.Tid, fmt.Sprintf("fux_%d", nb)), fmt.Sprintf("fux_%d", nb))
					} else {
						client.SendMessage(fmt.Sprintf("Entry2?t=%s&p=%s", client.Tid, fmt.Sprintf("fux_%d", nb)), fmt.Sprintf("fux_%d", nb))
					}
				}

				break outer
			case "J":
				//fmt.Println(client.Tid, "-", client.Fname, "- JOIN")
				// prepare workload
				i++
				client.Workload = client.Workload[i:]
				client.SendMessageHashed("Sink", "Sink")
				return nil
			case "I":
				magicBox.mu.Lock()

				if box, ok := magicBox.box[client.Tid]; ok {
					fmt.Println(client.Tid, "- caja: ", box.buf)

					if box.append(client.Writes) {
						fmt.Println(client.Tid, "box completed!")
						client.Writes = box.getWrites()
						delete(magicBox.box, client.Tid)
						magicBox.mu.Unlock()
						continue
					} else {
						magicBox.mu.Unlock()
						return nil
					}
				} else {
					magicBox.box[client.Tid] = newCaja(int(v.(float64)))
					//fmt.Println(client.Tid, "new box!")
					box = magicBox.box[client.Tid]
					box.append(client.Writes)
					magicBox.mu.Unlock()
					return nil
				}
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
		c, err := ln.Accept()

		if err != nil {
			panic(err)
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
