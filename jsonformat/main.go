// Description: A tool to format and validate JSON from a file or stdin.
// It reads JSON input, validates it, and writes it to stdout in a pretty-printed form.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	input := flag.String("file", "", "Path to JSON file (reads from stdin if not provided)")
	compact := flag.Bool("compact", false, "Output compact JSON instead of pretty-printed")
	validate := flag.Bool("validate", false, "Only validate JSON without printing output")
	indent := flag.String("indent", "  ", "Indentation string for pretty-printing")
	flag.Parse()

	var reader io.Reader
	if *input != "" {
		f, err := os.Open(*input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		reader = f
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if !json.Valid(data) {
		fmt.Fprintln(os.Stderr, "Error: invalid JSON")
		os.Exit(1)
	}

	if *validate {
		fmt.Println("JSON is valid.")
		return
	}

	var out bytes.Buffer
	if *compact {
		if err := json.Compact(&out, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error compacting JSON: %v\n", err)
			os.Exit(1)
		}
		out.WriteByte('\n')
	} else {
		if err := json.Indent(&out, data, "", *indent); err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		out.WriteByte('\n')
	}

	if _, err := os.Stdout.Write(out.Bytes()); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}
