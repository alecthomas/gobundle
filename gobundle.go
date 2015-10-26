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

package gobundle

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
)

var (
	// Bundles represents a collection of bundles
	Bundles = &BundleOfBundles{bundles: make(map[string]*Bundle)}
)

type file struct {
	data         []byte
	uncompressed []byte
}

// BundleOfBundles is a map of registered bundles
type BundleOfBundles struct {
	bundles map[string]*Bundle
}

// Bundle represents the contents of a bundle
type Bundle struct {
	Name               string
	names              []string
	files              map[string]*file
	compressed         bool
	retainUncompressed bool
}

// Builder represents a bundle and whether to decompress
// that bundle on init
type Builder struct {
	bundle           *Bundle
	uncompressOnInit bool
}

// NewBundle returns a new bundle
func NewBundle(name string) *Bundle {
	return &Bundle{
		Name:  name,
		files: make(map[string]*file),
	}
}

// NewBuilder returns a new builder
func NewBuilder(name string) *Builder {
	return &Builder{
		bundle: &Bundle{
			Name:  name,
			files: make(map[string]*file),
		},
	}
}

// Build builds a bundle
func (b *Builder) Build() *Bundle {
	names := make([]string, 0, len(b.bundle.files))
	for key := range b.bundle.files {
		names = append(names, key)
	}
	sort.Strings(names)
	b.bundle.names = names
	if b.bundle.compressed && b.uncompressOnInit {
		b.bundle.compressed = false
	}
	Bundles.Add(b.bundle)
	return b.bundle
}

// Compressed sets compression mode to true
func (b *Builder) Compressed() *Builder {
	b.bundle.compressed = true
	return b
}

// RetainUncompressed sets the retainUncompressed
// field to true
func (b *Builder) RetainUncompressed() *Builder {
	b.bundle.retainUncompressed = true
	return b
}

// UncompressOnInit sets the uncompressOnInit field
// to true
func (b *Builder) UncompressOnInit() *Builder {
	b.uncompressOnInit = true
	return b
}

// Add adds files to a builder.
func (b *Builder) Add(path string, data []byte) *Builder {
	if b.bundle.compressed && b.uncompressOnInit {
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		data, err = ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
	}
	b.bundle.files[path] = &file{
		data: data,
	}
	return b
}

// Files returns a list of files in bundle.
func (b *Bundle) Files() []string {
	return b.names
}

func errNotExist(path string) error {
	return fmt.Errorf("file does not exist: %s", path)
}

// Bytes returns the bytes for a file.
func (b *Bundle) Bytes(path string) ([]byte, error) {
	file := b.files[path]
	if file == nil {
		return nil, errNotExist(path)
	}
	if b.compressed {
		if file.uncompressed == nil {
			r, err := zlib.NewReader(bytes.NewReader(file.data))
			if err != nil {
				return nil, err
			}
			wb := &bytes.Buffer{}
			_, err = io.Copy(wb, r)
			if err != nil {
				return nil, err
			}
			if b.retainUncompressed {
				file.uncompressed = wb.Bytes()
			}
			return wb.Bytes(), nil
		}
		return file.uncompressed, nil
	}
	return file.data, nil
}

// Open opens a bundle file for reading.
func (b *Bundle) Open(path string) (io.Reader, error) {
	file := b.files[path]
	if file == nil {
		return nil, errNotExist(path)
	}
	if b.compressed {
		if file.uncompressed == nil {
			f, err := zlib.NewReader(bytes.NewReader(file.data))
			if err != nil {
				return nil, err
			}
			return f, nil
		}
		return bytes.NewReader(file.uncompressed), nil
	}
	return bytes.NewReader(file.data), nil
}

// Add adds a bundle of bundles to a bundle.
func (b *BundleOfBundles) Add(bundle *Bundle) {
	b.bundles[bundle.Name] = bundle
}

// Bundle returns the named Bundle, or nil if not found.
func (b *BundleOfBundles) Bundle(name string) *Bundle {
	if a, ok := b.bundles[name]; ok {
		return a
	}
	return nil
}

// Bundles returns a list of all available bundles.
func (b *BundleOfBundles) Bundles() []*Bundle {
	bundles := make([]*Bundle, 0, len(b.bundles))
	for _, n := range b.bundles {
		bundles = append(bundles, n)
	}
	return bundles
}
