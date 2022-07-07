package push

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/ARGOeu/argo-messaging/config"
	amsPb "github.com/ARGOeu/argo-messaging/push/grpc/proto"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// GrpcClient is used to interface with ams push server
type GrpcClient struct {
	psc          amsPb.PushServiceClient
	hsc          grpc_health_v1.HealthClient
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
func (st *GrpcClientStatus) Result(details bool) string {

	grpcStatus := status.Convert(st.err)

	if grpcStatus.Code() == codes.OK {
		return st.message
	}

	if grpcStatus.Code() == codes.Unavailable {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "ams-push-server",
			},
		).Error(grpcStatus.Message())
		if details {
			return grpcStatus.Message()
		} else {
			return "Push server is currently unavailable"
		}
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

// Target returns the grpc endpoint that the client is connected to
func (c *GrpcClient) Target() string {
	return c.pushEndpoint
}

// Dial connects to the specified grpc endpoint from the api config
func (c *GrpcClient) Dial() error {

	conn, err := grpc.Dial(c.pushEndpoint, c.dialOptions...)
	if err != nil {
		return err
	}

	c.conn = conn

	c.psc = amsPb.NewPushServiceClient(conn)
	c.hsc = grpc_health_v1.NewHealthClient(conn)

	return nil
}

func (c *GrpcClient) SubscriptionStatus(ctx context.Context, fullSub string) ClientStatus {

	statusSubR := &amsPb.SubscriptionStatusRequest{
		FullName: fullSub,
	}

	r, err := c.psc.SubscriptionStatus(ctx, statusSubR)

	return &GrpcClientStatus{
		err:     err,
		message: r.GetStatus(),
	}
}

// ActivateSubscription is a wrapper over the grpc ActivateSubscription call
func (c *GrpcClient) ActivateSubscription(ctx context.Context, subscription subscriptions.Subscription) ClientStatus {

	var pushType amsPb.PushType
	if subscription.PushCfg.Type == subscriptions.HttpEndpointPushConfig {
		pushType = amsPb.PushType_HTTP_ENDPOINT
	} else {
		pushType = amsPb.PushType_MATTERMOST
	}

	actSubR := &amsPb.ActivateSubscriptionRequest{
		Subscription: &amsPb.Subscription{
			FullName:  subscription.FullName,
			FullTopic: subscription.FullTopic,
			PushConfig: &amsPb.PushConfig{
				Type:                pushType,
				PushEndpoint:        subscription.PushCfg.Pend,
				MaxMessages:         subscription.PushCfg.MaxMessages,
				AuthorizationHeader: subscription.PushCfg.AuthorizationHeader.Value,
				RetryPolicy: &amsPb.RetryPolicy{
					Type:   subscription.PushCfg.RetPol.PolicyType,
					Period: uint32(subscription.PushCfg.RetPol.Period),
				},
				MattermostUrl:      subscription.PushCfg.MattermostUrl,
				MattermostChannel:  subscription.PushCfg.MattermostChannel,
				MattermostUsername: subscription.PushCfg.MattermostUsername,
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

func (c *GrpcClient) HealthCheck(ctx context.Context) ClientStatus {

	r, err := c.hsc.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: ""},
	)

	if err != nil {
		_, err = c.psc.Status(ctx, &amsPb.StatusRequest{})
	}

	return &GrpcClientStatus{
		err:     err,
		message: r.GetStatus().String(),
	}
}

// Close terminates the underlying grpc connection
func (c *GrpcClient) Close() {
	c.conn.Close()
}
