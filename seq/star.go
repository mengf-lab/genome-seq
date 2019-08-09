package seq

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// STAR algorithm
type STAR struct{}

// CheckIndexerAvailability returns an error if 'STAR' is not installed
func (algorithm *STAR) CheckIndexerAvailability() error {
	return checkBinary("STAR")
}

// BuildIndices for algorithm STAR
func (algorithm *STAR) BuildIndices(genomeAnnotations GenomeAnnotations) error {
	starIdxDir := filepath.Join(genomeAnnotations.BaseDir(), "star_idx")
	if err := os.Mkdir(starIdxDir, os.ModePerm); err == nil {
		starArgs := []string{
			"--runThreadN", "4", "--runMode", "genomeGenerate", "--genomeDir", starIdxDir, "--genomeSAindexNbases", "13", // changed to 13 because there is a warning
			"--genomeFastaFiles", filepath.Join(genomeAnnotations.BaseDir(), genomeAnnotations.FAFile()),
			"--sjdbGTFfile", filepath.Join(genomeAnnotations.BaseDir(), genomeAnnotations.GTFFile()),
		}
		log.Println("Running STAR indexing")
		starCmd := exec.Command("STAR", starArgs...)

		starCmd.Stdout = os.Stdout
		starCmd.Stderr = os.Stderr
		err := starCmd.Run()
		if err != nil {
			return err
		}
		log.Println("Finished STAR indexing")
	} else {
		return err
	}
	return nil
}
