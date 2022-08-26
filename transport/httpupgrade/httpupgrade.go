package httpupgrade

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strings"
)

type Config struct {
	Path string
	Host string
}

func NewHTTPUpgradeTransport(config *Config) (Transport, error) {
	return Transport{path: config.Path, host: config.Host}, nil
}

type Transport struct {
	path string
	host string
}

func (t Transport) Client(conn net.Conn) (net.Conn, error) {
	req, err := http.NewRequest("GET", "/"+t.path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Host = t.host

	err = req.Write(conn)
	if err != nil {
		return nil, err
	}

	//TODO The bufio usage here is unreliable
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	if resp.Status == "101 Switching Protocols" &&
		strings.ToLower(resp.Header.Get("Upgrade")) == "websocket" &&
		strings.ToLower(resp.Header.Get("Connection")) == "upgrade" {
		return conn, nil
	}
	return nil, errors.New("unrecognized reply")
}

func (t Transport) Server(conn net.Conn) (net.Conn, error) {
	connReader := bufio.NewReader(conn)
	req, err := http.ReadRequest(connReader)
	if err != nil {
		return nil, err
	}
	connection := strings.ToLower(req.Header.Get("Connection"))
	upgrade := strings.ToLower(req.Header.Get("Upgrade"))
	if connection != "upgrade" || upgrade != "websocket" {
		return nil, err
	}
	resp := &http.Response{
		Status:     "101 Switching Protocols",
		StatusCode: 101,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	resp.Header.Set("Connection", "upgrade")
	resp.Header.Set("Upgrade", "websocket")
	err = resp.Write(conn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
