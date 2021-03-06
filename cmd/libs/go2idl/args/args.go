/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package args has common command-line flags for generation programs.
package args

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/kubernetes/cmd/libs/go2idl/generator"
	"k8s.io/kubernetes/cmd/libs/go2idl/namer"
	"k8s.io/kubernetes/cmd/libs/go2idl/parser"

	"github.com/spf13/pflag"
)

// Default returns a defaulted GeneratorArgs. You may change the defaults
// before calling AddFlags.
func Default() *GeneratorArgs {
	generatorArgs := &GeneratorArgs{
		OutputBase:       DefaultSourceTree(),
		GoHeaderFilePath: filepath.Join(DefaultSourceTree(), "k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt"),
	}
	generatorArgs.AddFlags(pflag.CommandLine)
	return generatorArgs
}

// GeneratorArgs has arguments that are passed to generators.
type GeneratorArgs struct {
	// Which directories to parse.
	InputDirs []string

	// Source tree to write results to.
	OutputBase string

	// Package path within the source tree.
	OutputPackagePath string

	// Where to get copyright header text.
	GoHeaderFilePath string

	// If true, only verify, don't write anything.
	VerifyOnly bool

	// Any custom arguments go here
	CustomArgs interface{}
}

func (g *GeneratorArgs) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVarP(&g.InputDirs, "input-dirs", "i", g.InputDirs, "Comma-separated list of import paths to get input types from.")
	fs.StringVarP(&g.OutputBase, "output-base", "o", g.OutputBase, "Output base; defaults to $GOPATH/src/ or ./ if $GOPATH is not set.")
	fs.StringVarP(&g.OutputPackagePath, "output-package", "p", g.OutputPackagePath, "Base package path.")
	fs.StringVarP(&g.GoHeaderFilePath, "go-header-file", "h", g.GoHeaderFilePath, "File containing boilerplate header text. The string YEAR will be replaced with the current 4-digit year.")
	fs.BoolVar(&g.VerifyOnly, "verify-only", g.VerifyOnly, "If true, only verify existing output, do not write anything.")
}

// LoadGoBoilerplate loads the boilerplate file passed to --go-header-file.
func (g *GeneratorArgs) LoadGoBoilerplate() ([]byte, error) {
	b, err := ioutil.ReadFile(g.GoHeaderFilePath)
	if err != nil {
		return nil, err
	}
	b = bytes.Replace(b, []byte("YEAR"), []byte(strconv.Itoa(time.Now().Year())), -1)
	return b, nil
}

// NewBuilder makes a new parser.Builder and populates it with the input
// directories.
func (g *GeneratorArgs) NewBuilder() (*parser.Builder, error) {
	b := parser.New()
	for _, d := range g.InputDirs {
		if err := b.AddDir(d); err != nil {
			return nil, fmt.Errorf("unable to add directory %q: %v", d, err)
		}
	}
	return b, nil
}

// DefaultSourceTree returns the /src directory of the first entry in $GOPATH.
// If $GOPATH is empty, it returns "./". Useful as a default output location.
func DefaultSourceTree() string {
	paths := strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))
	if len(paths) > 0 && len(paths[0]) > 0 {
		return filepath.Join(paths[0], "src")
	}
	return "./"
}

// Execute implements main().
// If you don't need any non-default behavior, use as:
// args.Default().Execute(...)
func (g *GeneratorArgs) Execute(nameSystems namer.NameSystems, defaultSystem string, pkgs func(*generator.Context, *GeneratorArgs) generator.Packages) error {
	pflag.Parse()

	b, err := g.NewBuilder()
	if err != nil {
		return fmt.Errorf("Failed making a parser: %v", err)
	}

	c, err := generator.NewContext(b, nameSystems, defaultSystem)
	if err != nil {
		return fmt.Errorf("Failed making a context: %v", err)
	}

	c.Verify = g.VerifyOnly
	packages := pkgs(c, g)
	if err := c.ExecutePackages(g.OutputBase, packages); err != nil {
		return fmt.Errorf("Failed executing generator: %v", err)
	}

	return nil
}
