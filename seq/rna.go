package seq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/keqiang/filenet"
)

// RNAseqIndexBuilder is a general interface that has methods to build sequencing index
type RNAseqIndexBuilder interface {
	BaseDir() string
	GenomeAssembly() string
	FAFileName() string
	GTFFileName() string
	TXFAFileName() string
	BuildIndex() error
	DownloadGenomeFiles() error
}

// GencodeIndexBuilder builds index for Gencode
type GencodeIndexBuilder struct {
	Species, GencodeVersion string
}

// BaseDir returns the base directory if this builder runs builds
func (gb GencodeIndexBuilder) BaseDir() string {
	return "gencode_" + gb.Species + "_" + gb.GencodeVersion
}

// GenomeAssembly returns the corresponding genome assembly given the species and gencode version
func (gb GencodeIndexBuilder) GenomeAssembly() string {
	genomeAssemblyMapping := map[string]map[string]string{
		"hs": map[string]string{
			"30": "GRCh38",
		},
		"mm": map[string]string{
			"M22": "GRCm38",
		},
	}

	if subMapping, ok := genomeAssemblyMapping[gb.Species]; ok {
		if res, ok := subMapping[gb.GencodeVersion]; ok {
			return res
		}
	}

	panic(fmt.Sprintf("Can't find genome assembly for '%v' of Gencode version '%v'", gb.Species, gb.GencodeVersion))
}

// FAFileName returns the fa file name
func (gb GencodeIndexBuilder) FAFileName() string {
	return gb.GenomeAssembly() + ".primary_assembly.genome.fa"
}

// FilePrefix returns the file prefix string
func (gb GencodeIndexBuilder) FilePrefix() string {
	return fmt.Sprintf("gencode.v%v", gb.GencodeVersion)
}

// GTFFileName returns the GTF file name
func (gb GencodeIndexBuilder) GTFFileName() string {
	return gb.FilePrefix() + ".primary_assembly.annotation.gtf"
}

// TXFAFileName returns the transcript fa file name
func (gb GencodeIndexBuilder) TXFAFileName() string {
	return gb.FilePrefix() + ".transcripts.fa"
}

// DownloadGenomeFiles implements the interface method
func (gb GencodeIndexBuilder) DownloadGenomeFiles() error {
	genomeAssembly := gb.GenomeAssembly() // figure out the genome assembly

	log.Printf("Gencode downloading files for genome assembly '%v'\n", genomeAssembly)

	filePrefix := gb.FilePrefix()

	faFileName := gb.FAFileName()
	gtfFileName := gb.GTFFileName()
	txfaFileName := gb.TXFAFileName()

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

	ftpDir := fmt.Sprintf("pub/databases/gencode/%v/release_%v", speciesMapping[gb.Species], strings.ToUpper(gb.GencodeVersion))

	resDir := "gencode_" + gb.Species + "_" + gb.GencodeVersion

	fc := filenet.FTPDownloadConfig{
		URL:            "ftp.ebi.ac.uk",
		Port:           21,
		MaxConnection:  3,
		BaseDir:        ftpDir,
		DestDir:        resDir,
		Files2Download: files2Download,
	}

	err := fc.Download()

	if err != nil {
		return err
	}

	log.Println("Gencode files downloaded")

	files2Unzip := make(map[string]string)
	files2Unzip[filepath.Join(resDir, gtfGZFileName)] = filepath.Join(resDir, gtfFileName)
	files2Unzip[filepath.Join(resDir, faGZFileName)] = filepath.Join(resDir, faFileName)
	files2Unzip[filepath.Join(resDir, txfaGZFileName)] = filepath.Join(resDir, txfaFileName)

	log.Println("Gencode unzipping files")

	decompressFiles(files2Unzip, 3)

	log.Println("Gencode files unzipped")

	return nil
}

// BuildIndex implements RNAseqIndexBuilder
func (gb GencodeIndexBuilder) BuildIndex() error {

	err := gb.DownloadGenomeFiles() // download and uncompress genome files before proceeding

	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		BuildRNASeqIndex("star", gb)
	}()

	go func() {
		defer wg.Done()
		BuildRNASeqIndex("salmon", gb)
	}()

	wg.Wait()
	return nil
}

