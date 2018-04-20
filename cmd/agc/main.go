package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Elsewhen-Studios/go-agc/cpu"
	"github.com/Elsewhen-Studios/go-agc/memory"
)

var (
	yaAGCFormat = flag.Bool("yaagc", false, "Indicates that the memory file is in the yaAGC format")
	debug       = flag.Bool("debug", false, "Execute with debugger attached")
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		// wrong number of arguments provided
		fmt.Fprintln(os.Stderr, "incorrect number of arguments")
		flag.PrintDefaults()
		return
	}

	binFile := flag.Arg(0)

	var (
		coreMemReader io.Reader
		leftAligned   bool
	)
	if *yaAGCFormat {
		// the core memory file is in the yaAGC format so we
		// have to do some manipulation
		// see http://www.ibiblio.org/apollo/developer.html#CoreFormat
		raw, err := ioutil.ReadFile(binFile)
		if err != nil {
			fatal("failed to read core rope file", err)
		}

		// the order of the banks in the file is 2, 3, 0, 1, 4, 5, 6, etc
		// so we have to swap 0/1 with 2/3 since our loader expects them
		// in sequential order
		const bankSize = 1024 * 2
		src23 := raw[0 : bankSize*2]
		src01 := raw[bankSize*2 : bankSize*4]
		tmp := make([]byte, 1024*2*2)
		copy(tmp, src23)
		copy(src23, src01)
		copy(src01, tmp)
		coreMemReader = bytes.NewReader(raw)
		leftAligned = true
	} else {
		f, err := os.Open(binFile)
		if err != nil {
			fatal("failed to open core rope file", err)
		}
		defer f.Close()
		coreMemReader = f
	}

	mm := new(memory.Main)
	l := &memory.Loader{MM: mm, LeftAligned: leftAligned}
	if _, err := io.Copy(l, coreMemReader); err != nil {
		fatal("failed to load main memory", err)
	}

	theCPU := cpu.NewCPU(mm)

	if *debug {
		d := cpu.NewInteractiveDebugger()
		go d.Run()
		theCPU.Debugger = d
	}

	theCPU.Run()
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v", msg, err)
	os.Exit(1)
}
