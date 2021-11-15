package client

import (
	"bufio"
	"client/pkg/handler"
	"net"
)

type Client struct {
	connection net.Conn
}

func (c *Client) Run(network string, remoteAddr string, handler *handler.Handler) error {
	var err error
	c.connection, err = net.Dial(network, remoteAddr)
	if err != nil {
		return err
	}
	handler.Handle(bufio.NewReader(c.connection), bufio.NewWriter(c.connection))
	return nil
}

func (c *Client) Shutdown() error {
	return c.connection.Close()
}
