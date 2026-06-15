package common

import (
	"bytes"
	"encoding/gob"
	"net"
)

func DeepCopy(dist, src interface{}) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(src)
	CHECK(err)
	err = gob.NewDecoder(&buf).Decode(dist)
	CHECK(err)
}

func CHECK(err error) {
	if err != nil {
		panic(err)
	}
}

func GetLocalIPv4() ([]net.IP, error) {
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
