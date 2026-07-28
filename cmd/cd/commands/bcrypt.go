package commands

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	"github.com/hanzoai/deploy/util/cli"
)

// NewBcryptCmd represents the bcrypt command
func NewBcryptCmd() *cobra.Command {
	var password string
	bcryptCmd := &cobra.Command{
		Use:   "bcrypt",
		Short: "Generate bcrypt hash for any password",
		Example: `# Generate bcrypt hash for any password 
cd account bcrypt --password YOUR_PASSWORD

# Prompt for password input
cd account bcrypt

# Read password from stdin
echo -e "password" | cd account bcrypt`,
		Run: func(cmd *cobra.Command, _ []string) {
			password = cli.PromptPassword(password)
			bytePassword := []byte(password)
			// Hashing the password
			hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("Failed to generate bcrypt hash: %v", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), string(hash))
		},
	}

	bcryptCmd.Flags().StringVar(&password, "password", "", "Password for which bcrypt hash is generated")
	return bcryptCmd
}
