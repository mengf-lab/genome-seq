package seq

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/keqiang/filenet"
)

// Salmon algorithm
type Salmon struct{}

// CheckIndexerAvailability returns an error if 'salmon' is not installed
func (algorithm *Salmon) CheckIndexerAvailability() error {
	return filenet.CheckBinaryExistence("salmon")
}

// BuildIndices for algorithm Salmon
func (algorithm *Salmon) BuildIndices(genomeAnnotations GenomeAnnotations) error {
	log.Println("Running Salmon indexing")
	var wg sync.WaitGroup
	kmers := [6]string{"21", "23", "25", "27", "29", "31"} // all salmon Ks

	for _, kmer := range kmers {
		wg.Add(1)
		go buildSalmonIndexByKmer(kmer, genomeAnnotations, &wg)
	}

	wg.Wait()
	log.Println("Finished Salmon indexing")
	return nil
}

// BuildSalmonIndexByKmer builds index for a specific kmer
func BuildSalmonIndexByKmer(kmer string, genomeAnnotations GenomeAnnotations) error {
	salmonIdxDir := filepath.Join(genomeAnnotations.BaseDir(), fmt.Sprintf("salmon_k%v_idx", kmer))
	salmonArgs := []string{
		"index", "-t", filepath.Join(genomeAnnotations.BaseDir(), genomeAnnotations.TXFAFile()), "-i", salmonIdxDir, "--type", "quasi", "-k", kmer,
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

func buildSalmonIndexByKmer(kmer string, genomeAnnotations GenomeAnnotations, wg *sync.WaitGroup) error {
	defer wg.Done()
	return BuildSalmonIndexByKmer(kmer, genomeAnnotations)
}
