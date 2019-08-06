package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ahub.mbni.org/fileutil"
	"ahub.mbni.org/rnaseq"
	"github.com/jlaffaye/ftp"
)

type ftpConfig struct {
	URL            string
	Port           int
	BaseDir        string
	Files2Download []string
}

func main() {
	// parse command line arguments
	rVerPtr := flag.String("g", "30", "Gene code release version")
	speciesPtr := flag.String("s", "hs", "Species")
	flag.Parse()

	filePrefix := fmt.Sprintf("gencode.v%v", *rVerPtr)

	geneAsemblyMapping := map[string]map[string]string{
		"hs": map[string]string{
			"30": "GRCh38",
		},
		"mm": map[string]string{
			"M22": "GRCm38",
		},
	}

	faFileName := geneAsemblyMapping[*speciesPtr][*rVerPtr] + ".primary_assembly.genome.fa"
	gtfFileName := filePrefix + ".primary_assembly.annotation.gtf"
	txfaFileName := filePrefix + ".transcripts.fa"

	faGZFileName := faFileName + ".gz"
	gtfGZFileName := gtfFileName + ".gz"
	txfaGZFileName := txfaFileName + ".gz"

	files2Download := []string{
		gtfGZFileName,
		filePrefix + ".polyAs.gtf.gz",
		filePrefix + ".2wayconspseudos.gtf.gz",
		filePrefix + ".tRNAs.gtf.gz",
		txfaGZFileName,
		faGZFileName,
	}

	speciesMapping := map[string]string{
		"hs": "Gencode_human",
		"mm": "Gencode_mouse",
	} // map from species short name to full url

	ftpDir := fmt.Sprintf("pub/databases/gencode/%v/release_%v", speciesMapping[*speciesPtr], strings.ToUpper(*rVerPtr))

	fc := ftpConfig{"ftp.ebi.ac.uk", 21, ftpDir, files2Download}

	fileChannel := make(chan string) // make channel to add files

	var wg sync.WaitGroup
	wg.Add(1)
	go addFiles2Channel(fileChannel, files2Download, &wg) // must be async, otherwise the channel will block this statement

	maxConnection := 6    // number of connections allowed simultaneously
	wg.Add(maxConnection) // each work will call wg.Done() when it's done

	resDir := "gencode_" + *speciesPtr + "_" + *rVerPtr

	if err := os.Mkdir(resDir, os.ModePerm); err == nil {
		for i := 0; i < maxConnection; i++ {
			go startDownloadWorker(fileChannel, fc, resDir, &wg)
		}
	} else {
		log.Fatal(err)
	}

	wg.Wait()

	wg.Add(3) // wait decompressing threads

	faFilePath := filepath.Join(resDir, faFileName)
	gtfFilePath := filepath.Join(resDir, gtfFileName)
	txfaFilePath := filepath.Join(resDir, txfaFileName)

	// decompress gz files required to build index
	go decompressGZFile(filepath.Join(resDir, gtfGZFileName), gtfFilePath, &wg)
	go decompressGZFile(filepath.Join(resDir, faGZFileName), faFilePath, &wg)
	go decompressGZFile(filepath.Join(resDir, txfaGZFileName), txfaFilePath, &wg)

	wg.Wait()

	fmt.Println("All files downloaded and uncompressed")

	gc := rnaseq.GenomeConfig{BaseDir: resDir, FAFileName: faFileName, GTFFileName: gtfFileName, TXFAFileName: txfaFileName}

	rnaseq.BuildRNASeqIndex("salmon", &gc)
}

func addFiles2Channel(fileChannel chan string, files []string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(fileChannel)   // close the channel when done
	for _, fn := range files { // add all files to a channel
		fileChannel <- fn
	}
}

func startDownloadWorker(fileChannel chan string, fc ftpConfig, resDir string, wg *sync.WaitGroup) {
	defer wg.Done()
	for fn := range fileChannel {
		handleDownload(fn, resDir, fc)
	}
}

func handleDownload(fileName string, resDir string, fc ftpConfig) {
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

	outFile, err := os.Create(filepath.Join(resDir, fileName))
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

func decompressGZFile(gzFilePath, resFilePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Unzipping", gzFilePath)
	fileutil.DecompressGZFile(gzFilePath, resFilePath)
	fmt.Println("Unzipped to", resFilePath)
}
