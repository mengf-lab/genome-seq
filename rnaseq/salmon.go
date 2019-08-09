package rnaseq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// SalmonIndexer algorithm
type SalmonIndexer struct{}

// CheckBinary returns an error if 'salmon' is not installed
func (indexer *SalmonIndexer) CheckBinary() error {
	_, err := exec.LookPath("salmon") // check if 'salmon' binary is installed in the system
	if err != nil {
		return errors.New("Can not locate 'salmon' binary on your system; check if it's installed and is added to your PATH variable")
	}
	return nil
}

// BuildIndex for algorithm Salmon
func (indexer *SalmonIndexer) BuildIndex(gc IndexerRunner) error {
	log.Println("Running Salmon indexing")
	var wg sync.WaitGroup
	kmers := [6]string{"21", "23", "25", "27", "29", "31"} // all salmon Ks

	for _, kmer := range kmers {
		wg.Add(1)
		go buildSalmonIndexByKmer(kmer, gc, &wg)
	}

	wg.Wait()
	log.Println("Finished Salmon indexing")
	return nil
}

// BuildSalmonIndexByKmer builds index for a specific kmer
func BuildSalmonIndexByKmer(kmer string, indexRunner IndexerRunner) error {
	salmonIdxDir := filepath.Join(indexRunner.BaseDir(), fmt.Sprintf("salmon_k%v_idx", kmer))
	salmonArgs := []string{
		"index", "-t", filepath.Join(indexRunner.BaseDir(), indexRunner.TXFAFileName()), "-i", salmonIdxDir, "--type", "quasi", "-k", kmer,
	}
	log.Printf("Salmon indexing k%v\n", kmer)
	salmonCmd := exec.Command("salmon", salmonArgs...)

	salmonCmd.Stdout = os.Stdout
	salmonCmd.Stderr = os.Stderr
	err := salmonCmd.Run()

	if err != nil {
		return err
	}
	log.Printf("Finished Salmon indexing k%v\n", kmer)
	return nil
}

func buildSalmonIndexByKmer(kmer string, gc IndexerRunner, wg *sync.WaitGroup) error {
	defer wg.Done()
	return BuildSalmonIndexByKmer(kmer, gc)
}
