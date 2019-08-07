package rnaseq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"ahub.mbni.org/fileutil"
)

// GenomeConfig contains information about the genome directory
type GenomeConfig struct {
	BaseDir, FAFileName, GTFFileName, TXFAFileName string
}

func (gc *GenomeConfig) getFAFilePath() string {
	return filepath.Join(gc.BaseDir, gc.FAFileName)
}

func (gc *GenomeConfig) getGTFFilePath() string {
	return filepath.Join(gc.BaseDir, gc.GTFFileName)
}

func (gc *GenomeConfig) getTXFAFilePath() string {
	return filepath.Join(gc.BaseDir, gc.TXFAFileName)
}

// SequencingAlgorithm is a RNA-seq algorithm interface
type SequencingAlgorithm interface {
	BuildIndex(gc *GenomeConfig) error
}

// STAR algorithm
type STAR struct {
}

// BuildIndex for algorithm STAR
func (st STAR) BuildIndex(gc *GenomeConfig) error {
	starIdxDir := filepath.Join(gc.BaseDir, "star_idx")
	if err := os.Mkdir(starIdxDir, os.ModePerm); err == nil {
		starArgs := []string{
			"--runThreadN", "4", "--runMode", "genomeGenerate", "--genomeDir", starIdxDir,
			"--genomeFastaFiles", gc.getFAFilePath(),
			"--sjdbGTFfile", gc.getGTFFilePath(),
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
func (sa Salmon) BuildIndex(gc *GenomeConfig) error {
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
func BuildRNASeqIndex(algorithm string, gc *GenomeConfig) error {
	var seqAlgorithm SequencingAlgorithm

	if algorithm == "star" {
		seqAlgorithm = STAR{}
	} else if algorithm == "salmon" {
		seqAlgorithm = Salmon{}
	} else {
		return errors.New("Unsupported algorithm")
	}

	return seqAlgorithm.BuildIndex(gc)
}

// BuildSalmonIndexByKmer builds index for a specific kmer
func BuildSalmonIndexByKmer(kmer string, gc *GenomeConfig) error {
	salmonIdxDir := filepath.Join(gc.BaseDir, fmt.Sprintf("salmon_k%v_idx", kmer))
	salmonArgs := []string{
		"index", "-t", gc.getTXFAFilePath(), "-i", salmonIdxDir, "--type", "quasi", "-k", kmer,
	}
	fmt.Printf("Salmon indexing k%v\n", kmer)
	_, err := exec.Command("salmon", salmonArgs...).Output()
	if err != nil {
		return err
	}
	fmt.Println("Finished indexing", kmer)
	return nil
}

func buildSalmonIndexByKmer(kmer string, gc *GenomeConfig, wg *sync.WaitGroup) error {
	defer wg.Done()
	return BuildSalmonIndexByKmer(kmer, gc)
}

// SequencingIndexBuilder is an interface that has methods to build sequencing index
type SequencingIndexBuilder interface {
	BuildIndex() error
	DownloadGenomeFiles() error
}

// GencodeIndexBuilder builds index for Gencode
type GencodeIndexBuilder struct {
	Species, GencodeVersion string
}

// GenomeAssembly returns the corresponding genome assembly given the species and gencode version
func (gb *GencodeIndexBuilder) GenomeAssembly() (res string, err error) {
	genomeAssemblyMapping := map[string]map[string]string{
		"hs": map[string]string{
			"30": "GRCh38",
		},
		"mm": map[string]string{
			"M22": "GRCm38",
		},
	}

	if subMapping, ok := genomeAssemblyMapping[gb.Species]; ok {
		if res, ok = subMapping[gb.GencodeVersion]; ok {
			err = nil
			return
		}
	}

	err = fmt.Errorf("Can't find genome assembly for '%v' of Gencode version '%v'", gb.Species, gb.GencodeVersion)
	return
}

// DownloadGenomeFiles implements the interface method
func (gb *GencodeIndexBuilder) DownloadGenomeFiles() error {
	genomeAssembly, err := gb.GenomeAssembly() // figure out the genome assembly

	if err != nil {
		return err
	}

	log.Printf("Gencode downloading files for genome assembly '%v'\n", genomeAssembly)

	filePrefix := fmt.Sprintf("gencode.v%v", gb.GencodeVersion)

	faFileName := genomeAssembly + ".primary_assembly.genome.fa"
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

	ftpDir := fmt.Sprintf("pub/databases/gencode/%v/release_%v", speciesMapping[gb.Species], strings.ToUpper(gb.GencodeVersion))

	resDir := "gencode_" + gb.Species + "_" + gb.GencodeVersion

	fc := fileutil.FTPDownloadConfig{"ftp.ebi.ac.uk", 21, 3, ftpDir, resDir, files2Download}

	err = fc.Download()

	if err != nil {
		return err
	}

	faFilePath := filepath.Join(resDir, faFileName)
	gtfFilePath := filepath.Join(resDir, gtfFileName)
	txfaFilePath := filepath.Join(resDir, txfaFileName)

	var wg sync.WaitGroup
	wg.Add(3) // wait decompressing threads

	// decompress gz files required to build index
	go decompressGZFile(filepath.Join(resDir, gtfGZFileName), gtfFilePath, &wg)
	go decompressGZFile(filepath.Join(resDir, faGZFileName), faFilePath, &wg)
	go decompressGZFile(filepath.Join(resDir, txfaGZFileName), txfaFilePath, &wg)

	wg.Wait()

	fmt.Println("All files downloaded and uncompressed")
	return nil
}

// BuildIndex implements SequencingIndexBuilder
func (gb *GencodeIndexBuilder) BuildIndex() error {

	err := gb.DownloadGenomeFiles() // download genome files before proceeding

	if err != nil {
		return err
	}

	gc := GenomeConfig{BaseDir: resDir, FAFileName: faFileName, GTFFileName: gtfFileName, TXFAFileName: txfaFileName}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		rnaseq.BuildRNASeqIndex("star", &gc)
	}()

	go func() {
		defer wg.Done()
		rnaseq.BuildRNASeqIndex("salmon", &gc)
	}()

	wg.Wait()
	genomeAssembly, err := gb.GenomeAssembly()

	log.Printf("Gencode building index for genome assembly '%v'\n", genomeAssembly)

	return nil
}

func decompressGZFile(gzFilePath, resFilePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Unzipping", gzFilePath)
	fileutil.DecompressGZFile(gzFilePath, resFilePath)
	fmt.Println("Unzipped to", resFilePath)
}

// EnsemblIndexBuilder build index for Ensembl
type EnsemblIndexBuilder struct {
}

// BuildIndex implements SequencingIndexBuilder
func (eb *EnsemblIndexBuilder) BuildIndex() error {
	return nil
}
