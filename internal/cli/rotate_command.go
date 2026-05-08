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

func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
