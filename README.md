# Bundle static resources with Go binaries

This utility bundles a set of files into a virtual file system that can be compiled directly into a Go binary.

## Usage

First, get the command-line tool:

	go get github.com/alecthomas/gobundle/gobundle
	
Generate a bundle from all files under `/usr/share/doc/nasm`:

	gobundle --uncompressoninit --package=nasmdocs --bundle=nasmdocs \
		 --target=./nasmdocs.go /usr/share/doc/nasm
		 
This will generate a file something like this:

	