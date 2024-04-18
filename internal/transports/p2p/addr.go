package p2pexp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

func parseAddress(addr, port string) (*net.TCPAddr, error) {
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("could not parse port: %w", err)
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, errors.New("could not parse peer IP")
	}

	netAddr := &net.TCPAddr{
		IP:   ip,
		Port: portInt,
	}

	return netAddr, nil
}
