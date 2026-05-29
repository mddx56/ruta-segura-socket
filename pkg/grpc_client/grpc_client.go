package grpc_client

import (
	"context"
	"time"

	device_proto "github.com/waltherx/motos-socket/pkg/pb/device_proto"
	position_proto "github.com/waltherx/motos-socket/pkg/pb/position_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultTimeout = 5 * time.Second

type GRPCClient struct {
	conn           *grpc.ClientConn
	DeviceClient   device_proto.DeviceServiceClient
	PositionClient position_proto.PositionServiceClient
}

// New abre la conexion gRPC (sin TLS). Usar defer client.Close().
func New(target string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		conn:           conn,
		DeviceClient:   device_proto.NewDeviceServiceClient(conn),
		PositionClient: position_proto.NewPositionServiceClient(conn),
	}, nil
}

func (c *GRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *GRPCClient) ListDevices() (*device_proto.ListDevicesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.DeviceClient.ListDevicesSimple(ctx, &device_proto.ListDevicesRequest{})
}

func (c *GRPCClient) SavePosition(
	imei string, deviceTime int64,
	lat, lng float64,
	speed, course int32,
	attributes string,
) (*position_proto.SavePositionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.PositionClient.SavePosition(ctx, &position_proto.SavePositionRequest{
		Imei:       imei,
		DeviceTime: deviceTime,
		Latitude:   lat,
		Longitude:  lng,
		Speed:      speed,
		Course:     course,
		Attributes: attributes,
	})
}
