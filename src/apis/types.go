package apis

import (
	"encoding/json"
	"fmt"
)

// Message 接口返回
type Message struct {
	Error   string `json:"Error,omitempty"`
	Message string `json:"Message,omitempty"`
}

// NewJSONMessage 返回json格式的Message信息
func NewJSONMessage(msg string) string {
	data, _ := json.Marshal(Message{Message: msg})
	return string(data)
}

// NewJSONMessagef 支持format
func NewJSONMessagef(format string, v ...interface{}) string {
	data, _ := json.Marshal(Message{Message: fmt.Sprintf(format, v...)})
	return string(data)
}

// NewJSONError 返回json格式Message信息
func NewJSONError(err string) string {
	data, _ := json.Marshal(Message{Error: err})
	return string(data)
}

// NewJSONErrorf 支持format
func NewJSONErrorf(format string, v ...interface{}) string {
	data, _ := json.Marshal(Message{Error: fmt.Sprintf(format, v...)})
	return string(data)
}
