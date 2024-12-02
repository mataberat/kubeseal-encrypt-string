package encrypt

import "fmt"

type Config struct {
	Key            string
	Value          string
	Namespace      string
	ControllerNs   string
	ControllerName string
}

func (c *Config) Validate() error {
	if c.Key == "" || c.Value == "" || c.Namespace == "" {
		return fmt.Errorf("key, value, and namespace are required")
	}
	if c.ControllerNs == "" || c.ControllerName == "" {
		return fmt.Errorf("controller namespace and name are required")
	}
	return nil
}
