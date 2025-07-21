package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	var (
		outputFile  = flag.String("output", "/dev/stdout", "Output file to write")
		outputShort = flag.String("o", "/dev/stdout", "Output file to write (shorthand)")
		scopeDir    = flag.String("scope", "", "Directory containing all files eligible for concatenation")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <root>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConcatenates Markdown files intelligently.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <root>    Root markdown file to start from\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: exactly one root file must be specified\n")
		flag.Usage()
		os.Exit(1)
	}

	rootFile := args[0]

	output := *outputFile
	if *outputShort != "/dev/stdout" {
		output = *outputShort
	}

	if err := run(rootFile, output, *scopeDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(rootFile, outputFile, explicitScope string) error {
	if err := ValidateRootFile(rootFile); err != nil {
		return fmt.Errorf("invalid root file: %w", err)
	}

	scopeDir, err := DetermineScopeDir(rootFile, explicitScope)
	if err != nil {
		return fmt.Errorf("failed to determine scope directory: %w", err)
	}

	rootAbs, err := filepath.Abs(rootFile)
	if err != nil {
		return fmt.Errorf("failed to resolve root file path: %w", err)
	}

	traversal := NewFileTraversal(rootAbs, scopeDir)
	orderedFiles, err := traversal.Traverse()
	if err != nil {
		return fmt.Errorf("failed to traverse files: %w", err)
	}

	if len(orderedFiles) == 0 {
		return fmt.Errorf("no files found to process")
	}

	var writer io.Writer
	if outputFile == "/dev/stdout" {
		writer = os.Stdout
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", outputFile, err)
		}
		defer f.Close()
		writer = f
	}

	processor := NewFileProcessor(scopeDir, orderedFiles)

	for i, filename := range orderedFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", filename, err)
		}

		processedContent, err := processor.ProcessFile(filename, content)
		if err != nil {
			return fmt.Errorf("failed to process file %q: %w", filename, err)
		}

		if i > 0 {
			if _, err := writer.Write([]byte("\n\n")); err != nil {
				return fmt.Errorf("failed to write separator: %w", err)
			}
		}

		if _, err := writer.Write(processedContent); err != nil {
			return fmt.Errorf("failed to write processed content for file %q: %w", filename, err)
		}
	}

	return nil
}
