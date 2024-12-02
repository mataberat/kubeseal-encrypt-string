package encrypt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Encryptor struct {
	config Config
	tmpDir string
}

func NewEncryptor(config Config) *Encryptor {
	return &Encryptor{
		config: config,
	}
}

func (e *Encryptor) Execute() error {
	if err := e.createTempDir(); err != nil {
		return err
	}
	defer os.RemoveAll(e.tmpDir)

	if err := e.promptConfirmation(); err != nil {
		return err
	}

	if err := e.createUnsealedSecret(); err != nil {
		return err
	}

	if err := e.runKubeseal(); err != nil {
		return err
	}

	encryptedValue, err := e.extractEncryptedValue()
	if err != nil {
		return err
	}

	e.printResult(encryptedValue)
	return nil
}

func (e *Encryptor) createTempDir() error {
	tmpDir, err := os.MkdirTemp("", "kubeseal-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	e.tmpDir = tmpDir
	return nil
}

func (e *Encryptor) promptConfirmation() error {
	fmt.Printf("You will generate a vaulted string on namespace %s. Do you want to continue [Y/N]? ", e.config.Namespace)
	var response string
	fmt.Scanln(&response)
	if strings.ToUpper(response) != "Y" {
		return fmt.Errorf("operation cancelled by user")
	}
	return nil
}

func (e *Encryptor) createUnsealedSecret() error {
	unsealedPath := filepath.Join(e.tmpDir, "unsealed.yml")
	encodedValue, err := e.encodeBase64(e.config.Value)
	if err != nil {
		return err
	}

	secretYaml := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: temp-secret
  namespace: %s
type: Opaque
data:
  %s: %s
`, e.config.Namespace, e.config.Key, strings.TrimSpace(encodedValue))

	return os.WriteFile(unsealedPath, []byte(secretYaml), 0600)
}

func (e *Encryptor) findKubeseal() (string, error) {
	// Common paths to check
	paths := []string{
		"/opt/homebrew/bin/kubeseal",
		"/usr/local/bin/kubeseal",
		"/usr/bin/kubeseal",
		"/bin/kubeseal",
	}

	// Check if kubeseal exists in PATH
	if path, err := exec.LookPath("kubeseal"); err == nil {
		fmt.Printf("Using kubeseal from: %s\n", path)
		return path, nil
	}

	// Check common paths
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Using kubeseal from: %s\n", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("kubeseal not found in system")
}

func (e *Encryptor) runKubeseal() error {
	kubesealPath, err := e.findKubeseal()
	if err != nil {
		return err
	}

	unsealedPath := filepath.Join(e.tmpDir, "unsealed.yml")
	sealedPath := filepath.Join(e.tmpDir, "sealed.yml")

	cmd := exec.Command(kubesealPath,
		"--format", "yaml",
		"--controller-namespace", e.config.ControllerNs,
		"--controller-name", e.config.ControllerName,
		"-f", unsealedPath,
		"-o", sealedPath,
	)

	// Capture command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubeseal failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func (e *Encryptor) extractEncryptedValue() (string, error) {
	sealedPath := filepath.Join(e.tmpDir, "sealed.yml")
	content, err := os.ReadFile(sealedPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, e.config.Key+": ") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return "", fmt.Errorf("encrypted value not found in sealed secret")
}

func (e *Encryptor) encodeBase64(s string) (string, error) {
	cmd := exec.Command("base64", "-w", "0")
	cmd.Stdin = strings.NewReader(s)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (e *Encryptor) printResult(encryptedValue string) {
	fmt.Printf("\nUsing sealed-secrets in %s on %s deployment.\n\n", e.config.ControllerNs, e.config.ControllerName)
	fmt.Printf("String vaulted:\n%s\n\n", encryptedValue)
	fmt.Printf("This secret only valid in namespace %s.\n", e.config.Namespace)
}
