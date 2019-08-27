# genome-seq
This repo provides some basic tools related to genome sequencing.

## seqidx command
Run `go build` under seqidx folder to build a binary or just use `go run seqidx.go`. This tool downloads files required to build RNA-seq(STAR and Salmon) or ChIP-seq(BWA, Bowtie, Bowtie2) indices and invokes the binaries on your computer to actually build the index. You can also add your own algorithms under the seq directory.

```
Usage of seqidx:
  -a string
        Annotation source; specify either "gencode" or "ensembl" (default "gencode")
  -d string
        Use this option to indicate the annotation files already exist and specify the directory which contains all your annotations files; omit this option if you want to download annotation files
  -r string
        Annotation source release version; if you specified "gencode" as the annotation source, you need to specify version "30" or above for Human and version "M22" or above for Mouse. If you specified "ensembl", you just need to specify the ensembl release version "96" or above (default "30")
  -s string
        Species; specify "hs" for Human; See below for a full list of available species (default "hs")
  -t string
        Sequencing type; specify either "rna" or "chip" (default "rna")

Full list of available species
  hs -> Human (Homo sapiens)
  mm -> Mouse (Mus musculus)
  rn -> Rat (Rattus norvegicus) (only supports Ensembl)
  dr -> Zebrafish (Danio rerio) (only supports Ensembl)
  dm -> Fruitfly (Drosophila melanogaster) (only supports Ensembl)

Examples
  seqidx
    run using the default values, which is equivalent to seqidx -t rna -a gencode -s hs -r 30; this will download Gencode annotation files for Human Gencode version 30 and build RNA-seq index files

  seqidx -s mm -r M22
    downloads Gencode annotation files for Mouse Gencode version M22 and build RNA-seq index files

  seqidx -a ensembl -r 96 -s dm
    downloads Ensembl annotation files for Fruitfly Ensembl version 96 and build RNA-seq index files

  seqidx -t chip -a ensembl -r 96 -s dm -d ensembl_dm_96
    use existing Ensembl annotation files under directory 'ensembl_dm_96' for Fruitfly Ensembl version 96 to build CHiP-seq index files
```

## mirna command
Run `go build` under mirna folder to build a binary or just use `go run mirna.go`. This command only downloads the file 'mature.fa' and extracts the specified species from it. It can then be used to run MicroRNA-seq algorithms.

```
Usage of mirna:
  -r string
        Annotation source release version; specify "21" or later (default "21")
  -s string
        Species; specify "hs" for Human; default is "all" which will download all species this command supports. See below for a full list of available species (default "all")

Full list of available species
  hs -> Human (Homo sapiens)
  mm -> Mouse (Mus musculus)
  rn -> Rat (Rattus norvegicus)
  dr -> Zebrafish (Danio rerio)
  dm -> Fruitfly (Drosophila melanogaster)

Examples
  mirna
    run using the default values, which is equivalent to mirna -s all -r 21; this will download annotation files for version 21 and extract all species

  mirna -r 22 -s hs
    downloads version 22 and only extracts human annotations
```
