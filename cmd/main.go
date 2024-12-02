package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mataberat/kubeseal-encrypt-string/internal/encrypt"
)

func main() {
	config := encrypt.Config{}

	flag.StringVar(&config.Key, "key", "", "Secret key")
	flag.StringVar(&config.Value, "value", "", "Secret value (raw, will be base64 encoded)")
	flag.StringVar(&config.Namespace, "namespace", "", "Target namespace")
	flag.StringVar(&config.ControllerNs, "controller-ns", "kube-system", "Sealed secrets controller namespace")
	flag.StringVar(&config.ControllerName, "controller-name", "sealed-secrets-controller", "Sealed secrets controller deployment name")
	flag.Parse()

	if err := config.Validate(); err != nil {
		fmt.Println("Error:", err)
		flag.Usage()
		os.Exit(1)
	}

	encryptor := encrypt.NewEncryptor(config)
	if err := encryptor.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
