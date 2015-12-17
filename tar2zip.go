package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func convert(in io.Reader) {
	archive := tar.NewReader(in)
	for {
		hdr, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading archive: %s\n", err.Error())
			return
		}

		if hdr.Typeflag != tar.TypeDir {
			fmt.Printf("File: <%s> Size: %d\n", hdr.Name, hdr.Size)
		}
	}
}

func decompress(fn string, rdr io.Reader) io.Reader {
	var answer io.Reader

	if strings.HasSuffix(fn, ".tar.gz") ||
		strings.HasSuffix(fn, ".tgz") {
		gzReader, err := gzip.NewReader(rdr)
		if err == nil {
			answer = gzReader
		} else {
			fmt.Printf("%s can't be decompressed: %s\n", fn, err.Error())
			answer = rdr
		}

	} else {
		answer = rdr
	}

	return answer
}

func processFile(fn string) {
	input, err := os.Open(fn)
	if err != nil {
		fmt.Printf("Error opening <%s>: %s\n", fn, err.Error())
		return
	}
	defer input.Close()
	fmt.Printf("Converting %s...\n", fn)

	rdr := decompress(fn, input)
	convert(rdr)

	fmt.Println("Done!")
}

func main() {
	fmt.Println("Hello, soon there will be code here.")
	for _, filename := range os.Args[1:] {
		processFile(filename)
	}
}
