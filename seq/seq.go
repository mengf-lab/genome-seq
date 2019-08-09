package seq

import (
	"fmt"
	"os/exec"
)

// GenomeAnnotations is a general interface that specifies
// genome annotation file locations and methods to donload these files
type GenomeAnnotations interface {
	BaseDir() string
	GenomeAssembly() string
	FAFile() string
	GTFFile() string
	TXFAFile() string
	DownloadAnnotationFiles() error
}

// Algorithm is an interface representing a sequencing algorithm,
// such as STAR, salmon for RNA-seq or BWA for CHiP-seq
type Algorithm interface {
	// check if this algorithm's binary is installed on the system
	CheckIndexerAvailability() error
	// build index files for this algorithm using the specified GenomeAnnotations
	BuildIndices(genomeAnnotations GenomeAnnotations) error
}

// IndexGenomeAnnotations invokes each algorithm indexing method by passing the genome annotation files to the algorithm
func IndexGenomeAnnotations(genomeAnnotations GenomeAnnotations, algorithms []Algorithm, annotationFilesExisted bool) error {
	if !annotationFilesExisted { // download annotation files only when necessary
		err := genomeAnnotations.DownloadAnnotationFiles() // download and decompress genome annotation files before proceeding

		if err != nil {
			return err
		}
	}

	for _, algo := range algorithms { // for each specified RNA-seq algorithm
		err := algo.CheckIndexerAvailability() // check if the indexer binary is available
		if err != nil {
			return err
		}

		err = algo.BuildIndices(genomeAnnotations)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkBinary(binaryName string) error {
	_, err := exec.LookPath(binaryName) // check if the specified binary is installed in the system
	if err != nil {
		return fmt.Errorf("Can not locate '%v' binary on your system; check if it's installed and is added to your PATH variable", binaryName)
	}
	return nil
}
