package main

import (
	"context"
	"io"
	"os"
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
		WithLocalConnPort("4000"),
		WithLocalConnReader(lcPipeReader),
		WithLocalConnWriter(lcPipeWriter))
	if err != nil {
		return nil, err
	}
	orca := &Orca{tunnelConn: tc, localConn: lc}
	return orca, nil
}

// Testing reading/writing using pipe
func (orca *Orca) Run() {
	buf := make([]byte, 10000)
	orca.localConn.Write([]byte("Starting Local Connection"))
	for {
		n, err := orca.localConn.Read(buf)
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(buf[:n])
	}
}
