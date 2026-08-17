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
	Tid      string                   `json:"tid"`
	Fname    string                   `json:"fname"`
	Payload  string                   `json:"payload"`
	Deps     map[string]VC            `json:"deps"`
	Writes   map[string]string        `json:"writes"`
	Workload []map[string]interface{} `json:"workload"`
}

type MeshGoClient struct {
	Rpcc *MeshClient
	/* messages */
	mu        sync.Mutex
	Tid       string `json:"tid"`
	Fname     string `json:"fname"`
	ServerId  string `json:"serverid"`
	Payload   string `json:"payload"`
	Buffer    []Envelope
	Delivered []Envelope
	interChan chan struct{}
	/* causalmesh tcc */
	Workload []map[string]interface{} `json:"workload"`
	Writes   map[string]string        `json:"writes"`
	Deps     map[string]VC            `json:"deps"`
	Input    string                   `json:"input"`
	Abort    bool                     `json:"abort"`
}

func (c *MeshGoClient) SendMessage(to string, v string) error {
	fmt.Println("asking for ", to)
	addr, err := getAddr(c.Tid, to)
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
	fmt.Println("Sending request to: ", invokeurl)
	//go func() {
	response, err := http.Post(invokeurl, "*/*", output) // invokation

	if err != nil {
		return err
	}

	fmt.Println("response: ", response)
	//}()

	fmt.Println("Message sent to ", to, " for txn ", c.Tid, "through ", invokeurl)

	return nil
}

func getAddr(tid string, to string) (string, error) {
	if tid == "" {
		return NIGHTCORE_GW_ADDR + ":8080", nil
	}
	conn, err := net.Dial("tcp", NIGHTCORE_GW_ADDR+":"+SNITCH_PORT)

	if err != nil {
		return "", err
	}

	defer conn.Close()
	fmt.Println("Sending request ", tid, "for ", to)
	conn.Write([]byte("GET " + tid + " " + to + "\n"))
	buf := make([]byte, 256) // todo: hardcoded size
	n, err := conn.Read(buf)

	if err != nil {
		return "", err
	}

	addr := string(buf[:n])
	fmt.Println("recvd Addr: ", buf[:n], " -> ", addr)

	return addr, nil
}

func (c *MeshGoClient) subscribe(addr string) {
	conn, err := net.Dial("tcp", NIGHTCORE_GW_ADDR+":"+SNITCH_PORT)
	CHECK(err)

	defer conn.Close()

	payload := "PUT " + c.Tid + " " + c.Fname + " " + addr + "\n"
	n, err := conn.Write([]byte(payload))
	CHECK(err)
	fmt.Println("sent ", n, " bytes to the snitch...")
	buf := make([]byte, 32) // todo hardcoded
	n, err = conn.Read(buf)
	fmt.Println("recv'd ", buf[:n], " as response")
}

// todo: change to a condition variable
func (c *MeshGoClient) WaitForMessages(fromlist []string) error {
	received := make(map[string]bool)

	for _, p := range fromlist {
		received[p] = false
	}

	fmt.Println("received: ", received)

outer:
	for {
		fmt.Println("lock")
		c.mu.Lock()
		fmt.Println(received)
		for _, e := range c.Delivered {
			if received[e.Fname] {
				fmt.Println("already received from ", e.Fname)
				continue
			}

			for i := 0; i < len(fromlist); i++ {
				if e.Fname == fromlist[i] {
					fmt.Println("found msg from ", e.Fname)
					received[e.Fname] = true
					break
				}
			}

			out := true

			for _, r := range received {
				out = r && out
			}

			if out {
				fmt.Println("All messages received !")
				break outer
			}
		}
		fmt.Println("Unlock")
		c.mu.Unlock()
	}

	c.mu.Unlock()

	return nil
}

