package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lonegunmanb/avmfix/pkg"
)

const (
	folderFlag      = "folder"
	excludeFlag     = "exclude"
	helpFlag        = "help"
	
	folderUsage     = "The folder path to scan and apply fixes"
	excludeUsage    = "Glob matching pattern to exclude files/folders from processing"
	helpUsage       = "Show help information"
	
	errorMessage    = "Error during processing:"
	successMessage  = "Processing completed successfully"
)

func main() {
	var dirPath string
	var excludePattern string
	var showHelp bool

	flag.StringVar(&dirPath, folderFlag, "", folderUsage)
	flag.StringVar(&excludePattern, excludeFlag, "", excludeUsage)
	flag.BoolVar(&showHelp, helpFlag, false, helpUsage)
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\nOPTIONS:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --folder /path/to/terraform/files \n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --folder /path/to/terraform/files --exclude '**/test_*.tf'\n", os.Args[0])
	}
	
	flag.Parse()

	if showHelp {
		flag.Usage()
		return
	}

	if dirPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := pkg.DirectoryAutoFix(dirPath, excludePattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", errorMessage, err)
		os.Exit(1)
	}

	fmt.Println(successMessage)
}
