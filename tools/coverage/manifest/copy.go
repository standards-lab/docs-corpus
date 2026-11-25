package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copy copies each of m's files from under src to its path under dst, creating its directories,
// and checks the copy's size and sha256 against its entry. A file already at dst with its entry's
// size and sha256 is left as it is. A copy that differs from its entry is removed and is an error.
func Copy(m Manifest, src, dst string) error {
	var errs []error
	for _, f := range m.Files {
		target := filepath.Join(dst, filepath.FromSlash(f.Path))
		if matches(target, f) == nil {
			continue
		}
		if err := copyFile(filepath.Join(src, filepath.FromSlash(f.Path)), target); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := matches(target, f); err != nil {
			errs = append(errs, err, os.Remove(target))
		}
	}
	return errors.Join(errs...)
}

func matches(path string, f File) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	h := sha256.New()
	n, err := io.Copy(h, file)
	if err != nil {
		return err
	}
	if n != f.Bytes {
		return fmt.Errorf("%s: %d bytes, but the manifest has %d", path, n, f.Bytes)
	}
	if sum := hex.EncodeToString(h.Sum(nil)); sum != f.Sha256 {
		return fmt.Errorf("%s: sha256 %s, but the manifest has %s", path, sum, f.Sha256)
	}
	return nil
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, in.Close()) }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".copy-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(tmp.Name()))
		}
	}()
	if _, err := io.Copy(tmp, in); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Chmod(0o644); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}
