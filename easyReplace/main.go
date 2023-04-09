package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: easyReplace <input_file> <output_file> <search> <replace>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]
	search := os.Args[3]
	replace := os.Args[4]

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer file.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(outFile)

	for scanner.Scan() {
		line := scanner.Text()
		replacedLine := strings.ReplaceAll(line, search, replace)
		_, err := writer.WriteString(replacedLine + "\n")
		if err != nil {
			fmt.Println("Error writing to output file:", err)
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing writer:", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
		return
	}

	fmt.Println("Replacement operation completed successfully.")
}
