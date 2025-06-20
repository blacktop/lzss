package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blacktop/lzss"
)

func main() {
	var (
		compress   = flag.Bool("c", false, "compress input file")
		decompress = flag.Bool("d", false, "decompress input file")
		output     = flag.String("o", "", "output file (default: input + .lzss or input without .lzss)")
		help       = flag.Bool("h", false, "show help")
	)
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nLZSS compression/decompression tool\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -c input.txt              # compress to input.txt.lzss\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d input.txt.lzss         # decompress to input.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c -o output.lzss input.txt # compress with custom output\n", os.Args[0])
	}
	
	flag.Parse()
	
	if *help {
		flag.Usage()
		return
	}
	
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: Please specify exactly one input file\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	if !*compress && !*decompress {
		fmt.Fprintf(os.Stderr, "Error: Please specify either -c (compress) or -d (decompress)\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	if *compress && *decompress {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify both -c and -d\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	inputFile := flag.Arg(0)
	
	// Read input file
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}
	
	var result []byte
	var outputFile string
	
	if *compress {
		result = lzss.Compress(data)
		if *output != "" {
			outputFile = *output
		} else {
			outputFile = inputFile + ".lzss"
		}
		fmt.Printf("Compressed %d bytes to %d bytes (%.2f%% ratio)\n", 
			len(data), len(result), float64(len(result))/float64(len(data))*100)
	} else {
		result = lzss.Decompress(data)
		if *output != "" {
			outputFile = *output
		} else {
			// Remove .lzss extension if present
			if filepath.Ext(inputFile) == ".lzss" {
				outputFile = inputFile[:len(inputFile)-5]
			} else {
				outputFile = inputFile + ".decompressed"
			}
		}
		fmt.Printf("Decompressed %d bytes to %d bytes\n", len(data), len(result))
	}
	
	// Write output file
	err = ioutil.WriteFile(outputFile, result, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Output written to: %s\n", outputFile)
}