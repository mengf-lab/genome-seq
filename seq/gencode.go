package seq

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/keqiang/filenet"
)

// GencodeGenomeAnnotations contains file information for gencode annotations
type GencodeGenomeAnnotations struct {
	ExistingBaseDir, Species, Version string
}

// BaseDir returns the base directory if this builder runs builds
func (gb GencodeGenomeAnnotations) BaseDir() string {
	if gb.ExistingBaseDir != "" {
		return gb.ExistingBaseDir
	}
	return "gencode_" + gb.Species + "_" + gb.Version
}

// GenomeAssembly returns the corresponding genome assembly given the species and gencode version
func (gb GencodeGenomeAnnotations) GenomeAssembly() string {
	genomeAssemblyMapping := map[string]map[string]string{
		"hs": map[string]string{
			"30": "GRCh38",
		},
		"mm": map[string]string{
			"M22": "GRCm38",
		},
	}

	if subMapping, ok := genomeAssemblyMapping[gb.Species]; ok {
		if res, ok := subMapping[gb.Version]; ok {
			return res
		}
	}

	log.Fatalf("Can not determine genome assembly for '%v' of Gencode version '%v'", gb.Species, gb.Version)
	return ""
}

// FAFile returns the fa file name
func (gb GencodeGenomeAnnotations) FAFile() string {
	return gb.GenomeAssembly() + ".primary_assembly.genome.fa"
}

// FilePrefix returns the file prefix string
func (gb GencodeGenomeAnnotations) FilePrefix() string {
	return fmt.Sprintf("gencode.v%v", gb.Version)
}

// GTFFile returns the GTF file name
func (gb GencodeGenomeAnnotations) GTFFile() string {
	return gb.FilePrefix() + ".primary_assembly.annotation.gtf"
}

// TXFAFile returns the transcript fa file name
func (gb GencodeGenomeAnnotations) TXFAFile() string {
	return gb.FilePrefix() + ".transcripts.fa"
}

// DownloadAnnotationFiles implements the interface method
func (gb GencodeGenomeAnnotations) DownloadAnnotationFiles() error {
	genomeAssembly := gb.GenomeAssembly() // figure out the genome assembly

	log.Printf("Gencode downloading files for genome assembly '%v'\n", genomeAssembly)

	filePrefix := gb.FilePrefix()

	faFile := gb.FAFile()
	gtfFile := gb.GTFFile()
	txfaFile := gb.TXFAFile()

	faGZFile := faFile + ".gz"
	gtfGZFile := gtfFile + ".gz"
	txfaGZFile := txfaFile + ".gz"

	files2Download := []string{
		gtfGZFile,
		filePrefix + ".polyAs.gtf.gz",
		filePrefix + ".2wayconspseudos.gtf.gz",
		filePrefix + ".tRNAs.gtf.gz",
		txfaGZFile,
		faGZFile,
	}

	speciesMapping := map[string]string{
		"hs": "Gencode_human",
		"mm": "Gencode_mouse",
	} // map from species short name to full url

	ftpDir := fmt.Sprintf("pub/databases/gencode/%v/release_%v", speciesMapping[gb.Species], strings.ToUpper(gb.Version))

	resDir := "gencode_" + gb.Species + "_" + gb.Version

	fc := filenet.FTPDownloadConfig{
		URL:            "ftp.ebi.ac.uk",
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

	log.Println("Gencode files downloaded")

	files2Unzip := make(map[string]string)
	files2Unzip[filepath.Join(resDir, gtfGZFile)] = filepath.Join(resDir, gtfFile)
	files2Unzip[filepath.Join(resDir, faGZFile)] = filepath.Join(resDir, faFile)
	files2Unzip[filepath.Join(resDir, txfaGZFile)] = filepath.Join(resDir, txfaFile)

	log.Println("Gencode decompressing files")

	filenet.DecompressFiles(files2Unzip, 3)

	log.Println("Gencode files decompressed")

	return nil
}
