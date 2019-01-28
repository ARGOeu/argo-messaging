package push

import (
	"context"
)

// Client help us interface with any push backend mechanism
type Client interface {
	// Dial establishes a connection with the push backend
	Dial() error
	// ActivateSubscription provides the push backend
	// with all the necessary information to start the push functionality for the respective subscription
	ActivateSubscription(ctx context.Context, fullSub, fullTopic, pushEndpoint, retryType string, retryPeriod uint32) ClientStatus
	// DeactivateSubscription asks the push backend to stop the push functionality for the respective subscription
	DeactivateSubscription(ctx context.Context, fullSub string) ClientStatus
	// HealthCheck performs the grpc health check call
	HealthCheck(ctx context.Context) *GrpcClientStatus
	// Target returns the endpoint the client has been connected to
	Target() string
	// Close closes the connection with the push backend
	Close()
}

// ClientStatus represents responses from a push backend
type ClientStatus interface {
	// Result returns the string representation for the response from a push backend
	Result() string
}
