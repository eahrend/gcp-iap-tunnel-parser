package main

import "context"

// Orca handles the communication between the
// local port and the proxy port
type Orca struct {
	tunnelConn *TunnelConnection
}

func NewOrca(ctx context.Context, tcopts ...TunnelConnectionOption) (*Orca, error) {
	tc, err := NewTunnelConnection(ctx, tcopts...)
	if err != nil {
		return nil, err
	}
	orca := &Orca{tunnelConn: tc}
	return orca, nil
}

func (orca *Orca) Run() {
	orca.tunnelConn.Run()
}
