package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/user/envault/internal/audit"
)

// RunAudit prints the audit log for the vault at vaultPath.
func RunAudit(vaultPath string, out io.Writer) error {
	logDir := auditDirForVault(vaultPath)
	logPath := filepath.Join(logDir, "audit.jsonl")

	entries, err := audit.ReadAll(logPath)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(out, "No audit entries found.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tPROJECT\tACTION\tKEY\tSUCCESS")
	for _, e := range entries {
		success := "ok"
		if !e.Success {
			success = "fail"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.Project,
			e.Action,
			e.Key,
			success,
		)
	}
	return w.Flush()
}

// RecordAudit is a helper used by CLI commands to log an action.
func RecordAudit(vaultPath, action, key string, success bool) {
	logDir := auditDirForVault(vaultPath)
	project := filepath.Base(filepath.Dir(vaultPath))
	l, err := audit.NewLogger(logDir, project)
	if err != nil {
		// Audit logging is best-effort; don't block the user.
		return
	}
	l.Record(action, key, success) //nolint:errcheck
}

// auditDirForVault returns the directory where audit logs are stored
// alongside the vault file.
func auditDirForVault(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	if dir == "." {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			dir = "."
		}
	}
	return filepath.Join(dir, ".envault")
}
