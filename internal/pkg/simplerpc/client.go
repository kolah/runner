package simplerpc

import (
	"bufio"
	"errors"
	"fmt"
	"net"
)

type Client struct {
	host string
	port int
	conn net.Conn
}

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return err
	}
	c.conn = conn

	return nil
}

func (c *Client) readLn() (string, error) {
	if c.conn == nil {
		return "", errors.New("not connected")
	}
	b := bufio.NewReader(c.conn)

	line, err := b.ReadBytes('\n')

	if err != nil {
		return "", err
	}

	return string(line[0 : len(line)-1]), nil
}

func (c *Client) SendCommand(command string) (response string, error error) {
	_, err := fmt.Fprintln(c.conn, string([]byte(command)))

	if err != nil {
		return "", err
	}

	response, err = c.readLn()

	return response, err
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
