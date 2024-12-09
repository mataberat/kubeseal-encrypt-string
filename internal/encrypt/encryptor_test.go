package encrypt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptor_findKubeseal(t *testing.T) {
	e := NewEncryptor(Config{})
	got, err := e.findKubeseal()
	if err != nil {
		t.Skip("Skipping test: kubeseal not found in system")
	}
	if got == "" {
		t.Error("findKubeseal() returned empty path")
	}
	t.Logf("Found kubeseal at: %s", got)
}

func TestEncryptor_createTempDir(t *testing.T) {
	e := NewEncryptor(Config{})
	err := e.createTempDir()
	if err != nil {
		t.Errorf("createTempDir() error = %v", err)
	}
	if e.tmpDir == "" {
		t.Error("tmpDir should not be empty")
	}
	if _, err := os.Stat(e.tmpDir); os.IsNotExist(err) {
		t.Error("temporary directory was not created")
	}
	defer os.RemoveAll(e.tmpDir)
}

func TestEncryptor_createUnsealedSecret(t *testing.T) {
	e := NewEncryptor(Config{
		Key:        "test-key",
		Value:      "test-value",
		Namespace:  "test-ns",
		SecretName: "test-secret",
	})
	e.createTempDir()
	defer os.RemoveAll(e.tmpDir)

	err := e.createUnsealedSecret()
	if err != nil {
		t.Errorf("createUnsealedSecret() error = %v", err)
	}

	unsealedPath := filepath.Join(e.tmpDir, "unsealed.yml")
	content, err := os.ReadFile(unsealedPath)
	if err != nil {
		t.Errorf("Failed to read unsealed secret file: %v", err)
	}

	expectedContent := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: test-ns
type: Opaque
data:
  test-key: dGVzdC12YWx1ZQ==`

	if strings.TrimSpace(string(content)) != strings.TrimSpace(expectedContent) {
		t.Errorf("Unexpected content in unsealed secret file.\nGot:\n%s\nWant:\n%s", content, expectedContent)
	}
}

func TestEncryptor_encodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "simple string",
			input:   "hello",
			want:    "aGVsbG8=",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "special characters",
			input:   "hello!@#$%^&*()",
			want:    "aGVsbG8hQCMkJV4mKigp",
			wantErr: false,
		},
	}

	e := NewEncryptor(Config{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.encodeBase64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got = strings.TrimSpace(got)
			if got != tt.want {
				t.Errorf("encodeBase64() got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewEncryptor(t *testing.T) {
	config := Config{
		Key:            "test-key",
		Value:          "test-value",
		Namespace:      "test-ns",
		SecretName:     "test-secret",
		ControllerNs:   "kube-system",
		ControllerName: "sealed-secrets-controller",
	}

	e := NewEncryptor(config)
	if e == nil {
		t.Error("NewEncryptor() returned nil")
	}
	if e.config != config {
		t.Error("NewEncryptor() config not set correctly")
	}
}
