package rnaseq

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/keqiang/filenet"
)

// EnsemblIndexerRunner build index for Ensembl
type EnsemblIndexerRunner struct {
	Species, EnsemblVersion string
}

// BaseDir returns the base directory for the result directory of all index files
func (gb EnsemblIndexerRunner) BaseDir() string {
	return "ensembl_" + gb.Species + "_" + gb.EnsemblVersion
}

// GenomeAssembly returns the corresponding genome assembly given the species and gencode version
func (gb EnsemblIndexerRunner) GenomeAssembly() string {
	genomeAssemblyMapping := map[string]map[string]string{
		"hs": map[string]string{
			"96": "GRCh38",
			"97": "GRCh38",
		},
		"mm": map[string]string{
			"96": "GRCm38",
			"97": "GRCm38",
		},
		"rn": map[string]string{
			"96": "Rnor_6.0",
			"97": "Rnor_6.0",
		},
		"dr": map[string]string{
			"96": "GRCz11",
			"97": "GRCz11",
		},
		"dm": map[string]string{
			"96": "BDGP6.22",
			"97": "BDGP6.22",
		},
	}

	if subMapping, ok := genomeAssemblyMapping[gb.Species]; ok {
		if res, ok := subMapping[gb.EnsemblVersion]; ok {
			return res
		}
	}

	log.Fatalf("Can not determine genome assembly for '%v' of Ensembl version '%v'", gb.Species, gb.EnsemblVersion)
	return ""
}

// SpeciesString return the full species string
func (gb EnsemblIndexerRunner) SpeciesString() string {
	speciesMapping := map[string]string{
		"hs": "Homo_sapiens",            // human
		"mm": "Mus_musculus",            // mouse
		"rn": "Rattus_norvegicus",       // rat
		"dr": "Danio_rerio",             // zebrafish
		"dm": "Drosophila_melanogaster", // Fruitfly
	} // map from species short name to full species string

	return speciesMapping[gb.Species]
}

// FAFileName returns the fa file name
func (gb EnsemblIndexerRunner) FAFileName() string {
	var faType string
	if gb.Species == "rn" || gb.Species == "dm" { // rat and fruitfly don't have primary_assembly
		faType = "toplevel"
	} else {
		faType = "primary_assembly"
	}
	return fmt.Sprintf("%v.%v.dna.%v.fa", gb.SpeciesString(), gb.GenomeAssembly(), faType)
}

// GTFFileName returns the GTF file name
func (gb EnsemblIndexerRunner) GTFFileName() string {
	return fmt.Sprintf("%v.%v.%v.gtf", gb.SpeciesString(), gb.GenomeAssembly(), gb.EnsemblVersion)
}

// TXFAFileName returns the transcript fa file name
func (gb EnsemblIndexerRunner) TXFAFileName() string {
	return fmt.Sprintf("%v.%v.cdna.all.fa", gb.SpeciesString(), gb.GenomeAssembly())
}

// DownloadGenomeFiles implements the interface method
func (gb EnsemblIndexerRunner) DownloadGenomeFiles() error {
	genomeAssembly := gb.GenomeAssembly() // figure out the genome assembly

	log.Printf("Ensembl downloading files for genome assembly '%v'\n", genomeAssembly)

	faFileName := gb.FAFileName()
	gtfFileName := gb.GTFFileName()
	txfaFileName := gb.TXFAFileName()

	faGZFileName := faFileName + ".gz"
	gtfGZFileName := gtfFileName + ".gz"
	txfaGZFileName := txfaFileName + ".gz"

	files2Download := []string{
		fmt.Sprintf("fasta/%v/dna/%v", strings.ToLower(gb.SpeciesString()), faGZFileName),
		fmt.Sprintf("fasta/%v/cdna/%v", strings.ToLower(gb.SpeciesString()), txfaGZFileName),
		fmt.Sprintf("gtf/%v/%v", strings.ToLower(gb.SpeciesString()), gtfGZFileName),
	}

	ftpDir := fmt.Sprintf("pub/release-%v", gb.EnsemblVersion)

	resDir := "ensembl_" + gb.Species + "_" + gb.EnsemblVersion

	fc := filenet.FTPDownloadConfig{
		URL:            "ftp.ensembl.org",
		Port:           21,
		MaxConnection:  3,
		BaseDir:        ftpDir,
		DestDir:        resDir,
		Files2Download: files2Download,
	}

	err := fc.Download()

	if err != nil {
		return err
	}

	log.Println("Ensembl files downloaded")

	files2Unzip := make(map[string]string)
	files2Unzip[filepath.Join(resDir, gtfGZFileName)] = filepath.Join(resDir, gtfFileName)
	files2Unzip[filepath.Join(resDir, faGZFileName)] = filepath.Join(resDir, faFileName)
	files2Unzip[filepath.Join(resDir, txfaGZFileName)] = filepath.Join(resDir, txfaFileName)

	log.Println("Ensembl decompressing files")

	filenet.DecompressFiles(files2Unzip, 3)

	log.Println("Ensembl files decompressed")

	return nil
}
