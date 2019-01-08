package push

import (
	"context"
	"fmt"
)

type MockClient struct{}

func (*MockClient) Dial() error { return nil }

func (*MockClient) ActivateSubscription(ctx context.Context, fullSub, fullTopic, pushEndpoint, retryType string, retryPeriod uint32) ClientStatus {

	switch fullSub {
	case "/projects/ARGO/subscriptions/subNew":

		return &MockClientStatus{
			Status: fmt.Sprintf("Subscription %v activated", fullSub),
		}
	}

	return &MockClientStatus{
		Status: fmt.Sprintf("Subscription %v is already active", fullSub),
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

func (m *MockClientStatus) Result() string {
	return m.Status
}
