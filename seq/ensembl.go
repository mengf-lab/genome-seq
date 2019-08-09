package seq

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/keqiang/filenet"
)

// EnsemblGenomeAnnotations build index for Ensembl
type EnsemblGenomeAnnotations struct {
	ExistingBaseDir, Species, Version string
}

// BaseDir returns the base directory for the result directory of all index files
func (gb EnsemblGenomeAnnotations) BaseDir() string {
	if gb.ExistingBaseDir != "" {
		return gb.ExistingBaseDir
	}
	return "ensembl_" + gb.Species + "_" + gb.Version
}

// GenomeAssembly returns the corresponding genome assembly given the species and gencode version
func (gb EnsemblGenomeAnnotations) GenomeAssembly() string {
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
		if res, ok := subMapping[gb.Version]; ok {
			return res
		}
	}

	log.Fatalf("Can not determine genome assembly for '%v' of Ensembl version '%v'", gb.Species, gb.Version)
	return ""
}

// SpeciesString return the full species string
func (gb EnsemblGenomeAnnotations) SpeciesString() string {
	speciesMapping := map[string]string{
		"hs": "Homo_sapiens",            // human
		"mm": "Mus_musculus",            // mouse
		"rn": "Rattus_norvegicus",       // rat
		"dr": "Danio_rerio",             // zebrafish
		"dm": "Drosophila_melanogaster", // Fruitfly
	} // map from species short name to full species string

	return speciesMapping[gb.Species]
}

// FAFile returns the fa file name
func (gb EnsemblGenomeAnnotations) FAFile() string {
	var faType string
	if gb.Species == "rn" || gb.Species == "dm" { // rat and fruitfly don't have primary_assembly
		faType = "toplevel"
	} else {
		faType = "primary_assembly"
	}
	return fmt.Sprintf("%v.%v.dna.%v.fa", gb.SpeciesString(), gb.GenomeAssembly(), faType)
}

// GTFFile returns the GTF file name
func (gb EnsemblGenomeAnnotations) GTFFile() string {
	return fmt.Sprintf("%v.%v.%v.gtf", gb.SpeciesString(), gb.GenomeAssembly(), gb.Version)
}

// TXFAFile returns the transcript fa file name
func (gb EnsemblGenomeAnnotations) TXFAFile() string {
	return fmt.Sprintf("%v.%v.cdna.all.fa", gb.SpeciesString(), gb.GenomeAssembly())
}

// DownloadAnnotationFiles implements the interface method
func (gb EnsemblGenomeAnnotations) DownloadAnnotationFiles() error {
	genomeAssembly := gb.GenomeAssembly() // figure out the genome assembly

	log.Printf("Ensembl downloading files for genome assembly '%v'\n", genomeAssembly)

	faFile := gb.FAFile()
	gtfFile := gb.GTFFile()
	txfaFile := gb.TXFAFile()

	faGZFile := faFile + ".gz"
	gtfGZFile := gtfFile + ".gz"
	txfaGZFile := txfaFile + ".gz"

	files2Download := []string{
		fmt.Sprintf("fasta/%v/dna/%v", strings.ToLower(gb.SpeciesString()), faGZFile),
		fmt.Sprintf("fasta/%v/cdna/%v", strings.ToLower(gb.SpeciesString()), txfaGZFile),
		fmt.Sprintf("gtf/%v/%v", strings.ToLower(gb.SpeciesString()), gtfGZFile),
	}

	ftpDir := fmt.Sprintf("pub/release-%v", gb.Version)

	resDir := "ensembl_" + gb.Species + "_" + gb.Version

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
	files2Unzip[filepath.Join(resDir, gtfGZFile)] = filepath.Join(resDir, gtfFile)
	files2Unzip[filepath.Join(resDir, faGZFile)] = filepath.Join(resDir, faFile)
	files2Unzip[filepath.Join(resDir, txfaGZFile)] = filepath.Join(resDir, txfaFile)

	log.Println("Ensembl decompressing files")

	filenet.DecompressFiles(files2Unzip, 3)

	log.Println("Ensembl files decompressed")

	return nil
}
