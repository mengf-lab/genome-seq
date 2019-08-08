package main

import (
	"flag"
	"log"

	"github.com/keqiang/genome-seq/rnaseq"
)

func main() {
	// parse command line arguments
	genomeSourcePtr := flag.String("source", "gencode", "annotation source, specify either 'gencode' or 'ensembl'")
	rVerPtr := flag.String("g", "30", "release version")
	speciesPtr := flag.String("s", "hs", "Species")
	flag.Parse()

	var indexerRunner rnaseq.IndexerRunner
	if *genomeSourcePtr == "gencode" {
		indexerRunner = rnaseq.GencodeIndexerRunner{Species: *speciesPtr, GencodeVersion: *rVerPtr}
	} else if *genomeSourcePtr == "ensembl" {
		indexerRunner = rnaseq.EnsemblIndexerRunner{Species: *speciesPtr, EnsemblVersion: *rVerPtr}
	}

	err := rnaseq.RunIndexers(indexerRunner, []rnaseq.Algorithm{rnaseq.STAR, rnaseq.Salmon})
	if err != nil {
		log.Fatal(err)
	}
}
