package envcopy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyGlobs copies files matching globs (relative to srcRoot) into dstRoot.
// Never overwrites; missing matches are skipped. Returns copied relative paths.
func CopyGlobs(srcRoot, dstRoot string, globs []string) ([]string, error) {
	var copied []string
	for _, g := range globs {
		matches, err := filepath.Glob(filepath.Join(srcRoot, g))
		if err != nil {
			continue // bad pattern: skip
		}
		for _, src := range matches {
			info, err := os.Stat(src)
			if err != nil || info.IsDir() {
				continue
			}
			rel, err := filepath.Rel(srcRoot, src)
			if err != nil {
				continue
			}
			dst := filepath.Join(dstRoot, rel)
			if _, err := os.Stat(dst); err == nil {
				continue // never overwrite
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return copied, err
			}
			if err := copyFile(src, dst); err != nil {
				return copied, fmt.Errorf("copy %s: %w", rel, err)
			}
			copied = append(copied, rel)
		}
	}
	return copied, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
