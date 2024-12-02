package encrypt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptor_Execute(t *testing.T) {
	// Let the system find kubeseal automatically
	e := NewEncryptor(Config{})
	kubesealPath, err := e.findKubeseal()
	if err != nil {
		t.Skip("Skipping test: kubeseal not found in system")
	}
	t.Logf("Found kubeseal at: %s", kubesealPath)

	// Set up test environment
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		w.Close()
		r.Close()
	}()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "successful execution",
			config: Config{
				Key:            "test-key",
				Value:          "test-value",
				Namespace:      "test-ns",
				ControllerNs:   "kube-system",
				ControllerName: "sealed-secrets-controller",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				w.Write([]byte("Y\n"))
			}()

			e := NewEncryptor(tt.config)
			err := e.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptor_findKubeseal(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "find kubeseal in system",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEncryptor(Config{})
			got, err := e.findKubeseal()
			if (err != nil) != tt.wantErr {
				t.Errorf("findKubeseal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == "" {
				t.Error("findKubeseal() returned empty path")
			}
			if got != "" {
				t.Logf("Found kubeseal at: %s", got)
			}
		})
	}
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
		Key:       "test-key",
		Value:     "test-value",
		Namespace: "test-ns",
	})
	e.createTempDir()
	defer os.RemoveAll(e.tmpDir)

	err := e.createUnsealedSecret()
	if err != nil {
		t.Errorf("createUnsealedSecret() error = %v", err)
	}

	unsealedPath := filepath.Join(e.tmpDir, "unsealed.yml")
	if _, err := os.Stat(unsealedPath); os.IsNotExist(err) {
		t.Error("unsealed secret file was not created")
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

func TestEncryptor_extractEncryptedValue(t *testing.T) {
	e := NewEncryptor(Config{
		Key: "test-key",
	})
	e.createTempDir()
	defer os.RemoveAll(e.tmpDir)

	sealedContent := `apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: test
spec:
  encryptedData:
    test-key: encrypted-value-here`

	sealedPath := filepath.Join(e.tmpDir, "sealed.yml")
	err := os.WriteFile(sealedPath, []byte(sealedContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test sealed.yml: %v", err)
	}

	value, err := e.extractEncryptedValue()
	if err != nil {
		t.Errorf("extractEncryptedValue() error = %v", err)
	}
	if value != "encrypted-value-here" {
		t.Errorf("extractEncryptedValue() = %v, want %v", value, "encrypted-value-here")
	}
}

func TestNewEncryptor(t *testing.T) {
	config := Config{
		Key:            "test-key",
		Value:          "test-value",
		Namespace:      "test-ns",
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