func (c *MeshGoClient) listenIncommingMessages() {
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	CHECK(err)
	//defer ln.Close()

	lip, err := GetLocalIPv4()
	CHECK(err)
	caddr := lip[0].String() + ":" + strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	fmt.Println("going into the listening loop with addr ", caddr)
	fmt.Printf("endpoint: /function/%s\n", c.Fname)

	http.Handle("/function/"+c.Fname, http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			fmt.Println("message received!!!!")
			decoder := json.NewDecoder(r.Body)

			var envelope Envelope
			err := decoder.Decode(&envelope)
			CHECK(err) // maybe return error to sender here
			fmt.Fprint(rw, "Ok")

			c.recv(envelope)

		},
	))

	fmt.Println("deploying the http.")
	go func() {
		err = http.Serve(ln, nil)
		CHECK(err)
	}()

	fmt.Println("MailBoxService online!")
	c.subscribe(caddr)

}

func (c *MeshGoClient) handleConn(conn net.Conn) {
	fmt.Println("Handling conn: ", conn)
	defer conn.Close()

	var envelope Envelope
	err := json.NewDecoder(conn).Decode(&envelope)
	CHECK(err)
	c.recv(envelope)
}

func (c *MeshGoClient) recv(envelope Envelope) {
	//c.mu.Lock()
	fmt.Println("recv: ", envelope)

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
	fmt.Println("Delivering ", envelope)
	c.mu.Lock()
	c.Delivered = append(c.Delivered, envelope)

	for k, vc := range envelope.Deps {
		InsertOrMergeVC(&c.Deps, k, &vc)
	}

	for k, v := range envelope.Writes {
		if _, ok := c.Writes[k]; ok {
			panic("Operation not permited: concurrent write in txn.") // maybe return error to the sender
		}

		fmt.Println("merging ", k, ": ", v)
		c.Writes[k] = v
	}

	c.mu.Unlock()
}

func (client *MeshGoClient) Read(k string) string {
	fmt.Printf("read: %s ----\n", k)

	if v, ok := client.Writes[k]; ok {
		fmt.Printf("read-after-write: %s ----\n", v)
		return v
	}

	depsStr, err := json.Marshal(client.Deps)
	fmt.Println("deps:", depsStr)
	CHECK(err)
	res, err := (*client.Rpcc).ClientRead(context.Background(), &ClientReadRequest{Key: k, Deps: string(depsStr)})
	CHECK(err)
	fmt.Println("res:", res)
	var vc VC
	err = json.Unmarshal([]byte(res.Vc), &vc)
	CHECK(err)

	if res.Value != "None" {
		InsertOrMergeVC(&client.Deps, k, &vc)
	}

	fmt.Printf("end_read: %s -> %s----\n", k, res.Value)
	return res.Value
}

func (client *MeshGoClient) CommitTxn() {
	fmt.Println("Commit Txn ----", client.Deps, client.Writes)
	depsStr, err := json.Marshal(client.Deps)
	CHECK(err)
	writesStr, err := json.Marshal(client.Writes)
	CHECK(err)
	_, err = (*client.Rpcc).ClientCommitTxn(context.Background(), &ClientCommitTxnRequest{Deps: string(depsStr), Writes: string(writesStr)})

	if err != nil {
		client.Abort = true
	}

	if client.Tid == "" {
		fmt.Println("This is a solo buddy... returning")
		return
	}

	conn, err := net.Dial("tcp", NIGHTCORE_GW_ADDR+":"+SNITCH_PORT)
	CHECK(err)
	defer conn.Close()
	_, err = conn.Write([]byte("COMMIT " + client.Tid + "\n"))
	CHECK(err)
}

func (client *MeshGoClient) Write(k string, v string) {
	fmt.Printf("Write: %s -> %s ----\n", k, v)
	client.Writes[k] = v

	if client.Tid == "" {
		fmt.Println("WARNING: Txn context not initialized")
	}
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
		fmt.Println("fuck you!")
		return
	}

	if RPCC == nil {
		RPCC = CreateClient()
	}

	if client.Rpcc == nil {
		fmt.Println("InitRPCCClient: new rpcc")
		client.Rpcc = RPCC
	} else {
		fmt.Println("InitRPCCClient: uu")
	}
}

