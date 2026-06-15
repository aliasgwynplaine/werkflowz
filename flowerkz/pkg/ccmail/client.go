package ccmail

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	. "ccmeshclient/pkg/common"
)

var INPORT = "1283"
var PINGPORT = "1283"
var PONGPORT = "1284"

type Message struct {
	FuncId  int    `json:"func_id"`
	Vc      VC     `json:"vc"`
	Payload string `json:"payload"`
	Origin  string `json:"origin"`
}

type MailGoClient struct {
	mu          sync.Mutex
	FuncId      int
	Origin      string
	Buffer      []Message
	Delay       time.Duration
	Vc          VC
	interChan   chan struct{}
	realdeliver func(Message)
}

func (c *MailGoClient) tick() {
	c.Vc[c.FuncId]++
}

func (c *MailGoClient) merge(vc VC) {
	MergeIntoVC(&c.Vc, &vc)
}

func (c *MailGoClient) isDelivrable(m Message) bool {
	fmt.Println("cheking if it's deliverable")
	p := m.FuncId

	if m.Vc[p] != c.Vc[p]+1 {
		fmt.Println("non uu")
		return false
	}

	for k, v := range m.Vc {
		if k == p {
			continue
		}

		if v > c.Vc[k] {
			return false
		}
	}

	fmt.Println("oui :D")

	return true
}

func default_realdeliver(m Message) {
	fmt.Println("Id: ", m.FuncId)
	fmt.Println("Vc: ", m.Vc)
	fmt.Println("Payload: ", m.Payload)
	fmt.Println("Origin: ", m.Origin)
}

func (c *MailGoClient) deliver(m Message) {
	c.mu.Lock()
	c.merge(m.Vc)
	c.mu.Unlock()

	c.realdeliver(m)
}

func (c *MailGoClient) deliverCausalMessages() {
	for {
		unlocked := false
		rest := c.Buffer[0:0]

		for _, m := range c.Buffer {
			if c.isDelivrable(m) {
				c.mu.Unlock()
				fmt.Println("calling deliver...")
				c.deliver(m)
				c.mu.Lock()
				unlocked = true
			} else {
				rest = append(rest, m)
			}
		}

		c.Buffer = rest

		/* maybe something arrived whtn we were unlocked
		* if we were unlocked, then we make another round
		 */
		if !unlocked {
			break
		}
	}
}

func (c *MailGoClient) recv(m Message) {
	c.mu.Lock()
	fmt.Println("recv: ", m)

	if c.isDelivrable(m) {
		c.mu.Unlock()
		c.deliver(m)
		c.mu.Lock()
		c.deliverCausalMessages()
	} else {
		c.Buffer = append(c.Buffer, m)
	}

	c.mu.Unlock()
}

func (c *MailGoClient) handleConn(conn net.Conn) {
	fmt.Println("handerConn")

	defer conn.Close()

	var msg Message

	fmt.Println("decoding")
	err := json.NewDecoder(conn).Decode(&msg)

	CHECK(err)

	fmt.Println("msg: ", msg)

	c.recv(msg)
}

func (c *MailGoClient) waitForMessages(port string) {
	ln, err := net.Listen("tcp", "0.0.0.0:"+port)

	CHECK(err)
	/* // handle EADDRINUSE

	if err != nil && errors.Is(err, syscall.EADDRINUSE) {
		NEWPORT := strconv.Itoa(PORT + 1)
		ln, err = net.Listen("tcp", "0.0.0.0:"+NEWPORT)
	}
	*/

	defer ln.Close()

	fmt.Println("going into the loop...")

	for {
		select {
		case <-c.interChan:
			fmt.Println("going out !")
			return
		default:
			fmt.Println("default LOL")
		}

		// TODO: add a timeout ??
		conn, err := ln.Accept()
		fmt.Println("connection accepted...")

		if err != nil {
			select {
			case <-c.interChan:
				return
			default:
				fmt.Println("wfmerror:", err)
			}

			continue
		}

		go c.handleConn(conn)
	}

}

// address must contain the :port
// TODO: IMProve this :P
func (c *MailGoClient) SendMessage(addr string, v string) error {
	conn, err := net.Dial("tcp", addr)

	for err != nil {
		fmt.Println("err: ", err)
		time.Sleep(time.Duration(2) * time.Second)
		fmt.Println("Retrying...")
		conn, err = net.Dial("tcp", addr)
	}

	defer conn.Close()

	c.mu.Lock()

	c.tick()
	vc := c.Vc

	c.mu.Unlock()

	m := Message{c.FuncId, vc, v, c.Origin}

	encoder := json.NewEncoder(conn)

	return encoder.Encode(m)
}

/**
 * Maybe is not such a brilliant idea to receive all
 * the metadata in the deliver function.
 */
func CreateNewMailClient(funcid int, port string, delay int, fdeliver func(Message)) *MailGoClient {
	ip, err := GetLocalIPv4()

	CHECK(err)

	if fdeliver == nil {
		fdeliver = default_realdeliver
	}

	c := MailGoClient{
		FuncId:      funcid,
		Origin:      ip[0].String() + ":" + port,
		Buffer:      make([]Message, 0),
		Delay:       time.Duration(delay) * time.Millisecond,
		interChan:   make(chan struct{}),
		realdeliver: fdeliver,
	}

	fmt.Println("client crafted. creating thread to wait for messages")

	go c.waitForMessages(port)

	return &c
}
