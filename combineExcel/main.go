package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Command-line flag for mode selection
var fastMode bool

func init() {
	// Define a command-line flag to enable fast mode
	flag.BoolVar(&fastMode, "fast", false, "Enable fast mode (load all files into memory)")
}

// getExcelFiles searches for all .xlsx, .xlsm, and .xlsb files in the current directory.
func getExcelFiles() ([]string, error) {
	var files []string
	// Walk through the current directory to find Excel files
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isExcelFile(info.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// isExcelFile checks if a file has .xlsx, .xlsm, or .xlsb extensions.
func isExcelFile(fileName string) bool {
	extensions := []string{".xlsx", ".xlsm", ".xlsb"}
	for _, ext := range extensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

// normalizeHeader normalizes the header by converting to lowercase and removing special characters and spaces.
func normalizeHeader(header string) string {
	header = strings.ToLower(header)
	// Remove non-alphanumeric characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return reg.ReplaceAllString(header, "")
}

// getColumns reads the first row of a sheet (considered the header row) and normalizes it.
// Returns both normalized and original headers.
func getColumns(f *excelize.File, sheetName string) (map[string]string, []string, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	columns := make(map[string]string)
	var columnOrder []string
	if len(rows) > 0 {
		// Process the first row as headers
		for _, cell := range rows[0] {
			normalized := normalizeHeader(cell)
			columns[normalized] = cell
			columnOrder = append(columnOrder, normalized)
		}
	}
	return columns, columnOrder, nil
}

// combineFiles combines all Excel files into one large file, with an order of columns based on the input files.
func combineFiles(files []string, outputFileName string) error {
	outputFile := excelize.NewFile()
	sheetName := "Combined"
	outputFile.NewSheet(sheetName)

	currentRow := 1
	columnsInOrder := []string{}
	columnMap := make(map[string]string)

	if fastMode {
		// Fast Mode: Load all files into memory first
		fileData := loadAllFiles(files)
		for fileIndex, data := range fileData {
			if err := processFileData(outputFile, sheetName, data, files[fileIndex], fileIndex == 0, &currentRow, &columnsInOrder, &columnMap); err != nil {
				return err
			}
		}
	} else {
		// Memory-Efficient Mode: Process files one by one
		for fileIndex, file := range files {
			fmt.Printf("Processing file: %s\n", file)
			f, err := excelize.OpenFile(file)
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, sheet := range f.GetSheetList() {
				columns, colOrder, err := getColumns(f, sheet)
				if err != nil {
					fmt.Printf("Error reading columns from sheet %s in file %s: %v\n", sheet, file, err)
					continue
				}

				// Update column orders and map for merging
				updateColumns(&columnsInOrder, &columnMap, colOrder, fileIndex == 0)
				if currentRow == 1 {
					writeHeaders(outputFile, sheetName, columnMap, columnsInOrder)
					currentRow++
				}

				if err := writeRowsFromFile(f, outputFile, sheetName, sheet, columnsInOrder, currentRow, file); err != nil {
					fmt.Printf("Error writing rows from sheet %s in file %s: %v\n", sheet, file, err)
					continue
				}
				rows, _ := f.GetRows(sheet)
				currentRow += len(rows) - 1
			}

			_ = f.Close()
		}
	}

	if err := outputFile.SaveAs(outputFileName); err != nil {
		return fmt.Errorf("error saving combined file: %w", err)
	}

	fmt.Printf("Files combined and saved as %s\n", outputFileName)
	return nil
}

// loadAllFiles loads all files into memory for fast processing.
func loadAllFiles(files []string) []map[string]map[string][][]string {
	var allData []map[string]map[string][][]string
	for _, file := range files {
		f, err := excelize.OpenFile(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fileData := make(map[string]map[string][][]string)
		for _, sheet := range f.GetSheetList() {
			rows, err := f.GetRows(sheet)
			if err != nil {
				fmt.Printf("Error reading rows from sheet %s in file %s: %v\n", sheet, file, err)
				continue
			}

			columns, colOrder, err := getColumns(f, sheet)
			_ = columns
			
			if err != nil {
				fmt.Printf("Error reading columns from sheet %s in file %s: %v\n", sheet, file, err)
				continue
			}

			// Save both column order and rows for this sheet
			fileData[sheet] = map[string][][]string{
				"columns": {colOrder},
				"data":    rows,
			}
		}
		allData = append(allData, fileData)
	}

	return allData
}

// processFileData processes preloaded file data for fast mode.
func processFileData(outputFile *excelize.File, sheetName string, fileData map[string]map[string][][]string, fileName string, isFirstFile bool, currentRow *int, columnsInOrder *[]string, columnMap *map[string]string) error {
	for sheet, data := range fileData {
		columns := data["columns"][0]
		rows := data["data"]

		updateColumns(columnsInOrder, columnMap, columns, isFirstFile)
		if *currentRow == 1 {
			writeHeaders(outputFile, sheetName, *columnMap, *columnsInOrder)
			(*currentRow)++
		}

		for i, row := range rows {
			if i == 0 {
				continue // Skip the header row
			}
			cellMap := mapRowToHeader(row, rows[0])
			for colIndex, normalized := range *columnsInOrder {
				if value, ok := cellMap[normalized]; ok {
					colName, _ := excelize.ColumnNumberToName(colIndex + 2) // Adjust index for filename column
					outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, *currentRow+i), value)
				}
			}
			// Write the filename as the first column value
			colName, _ := excelize.ColumnNumberToName(1) // Filename column
			outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, *currentRow+i), fileName)
		}
	}

	return nil
}

