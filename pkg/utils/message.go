package utils

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Err    error
	Domain string
}

func (m *Message) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return fmt.Errorf("invalid json config")
	}
	return nil
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
