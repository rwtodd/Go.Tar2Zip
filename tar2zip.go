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
	"path/filepath"
	"strings"
	"time"
)

type compressionType int

const (
	gzComp compressionType = iota
	bz2Comp
	noComp
)

var verbose = flag.Bool("verbose", false, "print details about the files")

func convertOneFile(fn string, ft time.Time, in io.Reader, out *zip.Writer) {
	zipHeader := &zip.FileHeader{
		Name:   fn,
		Method: zip.Deflate,
	}
	zipHeader.SetModTime(ft)

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

func zipSingle(fn string, in io.Reader, out io.Writer) {
	outzip := zip.NewWriter(out)
	defer outzip.Close()

	if *verbose {
		fmt.Printf("Converting single non-tar file %s.\n", fn)
	}

	convertOneFile(fn, time.Now(), in, outzip)
}

func convertTar(in io.Reader, out io.Writer) {
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
			break
		}

		switch hdr.Typeflag {
		case tar.TypeReg, tar.TypeRegA, tar.TypeDir, tar.TypeGNUSparse:
			if *verbose {
				fmt.Printf("Converting <%s>, size %d\n", hdr.Name, hdr.Size)
			}
			convertOneFile(hdr.Name, hdr.ModTime, intar, outzip)
		default:
			fmt.Printf("Skipping entry: <%s> with Unsupported Type: %d\n", hdr.Name, hdr.Typeflag)
		}
	}
}

func decompress(comp compressionType, rdr io.Reader) io.Reader {
	var answer io.Reader

	switch comp {
	case gzComp:
		gzReader, err := gzip.NewReader(rdr)
		if err == nil {
			answer = gzReader
		} else {
			fmt.Printf("File can't be decompressed: %s\n", err.Error())
			answer = rdr
		}

	case bz2Comp:
		answer = bzip2.NewReader(rdr)

	default:
		answer = rdr
	}

	return answer
}

func analyzeInput(fn string) (isTar bool, comp compressionType, basefn string) {
	isTar = true
	comp = noComp
	var toStrip = 0

	switch {
	case strings.HasSuffix(fn, ".tar"):
		toStrip = 4
	case strings.HasSuffix(fn, ".tgz"):
		comp = gzComp
		toStrip = 4
	case strings.HasSuffix(fn, ".tar.gz"):
		comp = gzComp
		toStrip = 7
	case strings.HasSuffix(fn, ".tar.bz2"):
		comp = bz2Comp
		toStrip = 8
	case strings.HasSuffix(fn, ".bz2"):
		toStrip = 4
		comp = bz2Comp
		isTar = false
	case strings.HasSuffix(fn, ".gz"):
		toStrip = 3
		comp = gzComp
		isTar = false
	default:
		isTar = false
	}

	basefn = fn[:len(fn)-toStrip]

	return
}

func processFile(fn string) {
	input, err := os.Open(fn)
	if err != nil {
		fmt.Printf("Error opening <%s>: %s\n", fn, err.Error())
		return
	}
	defer input.Close()

	isTar, comp, basename := analyzeInput(fn)
	output, err := os.OpenFile(basename+".zip", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Error opening output file! %s\n", err.Error())
		return
	}
	defer output.Close()

	fmt.Printf("Converting %s...\n", fn)
	rdr := decompress(comp, input)
	if isTar {
		convertTar(rdr, output)
	} else {
		zipSingle(filepath.Base(basename), rdr, output)
	}

	fmt.Println("Done!")
}

func main() {
	flag.Parse()
	for _, filename := range flag.Args() {
		processFile(filename)
	}
}
