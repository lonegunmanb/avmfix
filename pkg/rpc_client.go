package pkg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/matt-FFFFFF/tfpluginschema/tfplugin5"
	"github.com/matt-FFFFFF/tfpluginschema/tfplugin6"
	"google.golang.org/grpc"
)

const (
	// providerPluginName is the name used to identify the provider plugin
	providerPluginName = "provider"
	// magicCookieKey is the key used for the magic cookie in the plugin handshake
	magicCookieKey = "TF_PLUGIN_MAGIC_COOKIE"
	// magicCookieValue is the value used for the magic cookie in the plugin handshake
	magicCookieValue = "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2"
)

var (
	// ErrNotImplemented is returned when a method is not implemented
	ErrNotImplemented = errors.New("not implemented")
)

// providerGRPCPlugin implements the plugin.GRPCPlugin interface for connecting to provider binaries
type providerGRPCPlugin struct {
	plugin.Plugin
	protocolVersion int // 5 or 6
}

// GRPCClient returns the client implementation using the gRPC connection.
// Must be exported for the plugin framework to use it.
func (p providerGRPCPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	if p.protocolVersion == 5 {
		return &providerGRPCClientV5{
			providerGRPCClient: &providerGRPCClient[*tfplugin5.GetProviderSchema_Request, *tfplugin5.GetProviderSchema_Response]{
				grpcClient: v5SchemaClient{client: tfplugin5.NewProviderClient(c)},
			},
		}, nil
	}
	return &providerGRPCClientV6{
		providerGRPCClient: &providerGRPCClient[*tfplugin6.GetProviderSchema_Request, *tfplugin6.GetProviderSchema_Response]{
			grpcClient: v6SchemaClient{client: tfplugin6.NewProviderClient(c)},
		},
	}, nil
}

// GRPCServer is not implemented as we're only acting as a client
func (p providerGRPCPlugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return ErrNotImplemented
}

// schemaClient defines the interface for clients that can retrieve schemas.
// Required as the v5 and v6 clients have different method signatures.
type schemaClient[TReq, TResp any] interface {
	getSchema(ctx context.Context, req TReq, opts ...grpc.CallOption) (TResp, error)
}

// v5SchemaClient adapts tfplugin5.ProviderClient to the schemaClient interface.
type v5SchemaClient struct {
	client tfplugin5.ProviderClient
}

// getSchema calls GetSchema on the V5 client and implements the schemaClient interface.
func (c v5SchemaClient) getSchema(ctx context.Context, req *tfplugin5.GetProviderSchema_Request, opts ...grpc.CallOption) (*tfplugin5.GetProviderSchema_Response, error) {
	return c.client.GetSchema(ctx, req, opts...)
}

// v6SchemaClient adapts tfplugin6.ProviderClient to the schemaClient interface.
type v6SchemaClient struct {
	client tfplugin6.ProviderClient
}

// getSchema calls GetProviderSchema on the V6 client and implements the schemaClient interface.
func (c v6SchemaClient) getSchema(ctx context.Context, req *tfplugin6.GetProviderSchema_Request, opts ...grpc.CallOption) (*tfplugin6.GetProviderSchema_Response, error) {
	return c.client.GetProviderSchema(ctx, req, opts...)
}

// providerGRPCClient is a generic wrapper for gRPC clients
type providerGRPCClient[TReq, TResp any] struct {
	grpcClient schemaClient[TReq, TResp]
}

// Schema calls GetSchema on the provider and returns the protobuf response
func (c *providerGRPCClient[TReq, TResp]) Schema(req TReq) (TResp, error) {
	var zeroResp TResp
	protoResp, err := c.grpcClient.getSchema(context.Background(), req)
	if err != nil {
		return zeroResp, fmt.Errorf("failed to get provider schema: %w", err)
	}
	return protoResp, nil
}

// providerGRPCClientV5 wraps the gRPC client for protocol v5
type providerGRPCClientV5 struct {
	*providerGRPCClient[*tfplugin5.GetProviderSchema_Request, *tfplugin5.GetProviderSchema_Response]
}

// v5Schema calls GetSchema on the provider and returns the protobuf response
func (c *providerGRPCClientV5) v5Schema() (*tfplugin5.GetProviderSchema_Response, error) {
	protoReq := &tfplugin5.GetProviderSchema_Request{} // Empty request
	return c.Schema(protoReq)
}

// providerGRPCClientV6 wraps the gRPC client for protocol v6
type providerGRPCClientV6 struct {
	*providerGRPCClient[*tfplugin6.GetProviderSchema_Request, *tfplugin6.GetProviderSchema_Response]
}

// v6Schema calls GetProviderSchema on the provider and returns the protobuf response
func (c *providerGRPCClientV6) v6Schema() (*tfplugin6.GetProviderSchema_Response, error) {
	protoReq := &tfplugin6.GetProviderSchema_Request{} // Empty request
	return c.Schema(protoReq)
}

// universalProvider provides a unified interface that works with both V5 and V6 protocols
type universalProvider interface {
	v5Schema() (*tfplugin5.GetProviderSchema_Response, error)
	v6Schema() (*tfplugin6.GetProviderSchema_Response, error)
	close()
}

// newGrpcClient creates a provider client that supports both V5 and V6 protocols.
func newGrpcClient(providerPath string) (universalProvider, error) {
	// No need for ProtocolVersion here as we are using VersionedPlugins
	handshakeConfig := plugin.HandshakeConfig{
		MagicCookieKey:   magicCookieKey,
		MagicCookieValue: magicCookieValue,
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		VersionedPlugins: map[int]plugin.PluginSet{
			5: {providerPluginName: providerGRPCPlugin{protocolVersion: 5}},
			6: {providerPluginName: providerGRPCPlugin{protocolVersion: 6}},
		},
		Cmd:              exec.Command(providerPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           hclog.New(&hclog.LoggerOptions{Level: hclog.Error}),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(providerPluginName)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense provider: %w", err)
	}

	// The plugin framework will return either a V5 or V6 client based on negotiation
	// We need to wrap it in a universal client that supports both interfaces
	if v5Client, ok := raw.(*providerGRPCClientV5); ok {
		return &universalProviderClient{
			v5:        v5Client,
			closeFunc: client.Kill,
		}, nil
	}
	if v6Client, ok := raw.(*providerGRPCClientV6); ok {
		return &universalProviderClient{
			v6:        v6Client,
			closeFunc: client.Kill,
		}, nil
	}

	client.Kill()
	return nil, fmt.Errorf("plugin returned unexpected type: %T", raw)
}

// universalProviderClient implements UniversalProvider and wraps either V5 or V6 clients
type universalProviderClient struct {
	v5        *providerGRPCClientV5
	v6        *providerGRPCClientV6
	closeFunc func()
}

func (c *universalProviderClient) v5Schema() (*tfplugin5.GetProviderSchema_Response, error) {
	if c.v5 != nil {
		return c.v5.v5Schema()
	}
	return nil, fmt.Errorf("V5 protocol not supported by this provider")
}

func (c *universalProviderClient) v6Schema() (*tfplugin6.GetProviderSchema_Response, error) {
	if c.v6 != nil {
		return c.v6.v6Schema()
	}
	return nil, fmt.Errorf("V6 protocol not supported by this provider")
}

func (c *universalProviderClient) close() {
	if c.closeFunc != nil {
		c.closeFunc()
	}
	c.v5 = nil
	c.v6 = nil
}
