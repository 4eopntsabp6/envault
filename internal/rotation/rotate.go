package rotation

import (
	"fmt"
	"time"

	"github.com/user/envault/internal/store"
)

// RotationResult holds the outcome of a key rotation operation.
type RotationResult struct {
	RotatedAt  time.Time
	KeysCount  int
	VaultPath  string
}

// Rotate re-encrypts all secrets in the vault using a new password.
// It loads all secrets with the old password, then saves them with the new one.
func Rotate(vaultPath, oldPassword, newPassword string) (*RotationResult, error) {
	if oldPassword == newPassword {
		return nil, fmt.Errorf("new password must differ from the old password")
	}

	oldVault, err := store.Load(vaultPath, oldPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to load vault with old password: %w", err)
	}

	keys := oldVault.Keys()
	if len(keys) == 0 {
		return nil, fmt.Errorf("vault is empty, nothing to rotate")
	}

	newVault := store.NewVault(vaultPath, newPassword)
	for _, k := range keys {
		v, ok := oldVault.Get(k)
		if !ok {
			continue
		}
		newVault.Set(k, v)
	}

	if err := newVault.Save(); err != nil {
		return nil, fmt.Errorf("failed to save rotated vault: %w", err)
	}

	return &RotationResult{
		RotatedAt: time.Now().UTC(),
		KeysCount: len(keys),
		VaultPath: vaultPath,
	}, nil
}
