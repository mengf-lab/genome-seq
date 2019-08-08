package main

import (
	"flag"
	"log"

	"github.com/keqiang/genome-seq/rnaseq"
)

func main() {
	// parse command line arguments
	genomeSourcePtr := flag.String("source", "gencode", "Genome source to build; specify either 'gencode' or 'ensemble'")
	rVerPtr := flag.String("g", "30", "Genome release version string")
	speciesPtr := flag.String("s", "hs", "Species")
	flag.Parse()

	var indexerRunner rnaseq.IndexerRunner
	if *genomeSourcePtr == "gencode" {
		indexerRunner = rnaseq.GencodeIndexerRunner{Species: *speciesPtr, GencodeVersion: *rVerPtr}
	} else if *genomeSourcePtr == "ensembl" {
		//indexerRunner = rnaseq.EnsemblIndexerRunner{Species: *speciesPtr, EnsemblVersion: *rVerPtr}
	}

	err := indexerRunner.RunIndexers([]rnaseq.Algorithm{rnaseq.STAR})
	if err != nil {
		log.Fatal(err)
	}
}
