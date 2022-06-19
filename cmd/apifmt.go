package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zeromicro/zero-api/format"
)

var (
	write = flag.Bool("w", false, "write result to (source) file instead of stdout")
)

func usage() {
	fmt.Fprint(os.Stderr, "usage: apifmt [flags] [path ...]\n")
	flag.PrintDefaults()
}

func report(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func processFile(fileName string, info fs.FileInfo, in io.Reader, out io.Writer) error {
	if in == nil {
		var err error
		in, err = os.Open(fileName)
		if err != nil {
			return err
		}
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	res, err := format.Source(src, fileName)
	if err != nil {
		return err
	}

	if *write {
		_, err := out.Write(res)
		return err
	}
	// write to file
	perm := info.Mode().Perm()
	backName, err := backupFile(fileName+".", src, perm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, res, perm)
	if err != nil {
		_ = os.Rename(backName, fileName)
		return err
	}

	err = os.Remove(backName)
	return err
}

const chmodSupported = runtime.GOOS != "windows"

func backupFile(filename string, data []byte, perm fs.FileMode) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}
	backname := f.Name()
	if chmodSupported {
		err := f.Chmod(perm)
		if err != nil {
			_ = f.Close()
			_ = os.Remove(backname)
			return backname, err
		}
	}

	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil {
		err = err1
	}
	return backname, err
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		if *write {
			report(errors.New("error: cannot use -w with standard input"))
			return
		}
		// *write not support standard input, so fs.FileInfo can be nil.
		if err := processFile("<standard input>", nil, os.Stdin, os.Stdout); err != nil {
			report(err)
		}
		return
	}

	for _, path := range args {
		walkSubDir := strings.HasSuffix(path, "/...")
		if walkSubDir {
			path = path[:len(path)-1]
		}
		filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else if d.IsDir() {
				if !walkSubDir {
					return filepath.SkipDir
				}
			} else {
				ext := filepath.Ext(path)
				if ext != ".api" {
					return nil
				}

				info, err := d.Info()
				if err != nil {
					return err
				}
				if err := processFile(path, info, nil, os.Stdout); err != nil {
					report(err)
				}
			}
			return err
		})
	}
}
