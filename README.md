# genome-seq
This repo provides some basic tools related to genome sequencing.

## seqidx command
Run `go build` under seqidx folder to build a binary or just use `go run seqidx.go`. This tool downloads files required to build RNA-seq(STAR and Salmon) or ChIP-seq(BWA, Bowtie, Bowtie2) indices and invokes the binaries on your computer to actually build the index. You can also add your own algorithms under the seq directory.

## mirna command
Run `go build` under mirna folder to build a binary or just use `go run mirna.go`. This command only downloads the file 'mature.fa' and extracts the specified species from it. It can then be used to run MicroRNA-seq algorithms.
