package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/keqiang/genome-seq/seq"
)

func main() {

	seqType := flag.String("t", "rna", `Sequencing type; specify either "rna" or "chip"`)
	// parse command line arguments
	genomeSourcePtr := flag.String("a", "gencode", `Annotation source; specify either "gencode" or "ensembl"`)

	speciesPtr := flag.String("s", "hs", `Species; specify "hs" for Human; See below for a full list of available species`)

	rVerPtr := flag.String("r", "30", `Annotation source release version; if you specified "gencode" as the annotation source, you need to specify version "30" or above for Human and version "M22" or above for Mouse. If you specified "ensembl", you just need to specify the ensembl release version "96" or above`)

	existingDir := flag.String("d", "", "Use this option to indicate the annotation files already exist and specify the directory which contains all your annotations files; omit this option if you want to download annotation files")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
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
    run using the default values, which is equivalent to seqidx -t rna -a gencode -s hs -r 30; this will download Gencode annotation files for Human Gencode version 30 and build RNA-seq index files

  seqidx -s mm -r M22
    downloads Gencode annotation files for Mouse Gencode version M22 and build RNA-seq index files

  seqidx -a ensembl -r 96 -s dm
    downloads Ensembl annotation files for Fruitfly Ensembl version 96 and build RNA-seq index files

  seqidx -t chip -a ensembl -r 96 -s dm -d ensembl_dm_96
    use existing Ensembl annotation files under directory 'ensembl_dm_96' for Fruitfly Ensembl version 96 to build ChIP-seq index files`)
		fmt.Println()
	}

	flag.Parse()

	if *existingDir != "" { // using existing files
		if _, err := os.Stat(*existingDir); os.IsNotExist(err) {
			log.Fatalf("Please use -d to specify an exsiting directory that contains all annotation files needed")
		}
	}

	var genomeAnnotations seq.GenomeAnnotations
	if *genomeSourcePtr == "gencode" {
		genomeAnnotations = seq.GencodeGenomeAnnotations{Species: *speciesPtr, Version: *rVerPtr, ExistingBaseDir: *existingDir}
	} else if *genomeSourcePtr == "ensembl" {
		genomeAnnotations = seq.EnsemblGenomeAnnotations{Species: *speciesPtr, Version: *rVerPtr, ExistingBaseDir: *existingDir}
	} else {
		log.Fatalf("Invalid annotation source '%v'\n", *genomeSourcePtr)
	}

	algorithms := make([]seq.Algorithm, 0, 3)
	if *seqType == "rna" {
		algorithms = append(algorithms, &seq.STAR{}, &seq.Salmon{})
	} else if *seqType == "chip" {
		algorithms = append(algorithms, &seq.Bowtie{}, &seq.Bowtie2{}, &seq.BWA{})
	} else {
		log.Fatalf("Invalid sequencing type '%v'\n", *seqType)
	}

	err := seq.IndexGenomeAnnotations(genomeAnnotations, algorithms, *existingDir != "")

	if err != nil {
		log.Fatal(err)
	}
}
