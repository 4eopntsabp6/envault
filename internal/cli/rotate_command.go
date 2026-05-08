package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/user/envault/internal/rotation"
	"github.com/user/envault/internal/store"
	"golang.org/x/term"
)

// RunRotate re-encrypts the vault at vaultPath with a new password.
// It reads old and new passwords from the terminal if not provided.
// If both passwords are supplied as arguments (e.g. in tests), no
// terminal prompts are shown.
func RunRotate(vaultPath, oldPassword, newPassword string, out io.Writer) error {
	if oldPassword == "" {
		p, err := readPassword("Enter current password: ")
		if err != nil {
			return fmt.Errorf("reading old password: %w", err)
		}
		oldPassword = p
	}

	if newPassword == "" {
		p, err := readPassword("Enter new password: ")
		if err != nil {
			return fmt.Errorf("reading new password: %w", err)
		}
		confirm, err := readPassword("Confirm new password: ")
		if err != nil {
			return fmt.Errorf("reading confirmation: %w", err)
		}
		if p != confirm {
			return fmt.Errorf("passwords do not match")
		}
		newPassword = p
	}

	if oldPassword == newPassword {
		return fmt.Errorf("new password must differ from the current password")
	}

	result, err := rotation.Rotate(vaultPath, oldPassword, newPassword)
	if err != nil {
		return fmt.Errorf("rotation failed: %w", err)
	}

	fmt.Fprintf(out, "Rotated %d secret(s) in %s at %s\n",
		result.KeysCount,
		result.VaultPath,
		result.RotatedAt.Format("2006-01-02T15:04:05Z"),
	)

	RecordAudit(store.NewVault(vaultPath, newPassword), "rotate", "")
	return nil
}

// readPassword prints prompt to stderr, reads a password from the terminal
// without echoing, then prints a newline to stderr before returning.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
