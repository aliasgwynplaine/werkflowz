package ccmesh

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "ccmeshclient/pkg/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var RPCC *MeshClient = nil

type MeshGoClient struct {
	Rpcc     *MeshClient
	Workload []map[string]interface{} `json:"workload"`
	Local    map[string]*Meta         `json:"local"`
	Deps     map[string]VC            `json:"deps"`
	Input    string                   `json:"input"`
	Abort    bool                     `json:"abort"`
	Origin   string                   `json:"origin"`
}

func (client *MeshGoClient) Read(k string) string {
	//start := time.Now()
	fmt.Printf("read: %s ----\n", k)

	if m, ok := client.Local[k]; ok {
		return m.Value
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
	//fmt.Println("read", k, ": ", vc, " time:", time.Since(start))
	fmt.Printf("end_read: %s ----\n", k)
	return res.Value
}

func (client *MeshGoClient) Write(k string, v string) {
	//start := time.Now()
	depsStr, err := json.Marshal(client.Deps)
	CHECK(err)
	localStr, err := json.Marshal(client.Local)
	CHECK(err)
	fmt.Println("Deps", client.Deps)
	fmt.Println("deps", depsStr)
	fmt.Println("Local", client.Local)
	fmt.Println("local", localStr)
	res, err := (*client.Rpcc).ClientWrite(context.Background(), &ClientWriteRequest{Key: k, Value: v, Deps: string(depsStr), Local: string(localStr)})
	CHECK(err)
	var vc VC
	err = json.Unmarshal([]byte(res.Vc), &vc)
	CHECK(err)
	deps := make(map[string]VC)
	for k, vc := range client.Deps {
		deps[k] = vc
	}
	for k, m := range client.Local {
		deps[k] = m.Vc
	}
	InsertOrMergeMeta(&client.Local, k, &Meta{Key: k, Value: v, Vc: vc, Deps: deps})
	//fmt.Println("write", k, " time:", time.Since(start))
}

func (client *MeshGoClient) Execute() []byte {
	if client.Local == nil || client.Deps == nil {
		panic("client not init")
	}
	//fmt.Println(client.Workload)
	abort := false
	for _, op := range client.Workload {
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
					break
				}
				//fmt.Println("read time:", time.Since(start))
			case "W":
				//start := time.Now()
				vs := v.([]interface{})
				client.Write(vs[0].(string), vs[1].(string))
				//fmt.Println("write time:", time.Since(start))
			}
		}
	}
	client.Abort = abort
	return nil
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

func NewMeshGoClient() *MeshGoClient {
	var client MeshGoClient
	InitClient()

	if client.Rpcc == nil {
		fmt.Println("newmeshgoclient: new rpcc")
		client.Rpcc = RPCC
	} else {
		fmt.Println("newmeshgoclient: uu")
	}

	client.Workload = make([]map[string]interface{}, 0)
	client.Local = make(map[string]*Meta, 0)
	client.Deps = make(map[string]VC, 0)

	if client.Rpcc == nil {
		fmt.Println("UUUUUUU")
	}

	return &client
}

func Run(input []byte) []byte {
	var client MeshGoClient
	err := json.Unmarshal(input, &client)
	CHECK(err)
	//conn, err := grpc.Dial(ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//CHECK(err)
	//
	//rpcc := NewMeshClient(conn)
	InitClient()
	client.Rpcc = RPCC

	client.Execute()
	CHECK(err)
	clientStr, err := json.Marshal(client)
	CHECK(err)
	//fmt.Println(string(clientStr))
	return clientStr
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
