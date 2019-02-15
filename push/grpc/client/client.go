package push

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/ARGOeu/argo-messaging/config"
	amsPb "github.com/ARGOeu/argo-messaging/push/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// GrpcClient is used to interface with ams push server
type GrpcClient struct {
	psc          amsPb.PushServiceClient
	dialOptions  []grpc.DialOption
	conn         *grpc.ClientConn
	pushEndpoint string
}

// GrpcClientStatus holds the outcome of a grpc request
type GrpcClientStatus struct {
	err     error
	message string
}

// Result prints the result of an grpc request
func (st *GrpcClientStatus) Result() string {

	grpcStatus := status.Convert(st.err)

	if grpcStatus.Code() == codes.OK {
		return fmt.Sprintf("Success: %v", st.message)
	}

	return fmt.Sprintf("Error: %v", grpcStatus.Message())
}

// NewGrpcClient returns a new client configured based on the provided api cfg
func NewGrpcClient(cfg *config.APICfg) *GrpcClient {

	client := new(GrpcClient)

	client.pushEndpoint = fmt.Sprintf("%v:%v", cfg.PushServerHost, cfg.PushServerPort)

	if cfg.PushTlsEnabled {

		cert, _ := tls.LoadX509KeyPair(cfg.Cert, cfg.CertKey)

		tlsConfig := &tls.Config{
			ServerName:         cfg.PushServerHost,
			Certificates:       []tls.Certificate{cert},
			RootCAs:            cfg.LoadCAs(),
			InsecureSkipVerify: !cfg.VerifyPushServer,
		}

		client.dialOptions = append(client.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

	} else {

		client.dialOptions = append(client.dialOptions, grpc.WithInsecure())
	}

	return client
}

// Dial connects to the specified grpc endpoint from the api config
func (c *GrpcClient) Dial() error {

	conn, err := grpc.Dial(c.pushEndpoint, c.dialOptions...)
	if err != nil {
		return err
	}

	c.conn = conn

	c.psc = amsPb.NewPushServiceClient(conn)

	return nil
}

// ActivateSubscription is a wrapper over the grpc ActivateSubscription call
func (c *GrpcClient) ActivateSubscription(ctx context.Context, fullSub, fullTopic, pushEndpoint, retryType string, retryPeriod uint32) ClientStatus {

	actSubR := &amsPb.ActivateSubscriptionRequest{
		Subscription: &amsPb.Subscription{
			FullName:  fullSub,
			FullTopic: fullTopic,
			PushConfig: &amsPb.PushConfig{
				PushEndpoint: pushEndpoint,
				RetryPolicy: &amsPb.RetryPolicy{
					Type:   retryType,
					Period: retryPeriod,
				},
			},
		}}

	r, err := c.psc.ActivateSubscription(ctx, actSubR)

	return &GrpcClientStatus{
		err:     err,
		message: r.GetMessage(),
	}

}

// DeactivateSubscription is a wrapper over the grpc DeactivateSubscription call
func (c *GrpcClient) DeactivateSubscription(ctx context.Context, fullSub string) ClientStatus {

	deActSubR := &amsPb.DeactivateSubscriptionRequest{
		FullName: fullSub,
	}

	r, err := c.psc.DeactivateSubscription(ctx, deActSubR)

	return &GrpcClientStatus{
		err:     err,
		message: r.GetMessage(),
	}
}

// Close terminates the underlying grpc connection
func (c *GrpcClient) Close() {
	c.conn.Close()
}
