package messages

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

// RecMsg holds info for a received message
type RecMsg struct {
	AckID string  `json:"ackId,omitempty"`
	Msg   Message `json:"message"`
}

// RecList holds the array of the receivedMessages - subscription related
type RecList struct {
	RecMsgs []RecMsg `json:"receivedMessages"`
}

// MsgList is used to hold a list of messages
type MsgList struct {
	Msgs []Message `json:"messages"`
}

// Message struct used to hold message information
type Message struct {
	ID      string     `json:"messageId,omitempty"`
	Attr    Attributes `json:"attributes,omitempty"`  // used to hold attribute key/value store
	Data    string     `json:"data"`                  // base64 encoded data payload
	PubTime string     `json:"publishTime,omitempty"` // publish timedate of message
}

// PushMsg contains structure for push messages
type PushMsg struct {
	Msg Message `json:"message"`
	Sub string  `json:"subscription"`
}

// MsgIDs utility struct
type MsgIDs struct {
	IDs []string `json:"messageIds"`
}

// Attributes representation as key/value
type Attributes map[string]string

//TotalSize returns the total bytesize of a message list
func (msgL RecList) TotalSize() int64 {
	sum := int64(0)
	for _, msg := range msgL.RecMsgs {
		// Convert data string to byte array
		bt := []byte(msg.Msg.Data)
		sum = sum + int64(len(bt))
	}

	return sum
}

//TotalSize returns the total bytesize of a message list
func (msgL MsgList) TotalSize() int64 {
	sum := int64(0)
	for _, msg := range msgL.Msgs {
		// Convert data string to byte array
		bt := []byte(msg.Data)
		sum = sum + int64(len(bt))
	}

	return sum
}

//Size returns the messages size in bytes
func (msg Message) Size() int64 {
	// Convert data string to byte array
	bt := []byte(msg.Data)
	size := int64(len(bt))
	return size
}

// MarshalJSON generates json string for Attributes type
func (attr Attributes) MarshalJSON() ([]byte, error) {
	var buff bytes.Buffer

	// sort attribute keys
	keys := make([]string, len(attr))
	i := 0
	for key := range attr {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	// iterate over sorted keys and marshal to json
	buff.WriteString("{")
	for i, key := range keys {
		if i != 0 {
			buff.WriteString(", ")
		}
		buff.WriteString(strings.Join([]string{"\"", key, "\":\"", attr[key], "\""}, ""))
	}

	buff.WriteString("}")

	return buff.Bytes(), nil
}

// type Attribute struct {
// 	Key   string `json:"key"`
// 	Value string `json:"value"`
// }

// Construct functions
//////////////////////

// New creates a new Message based on data string provided
func New(data string) Message {
	msg := Message{ID: "0", Attr: make(map[string]string), Data: data}

	return msg
}

// LoadMsgListJSON creates a MsgList from a json definition
func LoadMsgListJSON(input []byte) (MsgList, error) {
	m := MsgList{}
	err := json.Unmarshal([]byte(input), &m)
	return m, err
}

// LoadMsgJSON creates a new Message from a json string represenatation
func LoadMsgJSON(input []byte) (Message, error) {
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

// AttrExists checks if an attribute exists based on key. Returns also a boolean
// if the attribute exists
func (msg *Message) AttrExists(key string) (bool, string) {

	for k, value := range msg.Attr {
		if k == key {
			return true, value
		}
	}

	return false, ""

}

// InsertAttribute takes a key/value item and appends it in Message's attributes
func (msg *Message) InsertAttribute(key string, value string) error {
	exists, _ := msg.AttrExists(key)
	if exists {
		return errors.New("Attribute already exists")
	}

	msg.Attr[key] = value
	return nil
}

// UpdateAttribute updates an existing attribute based on key and new value
func (msg *Message) UpdateAttribute(key string, value string) error {
	exists, _ := msg.AttrExists(key)
	if exists {
		msg.Attr[key] = value
		return nil
	}

	return errors.New("Attribute doesn't exist")
}

// RemoveAttribute takes a key and removes attribute if exists (based on key)
func (msg *Message) RemoveAttribute(key string) error {
	exists, _ := msg.AttrExists(key)
	if exists {
		delete(msg.Attr, key)
		return nil
	}

	return errors.New("Attribute doesn't exist")
}

// GetAttribute takes a key and return attribute value if exists (based on key)
func (msg *Message) GetAttribute(key string) (string, error) {
	exists, value := msg.AttrExists(key)
	if exists {
		return value, nil
	}

	return "", errors.New("Attribute doesn't exist")

}

// ExportJSON exports whole Message Structure as a json string
func (pMsg *PushMsg) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(pMsg, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole Message Structure as a json string
func (msg *Message) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(msg, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole msgId  Structure as a json string
func (msgIDs *MsgIDs) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(msgIDs, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole msgId  Structure as a json string
func (recList *RecList) ExportJSON() (string, error) {
	if recList.RecMsgs == nil {
		recList.RecMsgs = []RecMsg{}
	}
	output, err := json.MarshalIndent(recList, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports whole MsgList as a json string
func (msgList *MsgList) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(msgList, "", "   ")
	return string(output[:]), err
}
