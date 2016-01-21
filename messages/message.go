package messages

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
)

// Message struct used to hold message information
type Message struct {
	ID   int64       `json:"messageId"`
	Attr []Attribute `json:"attributes"` // used to hold attribute key/value store
	Data string      `json:"data"`       // base64 encoded data payload
}

// Attribute representation as key/value
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Construct functions
//////////////////////

// New creates a new Message based on data string provided
func New(data string) Message {
	msg := Message{ID: 0, Attr: []Attribute{}, Data: data}

	return msg
}

// LoadJSON creates a new Message from a json string represenatation
func LoadJSON(input string) (Message, error) {
	m := Message{}
	err := json.Unmarshal([]byte(input), &m)
	return m, err
}

// Message Methods
//////////////////

// GetDecoded returns the base64 payload in it's original form
func (msg *Message) GetDecoded() string {
	decoded, _ := b64.StdEncoding.DecodeString(msg.Data)
	return string(decoded[:])
}

// AttrExists checks if an attribute exists based on key. Returns also the index
// of the attribute item in Attributes slice
func (msg *Message) AttrExists(key string) (int, string) {

	for i, a := range msg.Attr {
		if a.Key == key {
			return i, a.Value
		}
	}

	return -1, ""

}

// InsertAttribute takes a key/value item and appends it in Message's attributes
func (msg *Message) InsertAttribute(key string, value string) error {
	i, _ := msg.AttrExists(key)
	if i > -1 {
		return errors.New("Attribute already exists")
	}

	msg.Attr = append(msg.Attr, Attribute{Key: key, Value: value})
	return nil
}

// UpdateAttribute updates an existing attribute based on key and new value
func (msg *Message) UpdateAttribute(key string, value string) error {
	i, _ := msg.AttrExists(key)
	if i > -1 {
		msg.Attr[i] = Attribute{Key: key, Value: value}
		return nil
	}

	return errors.New("Attribute doesn't exist")
}

// RemoveAttribute takes a key and removes attribute if exists (based on key)
func (msg *Message) RemoveAttribute(key string) error {
	i, _ := msg.AttrExists(key)
	if i > -1 {
		msg.Attr = append(msg.Attr[:i], msg.Attr[i+1:]...)
		return nil
	}

	return errors.New("Attribute doesn't exist")
}

// GetAttribute takes a key and return attribute value if exists (based on key)
func (msg *Message) GetAttribute(key string) (string, error) {
	i, value := msg.AttrExists(key)
	if i > -1 {
		return value, nil
	}

	return "", errors.New("Attribute doesn't exist")

}

// ExportJSON exports whole Message Structure as a json string
func (msg *Message) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(msg, "", "   ")
	return string(output[:]), err
}
