package fileutil

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(url, resFilePath string) error {
	// Get data at the url
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the body to file
	return writeToFile(resFilePath, resp.Body)
}

// DecompressGZFile decompress a .gz file and save to another
func DecompressGZFile(gzFilePath, resFilePath string) error {
	fi, err := os.Open(gzFilePath) // open file as a file handler
	if err != nil {
		return err
	}
	defer fi.Close()
	fz, err := gzip.NewReader(fi)
	if err != nil {
		return err
	}
	defer fz.Close()

	return writeToFile(resFilePath, fz)
}

// reader from a Reader object and write to a file
func writeToFile(filePath string, src io.Reader) error {
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, src)
	return err
}
