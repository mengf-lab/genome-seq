package main

import (
	"flag"
)

func main() {
	// parse command line arguments
	genomeSourcePtr := flag.String("source", "gencode", "Genome source to build; specify either 'gencode' or 'ensemble'")
	rVerPtr := flag.String("g", "30", "Gene code release version")
	speciesPtr := flag.String("s", "hs", "Species")
	flag.Parse()
}
