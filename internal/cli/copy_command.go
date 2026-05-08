package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/envault/internal/copy"
	"github.com/user/envault/internal/store"
)

// RunCopy copies secrets from one vault file into another.
// srcPath and dstPath are the paths to the respective vault files.
// srcPass and dstPass are their passwords.
// keys is an optional comma-separated list of keys to copy; empty means all.
// overwrite controls whether existing keys in dst are replaced.
func RunCopy(w io.Writer, srcPath, srcPass, dstPath, dstPass string, keys string, overwrite bool) error {
	src, err := store.Load(srcPath, srcPass)
	if err != nil {
		return fmt.Errorf("opening source vault: %w", err)
	}

	dst, err := store.Load(dstPath, dstPass)
	if err != nil {
		return fmt.Errorf("opening destination vault: %w", err)
	}

	var keyList []string
	if keys != "" {
		for _, k := range strings.Split(keys, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				keyList = append(keyList, k)
			}
		}
	}

	res, err := copy.Copy(src, dst, copy.Options{
		Overwrite: overwrite,
		Keys:      keyList,
	})
	if err != nil {
		return err
	}

	if err := dst.Save(dstPath, dstPass); err != nil {
		return fmt.Errorf("saving destination vault: %w", err)
	}

	fmt.Fprintf(w, "Copied %d key(s)", len(res.Copied))
	if len(res.Skipped) > 0 {
		fmt.Fprintf(w, ", skipped %d existing key(s)", len(res.Skipped))
	}
	fmt.Fprintln(w)

	for _, k := range res.Copied {
		fmt.Fprintf(w, "  + %s\n", k)
	}
	for _, k := range res.Skipped {
		fmt.Fprintf(w, "  ~ %s (skipped)\n", k)
	}

	return nil
}
