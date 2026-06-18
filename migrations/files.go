package migrations

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

const defaultFileMode = 0o644

func createFile(path string, code []byte) error {
	src, err := formatGoSource(code)
	if err != nil {
		return fmt.Errorf("format generated file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create generated file directory: %w", err)
	}

	// #nosec G304 -- path is a caller-provided destination for generated Go source.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultFileMode)
	if err != nil {
		return fmt.Errorf("create generated file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if _, err := file.Write(src); err != nil {
		return fmt.Errorf("write generated file: %w", err)
	}
	return nil
}

func formatGoSource(code []byte) ([]byte, error) {
	src, err := format.Source(code)
	if err != nil {
		return nil, err
	}
	return src, nil
}
