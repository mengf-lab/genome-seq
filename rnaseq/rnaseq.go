package rnaseq

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Algorithm is an enum type for RNA-seq algorithm names
type Algorithm string

// Available RNA-seq algorithms
const (
	STAR   Algorithm = "star"
	Salmon Algorithm = "salmon"
)

// NewIndexer returns actual indexer of specified algorithm
func (algo Algorithm) NewIndexer() (indexer AlgorithmIndexer, err error) {
	if algo == STAR {
		return STARIndexer{}, nil
	} else if algo == Salmon {
		return SalmonIndexer{}, nil
	}

	return nil, fmt.Errorf("No indexer found for algorithm '%v'", algo)
}

// IndexerRunner is a general interface that runs specified sequencing algorithm indexers
type IndexerRunner interface {
	BaseDir() string
	GenomeAssembly() string
	FAFileName() string
	GTFFileName() string
	TXFAFileName() string
	DownloadGenomeFiles() error
	RunIndexers(algorithms []Algorithm) error // invoke actual indexers based on given algorithms
}

// AlgorithmIndexer is a RNA-seq algorithm interface
type AlgorithmIndexer interface {
	BuildIndex(gc IndexerRunner) error
}

// STARIndexer algorithm
type STARIndexer struct {
}

// BuildIndex for algorithm STAR
func (st STARIndexer) BuildIndex(sqb IndexerRunner) error {
	_, err := exec.LookPath("STAR") // check if 'salmon' binary is installed in the system
	if err != nil {
		return errors.New("Can't find 'STAR' binary on your system; check if it's installed and is added to your PATH env variable")
	}
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

// SalmonIndexer algorithm
type SalmonIndexer struct {
}

// BuildIndex for algorithm Salmon
func (sa SalmonIndexer) BuildIndex(gc IndexerRunner) error {
	_, err := exec.LookPath("salmon") // check if 'salmon' binary is installed in the system
	if err != nil {
		return errors.New("Can't find 'salmon' binary on your system; check if it's installed and is added to your PATH env variable")
	}
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

// BuildSalmonIndexByKmer builds index for a specific kmer
func BuildSalmonIndexByKmer(kmer string, sqb IndexerRunner) error {
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

func buildSalmonIndexByKmer(kmer string, gc IndexerRunner, wg *sync.WaitGroup) error {
	defer wg.Done()
	return BuildSalmonIndexByKmer(kmer, gc)
}
