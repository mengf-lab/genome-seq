package rnaseq

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/keqiang/filenet"
)

// GencodeIndexerRunner runs indexers on Gencode genome files
type GencodeIndexerRunner struct {
	Species, GencodeVersion string
}

// BaseDir returns the base directory if this builder runs builds
func (gb GencodeIndexerRunner) BaseDir() string {
	return "gencode_" + gb.Species + "_" + gb.GencodeVersion
}

// GenomeAssembly returns the corresponding genome assembly given the species and gencode version
func (gb GencodeIndexerRunner) GenomeAssembly() string {
	genomeAssemblyMapping := map[string]map[string]string{
		"hs": map[string]string{
			"30": "GRCh38",
		},
		"mm": map[string]string{
			"M22": "GRCm38",
		},
	}

	if subMapping, ok := genomeAssemblyMapping[gb.Species]; ok {
		if res, ok := subMapping[gb.GencodeVersion]; ok {
			return res
		}
	}

	log.Fatalf("Can not determine genome assembly for '%v' of Gencode version '%v'", gb.Species, gb.GencodeVersion)
	return ""
}

// FAFileName returns the fa file name
func (gb GencodeIndexerRunner) FAFileName() string {
	return gb.GenomeAssembly() + ".primary_assembly.genome.fa"
}

// FilePrefix returns the file prefix string
func (gb GencodeIndexerRunner) FilePrefix() string {
	return fmt.Sprintf("gencode.v%v", gb.GencodeVersion)
}

// GTFFileName returns the GTF file name
func (gb GencodeIndexerRunner) GTFFileName() string {
	return gb.FilePrefix() + ".primary_assembly.annotation.gtf"
}

// TXFAFileName returns the transcript fa file name
func (gb GencodeIndexerRunner) TXFAFileName() string {
	return gb.FilePrefix() + ".transcripts.fa"
}

// DownloadGenomeFiles implements the interface method
func (gb GencodeIndexerRunner) DownloadGenomeFiles() error {
	genomeAssembly := gb.GenomeAssembly() // figure out the genome assembly

	log.Printf("Gencode downloading files for genome assembly '%v'\n", genomeAssembly)

	filePrefix := gb.FilePrefix()

	faFileName := gb.FAFileName()
	gtfFileName := gb.GTFFileName()
	txfaFileName := gb.TXFAFileName()

	faGZFileName := faFileName + ".gz"
	gtfGZFileName := gtfFileName + ".gz"
	txfaGZFileName := txfaFileName + ".gz"

	files2Download := []string{
		gtfGZFileName,
		filePrefix + ".polyAs.gtf.gz",
		filePrefix + ".2wayconspseudos.gtf.gz",
		filePrefix + ".tRNAs.gtf.gz",
		txfaGZFileName,
		faGZFileName,
	}

	speciesMapping := map[string]string{
		"hs": "Gencode_human",
		"mm": "Gencode_mouse",
	} // map from species short name to full url

	ftpDir := fmt.Sprintf("pub/databases/gencode/%v/release_%v", speciesMapping[gb.Species], strings.ToUpper(gb.GencodeVersion))

	resDir := "gencode_" + gb.Species + "_" + gb.GencodeVersion

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
	files2Unzip[filepath.Join(resDir, gtfGZFileName)] = filepath.Join(resDir, gtfFileName)
	files2Unzip[filepath.Join(resDir, faGZFileName)] = filepath.Join(resDir, faFileName)
	files2Unzip[filepath.Join(resDir, txfaGZFileName)] = filepath.Join(resDir, txfaFileName)

	log.Println("Gencode decompressing files")

	filenet.DecompressFiles(files2Unzip, 3)

	log.Println("Gencode files decompressed")

	return nil
}
