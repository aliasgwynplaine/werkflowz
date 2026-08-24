package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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

func Hash2N(a string, b string, n int) int {
	if n <= 0 {
		panic("fuck you")
	}

	h := sha256.New()

	h.Write([]byte(a))
	h.Write([]byte{0})
	h.Write([]byte(b))

	sum := h.Sum(nil)

	x := binary.BigEndian.Uint64(sum[:8])

	return int(x % uint64(n))
}

func GetHost(a string, b string) string {
	idx := Hash2N(a, b, T)

	return PEERS[idx]
}
