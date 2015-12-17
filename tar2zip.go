package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func convertOneFile(hdr *tar.Header, in *tar.Reader, out *zip.Writer) {
	zipHeader := &zip.FileHeader{
		Name:   hdr.Name,
		Method: zip.Deflate,
	}
	zipHeader.SetModTime(hdr.ModTime)

	content, err := out.CreateHeader(zipHeader)
	if err != nil {
		fmt.Printf("Error creating zip header: %s\n", err.Error())
		return
	}

	_, err = io.Copy(content, in)
	if err != nil {
		fmt.Printf("Error copying file to zip! %s\n", err.Error())
	}
}

func convert(in io.Reader, out io.Writer) {
	intar := tar.NewReader(in)
	outzip := zip.NewWriter(out)
	defer outzip.Close()

	for {
		hdr, err := intar.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading archive: %s\n", err.Error())
			return
		}

		if hdr.Typeflag == tar.TypeReg ||
			hdr.Typeflag == tar.TypeRegA ||
			hdr.Typeflag == tar.TypeGNUSparse {
			fmt.Printf("File: <%s> Size: %d\n", hdr.Name, hdr.Size)
			convertOneFile(hdr, intar, outzip)
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

func zipname(fn string) string {
	return fn + ".zip" // temporary! FIXME
}

func processFile(fn string) {
	input, err := os.Open(fn)
	if err != nil {
		fmt.Printf("Error opening <%s>: %s\n", fn, err.Error())
		return
	}
	defer input.Close()

	output, err := os.OpenFile(zipname(fn), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Error opening output file! %s\n", err.Error())
		return
	}
	defer output.Close()

	fmt.Printf("Converting %s...\n", fn)
	rdr := decompress(fn, input)
	convert(rdr, output)

	fmt.Println("Done!")
}

func main() {
	fmt.Println("Hello, soon there will be code here.")
	for _, filename := range os.Args[1:] {
		processFile(filename)
	}
}
