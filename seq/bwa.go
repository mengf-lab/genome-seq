package seq

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/keqiang/filenet"
)

// BWA algorithm
type BWA struct{}

// CheckIndexerAvailability returns an error if 'bwa' is not installed
func (algorithm *BWA) CheckIndexerAvailability() error {
	return filenet.CheckBinaryExistence("bwa")
}

// BuildIndices for algorithm BWA
func (algorithm *BWA) BuildIndices(genomeAnnotations GenomeAnnotations) error {
	bwaIdxDir := filepath.Join(genomeAnnotations.BaseDir(), "bwa_idx")
	if err := os.Mkdir(bwaIdxDir, os.ModePerm); err == nil {
		bwaArgs := []string{
			"index", "-a", "bwtsw", "-p", filepath.Join(bwaIdxDir, "bwa_idx"), filepath.Join(genomeAnnotations.BaseDir(), genomeAnnotations.FAFile()),
		}
		log.Println("Running BWA indexing")
		bwaCmd := exec.Command("bwa", bwaArgs...)

		bwaCmd.Stdout = os.Stdout
		bwaCmd.Stderr = os.Stderr
		err := bwaCmd.Run()
		if err != nil {
			return err
		}
		log.Println("Finished BWA indexing")
	} else {
		return err
	}
	return nil
}
