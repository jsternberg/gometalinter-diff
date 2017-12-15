package git

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

// Root returns the root of the git repository.
func Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}

// Checkout checks out the git revision.
func Checkout(rev string) error {
	cmd := exec.Command("git", "checkout", rev)
	return cmd.Run()
}

// CopyRevision copies the specified revision into the destination directory.
func CopyRevision(rev, dest string) error {
	var buf bytes.Buffer
	ls := exec.Command("git", "ls-tree", "-r", "--name-only", rev)
	ls.Stdout = &buf
	if err := ls.Run(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		path := scanner.Text()
		if err := func() error {
			fpath := filepath.Join(dest, path)
			if err := os.MkdirAll(filepath.Dir(fpath), 0700); err != nil {
				return err
			}

			f, err := os.Create(fpath)
			if err != nil {
				return err
			}
			defer f.Close()

			cmd := exec.Command("git", "show", rev+":"+path)
			cmd.Stdout = f
			if err := cmd.Run(); err != nil {
				return err
			}
			return f.Close()
		}(); err != nil {
			return err
		}
	}
	return nil
}
