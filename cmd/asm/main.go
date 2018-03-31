package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Elsewhen-Studios/go-agc/assembler"
)

func main() {
	sourceFile := flag.String("source", "", "the assembly source file")
	outputFile := flag.String("output", "image.bin", "the binary output file")

	flag.Parse()

	if *sourceFile == "" {
		log.Fatal("no source file specified")
	}

	if *outputFile == "" {
		log.Fatal("no output file specified")
	}

	var a assembler.Assembler
	if a.Assemble(*sourceFile) {
		output, err := os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { output.Close() }()

		if err := a.WriteOut(output); err != nil {
			log.Println(err)

			if err := output.Close(); err != nil {
				log.Println(err)
			}

			if err := os.Remove(*outputFile); err != nil {
				log.Println(err)
			}
		}
	}

	for _, p := range a.Problems {
		fmt.Println(p)
	}
}