func decompressFiles(files2Uncompress map[string]string, maxWorkerNumber int) {
	// TODO implement actuall worker logic
	if maxWorkerNumber > 5 {
		maxWorkerNumber = 5
	}
	var wg sync.WaitGroup
	wg.Add(len(files2Uncompress))
	for gzFile, unzippedFile := range files2Uncompress {
		go func(src, dst string) {
			defer wg.Done()
			log.Printf("Unzipping file '%v'\n", src)
			filenet.GZipDecompress(src, dst)
			log.Printf("Unzipped to file '%v'\n", dst)
		}(gzFile, unzippedFile)
	}
	wg.Wait()
}

// EnsemblIndexBuilder build index for Ensembl
type EnsemblIndexBuilder struct {
}

// BuildIndex implements RNAseqIndexBuilder
func (eb *EnsemblIndexBuilder) BuildIndex() error {
	return nil
}

// SequencingAlgorithm is a RNA-seq algorithm interface
type SequencingAlgorithm interface {
	BuildIndex(gc RNAseqIndexBuilder) error
}

// STAR algorithm
type STAR struct {
}

// BuildIndex for algorithm STAR
func (st STAR) BuildIndex(sqb RNAseqIndexBuilder) error {
	starIdxDir := filepath.Join(sqb.BaseDir(), "star_idx")
	if err := os.Mkdir(starIdxDir, os.ModePerm); err == nil {
		starArgs := []string{
			"--runThreadN", "4", "--runMode", "genomeGenerate", "--genomeDir", starIdxDir,
			"--genomeFastaFiles", filepath.Join(sqb.BaseDir(), sqb.FAFileName()),
			"--sjdbGTFfile", filepath.Join(sqb.BaseDir(), sqb.GTFFileName()),
		}
		fmt.Println("Running STAR indexing")
		_, err := exec.Command("STAR", starArgs...).Output()
		if err != nil {
			return err
		}
		fmt.Println("Finished indexing")
	} else {
		return err
	}
	return nil
}

// Salmon algorithm
type Salmon struct {
}

// BuildIndex for algorithm Salmon
func (sa Salmon) BuildIndex(gc RNAseqIndexBuilder) error {
	fmt.Println("Running Salmon indexing")
	var wg sync.WaitGroup
	kmers := [6]string{"21", "23", "25", "27", "29", "31"} // all salmon Ks

	for _, kmer := range kmers {
		wg.Add(1)
		go buildSalmonIndexByKmer(kmer, gc, &wg)
	}

	wg.Wait()
	fmt.Println("Finished Salmon indexing")
	return nil
}

// BuildRNASeqIndex builds index for specified algorithm
func BuildRNASeqIndex(algorithm string, sqb RNAseqIndexBuilder) error {
	var seqAlgorithm SequencingAlgorithm

	if algorithm == "star" {
		seqAlgorithm = STAR{}
	} else if algorithm == "salmon" {
		seqAlgorithm = Salmon{}
	} else {
		return errors.New("Unsupported algorithm")
	}

	return seqAlgorithm.BuildIndex(sqb)
}

// BuildSalmonIndexByKmer builds index for a specific kmer
func BuildSalmonIndexByKmer(kmer string, sqb RNAseqIndexBuilder) error {
	salmonIdxDir := filepath.Join(sqb.BaseDir(), fmt.Sprintf("salmon_k%v_idx", kmer))
	salmonArgs := []string{
		"index", "-t", filepath.Join(sqb.BaseDir(), sqb.TXFAFileName()), "-i", salmonIdxDir, "--type", "quasi", "-k", kmer,
	}
	fmt.Printf("Salmon indexing k%v\n", kmer)
	_, err := exec.Command("salmon", salmonArgs...).Output()
	if err != nil {
		return err
	}
	fmt.Println("Finished indexing", kmer)
	return nil
}

func buildSalmonIndexByKmer(kmer string, gc RNAseqIndexBuilder, wg *sync.WaitGroup) error {
	defer wg.Done()
	return BuildSalmonIndexByKmer(kmer, gc)
}
