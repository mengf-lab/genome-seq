package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/keqiang/genome-seq/rnaseq"
)

func main() {
	// parse command line arguments
	genomeSourcePtr := flag.String("a", "gencode", `Annotation source, specify either "gencode" or "ensembl"`)

	speciesPtr := flag.String("s", "hs", `Species; Example: specify "hs" for Human; See below for a full list of available species`)

	rVerPtr := flag.String("r", "30", `Annotation source release version; If you specified "gencode" for annotation source flag, then you need to specify version "30" or above for Human and version "M22" or above for Mouse. If you specified "ensembl", then you just need to specify the ensembl release version "96" or above`)

	flag.Usage = func() {
		fmt.Println("Argument list")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Print(`Full list of available species
  hs -> Human (Homo sapiens)
  mm -> Mouse (Mus musculus)
  rn -> Rat (Rattus norvegicus) (only supports Ensembl)
  dr -> Zebrafish (Danio rerio) (only supports Ensembl)
  dm -> Fruitfly (Drosophila melanogaster) (only supports Ensembl)
		`)
		fmt.Println()
		fmt.Print(`Examples
  seqidx
    will run using the default values, which is equivalent to seqidx -a gencode -s hs -r 30; this will download Gencode files for Human Gencode version 30 and build index files

  seqidx -s mm -r M22
    downloads Gencode files for Mouse Gencode version M22 and build index files

  seqidx -a ensembl -r 96 -s dm
    downloads Ensembl files for Fruitfly Ensembl version 96 and build index files`)
		fmt.Println()
	}

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
