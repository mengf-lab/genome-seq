# genome-seq
This repo provides some basic tools related to genome sequencing.

## seqidx command
Run `go build` under seqidx folder to build a binary or just use `go run seqidx.go`. This tool downloads files required to build RNA-seq indices for Salmon and STAR (or you can add your own Algorithms) and invokes the binaries on your computer to actually build the index. 

## mirna command
Run `go build` under mirna folder to build a binary or just use `go run mirna.go`. This command only downloads the file 'mature.fa' and extracts the specified species from it. It can then be used to run MicroRNA-seq algorithms.
