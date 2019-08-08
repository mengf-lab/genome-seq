package main

import (
	"flag"

	"github.com/keqiang/genome-seq/seq"
)

func main() {
	// parse command line arguments
	genomeSourcePtr := flag.String("source", "gencode", "Genome source to build; specify either 'gencode' or 'ensemble'")
	rVerPtr := flag.String("g", "30", "Genome release version string")
	speciesPtr := flag.String("s", "hs", "Species")
	flag.Parse()

	if *genomeSourcePtr == "gencode" {
		indexBuilder := seq.GencodeIndexBuilder{Species: *speciesPtr, GencodeVersion: *rVerPtr}
		indexBuilder.BuildIndex()
	}
}
