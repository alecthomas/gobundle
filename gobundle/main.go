// Copyright 2012 Alec Thomas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	flag "github.com/ogier/pflag"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const (
	USAGE = `usage: {{.Name}} <path> [<path> ...]

This utility and library can be used to compile static files into a binary.

To bundle a set of files:

    {{.Name}} --package=entries --target=entries.go \
        /etc/passwd /etc/hosts /etc/services

To bundle a whole directory:

    {{.Name}} --recursive --package=etc --target=etc.go /etc

For larger bundles it is highly recommended that the files be compressed,
possibly in conjunction with other flags that modify the bundle behavior:

    {{.Name}} --recursive --compress --uncompress_on_init --retain_uncompressed \
        --package=etc --target=etc.go /etc

Flags:
`
	HEADER_TEMPLATE = `package {{ .Package }}

import (
    "github.com/alecthomas/gobundle"
)

var {{ .Bundle | ToTitle }}Bundle *gobundle.Bundle = gobundle.NewBuilder("{{ .Bundle }}"){{if .Compressed}}.Compressed(){{ if .RetainUncompressed }}.RetainUncompressed(){{ end }}{{ if .UncompressOnInit }}.UncompressOnInit(){{ end }}{{end}}`
	FILE_TEMPLATE = `.Add(
    "{{ .Name }}", {{ .Content }},
)`
	FOOTER_TEMPLATE = `.Build()
`
)

var (
	targetFlag             = flag.String("target", "", "target bundle filename to generate")
	pkgFlag                = flag.String("package", "", "target package name (inferred from --target if not provided)")
	bundleFlag             = flag.String("bundle", "", "bundle name (inferred from --package if not provided)")
	excludeFlag            = flag.String("exclude", "", "list of globs to exclude from")
	recurseFlag            = flag.BoolP("recursive", "r", false, "recursively add files")
	compressFlag           = flag.BoolP("compress", "c", false, "compress files before encoding")
	retainUncompressedFlag = flag.BoolP("retain_uncompressed", "u", false, "whether to retain the uncompressed copy on initial access")
	uncompressOnInitFlag   = flag.BoolP("uncompress_on_init", "i", false, "whether to uncompress files on package init()")
	encodeAsBytesFlag      = flag.BoolP("encode_as_bytes", "b", false, "whether to encode as bytes or the default, escaped strings")

	funcMap = map[string]interface{}{"ToTitle": strings.Title}
	usage   = template.Must(template.New("usage").Parse(USAGE))
	header  = template.Must(template.New("header").Funcs(funcMap).Parse(HEADER_TEMPLATE))
	file    = template.Must(template.New("file").Parse(FILE_TEMPLATE))
	footer  = template.Must(template.New("footer").Parse(FOOTER_TEMPLATE))
)

type BundleContext struct {
	Package            string
	Bundle             string
	Compressed         bool
	RetainUncompressed bool
	UncompressOnInit   bool
}

type FileContext struct {
	Name    string
	Id      int
	Content string
}

type BufferCloser struct {
	bytes.Buffer
}

func (b *BufferCloser) Close() error {
	return nil
}

func writeEntry(f *os.File, base, path string) {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	buffer := &BufferCloser{}
	var write io.WriteCloser = buffer
	if *compressFlag {
		write, err = zlib.NewWriterLevel(buffer, 9)
		if err != nil {
			panic(err)
		}
	}
	_, err = io.Copy(write, fp)
	if err != nil {
		panic(err)
	}
	fp.Close()
	write.Close()
	ctx := &FileContext{
		Name:    strings.TrimLeft(path, "/"),
		Content: formatContent(buffer.Bytes()),
	}
	file.Execute(f, ctx)
}

func formatContent(buffer []byte) string {
	if *encodeAsBytesFlag {
		return fmt.Sprintf("%#v", buffer)
	}
	return "[]byte(" + strconv.QuoteToASCII(string(buffer)) + ")"
}

func fatal(f string, args ...interface{}) {
	fmt.Errorf(f, args...)
	os.Exit(1)
}

func fatalif(state bool, f string, args ...interface{}) {
	if state {
		fatal(f, args...)
	}
}

func exclude(path string, excludes []string) bool {
	for _, e := range excludes {
		if match, _ := filepath.Match(e, path); match {
			return true
		}
	}
	return false
}

func main() {
	flag.Usage = func() {
		type UsageContext struct {
			Name string
		}
		ctx := &UsageContext{
			Name: filepath.Base(os.Args[0]),
		}
		usage.Execute(os.Stderr, ctx)
		flag.PrintDefaults()
	}
	flag.Parse()
	fatalif(*targetFlag == "", "")
	f, err := os.Create(*targetFlag)
	if err != nil {
		panic(err)
	}
	target, err := filepath.Abs(*targetFlag)
	if err != nil {
		panic(err)
	}
	pkg := path.Base(path.Dir(target))
	if *pkgFlag != "" {
		pkg = *pkgFlag
	}
	name := pkg
	if *bundleFlag != "" {
		name = *bundleFlag
	}
	ctx := &BundleContext{
		Package:            pkg,
		Bundle:             name,
		Compressed:         *compressFlag,
		RetainUncompressed: *retainUncompressedFlag,
		UncompressOnInit:   *uncompressOnInitFlag,
	}
	header.Execute(f, ctx)
	excludes := strings.Split(*excludeFlag, ",")

	for _, root := range flag.Args() {
		if exclude(root, excludes) {
			continue
		}
		if *recurseFlag {
			filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if exclude(path, excludes) {
					return filepath.SkipDir
				}
				if !info.IsDir() {
					fmt.Println(path)
					writeEntry(f, root, path)
				}
				return nil
			})
		} else {
			fmt.Println(root)
			writeEntry(f, root, root)
		}
	}
	footer.Execute(f, ctx)
}
