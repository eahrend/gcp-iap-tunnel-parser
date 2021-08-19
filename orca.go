package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// Orca handles the communication between the
// local port and the proxy port
type Orca struct {
	tunnelConn *TunnelConnection
	localConn  *LocalConn
}

func NewOrca(ctx context.Context) (*Orca, error) {
	project := os.Getenv("PROJECT_ID")
	zone := os.Getenv("ZONE")
	instance := os.Getenv("INSTANCE")
	port := os.Getenv("PORT")
	localPort := os.Getenv("LOCAL_PORT")
	tcPipeReader, tcPipeWriter := io.Pipe()
	tc, err := NewTunnelConnection(ctx,
		WithProject(project),
		WithZone(zone),
		WithPort(port),
		WithInstanceName(instance),
		WithTunnelReader(tcPipeReader),
		WithTunnelWriter(tcPipeWriter))
	if err != nil {
		return nil, err
	}
	lcPipeReader, lcPipeWriter := io.Pipe()
	lc, err := NewLocalConn(ctx,
		WithLocalConnPort(localPort),
		WithLocalConnReader(lcPipeReader),
		WithLocalConnWriter(lcPipeWriter))
	if err != nil {
		return nil, err
	}
	orca := &Orca{tunnelConn: tc, localConn: lc}
	return orca, nil
}

func (orca *Orca) Run() error {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// going off of max size from the python library
	localbuf := make([]byte, 16384)
	socketbuf := make([]byte, 16384)
	fmt.Println("Listening on port 4000")
	err := orca.localConn.Accept()
	if err != nil {
		return err
	}
	fmt.Println("Client connected")
	go func() {
		for {
			n, err := orca.localConn.Read(localbuf)
			if err != nil {
				panic(err)
			}
			msg := localbuf[:n]
			fmt.Println("Sending data back", msg)
			newMsg := NewIAPDataMessage(msg)
			err = newMsg.CreateDataFrame()
			if err != nil {
				panic(err)
			}
			fmt.Println("Modified Data:", newMsg.data)
			_, err = orca.tunnelConn.Write(newMsg.data)
			if err != nil {
				panic(err)
			}
		}
	}()
	go func() {
		for {
			n, err := orca.tunnelConn.Read(socketbuf)
			if err != nil {
				panic(err)
			}
			d := socketbuf[:n]
			msg := NewIAPMessage(d)
			tag := msg.PeekMessageTag()
			switch tag {
			case MessageAck:
				fmt.Println("Got Message Ack From IAP")
				continue
			case MessageConnectSuccessSid:
				fmt.Println("Got Success SID")
				newMsg := msg.AsConnectSIDMessage()
				orca.tunnelConn.SetSid(newMsg.GetSID())
				continue
			case MessageData:
				fmt.Println("Got Data Message")
				newMsg := msg.AsDataMessage()
				orca.tunnelConn.bytesReceived += newMsg.GetDataLength()
				ackBytes := make([]byte, 10)
				ackMsg := NewIAPAckMessage(ackBytes)
				ackMsg.SetTag(MessageAck)
				ackMsg.SetAck(uint64(orca.tunnelConn.bytesReceived))
				err = orca.tunnelConn.websocketConn.WriteMessage(websocket.BinaryMessage, ackMsg.data)
				if err != nil {
					panic(err)
				}
				orca.tunnelConn.bytesAcked += orca.tunnelConn.bytesReceived
				_, err = orca.localConn.Write(newMsg.data)
				if err != nil {
					panic(err)
				}
				continue
			default:
				fmt.Printf("Unknown tag: %d", tag)
				break
			}
		}
	}()
	<-c
	return nil
}
