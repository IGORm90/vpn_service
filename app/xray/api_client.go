package xray

import (
	"context"
	"fmt"
	"time"
	"vpn-service/database"

	handlerService "github.com/xtls/xray-core/app/proxyman/command"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/common/uuid"
	"github.com/xtls/xray-core/proxy/vless"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// APIClient provides access to Xray HandlerService.
type APIClient struct {
	address    string
	inboundTag string
	timeout    time.Duration
}

// NewAPIClient creates a new API client for Xray.
func NewAPIClient(address, inboundTag string, timeout time.Duration) *APIClient {
	return &APIClient{
		address:    address,
		inboundTag: inboundTag,
		timeout:    timeout,
	}
}

// AddUser adds a single user to the inbound via API.
func (c *APIClient) AddUser(user *database.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	protoUser, err := buildVlessProtocolUser(user)
	if err != nil {
		return err
	}

	return c.alterInbound(func(ctx context.Context, client handlerService.HandlerServiceClient) error {
		_, err := client.AlterInbound(ctx, &handlerService.AlterInboundRequest{
			Tag: c.inboundTag,
			Operation: serial.ToTypedMessage(&handlerService.AddUserOperation{
				User: protoUser,
			}),
		})
		return err
	})
}

// RemoveUser removes a single user from the inbound via API.
func (c *APIClient) RemoveUser(user *database.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	return c.alterInbound(func(ctx context.Context, client handlerService.HandlerServiceClient) error {
		_, err := client.AlterInbound(ctx, &handlerService.AlterInboundRequest{
			Tag: c.inboundTag,
			Operation: serial.ToTypedMessage(&handlerService.RemoveUserOperation{
				Email: user.Username,
			}),
		})
		return err
	})
}

func (c *APIClient) alterInbound(
	action func(ctx context.Context, client handlerService.HandlerServiceClient) error,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial xray api %s: %w", c.address, err)
	}
	defer conn.Close()

	client := handlerService.NewHandlerServiceClient(conn)
	if err := action(ctx, client); err != nil {
		return fmt.Errorf("xray api alter inbound failed: %w", err)
	}
	return nil
}

func buildVlessProtocolUser(user *database.User) (*protocol.User, error) {
	parsedUUID, err := uuid.ParseString(user.UUID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	account := &vless.Account{
		Id:   parsedUUID.String(),
		Flow: "",
	}

	return &protocol.User{
		Email:   user.Username,
		Level:   0,
		Account: serial.ToTypedMessage(account),
	}, nil
}
