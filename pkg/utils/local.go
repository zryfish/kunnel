package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Local struct {
	LocalHost string
	LocalPort int
}

var ErrInvalidLocal = errors.New("invalid local")

func NewLocal(s string) (*Local, error) {
	parts := strings.Split(s, ":")
	if len(parts) == 0 || len(parts) > 2 {
		return nil, ErrInvalidLocal
	}

	local := &Local{
		LocalHost: "127.0.0.1",
	}

	var host, port string

	if len(parts) == 1 {
		port = parts[0]
	} else if len(parts) == 2 {
		host, port = parts[0], parts[1]
	}

	if len(host) != 0 {
		local.LocalHost = host
	}

	portNumber, ok := isValidPort(port)
	if !ok {
		return nil, ErrInvalidLocal
	}
	local.LocalPort = portNumber

	return local, nil
}

func isValidPort(s string) (int, bool) {
	port, err := strconv.Atoi(s)
	if err != nil || (port <= 0 || port >= 65535) {
		return 0, false
	}
	return port, true
}

func (l *Local) String() string {
	return fmt.Sprintf("%s:%d", l.LocalHost, l.LocalPort)
}
