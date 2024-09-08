package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// getExcelFiles sucht nach allen .xlsx, .xlsm und .xlsb Dateien im aktuellen Verzeichnis.
func getExcelFiles() ([]string, error) {
	var files []string
	// Walk durch das aktuelle Verzeichnis, um alle Excel-Dateien zu finden
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Prüft, ob die Datei die gesuchten Endungen hat
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".xlsx") || strings.HasSuffix(info.Name(), ".xlsm") || strings.HasSuffix(info.Name(), ".xlsb")) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// normalizeHeader nimmt einen Header-String und normalisiert ihn.
// Dies bedeutet: Kleinbuchstaben, Entfernen von Sonderzeichen und Leerzeichen.
func normalizeHeader(header string) string {
	// Kleinbuchstaben umwandeln
	header = strings.ToLower(header)
	// Entfernen von Sonderzeichen und Leerzeichen mit einem regulären Ausdruck
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return reg.ReplaceAllString(header, "")
}

// getColumns liest die erste Zeile eines Tabellenblatts, welche die Spaltenüberschriften enthält,
// und normalisiert diese. Gibt sowohl die normalisierten als auch die Originalüberschriften zurück.
func getColumns(f *excelize.File, sheetName string) (map[string]string, []string, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	// Mapping von normalisierten Headern auf die Originale und die Reihenfolge der Spalten
	columns := make(map[string]string)
	var columnOrder []string
	if len(rows) > 0 {
		for _, cell := range rows[0] {
			normalized := normalizeHeader(cell)
			columns[normalized] = cell  // Speichere das Original unter dem normalisierten Header
			columnOrder = append(columnOrder, normalized) // Die Reihenfolge der Spalten speichern
		}
	}
	return columns, columnOrder, nil
}

// combineFiles kombiniert alle Excel-Dateien zu einer großen Datei und berücksichtigt
// dabei nur die gemeinsamen Spalten. Die Spaltenüberschriften werden normalisiert.
func combineFiles(files []string, outputFileName string) error {
	// Erstellen einer neuen Datei für die kombinierten Daten
	outputFile := excelize.NewFile()
	sheetName := "Combined"
	outputFile.NewSheet(sheetName)

	// Aktuelle Zeile in der neuen Datei, ab der geschrieben wird
	currentRow := 1
	commonColumns := make(map[string]string)  // Mapping von normalisierten Headern zu Original-Headern
	columnOrder := []string{}                 // Speichert die Reihenfolge der gemeinsamen Spalten

	for _, file := range files {
		fmt.Printf("Processing file: %s\n", file)
		f, err := excelize.OpenFile(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, sheet := range f.GetSheetList() {
			columns, colOrder, err := getColumns(f, sheet)
			_ = colOrder
			if err != nil {
				fmt.Printf("Error reading columns from sheet %s in file %s: %v\n", sheet, file, err)
				continue
			}

			// Bestimmen der gemeinsamen Spalten über alle Dateien hinweg
			if len(commonColumns) == 0 {
				// Beim ersten Durchlauf fügen wir alle Spalten zur gemeinsamen Liste hinzu
				for normalized, original := range columns {
					commonColumns[normalized] = original
					columnOrder = append(columnOrder, normalized)
				}
			} else {
				// Entfernen von Spalten, die in diesem Blatt nicht vorhanden sind
				for col := range commonColumns {
					if _, ok := columns[col]; !ok {
						delete(commonColumns, col)
					}
				}
			}

			// Wenn es die erste Datei/Sheet ist, schreibe die Original-Header in die Ausgabe
			if currentRow == 1 {
				for i, normalized := range columnOrder {
					if original, ok := commonColumns[normalized]; ok {
						// Setze die Original-Header in die erste Zeile
						colName, _ := excelize.ColumnNumberToName(i + 1)
						outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, currentRow), original)
					}
				}
				currentRow++ // Nächste Zeile nach der Header-Zeile
			}

			// Lesen aller Zeilen dieses Tabellenblattes
			rows, err := f.GetRows(sheet)
			if err != nil {
				fmt.Printf("Error reading rows from sheet %s in file %s: %v\n", sheet, file, err)
				continue
			}

			// Iterieren durch alle Zeilen außer der ersten (Header)
			for i, row := range rows {
				if i == 0 {
					// Die Header-Zeile wird übersprungen, da sie schon geschrieben wurde
					continue
				}

				// Erstellen einer Map für die aktuellen Spalten dieser Zeile
				cellMap := make(map[string]string)
				for j, cell := range row {
					currentHeader := normalizeHeader(rows[0][j])
					cellMap[currentHeader] = cell
				}

				// Schreibe die Zellen in die entsprechenden Spalten in der kombinierten Datei
				for colIndex, normalized := range columnOrder {
					if value, ok := cellMap[normalized]; ok {
						colName, _ := excelize.ColumnNumberToName(colIndex + 1)
						outputFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, currentRow), value)
					}
				}
				currentRow++ // Nächste Zeile
			}
		}

		// Schließen der Datei
		_ = f.Close()
	}

	// Speichern der kombinierten Datei
	if err := outputFile.SaveAs(outputFileName); err != nil {
		return fmt.Errorf("error saving combined file: %w", err)
	}

	fmt.Printf("Files combined and saved as %s\n", outputFileName)
	return nil
}

func main() {
	// 1. Excel-Dateien aus dem aktuellen Verzeichnis finden
	files, err := getExcelFiles()
	if err != nil {
		fmt.Printf("Error finding Excel files: %v\n", err)
		return
	}

	// Überprüfen, ob Dateien gefunden wurden
	if len(files) == 0 {
		fmt.Println("No Excel files found in the current directory.")
		return
	}

	// 2. Kombinieren der gefundenen Dateien in eine große Datei
	err = combineFiles(files, "CombinedOutput.xlsx")
	if err != nil {
		fmt.Printf("Error combining files: %v\n", err)
	}
}
