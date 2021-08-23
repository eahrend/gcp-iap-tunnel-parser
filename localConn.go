package main

import (
	"context"
	"fmt"
	"io"
	"net"
)

// LocalConn represents a local tcp connection
type LocalConn struct {
	conn          net.Conn
	localListener net.Listener
	port          string
	bytesReceived uint32
	bytesSent     uint32
	reader        io.Reader
	writer        io.Writer
}

// LocalConnOption is the configuration option for LocalConn
// used in the constructor.
type LocalConnOption func(conn *LocalConn)

func NewLocalConn(ctx context.Context, opts ...LocalConnOption) (*LocalConn, error) {
	lc := &LocalConn{}
	for _, opt := range opts {
		opt(lc)
	}
	// for testing
	// TODO: Add local port checking so you don't need to be explicit about
	//  what port you're looking for, also _maybe_ add ipv6?
	localListener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", lc.port))
	if err != nil {
		return nil, err
	}
	lc.localListener = localListener
	return lc, nil
}

func WithLocalConnPort(port string) LocalConnOption {
	return func(conn *LocalConn) {
		conn.port = port
	}
}

func WithLocalConnWriter(writer io.Writer) LocalConnOption {
	return func(conn *LocalConn) {
		conn.writer = writer
	}
}

func WithLocalConnReader(reader io.Reader) LocalConnOption {
	return func(conn *LocalConn) {
		conn.reader = reader
	}
}

// Accept is blocking, only start accepting once we can confirm that
// the websocket connection is valid
func (lc *LocalConn) Accept() error {
	c, err := lc.localListener.Accept()
	lc.conn = c
	return err
}

func (lc *LocalConn) Read(buf []byte) (n int, err error) {
	return lc.conn.Read(buf)
}

func (lc *LocalConn) Write(buf []byte) (n int, err error) {
	return lc.conn.Write(buf)
}
