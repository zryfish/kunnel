package agent

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	Name string

	LocalPort int

	LocalHost string

	Domain string

	Host string

	Hedaers map[string]string
}

func (c *Config) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, c); err != nil {
		return fmt.Errorf("invalid json config")
	}
	return nil
}

func (c *Config) Marshal() ([]byte, error) {
	return json.Marshal(c)
}
