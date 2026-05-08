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

func (c *Client) DebugBundle(ctx context.Context, anonymize, systemInfo bool) (string, error) {
	resp, err := c.daemon.DebugBundle(ctx, &proto.DebugBundleRequest{
		Anonymize:  anonymize,
		SystemInfo: systemInfo,
	})
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

// GetLogLevel returns the current daemon log level.
func (c *Client) GetLogLevel(ctx context.Context) (proto.LogLevel, error) {
	resp, err := c.daemon.GetLogLevel(ctx, &proto.GetLogLevelRequest{})
	if err != nil {
		return proto.LogLevel_UNKNOWN, err
	}
	return resp.Level, nil
}

// SetLogLevel sets the daemon log level.
func (c *Client) SetLogLevel(ctx context.Context, level proto.LogLevel) error {
	_, err := c.daemon.SetLogLevel(ctx, &proto.SetLogLevelRequest{Level: level})
	return err
}

// ListStates returns internal daemon states.
func (c *Client) ListStates(ctx context.Context) ([]*proto.State, error) {
	resp, err := c.daemon.ListStates(ctx, &proto.ListStatesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.States, nil
}

// CleanState clears a specific state.
func (c *Client) CleanState(ctx context.Context, stateName string) error {
	_, err := c.daemon.CleanState(ctx, &proto.CleanStateRequest{StateName: stateName})
	return err
}

// DeleteState deletes a specific state.
func (c *Client) DeleteState(ctx context.Context, stateName string) error {
	_, err := c.daemon.DeleteState(ctx, &proto.DeleteStateRequest{StateName: stateName})
	return err
}

// TracePacket traces a packet through the firewall rules.
func (c *Client) TracePacket(ctx context.Context, req *proto.TracePacketRequest) (*proto.TracePacketResponse, error) {
	return c.daemon.TracePacket(ctx, req)
}

// GetEvents retrieves system events.
func (c *Client) GetEvents(ctx context.Context) ([]*proto.SystemEvent, error) {
	resp, err := c.daemon.GetEvents(ctx, &proto.GetEventsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Events, nil
}

// ListProfiles returns all configured profiles.
func (c *Client) ListProfiles(ctx context.Context) ([]*proto.Profile, error) {
	resp, err := c.daemon.ListProfiles(ctx, &proto.ListProfilesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Profiles, nil
}

// GetActiveProfile returns the name of the currently active profile.
func (c *Client) GetActiveProfile(ctx context.Context) (string, error) {
	resp, err := c.daemon.GetActiveProfile(ctx, &proto.GetActiveProfileRequest{})
	if err != nil {
		return "", err
	}
	return resp.ProfileName, nil
}

// SwitchProfile switches to the given profile name.
func (c *Client) SwitchProfile(ctx context.Context, name string) error {
	_, err := c.daemon.SwitchProfile(ctx, &proto.SwitchProfileRequest{ProfileName: &name})
	return err
}

// AddProfile creates a new profile.
func (c *Client) AddProfile(ctx context.Context, name string) error {
	_, err := c.daemon.AddProfile(ctx, &proto.AddProfileRequest{
		ProfileName: name,
	})
	return err
}

// RemoveProfile deletes a profile.
func (c *Client) RemoveProfile(ctx context.Context, name string) error {
	_, err := c.daemon.RemoveProfile(ctx, &proto.RemoveProfileRequest{ProfileName: name})
	return err
}

// GetFeatures returns feature flags from the management server.
func (c *Client) GetFeatures(ctx context.Context) (*proto.GetFeaturesResponse, error) {
	return c.daemon.GetFeatures(ctx, &proto.GetFeaturesRequest{})
}

// SetConfig applies new configuration settings.
func (c *Client) SetConfig(ctx context.Context, req *proto.SetConfigRequest) error {
	_, err := c.daemon.SetConfig(ctx, req)
	return err
}

// GetPeerSSHHostKey retrieves the SSH host key for a peer by address.
func (c *Client) GetPeerSSHHostKey(ctx context.Context, peerAddress string) (string, error) {
	resp, err := c.daemon.GetPeerSSHHostKey(ctx, &proto.GetPeerSSHHostKeyRequest{PeerAddress: peerAddress})
	if err != nil {
		return "", err
	}
	return string(resp.SshHostKey), nil
}

// RequestJWTAuth initiates browser-based SSO login flow.
func (c *Client) RequestJWTAuth(ctx context.Context) (*proto.RequestJWTAuthResponse, error) {
	return c.daemon.RequestJWTAuth(ctx, &proto.RequestJWTAuthRequest{})
}

// WaitJWTToken waits for the SSO login to complete.
func (c *Client) WaitJWTToken(ctx context.Context, deviceCode string) error {
	_, err := c.daemon.WaitJWTToken(ctx, &proto.WaitJWTTokenRequest{
		DeviceCode: deviceCode,
	})
	return err
}

// SetSyncResponsePersistence enables/disables sync response persistence (debug).
func (c *Client) SetSyncResponsePersistence(ctx context.Context, enable bool) error {
	_, err := c.daemon.SetSyncResponsePersistence(ctx, &proto.SetSyncResponsePersistenceRequest{
		Enabled: enable,
	})
	return err
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
