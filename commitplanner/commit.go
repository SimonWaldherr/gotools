// Description: Commit planner is a tool to plan commits for a git repository.
// It reads a file with the commits and applies them to the git repository.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Commit struct {
	Author  string
	Email   string
	Date    time.Time
	Message string
	Files   []*File
}

type File struct {
	Name    string
	Content string
}

func main() {
	commits, err := parseCommitFile("commit-file.txt")
	if err != nil {
		fmt.Println("Fehler beim Parsen der commit-file.txt:", err)
		return
	}

	fmt.Printf("Es wurden %d Commits gefunden.\n", len(commits))
	fmt.Printf("Der erste Commit hat %d Dateien.\n", len(commits[0].Files))
	fmt.Printf("%#v, %#v\n", commits, commits[0].Files)

	if len(commits) > 0 {
		fmt.Println("Der erste Commit hat folgende Dateien:")
		for _, file := range commits[0].Files {
			fmt.Println(file.Name)
		}
	} else {
		fmt.Println("Es wurden keine Commits gefunden.")
		return
	}

	//if true {
	//	fmt.Println("Es werden keine Commits angewendet.")
	//	return
	//}

	err = applyCommits(commits)
	if err != nil {
		fmt.Println("Fehler beim Anwenden der Commits:", err)
		return
	}
}

func parseCommitFile(filename string) ([]Commit, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var commits []Commit
	var currentCommit *Commit
	var currentFile *File

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Line X: %s\n", line)

		if strings.HasPrefix(line, "--- Commit") {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}
			currentCommit = &Commit{}
		} else if strings.HasPrefix(line, "--- Datei") {
			if currentFile != nil {
				currentCommit.Files = append(currentCommit.Files, currentFile)
			}
			currentFile = &File{}
		} else if strings.HasPrefix(line, "Autor:") {
			currentCommit.Author = strings.TrimSpace(strings.TrimPrefix(line, "Autor:"))
		} else if strings.HasPrefix(line, "Email:") {
			currentCommit.Email = strings.TrimSpace(strings.TrimPrefix(line, "Email:"))
		} else if strings.HasPrefix(line, "Datum:") {
			dateStr := strings.TrimSpace(strings.TrimPrefix(line, "Datum:"))
			date, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				fmt.Printf("Fehler beim Parsen des Datums: %s\n", dateStr)
				return nil, err
			}
			currentCommit.Date = date
		} else if strings.HasPrefix(line, "Message:") {
			currentCommit.Message = strings.TrimSpace(strings.TrimPrefix(line, "Message:"))
		} else if strings.HasPrefix(line, "Name:") {
			currentFile.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
		} else if strings.HasPrefix(line, "Inhalt:") {
			content := strings.Builder{}
			for scanner.Scan() {
				line = scanner.Text()
				fmt.Printf("Line Y: %s\n", line)

				if strings.HasPrefix(line, "---") {
					fmt.Printf("Line Z (end): %s\n", line)
					currentFile.Content = content.String()
					currentCommit.Files = append(currentCommit.Files, currentFile)
					break
				}

				if strings.HasPrefix(line, "{{.PrevContent}}") {
					breakVar := false
					for i := len(commits) - 1; i >= 0; i-- {
						for _, files := range commits[i].Files {
							if files.Name == currentFile.Name {
								content.WriteString(files.Content)
								breakVar = true
								break
							}
						}
						if breakVar {
							break
						}
					}
				} else {
					content.WriteString(line)
				}
				content.WriteString("\n")
			}
			currentFile.Content = content.String()
		}
	}

	if currentCommit != nil {
		if currentFile != nil {
			currentCommit.Files = append(currentCommit.Files, currentFile)
		}
		commits = append(commits, *currentCommit)
	}

	return commits, scanner.Err()
}

func applyCommits(commits []Commit) error {
	for _, commit := range commits {
		for _, file := range commit.Files {
			err := os.WriteFile(file.Name, []byte(file.Content), 0644)
			if err != nil {
				return err
			}

			cmd := exec.Command("git", "add", file.Name)
			err = cmd.Run()
			if err != nil {
				return err
			}
		}

		commitDate := commit.Date.Format(time.RFC3339)
		env := fmt.Sprintf("GIT_COMMITTER_DATE=%s", commitDate)
		cmd := exec.Command("git", "commit", "-m", commit.Message, "--author", fmt.Sprintf("%s <%s>", commit.Author, commit.Email), "--date", commitDate)
		cmd.Env = append(os.Environ(), env)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
