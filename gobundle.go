package gobundle

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

type file struct {
	data         []byte
	uncompressed []byte
}

type Bundle struct {
	Name               string
	names              []string
	files              map[string]*file
	compressed         bool
	retainUncompressed bool
}

type Builder struct {
	bundle           *Bundle
	uncompressOnInit bool
}

func NewBundle(name string) *Bundle {
	return &Bundle{
		Name:  name,
		files: make(map[string]*file),
	}
}

func NewBuilder(name string) *Builder {
	return &Builder{
		bundle: &Bundle{
			Name:  name,
			files: make(map[string]*file),
		},
	}
}

func (b *Builder) Build() *Bundle {
	names := make([]string, 0, len(b.bundle.files))
	for key, _ := range b.bundle.files {
		names = append(names, key)
	}
	names = sort.Strings(names)
	b.bundle.names = names
	if b.bundle.compressed && b.uncompressOnInit {
		b.bundle.compressed = false
	}
	Bundles[b.bundle.Name] = b.bundle
	return b.bundle
}

func (b *Builder) Compressed() *Builder {
	b.bundle.compressed = true
	return b
}

func (b *Builder) RetainUncompressed() *Builder {
	b.bundle.retainUncompressed = true
	return b
}

func (b *Builder) UncompressOnInit() *Builder {
	b.uncompressOnInit = true
	return b
}

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

// Return list of files in bundle.
func (b *Bundle) Files() []string {
	return b.names
}

// Return the bytes for a file.
func (b *Bundle) Bytes(path string) ([]byte, error) {
	file := b.files[path]
	if file == nil {
		return nil, os.ErrNotExist
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
				file.data = wb.Bytes()
			}
			return wb.Bytes(), nil
		} else {
			return file.uncompressed, nil
		}
	}
	return file.data, nil
}

// Open a bundle file for reading.
func (b *Bundle) Open(path string) (io.Reader, error) {
	file := b.files[path]
	if file == nil {
		return nil, os.ErrNotExist
	}
	if b.compressed {
		if file.uncompressed == nil {
			f, err := zlib.NewReader(bytes.NewReader(file.data))
			if err != nil {
				return nil, err
			}
			return f, nil
		} else {
			return bytes.NewReader(file.uncompressed), nil
		}
	}
	return bytes.NewReader(file.data), nil
}

// Map of registered bundles
var Bundles map[string]*Bundle
