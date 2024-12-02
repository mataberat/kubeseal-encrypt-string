package encrypt

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

type Encryptor struct {
	config   Config
	tmpDir   string
	testName string
}

func NewEncryptor(config Config) *Encryptor {
	return &Encryptor{
		config:   config,
		testName: fmt.Sprintf("test-%s", uuid.New().String()[:8]),
	}
}

func (e *Encryptor) Execute() error {
	if err := e.validateNamespace(); err != nil {
		return err
	}

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

	if err := e.validateSecret(encryptedValue); err != nil {
		return err
	}

	e.printResult(encryptedValue)
	return nil
}

func (e *Encryptor) validateNamespace() error {
	kubectlPath, err := e.findKubectl()
	if err != nil {
		return err
	}

	cmd := exec.Command(kubectlPath, "get", "namespace", e.config.Namespace)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ namespace %s not found or not accessible: %v\nOutput: %s",
			e.config.Namespace, err, string(output))
	}

	cmd = exec.Command(kubectlPath, "get", "deployment",
		e.config.ControllerName,
		"-n", e.config.ControllerNs)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ sealed-secrets controller not found in namespace %s: %v\nOutput: %s",
			e.config.ControllerNs, err, string(output))
	}

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
	titleColor := color.New(color.FgCyan, color.Bold)
	titleColor.Printf("\n🔒 Generating Sealed Secret\n")
	color.New(color.FgYellow).Printf("📍 Target Namespace: %s\n", e.config.Namespace)
	color.New(color.FgBlue).Printf("🎯 Controller Namespace: %s\n", e.config.ControllerNs)
	color.New(color.FgWhite).Print("Continue? [Y/N]: ")

	var response string
	fmt.Scanln(&response)
	if strings.ToUpper(response) != "Y" {
		return fmt.Errorf("❌ Operation cancelled by user")
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
  name: %s
  namespace: %s
type: Opaque
data:
  %s: %s
`, e.testName, e.config.Namespace, e.config.Key, strings.TrimSpace(encodedValue))

	return os.WriteFile(unsealedPath, []byte(secretYaml), 0600)
}

func (e *Encryptor) findKubeseal() (string, error) {
	paths := []string{
		"/opt/homebrew/bin/kubeseal",
		"/usr/local/bin/kubeseal",
		"/usr/bin/kubeseal",
		"/bin/kubeseal",
	}

	if path, err := exec.LookPath("kubeseal"); err == nil {
		color.New(color.FgBlue).Printf("🔧 Using kubeseal from: %s\n", path)
		return path, nil
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			color.New(color.FgBlue).Printf("🔧 Using kubeseal from: %s\n", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("❌ kubeseal not found in system")
}

func (e *Encryptor) findKubectl() (string, error) {
	paths := []string{
		"/opt/homebrew/bin/kubectl",
		"/usr/local/bin/kubectl",
		"/usr/bin/kubectl",
		"/bin/kubectl",
	}

	if path, err := exec.LookPath("kubectl"); err == nil {
		color.New(color.FgBlue).Printf("🔧 Using kubectl from: %s\n", path)
		return path, nil
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			color.New(color.FgBlue).Printf("🔧 Using kubectl from: %s\n", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("❌ kubectl not found in system")
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
		"--scope", "strict",
		"--controller-namespace", e.config.ControllerNs,
		"--controller-name", e.config.ControllerName,
		"--secret-file", unsealedPath,
		"--sealed-secret-file", sealedPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("❌ kubeseal failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func (e *Encryptor) validateSecret(encryptedValue string) error {
	color.New(color.FgCyan).Printf("\n🔍 Validating sealed secret: %s\n", e.testName)

	sealedSecret := fmt.Sprintf(`apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: %s
  namespace: %s
spec:
  encryptedData:
    %s: %s`, e.testName, e.config.Namespace, e.config.Key, encryptedValue)

	if err := e.kubectlApply([]byte(sealedSecret)); err != nil {
		return fmt.Errorf("❌ failed to apply test sealed secret: %w", err)
	}
	defer e.cleanupTestResources(e.testName)

	color.New(color.FgYellow).Println("⏳ Waiting for secret creation...")
	for i := 0; i < 10; i++ {
		if secretValue, err := e.getSecretValue(e.testName, e.config.Key); err == nil {
			decodedValue, err := e.decodeBase64(secretValue)
			if err != nil {
				return fmt.Errorf("❌ failed to decode secret value: %w", err)
			}
			if decodedValue == e.config.Value {
				color.New(color.FgGreen).Println("✅ Secret validation successful")
				return nil
			}
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("❌ failed to validate secret: timeout waiting for secret creation")
}

func (e *Encryptor) kubectlApply(manifest []byte) error {
	kubectlPath, err := e.findKubectl()
	if err != nil {
		return err
	}

	cmd := exec.Command(kubectlPath, "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(manifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}

func (e *Encryptor) cleanupTestResources(name string) {
	kubectlPath, _ := e.findKubectl()
	exec.Command(kubectlPath, "delete", "sealedsecret", name, "-n", e.config.Namespace).Run()
	exec.Command(kubectlPath, "delete", "secret", name, "-n", e.config.Namespace).Run()
}

func (e *Encryptor) getSecretValue(name string, key string) (string, error) {
	kubectlPath, err := e.findKubectl()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(kubectlPath, "get", "secret", name, "-n", e.config.Namespace, "-o", fmt.Sprintf("jsonpath={.data.%s}", key))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kubectl get secret failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
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

func (e *Encryptor) decodeBase64(s string) (string, error) {
	cmd := exec.Command("base64", "--decode")
	cmd.Stdin = strings.NewReader(s)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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
	return "", fmt.Errorf("❌ encrypted value not found in sealed secret")
}

func (e *Encryptor) printResult(encryptedValue string) {
	titleColor := color.New(color.FgCyan, color.Bold)
	infoColor := color.New(color.FgGreen)
	valueColor := color.New(color.FgYellow)

	titleColor.Printf("\n🎉 Sealed Secret Generated Successfully\n")
	infoColor.Printf("\n📦 Controller: ")
	fmt.Printf("%s/%s\n", e.config.ControllerNs, e.config.ControllerName)
	infoColor.Printf("🌐 Target Namespace: ")
	fmt.Printf("%s\n", e.config.Namespace)
	infoColor.Printf("\n🔑 Encrypted Value:\n")
	valueColor.Printf("%s\n", encryptedValue)
	titleColor.Printf("\n✨ Secret is ready to use!\n")
}
