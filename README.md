# tar2zip
Utility to convert tar.(gz|bz2) files to ZIP files

When working on Windows, I prefer to deal with ZIP files
over tar.gz archives.  It's just nicer, since ZIP files
are integrated into the Explorer by default, without needing
to install any programs like 7-zip.

Since golang comes with packages for tar and zip, it was 
very easy to throw together this converter. Now, when I
get a typical tar archive, I can just convert it and 
go about my business.

## Go Get

You should be able to install this program by saying:

    go get github.com/rwtodd/tar2zip

## Usage:

`tar2zip [-verbose] infile1 infile2 ...`

The `verbose` flag causes the program to write a line about 
every converted file in the archives

