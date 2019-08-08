package rnaseq

// EnsemblIndexerRunner build index for Ensembl
type EnsemblIndexerRunner struct {
	Species, EnsemblVersion string
}

// RunIndexers implements IndexerRunner
func (eb *EnsemblIndexerRunner) RunIndexers(algorithms []Algorithm) error {
	return nil
}
