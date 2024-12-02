package main

import (
	"fmt"
	"os"

	"github.com/mataberat/kubeseal-encrypt-string/internal/encrypt"
	"github.com/spf13/cobra"
)

func main() {
	config := encrypt.NewConfig()

	rootCmd := &cobra.Command{
		Use:   "kubeseal-encrypt-string",
		Short: "Encrypt strings using sealed-secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Validate(); err != nil {
				return fmt.Errorf("validation error: %w", err)
			}

			encryptor := encrypt.NewEncryptor(*config)
			return encryptor.Execute()
		},
	}

	// Add flags
	flags := rootCmd.Flags()
	flags.StringVar(&config.Key, "key", "", "Secret key")
	flags.StringVar(&config.Value, "value", "", "Secret value (raw, will be base64 encoded)")
	flags.StringVar(&config.Namespace, "namespace", "", "Target namespace")
	flags.StringVar(&config.ControllerNs, "controller-namespace", config.ControllerNs, "Sealed secrets controller namespace")
	flags.StringVar(&config.ControllerName, "controller-name", config.ControllerName, "Sealed secrets controller deployment name")

	// Add completion command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate completion script",
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				rootCmd.GenFishCompletion(os.Stdout, true)
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
