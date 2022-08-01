package push

import (
	"context"
	"fmt"
	"github.com/ARGOeu/argo-messaging/subscriptions"
)

type MockClient struct{}

func (*MockClient) SubscriptionStatus(ctx context.Context, fullSub string) ClientStatus {

	return &MockClientStatus{
		Status: fmt.Sprintf("Subscription %v is currently active", fullSub),
	}
}

func (*MockClient) HealthCheck(ctx context.Context) ClientStatus {
	return &MockClientStatus{
		Status: "SERVING",
	}
}

func (*MockClient) Target() string {
	return "localhost:5555"
}

func (*MockClient) Dial() error { return nil }

func (*MockClient) ActivateSubscription(ctx context.Context, subscription subscriptions.Subscription) ClientStatus {

	switch subscription.FullName {
	case "/projects/ARGO/subscriptions/subNew":

		return &MockClientStatus{
			Status: fmt.Sprintf("Subscription %v activated", subscription.FullName),
		}

	case "/projects/ARGO/subscriptions/sub1":

		return &GrpcClientStatus{
			err:     nil,
			message: fmt.Sprintf("Subscription %v activated", subscription.FullName),
		}

	case "/projects/ARGO/subscriptions/sub4":

		return &GrpcClientStatus{
			err:     nil,
			message: fmt.Sprintf("Subscription %v activated", subscription.FullName),
		}
	}

	return &MockClientStatus{
		Status: fmt.Sprintf("Subscription %v is already active", subscription.FullName),
	}

}

func (*MockClient) DeactivateSubscription(ctx context.Context, fullSub string) ClientStatus {

	switch fullSub {
	case "/projects/ARGO/subscriptions/sub4":

		return &MockClientStatus{
			Status: fmt.Sprintf("Subscription %v deactivated", fullSub),
		}
	}

	return &MockClientStatus{
		Status: fmt.Sprintf("Subscription %v is not active", fullSub),
	}
}

func (*MockClient) Close() {}

type MockClientStatus struct {
	Status string
}

func (m *MockClientStatus) Result(details bool) string {
	return m.Status
}
