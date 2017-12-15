package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"os/exec"

	"bytes"

	"bufio"

	"github.com/jsternberg/gometalinter-diff/git"
	"github.com/spf13/pflag"
)

func LintRevision(rev string) (*os.File, error) {
	workdir, err := ioutil.TempDir(os.TempDir(), "lint")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workdir)

	// Copy the revision into the directory.
	if err := git.CopyRevision(rev, workdir); err != nil {
		return nil, err
	}
	return LintDir(workdir)
}

func LintDir(dir string) (*os.File, error) {
	f, err := ioutil.TempFile(os.TempDir(), "lint.txt.")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("gometalinter")
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			os.Remove(f.Name())
			return nil, err
		}
	}
	return f, nil
}

func realMain() error {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <rev>\n", os.Args[0])
		pflag.PrintDefaults()
	}
	pflag.Parse()

	args := pflag.Args()
	if len(args) != 1 {
		pflag.Usage()
		os.Exit(1)
	}

	// Change to the root directory of the git repository.
	dir, err := git.Root()
	if err != nil {
		return err
	}
	os.Chdir(dir)

	// Run gometalinter.
	lintfile1, err := LintRevision(args[0])
	if err != nil {
		return err
	}
	defer os.Remove(lintfile1.Name())

	lintfile2, err := LintDir(".")
	if err != nil {
		return err
	}
	defer os.Remove(lintfile2.Name())

	var diff bytes.Buffer
	cmd := exec.Command("diff", "-u", lintfile1.Name(), lintfile2.Name())
	cmd.Stdout = &diff
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return err
		}
	}

	// Scan the output for lines beginning with a plus sign and output them.
	scanner := bufio.NewScanner(&diff)
	// Discard the first two lines of the diff.
	scanner.Scan()
	scanner.Scan()
	errors := 0
	for scanner.Scan() {
		if bytes.HasPrefix(scanner.Bytes(), []byte("+")) {
			os.Stdout.Write(scanner.Bytes())
			os.Stdout.Write([]byte("\n"))
			errors++
		}
	}

	if errors > 0 {
		os.Exit(1)
	}
	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
