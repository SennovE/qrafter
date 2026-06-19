package migrations

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

// Migration registers one generated migration version and its up/down DDL.
type Migration struct {
	Version string
	Up      func() ddl.Statements
	Down    func() ddl.Statements
}

const migrationElem = `    qmig.New("%s", Up%s, Down%s),
`

func New(version string, up, down func() ddl.Statements) Migration {
	return Migration{
		Version: version,
		Up: up,
		Down: down,
	}
}

func appendMigrationToRegistry(registryPath, version string) error {
	path := filepath.Join(registryPath, configFilename)
	// #nosec G304 -- path is the registry inside the caller-provided migrations directory.
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read registry file: %w", err)
	}

	content := string(b)

	entry := fmt.Sprintf(migrationElem, version, version, version)

	startMarker := "var Registry = []qmig.Migration{"
	start := strings.Index(content, startMarker)
	if start == -1 {
		return fmt.Errorf("registry declaration not found")
	}

	bodyStart := start + len(startMarker)
	end := findMatchingBrace(content, bodyStart-1)
	if end == -1 {
		return fmt.Errorf("registry closing brace not found")
	} else if bodyStart == end {
		entry = "\n" + entry
	}

	content = content[:end] + entry + content[end:]
	b = bytes.NewBufferString(content).Bytes()
	b, err = formatGoSource(b)
	if err != nil {
		return fmt.Errorf("format registry: %w", err)
	}

	return os.WriteFile(path, b, defaultFileMode) // #nosec G304,G306,G703 -- registry path is caller-controlled generated Go source.
}

func findMatchingBrace(s string, openBracePos int) int {
	depth := 0

	for i := openBracePos; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}
