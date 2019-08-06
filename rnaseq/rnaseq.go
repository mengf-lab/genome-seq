package rnaseq

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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
			"--runThreadN", "16", "--runMode", "genomeGenerate", "--genomeDir", starIdxDir,
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
