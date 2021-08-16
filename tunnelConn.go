package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"net/http"
	"net/url"
	"strings"
)

// TunnelConnection represents the connection between your local
// machine and the IAP
type TunnelConnection struct {
	websocketConn *websocket.Conn
	bytesAcked    int
	connected     bool
	sid           string
	project       string
	zone          string
	instanceName  string
	port          string
	nic           string
}

const (
	tlsBaseUri       = "tunnel.cloudproxy.app"
	wssScheme        = "wss"
	mtlsScheme       = "mtls"
	webSocketVersion = "v4"
	connectEndpoint  = "connect"
	// Currently not used, I can check what's required to trigger this...
	mtlsBaseURi             = "mtls.tunnel.cloudproxy.app"
	subProtocolName         = "relay.tunnel.cloudproxy.app"
	origin                  = "bot:iap-tunneler"
	defaultNetworkInterface = "nic0"
)

// TunnelConnectionOption acts as a configuration wrapper to our tunnel connection
type TunnelConnectionOption func(connection *TunnelConnection)

// NewTunnelConnection creates a tunnel connection object
func NewTunnelConnection(ctx context.Context, opts ...TunnelConnectionOption) (*TunnelConnection, error) {
	tc := &TunnelConnection{}
	for _, opt := range opts {
		opt(tc)
	}
	if tc.nic == "" {
		tc.nic = defaultNetworkInterface
	}
	computeService, err := compute.NewService(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	instanceService := computeService.Instances
	instanceListCall := instanceService.List(tc.project, tc.zone)
	filters := []string{
		"status = RUNNING",
		fmt.Sprintf("name = %s", tc.instanceName),
	}
	instanceListCall.Filter(strings.Join(filters[:], " "))
	instanceList, err := instanceListCall.Do()
	if err != nil {
		return nil, err
	}
	// verify instance exists
	instanceVerify := false
	for _, instance := range instanceList.Items {
		nicVerify := false
		for _, nic := range instance.NetworkInterfaces {
			if nic.Name == tc.nic {
				nicVerify = true
				break
			}
		}
		if !nicVerify {
			break
		}
		if instance.Name == tc.instanceName {
			instanceVerify = true
			break
		}
	}
	if !instanceVerify {
		return nil, errors.New("failed to find instance")
	}
	// currently it doesn't give me an issue with scopes, in the future
	// I may want to be explicit
	scopes := []string{}
	cred, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return nil, err
	}
	ts, err := cred.TokenSource.Token()
	// may want to be more variable down the road, but for now this works
	u := url.URL{Scheme: wssScheme, Host: tlsBaseUri, Path: fmt.Sprintf("/%s/%s", webSocketVersion, connectEndpoint)}
	q := u.Query()
	//"project=%s&zone=%s&instance=%s&interface=%s&port=%s", project, zone, instance, interfaceName, port
	q.Add("project", tc.project)
	q.Add("zone", tc.zone)
	q.Add("instance", tc.instanceName)
	q.Add("interface", tc.nic)
	q.Add("port", tc.port)
	u.RawQuery = q.Encode()
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Origin":                 []string{origin},
		"Sec-Websocket-Protocol": []string{subProtocolName},
		"Authorization":          []string{fmt.Sprintf("Bearer %s", ts.AccessToken)},
	})
	tc.websocketConn = c
	return tc, nil
}

func WithProject(project string) TunnelConnectionOption {
	return func(tc *TunnelConnection) {
		tc.project = project
	}
}

func WithInstanceName(instanceName string) TunnelConnectionOption {
	return func(tc *TunnelConnection) {
		tc.instanceName = instanceName
	}
}

func WithZone(zone string) TunnelConnectionOption {
	return func(tc *TunnelConnection) {
		tc.zone = zone
	}
}

func WithPort(port string) TunnelConnectionOption {
	return func(tc *TunnelConnection) {
		tc.port = port
	}
}

func WithNic(nic string) TunnelConnectionOption {
	return func(tc *TunnelConnection) {
		tc.nic = nic
	}
}

// Run starts reading/writing from/to the connection
func (tc *TunnelConnection) Run() {

}
