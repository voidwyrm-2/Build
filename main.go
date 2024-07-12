package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/akamensky/argparse"
)

var COMPILER string = "gcc"

func compile(files []string, xargs []string, outpath string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := files
	args = append(args, xargs...)
	args = append(args, "-o", outpath)

	cmd := exec.Command(COMPILER, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

func readFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	content := ""
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return content, nil
}

type NullableString any

type CBuildfileInfo struct {
	output, compiler NullableString
	xargs            []string
}

func parseCbuildFile() ([]string, CBuildfileInfo, error) {
	content, ferr := readFile("./CBuildfile")
	if ferr != nil {
		return []string{}, CBuildfileInfo{}, ferr
	}

	lines := strings.Split(content, "\n")

	var paths []string

	var outpath NullableString = nil
	var compiler NullableString = nil
	var xargs []string

	for ln, l := range lines {
		commentIdx := strings.Index(l, "//")
		if commentIdx != -1 {
			l = l[:commentIdx]
		}
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		if l[0] == '%' {
			dir := l[1:]
			if len(dir) >= 4 {
				if dir[:len(dir)-(len(dir)-4)] == "OUT " {
					if outpath != nil {
						return []string{}, CBuildfileInfo{}, fmt.Errorf("error on line %d: directive 'OUT' has already been declared", ln+1)
					}
					outpath = strings.TrimSpace(dir[3:])
					continue
				} else if dir[:len(dir)-(len(dir)-4)] == "ARG " {
					xargs = append(xargs, strings.TrimSpace(dir[3:]))
					continue
				} else {
					return []string{}, CBuildfileInfo{}, fmt.Errorf("error on line %d: invalid directive '%s'", ln+1, l)
				}
			} else if len(dir) >= 9 {
				if dir[:len(dir)-(len(dir)-9)] == "COMPILER " {
					if compiler != nil {
						return []string{}, CBuildfileInfo{}, fmt.Errorf("error on line %d: directive 'COMPILER' has already been declared", ln+1)
					}
					compiler = strings.TrimSpace(dir[8:])
					continue
				} else {
					return []string{}, CBuildfileInfo{}, fmt.Errorf("error on line %d: invalid directive '%s'", ln+1, l)
				}
			} else {
				return []string{}, CBuildfileInfo{}, fmt.Errorf("error on line %d: invalid directive '%s'", ln+1, l)
			}
		}

		if _, patherr := os.Stat(l); patherr == nil {
			//fmt.Printf("(l %d) file '%s' exists\n", ln+1, l)
			paths = append(paths, l)
		} else if errors.Is(patherr, os.ErrNotExist) {
			return []string{}, CBuildfileInfo{}, fmt.Errorf("error on line %d: path '%s' does not exist", ln+1, l)
		} else {
			// Schrodinger: file may or may not exist. See err for details.
			// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
			return []string{}, CBuildfileInfo{}, patherr
		}
	}

	return paths, CBuildfileInfo{output: outpath, compiler: compiler, xargs: xargs}, nil
}

const extraNote string = "\n              Note: if -b/--build and -c/--cbuild are omitted, then CBuild will search for a CBuildfile in the current directory"

func main() {
	parser := argparse.NewParser("CBuild", "A simple C build system"+extraNote)

	ptr_paths := parser.StringList("b", "build", &argparse.Options{Required: false, Help: "Paths of the C files to build"})

	ptr_cbuildfile := parser.String("c", "cbuild", &argparse.Options{Required: false, Help: "Can be used instead of -b/--build to specify the CBuild file", Default: ""})

	ptr_outpath := parser.String("o", "out", &argparse.Options{Required: false, Help: "Optionally specify the output path of the compile executable", Default: ""})

	//__showInvocation := parser.Flag("v", "invocation", &//argparse.Options{Required: false, Help: "Compiler 'show invocation' flag", Default: false})

	parserErr := parser.Parse(os.Args)
	if parserErr != nil {
		fmt.Print(parser.Usage(parserErr))
		return
	}

	var paths []string = *ptr_paths
	var cbuildfile string = *ptr_cbuildfile
	var outpath string = *ptr_outpath

	var xargs []string

	cbuildfile = path.Clean(strings.TrimSpace(cbuildfile))
	if cbuildfile == "" || cbuildfile == "." {
		if len(paths) == 0 {
			cbuildfile = "./CBuildfile"
		} else {
			cbuildfile = ""
		}
	}

	if len(cbuildfile) >= 10 {
		if cbuildfile[len(cbuildfile)-10:] != "CBuildfile" {
			fmt.Println("CBuild: given CBuildfile path is not a CBuildfile")
			return
		}
	} else {
		fmt.Println("CBuild: given CBuildfile path is not a CBuildfile")
		return
	}

	outpath = path.Clean(strings.TrimSpace(outpath))

	for p := range paths {
		pathStr := path.Clean(strings.TrimSpace(paths[p]))
		if _, pathsErr := os.Stat(pathStr); errors.Is(pathsErr, os.ErrNotExist) {
			fmt.Printf("file '%s' does not exist\n", pathStr)
			return
		} else if pathsErr != nil {
			// Schrodinger: file may or may not exist. See err for details.
			// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
			fmt.Println(pathsErr.Error())
			return
		}
		paths[p] = pathStr
	}

	if outpath == "" || outpath == "." {
		outpath = "compiled_output"
	}

	if len(paths) > 0 && cbuildfile != "" {
		fmt.Println("CBuild: -b/--build and -c/--cbuild cannot be given at the same time")
		return
	}

	if len(paths) == 0 {
		cbuildfile_abs, _ := filepath.Abs("./CBuildfile")
		if _, cbuildFerr := os.Stat(cbuildfile_abs); cbuildFerr == nil {
			//fmt.Println("CBuildfile exists")
			var cbuildParseErr error
			var info CBuildfileInfo
			paths, info, cbuildParseErr = parseCbuildFile()
			if cbuildParseErr != nil {
				fmt.Println(cbuildParseErr.Error())
				return
			}
			if info.output != nil {
				outpath = info.output.(string)
			}
			if info.compiler != nil {
				COMPILER = info.compiler.(string)
			}
			xargs = info.xargs
		} else if errors.Is(cbuildFerr, os.ErrNotExist) {
			fmt.Println("CBuild: no CBuildfile found")
			return
		} else {
			// Schrodinger: file may or may not exist. See err for details.
			// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
			fmt.Println(cbuildFerr.Error())
			return
		}
	}

	stdout, stderr, _ := compile(paths, xargs, outpath)
	stderr = strings.TrimSpace(stderr)
	//if compileErr != nil {
	//	fmt.Println()
	//}
	if stderr != "" {
		fmt.Println(stderr)
	}
	if stdout != "" {
		fmt.Println(stdout)
	}
}
