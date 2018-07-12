package main

import (
	"flag"
	"c2"
	"fmt"
)

// Main

func main() {
	genFiles := flag.String("g", "", "Generates required files prior to language creation.")
	parseLang := flag.String("p", "", "Parses a language file, creating symbols and a parse table.")
	flag.Parse()

	if *genFiles != "" && flag.NArg() == 1 {
		c2.GenerateFiles(*genFiles, flag.Arg(0))
	} else if *parseLang != "" {
		err := c2.ParseFile(*parseLang)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Compiler-Compiler (C2)" +
			"\n\tCompiles a language file (.C2L) into go source." +
			"\n\t-g <path/to/create/lang.c2l> : Generates required files prior to language creation." +
			"\n\t-p <path/to/lang.c2l>        : Parses a language file, creating symbols and a parse table.")
	}
}