func (c *MeshGoClient) InitTxn() {
	c.Tid = uuid.New().String()
}

func NewMeshGoClient(fname string) *MeshGoClient {
	var client MeshGoClient
	InitClient()

	if client.Rpcc == nil {
		fmt.Println("newmeshgoclient: new rpcc")
		client.Rpcc = RPCC
	} else {
		fmt.Println("newmeshgoclient: uu")
	}

	client.Writes = make(map[string]string, 0)
	client.Deps = make(map[string]VC, 0)
	//client.Buffer = append(client.Buffer, 0)
	client.Fname = fname
	client.Delivered = make([]Envelope, 0)
	client.interChan = make(chan struct{})

	return &client
}

func (c *MeshGoClient) InitMailBoxService() {
	if c.Tid == "" {
		panic("No Tid")
	}

	c.listenIncommingMessages()
}

func (c *MeshGoClient) OpenEnvelope(envelope Envelope) {
	fmt.Println("Opening envelope with message from ", envelope.Fname)
	c.mu.Lock()
	c.Tid = envelope.Tid
	c.Deps = envelope.Deps
	fmt.Println("Writes: ", envelope.Writes)
	c.Writes = envelope.Writes
	c.Workload = envelope.Workload
	c.Delivered = append(c.Delivered, envelope) // maybe too much data
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
	var client MeshGoClient
	InitClient()
	client.Rpcc = RPCC

	client.OpenEnvelope(envelope)
	client.Execute()
	CHECK(err)
	envelopeStr, err := json.Marshal(envelope)
	CHECK(err)
	//fmt.Println(string(clientStr))
	return envelopeStr
}

func (client *MeshGoClient) Execute() []byte {
	if client.Writes == nil || client.Deps == nil {
		panic("client not init")
	}
	//fmt.Println(client.Workload)
	workload := client.Workload
	abort := false
outer:
	for i := 0; i < len(workload); i++ {
		op := workload[i]

		if len(op) != 1 {
			panic("op is not 1")
		}

		for k, v := range op {
			switch k {
			case "R":
				//start := time.Now()
				res := client.Read(v.(string))
				if res == "None" {
					abort = true
					break outer
				}
				//fmt.Println("read time:", time.Since(start))
			case "W":
				//start := time.Now()
				vs := v.([]interface{})
				client.Write(vs[0].(string), vs[1].(string))
				//fmt.Println("write time:", time.Since(start))
			case "O":
				// handle fanout
				client.InitTxn()
				grado := v.(int)
				payload := make([][]map[string]interface{}, grado)

				for nb := 0; nb < grado; i++ {
					nextop := workload[i]

					if len(nextop) != 1 {
						panic("len is not 1 uu")
					}

					ready := false

					for kk, _ := range nextop {
						payload[nb] = append(payload[nb], nextop)

						if kk == "J" {
							ready = true
						}
					}

					if ready {
						nb++
					}
				}

				for nb := 0; nb < grado; nb++ {
					payload[nb] = append(payload[nb], workload[i:]...)
					client.Workload = payload[nb]
					client.SendMessage("Entry", fmt.Sprintf("fux_%d", i))
				}
			case "J":
				// prepare workload
				client.Workload = client.Workload[i:]
				client.SendMessage("Entry1", client.Fname)
			case "I":
				// handle fanin
				client.InitMailBoxService()
				var fromlist []string

				for i := 0; i < v.(int); i++ {
					fromlist = append(fromlist, fmt.Sprintf("fux_%d", i))
				}

				client.WaitForMessages(fromlist)
			case "C":
				client.CommitTxn()
			}

		}
	}
	client.Abort = abort
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
