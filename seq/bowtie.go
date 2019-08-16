package seq

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/keqiang/filenet"
)

func checkBowtieIndexerAvailability(versionedBowtie string) error {
	return filenet.CheckBinaryExistence(versionedBowtie + "-build")
}

func buildBowtieIndices(versionedBowtie string, genomeAnnotations GenomeAnnotations) error {
	bowtieIdxDir := filepath.Join(genomeAnnotations.BaseDir(), versionedBowtie+"_idx")
	if err := os.Mkdir(bowtieIdxDir, os.ModePerm); err == nil {
		bowtieArgs := []string{
			"-f", filepath.Join(genomeAnnotations.BaseDir(), genomeAnnotations.FAFile()), filepath.Join(bowtieIdxDir, versionedBowtie+"_idx"),
		}
		log.Println("Running", versionedBowtie, "indexing")
		bowtieCmd := exec.Command(versionedBowtie+"-build", bowtieArgs...)

		bowtieCmd.Stdout = os.Stdout
		bowtieCmd.Stderr = os.Stderr
		err := bowtieCmd.Run()
		if err != nil {
			return err
		}
		log.Println("Finished", versionedBowtie, "indexing")
	} else {
		return err
	}
	return nil
}

// Bowtie algorithm
type Bowtie struct{}

// CheckIndexerAvailability returns an error if 'bowtie' is not installed
func (algorithm *Bowtie) CheckIndexerAvailability() error {
	return checkBowtieIndexerAvailability("bowtie")
}

// BuildIndices for algorithm Bowtie
func (algorithm *Bowtie) BuildIndices(genomeAnnotations GenomeAnnotations) error {
	return buildBowtieIndices("bowtie", genomeAnnotations)
}

// Bowtie2 algorithm
type Bowtie2 struct{}

// CheckIndexerAvailability returns an error if 'bowtie2' is not installed
func (algorithm *Bowtie2) CheckIndexerAvailability() error {
	return checkBowtieIndexerAvailability("bowtie2")
}

// BuildIndices for algorithm Bowtie2
func (algorithm *Bowtie2) BuildIndices(genomeAnnotations GenomeAnnotations) error {
	return buildBowtieIndices("bowtie2", genomeAnnotations)
}
