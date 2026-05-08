package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yourusername/envault/internal/store"
	"github.com/yourusername/envault/internal/template"
)

// RunTemplate renders a template file using secrets from the vault and
// writes the result to the given writer. If strict is true, any
// unresolved placeholder causes an error.
func RunTemplate(vaultPath, password, tmplPath string, strict bool, out io.Writer) error {
	v := store.NewVault(vaultPath, password)
	if err := v.Load(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load vault: %w", err)
		}
	}

	res, err := template.RenderFile(tmplPath, v, strict)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	if len(res.Missing) > 0 {
		fmt.Fprintf(out, "# warning: unresolved keys: %s\n", strings.Join(res.Missing, ", "))
	}

	fmt.Fprint(out, res.Output)

	if !strings.HasSuffix(res.Output, "\n") {
		fmt.Fprintln(out)
	}

	RecordAudit(vaultPath, fmt.Sprintf("template rendered: %s (resolved=%d missing=%d)",
		tmplPath, len(res.Resolved), len(res.Missing)))

	return nil
}
