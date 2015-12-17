package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var verbose = flag.Bool("verbose", false, "print details about the files")

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

		switch hdr.Typeflag {
		case tar.TypeReg, tar.TypeRegA, tar.TypeDir, tar.TypeGNUSparse:
			if *verbose {
				fmt.Printf("Converting <%s>, size %d\n", hdr.Name, hdr.Size)
			}
			convertOneFile(hdr, intar, outzip)
		default:
			fmt.Printf("Skipping entry: <%s> with Unsupported Type: %d\n", hdr.Name, hdr.Typeflag)
		}
	}
}

func decompress(fn string, rdr io.Reader) io.Reader {
	var answer io.Reader

	switch {
	case strings.HasSuffix(fn, ".tar.gz"), strings.HasSuffix(fn, ".tgz"):
		gzReader, err := gzip.NewReader(rdr)
		if err == nil {
			answer = gzReader
		} else {
			fmt.Printf("%s can't be decompressed: %s\n", fn, err.Error())
			answer = rdr
		}

	case strings.HasSuffix(fn, ".tar.bz2"):
		answer = bzip2.NewReader(rdr)

	default:
		answer = rdr
	}

	return answer
}

func zipname(fn string) (zfn string) {
	switch {
	case strings.HasSuffix(fn, ".tgz"), strings.HasSuffix(fn, ".tar"):
		zfn = fn[:len(fn)-3] + "zip"
	case strings.HasSuffix(fn, ".tar.gz"):
		zfn = fn[:len(fn)-6] + "zip"
	case strings.HasSuffix(fn, ".tar.bz2"):
		zfn = fn[:len(fn)-7] + "zip"
	default:
		zfn = fn + ".zip"
	}
	return
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
	flag.Parse()
	for _, filename := range flag.Args() {
		processFile(filename)
	}
}
