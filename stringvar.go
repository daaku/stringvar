// Command stringvar can generate go code to embed static assets. It's
// just a way to store the contents of files into variables.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func run() error {
	prg := os.Args[0]
	fs := flag.NewFlagSet(prg, flag.ExitOnError)
	out := fs.String("out", "", "output filename")
	pkg := fs.String("pkg", "", "package name in output file")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage:", prg, "-out [filename] -pkg [package] var1:file1 [var2:file2 ...]")
		fs.PrintDefaults()
		os.Exit(1)
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return errors.WithStack(err)
	}
	pairs := fs.Args()
	if len(pairs) == 0 {
		return errors.New("no variable:file arguments specified")
	}
	wd, err := os.Getwd()
	if err != nil {
		return errors.WithStack(err)
	}
	outFile, err := ioutil.TempFile(wd, fmt.Sprintf("stringvar-%s", *out))
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.Remove(outFile.Name())
	defer outFile.Close()
	fmt.Fprintf(outFile, "package %s\n\n", *pkg)
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			return errors.Errorf("unexpected argument %q", pair)
		}
		file, err := ioutil.ReadFile(parts[1])
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = fmt.Fprintf(outFile, "const %s = %q\n\n", parts[0], file)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if err := outFile.Close(); err != nil {
		return errors.WithStack(err)
	}
	if err := os.Rename(outFile.Name(), *out); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