// writeRowsFromFile writes rows from a single file to the output file.
func writeRowsFromFile(inputFile *excelize.File, outputFile *excelize.File, sheetName, inputSheet string, columnOrder []string, currentRow int, fileName string) error {
	rows, err := inputFile.GetRows(inputSheet)
	if err != nil {
		return fmt.Errorf("error reading rows from sheet %s: %w", inputSheet, err)
	}

	for i, row := range rows {
		if i == 0 {
			continue // Skip the header row
		}
		cellMap := mapRowToHeader(row, rows[0])
		for colIndex, normalized := range columnOrder {
			if value, ok := cellMap[normalized]; ok {
				colName, _ := excelize.ColumnNumberToName(colIndex + 2) // Adjust index for filename column
				outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, currentRow+i), value)
			}
		}
		// Write the filename as the first column value
		colName, _ := excelize.ColumnNumberToName(1) // Filename column
		outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, currentRow+i), fileName)
	}

	return nil
}

// updateColumns updates the final column order and map for each file processed.
func updateColumns(finalOrder *[]string, columnMap *map[string]string, currentColumns []string, isFirstFile bool) {
	if isFirstFile {
		for _, col := range currentColumns {
			*finalOrder = append(*finalOrder, col)
			(*columnMap)[normalizeHeader(col)] = col
		}
	} else {
		for _, col := range currentColumns {


			normalized := normalizeHeader(col)
			if _, exists := (*columnMap)[normalized]; !exists {
				*finalOrder = append(*finalOrder, normalized)
				(*columnMap)[normalized] = col
			}
		}
	}
}

// writeHeaders writes headers to the output file.
func writeHeaders(outputFile *excelize.File, sheetName string, columnMap map[string]string, columnOrder []string) {
	for i, normalized := range columnOrder {
		if original, ok := columnMap[normalized]; ok {
			colName, _ := excelize.ColumnNumberToName(i + 2) // Adjust index for filename column
			outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 1), original)
		}
	}
}

// mapRowToHeader creates a map between the row's cells and the header names.
func mapRowToHeader(row, header []string) map[string]string {
	cellMap := make(map[string]string)
	for j, cell := range row {
		currentHeader := normalizeHeader(header[j])
		cellMap[currentHeader] = cell
	}
	return cellMap
}

func main() {
	flag.Parse()

	files, err := getExcelFiles()
	if err != nil {
		fmt.Printf("Error finding Excel files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No Excel files found in the current directory.")
		return
	}

	err = combineFiles(files, "CombinedOutput.xlsx")
	if err != nil {
		fmt.Printf("Error combining files: %v\n", err)
	}
}