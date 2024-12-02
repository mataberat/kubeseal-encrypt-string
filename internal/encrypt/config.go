package encrypt

import (
	"fmt"
	"os"
)

type Config struct {
	Key            string
	Value          string
	Namespace      string
	ControllerNs   string
	ControllerName string
}

func NewConfig() *Config {
	return &Config{
		ControllerNs:   getEnvOrDefault("SEALED_SECRETS_CONTROLLER_NAMESPACE", "kube-system"),
		ControllerName: getEnvOrDefault("SEALED_SECRETS_CONTROLLER_NAME", "sealed-secrets-controller"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
