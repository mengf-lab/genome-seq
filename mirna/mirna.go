package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/keqiang/filenet"
	"github.com/keqiang/filenet/ftp"
)

func main() {
	speciesPtr := flag.String("s", "all", `Species; specify "hs" for Human; default is "all" which will download all species this command supports. See below for a full list of available species`)

	rVerPtr := flag.String("r", "21", `Annotation source release version; specify "21" or later`)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println()
		fmt.Print(`Full list of available species
  hs -> Human (Homo sapiens)
  mm -> Mouse (Mus musculus)
  rn -> Rat (Rattus norvegicus)
  dr -> Zebrafish (Danio rerio)
  dm -> Fruitfly (Drosophila melanogaster)
		`)
		fmt.Println()
		fmt.Print(`Examples
  mirna
    run using the default values, which is equivalent to mirna -s all -r 21; this will download annotation files for version 22 and extract all species

  mirna -r 22 -s hs
    downloads version 22 and only extracts human annotations`)
		fmt.Println()
	}

	flag.Parse()

	versionMapping := map[string]string{
		"21": "21",
		"22": "22.1",
	}

	ftpDir := "/pub/mirbase/" + versionMapping[*rVerPtr]
	resDir := "mirbase_v" + *rVerPtr

	faFile := "mature.fa"
	faGZFile := faFile + ".gz"

	fc := ftp.NewDownloadConfig("mirbase.org", ftpDir, resDir, []string{faGZFile})
	err := fc.Download()

	if err != nil {
		log.Fatal(err)
	}

	resFAFile := filepath.Join(resDir, faFile)

	err = filenet.GZipDecompress(filepath.Join(resDir, faGZFile), resFAFile)

	if err != nil {
		log.Fatal(err)
	}

	speciesMapping := map[string]string{
		"hs": "Homo sapiens",            // human
		"mm": "Mus musculus",            // mouse
		"rn": "Rattus norvegicus",       // rat
		"dr": "Danio rerio",             // zebrafish
		"dm": "Drosophila melanogaster", // Fruitfly
	} // map from species short name to full species string

	for species, speciesFullName := range speciesMapping {
		if *speciesPtr == "all" || *speciesPtr == species {
			log.Println("Extracting annotations for species:", species)
			err = writeSpeciesAnnotationFile(resFAFile, resDir, species, speciesFullName)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func writeSpeciesAnnotationFile(annotationFile, resDir, species, speciesFullName string) error {
	speciesAnnotationFile, err := os.Create(filepath.Join(resDir, species+".fa"))
	if err != nil {
		return err
	}
	defer speciesAnnotationFile.Close()
	args := []string{"-A", "1", speciesFullName, annotationFile}
	cmd := exec.Command("grep", args...)
	cmd.Stdout = speciesAnnotationFile
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return err
}
