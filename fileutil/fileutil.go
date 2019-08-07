package fileutil

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
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

// FTPDownloadConfig is a config for files to download from a public FTP server
type FTPDownloadConfig struct {
	URL            string
	Port           int
	MaxConnection  int
	BaseDir        string
	DestDir        string
	Files2Download []string
}

// Download downloads files based on the config
func (fc *FTPDownloadConfig) Download() error {
	var wg sync.WaitGroup

	fileChannel := make(chan string)                         // init channel to add files
	wg.Add(1)                                                // adding files to channel
	go addFiles2Channel(fileChannel, fc.Files2Download, &wg) // must be async, otherwise the channel will block this statement

	wg.Add(fc.MaxConnection) // each worker will call wg.Done() when it's done

	if err := os.Mkdir(fc.DestDir, os.ModePerm); err == nil { // create the result directory
		for i := 0; i < fc.MaxConnection; i++ {
			go startDownloadWorker(fileChannel, fc, &wg)
		}
	} else {
		log.Fatal(err)
	}

	wg.Wait()
	return nil
}

func addFiles2Channel(fileChannel chan string, files []string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(fileChannel)   // close the channel when done (so the range operator can operates on channel)
	for _, fn := range files { // add all files to a channel
		fileChannel <- fn
	}
}

func startDownloadWorker(fileChannel chan string, fc *FTPDownloadConfig, wg *sync.WaitGroup) {
	defer wg.Done()
	for fileName := range fileChannel {
		handleDownload(fileName, fc)
	}
}

func handleDownload(fileName string, fc *FTPDownloadConfig) {
	fmt.Println("Downloading:", fileName)

	ftpURL := fmt.Sprintf("%v:%v", fc.URL, fc.Port)
	c, err := ftp.Dial(ftpURL, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	defer c.Quit()

	err = c.Login("anonymous", "anonymous")
	if err != nil {
		log.Fatal(err)
	}

	ftpDir := fc.BaseDir

	err = c.ChangeDir(ftpDir)
	if err != nil {
		log.Fatal(err)
	}

	res, err := c.Retr(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()

	outFile, err := os.Create(filepath.Join(fc.DestDir, fileName))
	defer outFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(outFile, res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Downloaded:", fileName)
}
