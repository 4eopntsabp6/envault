package cli

import (
	"fmt"
	"os"

	"github.com/user/envault/internal/env"
	"github.com/user/envault/internal/shell"
	"github.com/user/envault/internal/store"
)

// RunSet sets a secret in the vault for the current project.
func RunSet(vaultPath, key, value string) error {
	v, err := loadOrNew(vaultPath)
	if err != nil {
		return err
	}
	v.Set(key, value)
	if err := v.Save(vaultPath); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}
	fmt.Printf("Set %s\n", key)
	return nil
}

// RunGet retrieves and prints a secret from the vault.
func RunGet(vaultPath, key string) error {
	v, err := loadOrNew(vaultPath)
	if err != nil {
		return err
	}
	val, ok := v.Get(key)
	if !ok {
		return fmt.Errorf("key %q not found", key)
	}
	fmt.Println(val)
	return nil
}

// RunDelete removes a secret from the vault.
func RunDelete(vaultPath, key string) error {
	v, err := loadOrNew(vaultPath)
	if err != nil {
		return err
	}
	v.Delete(key)
	if err := v.Save(vaultPath); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}
	fmt.Printf("Deleted %s\n", key)
	return nil
}

// RunList prints all keys stored in the vault.
func RunList(vaultPath string) error {
	v, err := loadOrNew(vaultPath)
	if err != nil {
		return err
	}
	keys := v.Keys()
	if len(keys) == 0 {
		fmt.Println("(no secrets stored)")
		return nil
	}
	for _, k := range keys {
		fmt.Println(k)
	}
	return nil
}

// RunExport prints shell export statements for all secrets.
func RunExport(vaultPath, format string) error {
	v, err := loadOrNew(vaultPath)
	if err != nil {
		return err
	}
	fmt.Fprint(os.Stdout, shell.ExportEnv(v, shell.ParseFormat(format)))
	return nil
}

// RunImport imports secrets from a .env file into the vault.
func RunImport(vaultPath, filePath string) error {
	v, err := loadOrNew(vaultPath)
	if err != nil {
		return err
	}
	count, err := env.ImportFile(v, filePath)
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}
	if err := v.Save(vaultPath); err != nil {
		return fmt.Errorf("save vault: %w", err)
	}
	fmt.Printf("Imported %d secret(s) from %s\n", count, filePath)
	return nil
}

func loadOrNew(vaultPath string) (*store.Vault, error) {
	v, err := store.Load(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("load vault: %w", err)
	}
	return v, nil
}
