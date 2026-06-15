package migrations

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

type Migration struct {
	Version string
	Up      func() ddl.Statements
	Down    func() ddl.Statements
}

var registryFilename = "registry.go"

const registryTemplate = `package migrations

import qmig "github.com/SennovE/qrafter/migrations"

var Registry = []qmig.Migration{
}

`

const migrationElem = `	{
		Version: "%s",
		Up:      Up%s,
		Down:    Down%s,
	},
`

func createRegistryFile(outDir string) error {
	path := filepath.Join(outDir, registryFilename)

	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat registry file: %w", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create migrations dir: %w", err)
	}

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		defaultMigrationFileMode,
	)
	if err != nil {
		return fmt.Errorf("create registry file: %w", err)
	}
	defer file.Close()

	code, err := renderGoTemplate("registry", registryTemplate, struct{}{})
	if err != nil {
		return fmt.Errorf("render registry template: %w", err)
	}

	if _, err := file.Write(code); err != nil {
		return fmt.Errorf("write registry file: %w", err)
	}

	return nil
}

func appendMigrationToRegistry(registryPath string, version string) error {
	path := filepath.Join(registryPath, registryFilename)
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
	b, err = format.Source(b)
	if err != nil {
		return fmt.Errorf("format registry: %w", err)
	}

	return os.WriteFile(path, b, 0o644)
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
