package util

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// FromAferoToOS will copy the content from fs on path fspath to os(stdlib) to the ospath
func FromAferoToOS(afs afero.Fs, fspath, ospath string) error {
	err := afero.Walk(afs, fspath, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk path %q: %w", p, err)
		}
		relpath, _ := filepath.Rel(fspath, p)
		tmppath := filepath.Join(ospath, relpath)
		if info.IsDir() {
			err := os.MkdirAll(tmppath, info.Mode())
			if err != nil {
				return fmt.Errorf("could not create path %q: %w", tmppath, err)
			}
			return nil
		}
		f, err := os.Create(tmppath)
		if err != nil {
			return fmt.Errorf("failed to create %s into the FS: %w", info.Name(), err)
		}

		af, err := afs.Open(p)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", p, err)
		}

		_, err = io.Copy(f, af)
		if err != nil {
			return fmt.Errorf("failed copying data from %q: %w", p, err)
		}
		f.Close()
		af.Close()
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk path %q: %w", fspath, err)
	}
	return nil
}

// FromOSToAfero will copy the content on ospath to fs on the fspath
func FromOSToAfero(afs afero.Fs, ospath, fspath string) error {
	err := filepath.WalkDir(ospath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk path %q: %w", p, err)
		}
		relpath, _ := filepath.Rel(ospath, p)
		tmppath := filepath.Join(fspath, relpath)
		if d.IsDir() {
			err = afs.MkdirAll(tmppath, d.Type())
			if err != nil {
				return fmt.Errorf("could not create path %q: %w", tmppath, err)
			}
			return nil
		}
		f, err := afs.Create(tmppath)
		if err != nil {
			return fmt.Errorf("failed to create %s into the FS: %w", d.Name(), err)
		}

		osf, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", p, err)
		}

		_, err = io.Copy(f, osf)
		if err != nil {
			return fmt.Errorf("failed copying data from %q: %w", p, err)
		}
		f.Close()
		osf.Close()
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to copy the module to %s: %w", ospath, err)
	}
	return nil
}
