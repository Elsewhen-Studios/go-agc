package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	sourceFile := flag.String("source", "", "the assembly source file")

	flag.Parse()

	if *sourceFile == "" {
		log.Fatal("no source file specified")
	}

	r, err := os.Open(*sourceFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { r.Close() }()

	var l int64
	if info, err := r.Stat(); err != nil {
		log.Fatal(err)
	} else {
		l = info.Size()
	}

	off := 0
	for ; l > 1; l -= 2 {
		var v int16
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			log.Fatal(err)
		}

		if _, err := fmt.Printf("%05o: %05o\n", off, v); err != nil {
			log.Fatal(err)
		}
		off++
	}

	if l > 0 {
		var v int8
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			log.Fatal(err)
		}

		if _, err := fmt.Printf("%05o: %05o (truncated)\n", off, uint16(v)<<8); err != nil {
			log.Fatal(err)
		}
	}
}
