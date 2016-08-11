package push

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"
)

// Sender is inteface for sending messages to remote endpoints
type Sender interface {
	Send(msg string, endpoint string) error
}

// HTTPSender sends msgs through http
type HTTPSender struct {
	Client http.Client
}

// MockSender mocks sending messages to remote endpoints
type MockSender struct {
	ClientFail   bool
	LastMsg      string
	LastEndpoint string
}

// NewHTTPSender creates a new HTTPSender. Specify timeout in seconds
func NewHTTPSender(timeoutSec int) *HTTPSender {
	timeout := time.Duration(timeoutSec) * time.Second
	client := http.Client{Timeout: timeout}
	snd := HTTPSender{Client: client}
	return &snd
}

// Send sends a message through HTTP
func (hs *HTTPSender) Send(msg string, endpoint string) error {
	var jsonStr = []byte(msg)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	log.Println("Sending to endpoint:", endpoint)
	log.Println("message contents:", msg)

	resp, err := hs.Client.Do(req)

	if err == nil {
		if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 && resp.StatusCode != 101 {
			err = errors.New("Endpoint Responded: not delivered")
		}
		log.Println("message Delivered")
	}

	return err
}

// NewMockSender creates a new Mock sender
func NewMockSender(fail bool) *MockSender {

	snd := MockSender{ClientFail: fail}
	return &snd
}

// Send sends a message through HTTP
func (ms *MockSender) Send(msg string, endpoint string) error {
	if ms.ClientFail == true {
		return errors.New("endpoint not reachable")
	}

	ms.LastMsg = msg
	ms.LastEndpoint = endpoint

	return nil
}
