package client

import (
	"context"

	"github.com/netbirdio/netbird/client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	daemon proto.DaemonServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		daemon: proto.NewDaemonServiceClient(conn),
	}, nil
}

func (c *Client) Status(ctx context.Context) (*proto.StatusResponse, error) {
	return c.daemon.Status(ctx, &proto.StatusRequest{GetFullPeerStatus: true})
}

func (c *Client) Up(ctx context.Context) error {
	_, err := c.daemon.Up(ctx, &proto.UpRequest{})
	return err
}

func (c *Client) Down(ctx context.Context) error {
	_, err := c.daemon.Down(ctx, &proto.DownRequest{})
	return err
}

func (c *Client) ListNetworks(ctx context.Context) ([]*proto.Network, error) {
	resp, err := c.daemon.ListNetworks(ctx, &proto.ListNetworksRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Routes, nil
}

func (c *Client) SelectNetworks(ctx context.Context, ids []string) error {
	_, err := c.daemon.SelectNetworks(ctx, &proto.SelectNetworksRequest{
		NetworkIDs: ids,
		Append:     false,
	})
	return err
}

func (c *Client) DeselectNetworks(ctx context.Context, ids []string) error {
	_, err := c.daemon.DeselectNetworks(ctx, &proto.SelectNetworksRequest{
		NetworkIDs: ids,
	})
	return err
}

func (c *Client) ForwardingRules(ctx context.Context) ([]*proto.ForwardingRule, error) {
	resp, err := c.daemon.ForwardingRules(ctx, &proto.EmptyRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Rules, nil
}

func (c *Client) DebugBundle(ctx context.Context) (string, error) {
	resp, err := c.daemon.DebugBundle(ctx, &proto.DebugBundleRequest{})
	if err != nil {
		return "", err
	}
	return resp.Path, nil
}

func (c *Client) Logout(ctx context.Context) error {
	_, err := c.daemon.Logout(ctx, &proto.LogoutRequest{})
	return err
}

func (c *Client) GetConfig(ctx context.Context) (*proto.GetConfigResponse, error) {
	return c.daemon.GetConfig(ctx, &proto.GetConfigRequest{})
}

func (c *Client) Login(ctx context.Context, setupKey, managementURL string) error {
	_, err := c.daemon.Login(ctx, &proto.LoginRequest{
		SetupKey:      setupKey,
		ManagementUrl: managementURL,
	})
	return err
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
