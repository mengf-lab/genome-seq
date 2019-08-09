package rnaseq

import (
	"fmt"
)

// IndexerRunner is a general interface that runs specified sequencing algorithm indexers
type IndexerRunner interface {
	BaseDir() string
	GenomeAssembly() string
	FAFileName() string
	GTFFileName() string
	TXFAFileName() string
	DownloadGenomeFiles() error
}

// AlgorithmIndexer is a RNA-seq algorithm interface
type AlgorithmIndexer interface {
	CheckBinary() error
	BuildIndex(gc IndexerRunner) error
}

// Algorithm is an enum type for RNA-seq algorithm names
type Algorithm string

// Available RNA-seq algorithms
const (
	STAR   Algorithm = "star"
	Salmon Algorithm = "salmon"
)

// NewAlgorithmIndexer returns actual indexer of specified algorithm
func (algo Algorithm) NewAlgorithmIndexer() (indexer AlgorithmIndexer, err error) {
	err = nil
	switch algo {
	case STAR:
		indexer = &STARIndexer{}
	case Salmon:
		indexer = &SalmonIndexer{}
	default:
		err = fmt.Errorf("No indexer found for algorithm '%v'", algo)
	}

	if err = indexer.CheckBinary(); err != nil { // binary not installed
		indexer = nil
	}
	return
}

// RunIndexers invokes each indexer by passing the runner to it
func RunIndexers(indexerRunner IndexerRunner, algorithms []Algorithm) error {
	err := indexerRunner.DownloadGenomeFiles() // download and decompress genome files before proceeding

	if err != nil {
		return err
	}

	for _, algo := range algorithms { // for each specified RNA-seq algorithm
		indexer, err := algo.NewAlgorithmIndexer() // instantiate an indexer
		if err != nil {
			return err
		}
		err = indexer.BuildIndex(indexerRunner)
		if err != nil {
			return err
		}
	}

	return nil
}
