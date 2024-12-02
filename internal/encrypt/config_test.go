package encrypt

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Key:            "test-key",
				Value:          "test-value",
				Namespace:      "test-ns",
				ControllerNs:   "kube-system",
				ControllerName: "sealed-secrets-controller",
			},
			wantErr: false,
		},
		{
			name: "missing key",
			config: Config{
				Value:          "test-value",
				Namespace:      "test-ns",
				ControllerNs:   "kube-system",
				ControllerName: "sealed-secrets-controller",
			},
			wantErr: true,
		},
		{
			name: "missing value",
			config: Config{
				Key:            "test-key",
				Namespace:      "test-ns",
				ControllerNs:   "kube-system",
				ControllerName: "sealed-secrets-controller",
			},
			wantErr: true,
		},
		{
			name: "missing namespace",
			config: Config{
				Key:            "test-key",
				Value:          "test-value",
				ControllerNs:   "kube-system",
				ControllerName: "sealed-secrets-controller",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateControllerSettings(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "default controller settings",
			config: Config{
				Key:            "test-key",
				Value:          "test-value",
				Namespace:      "test-ns",
				ControllerNs:   "kube-system",
				ControllerName: "sealed-secrets-controller",
			},
			wantErr: false,
		},
		{
			name: "custom controller settings",
			config: Config{
				Key:            "test-key",
				Value:          "test-value",
				Namespace:      "test-ns",
				ControllerNs:   "custom-ns",
				ControllerName: "custom-controller",
			},
			wantErr: false,
		},
		{
			name: "empty controller settings",
			config: Config{
				Key:       "test-key",
				Value:     "test-value",
				Namespace: "test-ns",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
