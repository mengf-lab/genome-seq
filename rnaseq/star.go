package rnaseq

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// STARIndexer algorithm
type STARIndexer struct{}

// CheckBinary returns an error if 'STAR' is not installed
func (indexer *STARIndexer) CheckBinary() error {
	_, err := exec.LookPath("STAR") // check if 'STAR' binary is installed in the system
	if err != nil {
		return errors.New("Can't find 'STAR' binary on your system; check if it's installed and is added to your PATH variable")
	}
	return nil
}

// BuildIndex for algorithm STAR
func (indexer *STARIndexer) BuildIndex(indexRunner IndexerRunner) error {
	starIdxDir := filepath.Join(indexRunner.BaseDir(), "star_idx")
	if err := os.Mkdir(starIdxDir, os.ModePerm); err == nil {
		starArgs := []string{
			"--runThreadN", "4", "--runMode", "genomeGenerate", "--genomeDir", starIdxDir,
			"--genomeFastaFiles", filepath.Join(indexRunner.BaseDir(), indexRunner.FAFileName()),
			"--sjdbGTFfile", filepath.Join(indexRunner.BaseDir(), indexRunner.GTFFileName()),
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
