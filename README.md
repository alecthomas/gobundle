# Bundle static resources with Go binaries

This utility bundles a set of files into a virtual file system that can be compiled directly into a Go binary.

## Installation

First, get the command-line tool:

    go get github.com/alecthomas/gobundle \
        github.com/alecthomas/gobundle/gobundle

## Usage

### Generating a bundle

Generate a bundle from all files under `/usr/share/doc/nasm`:

    gobundle --recursive --compress --uncompress_on_init --package=mybundle \
        --target=/tmp/gobundle.go --exclude='.git/*' .
		 
This will generate a file something like this:

    package mybundle

    import (
        "github.com/alecthomas/gobundle"
    )

    var MybundleBundle *gobundle.Bundle = gobundle.NewBuilder("mybundle").Compressed().UncompressOnInit().Add(
        "README.md", "x\xdaL\x91\xb1\xae\xdb...
    ).Add(
        "gobundle/main.go", "x\u0694W[o\xdb\...
    ).Add(
        "gobundle.go", "x\xda\xccVMo\x9b@\x1...
    ).Build()

### Accessing a bundle

The bundle can then be accessed either directly via the generated exported variable `mybundle.MybundleBundle`:

    bundle := mybundle.MybundleBundle

Or via the global registry (which can also be used to iterate over all bundles):

    bundle := gobundle.Bundles.Bundle("mybundle")

Once you have a bundle you can traverse the files in the bundle, retrieve the raw bytes for a file, or open a file as an `io.Reader`:

    for _, name := range bundle.Files() {
        r, _ := bundle.Open(name)
        b, _ := ioutil.ReadAll(r)
        fmt.Printf("file %s has length %d\n", name, len(b))
    }