package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/oauth2/google"
)

func connectToWSS() (*websocket.Conn, error) {
	ctx := context.Background()
	scopes := []string{}
	project := os.Getenv("PROJECT_ID")
	zone := os.Getenv("ZONE")
	instance := os.Getenv("INSTANCE")
	interfaceName := os.Getenv("INTERFACE")
	port := os.Getenv("PORT")
	queryData := fmt.Sprintf("project=%s&zone=%s&instance=%s&interface=%s&port=%s", project, zone, instance, interfaceName, port)
	cred, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		panic(err)
	}
	ts, err := cred.TokenSource.Token()
	if err != nil {
		panic(err)
	}
	u := url.URL{Scheme: "wss", Host: "tunnel.cloudproxy.app", Path: "/v4/connect", RawQuery: queryData}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Origin":        []string{"bot:iap-tunneler"},
		"Authorization": []string{fmt.Sprintf("Bearer %s", ts.AccessToken)},
	})
	return c, err
}

func main() {
	writeToPortChan := make(chan []byte)
	writeToSockChan := make(chan []byte)
	var bytesReceived int
	defer close(writeToPortChan)
	defer close(writeToSockChan)
	localListener, err := net.Listen("tcp4", "127.0.0.1:4000")
	if err != nil {
		panic(err)
	}
	defer localListener.Close()
	c, err := connectToWSS()
	if err != nil {
		panic(err)
	}

	defer c.Close()
	go func() {
		localConn, err := localListener.Accept()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, 16384)
		defer localConn.Close()
		for {
			select {
			case msg := <-writeToPortChan:
				fmt.Println("received message: ", string(msg))
				_, err := localConn.Write(msg)
				if err != nil {
					panic(err)
				}
			default:
				n, err := localConn.Read(buf)
				if err != nil {
					panic(err)
				}
				if err == io.EOF {
					fmt.Println("connection closed")
				} else if err != nil {
					panic(err)
				}
				fmt.Println("sending data to socket:", string(buf[:n]))
				dataToSend := createSubprotocolDataFrame(buf[:n])
				if len(dataToSend) == 0 {
					continue
				}
				fmt.Println(dataToSend)
				writeToSockChan <- dataToSend
				fmt.Println("sent data to socket")
			}
		}
	}()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case msg := <-writeToSockChan:
				fmt.Println("sending message to socket:", string(msg))
				err := c.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					panic(err)
				}
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				tag, data, err := extractSubProtocolTag(message)
				fmt.Println(data)
				if err != nil {
					panic(err)
				}
				if tag == 4 {
					msg, _, err := handleSubprotocolData(data)
					if err != nil {
						panic(err)
					}
					bytesReceived += len(msg)
					ackData := sendAck(bytesReceived)
					fmt.Println(ackData)
					err = c.WriteMessage(websocket.BinaryMessage, ackData)
					if err != nil {
						panic(err)
					}
					writeToPortChan <- msg
					fmt.Println("sent message to channel:", msg)
				} else if tag == 7 {
					// currently I'm not hitting a 7 from the server, so I'm killing everything until I get there
					panic("got ack message from server")
				} else if tag == 1 {
					_, _, err := handleSubprotocolConnectSuccessSid(data)
					if err != nil {
						panic(err)
					}
				} else if tag == 2 {
					// _HandleSubprotocolReconnectSuccessAck
				} else {
					panic(fmt.Errorf("unknown tag: %v", tag))
				}
			}
		}
	}()

	// templated code from gorilla wss example to let me cleanly exit
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println(t)
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
