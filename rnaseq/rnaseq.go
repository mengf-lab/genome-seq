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
	RunIndexers(algorithms []Algorithm) error // invoke actual indexers based on given algorithms
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
