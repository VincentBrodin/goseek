package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/windows"
)

type FileMatch struct {
	name      string
	extension string
	dir       string
	match     float32
}

func split_file_name(file string) (name, extension string) {
	index := -1
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '.' {
			index = i
			break
		}
	}

	if index == -1 {
		return file, "*" // No extension found
	}

	return file[:index], file[index+1:]
}

func levenshtein_distance(string_a, string_b string) int {
	a_len := len(string_a)
	b_len := len(string_b)

	if a_len == 0 {
		// If a is empty, the distance is the number of characters in b
		return b_len
	} else if b_len == 0 {
		// If b is empty, the distance is the number of characters in a
		return a_len
	}

	// Create a matrix and set all values to 0
	matrix := make([][]int, a_len+1)
	for i := range matrix {
		matrix[i] = make([]int, b_len+1)
	}

	// Initialization of the first row and column
	for i := 0; i <= a_len; i++ {
		matrix[i][0] = i // Distance from string_a to an empty string
	}
	for j := 0; j <= b_len; j++ {
		matrix[0][j] = j // Distance from an empty string to string_b
	}

	// Calculate rows and columns distances
	for i := 1; i <= a_len; i++ {
		for j := 1; j <= b_len; j++ {
			cost := 0
			if string_a[i-1] != string_b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // Deletion, Insertion
				matrix[i-1][j-1]+cost,                   // Substitution
			)
		}
	}

	return matrix[a_len][b_len]
}

// levenshtein_match() test
func levenshtein_match(string_a, string_b string) float32 {
	distance := levenshtein_distance(string_a, string_b)
	match := float32(distance) / float32(max(len(string_a), len(string_b)))
	return 1 - match
}

func search_all_drives(file_name, extension string, ch chan<- FileMatch, wg *sync.WaitGroup) {
	drives := get_drives()
	for drive := range drives {
		wg.Add(1)
		go search_folder(drives[drive], file_name, extension, ch, wg)
	}

}

func hasReadPermission(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false // File doesn't exist or error occurred
	}

	return fileInfo.Mode().Perm()&os.FileMode(0400) != 0
}

func search_folder(folder, s_file_name, s_extension string, ch chan<- FileMatch, wg *sync.WaitGroup) {
	defer wg.Done()
	// if !hasReadPermission(folder) {
	// 	return
	// }
	entry, err := os.ReadDir(folder)

	if err != nil {
		// fmt.Println("Error reading:", folder)
		return
	}

	extension_sensitive := s_extension != "*"

	for _, file := range entry {
		if file.IsDir() {
			wg.Add(1)
			search_folder(filepath.Join(folder, file.Name()), s_file_name, s_extension, ch, wg)
		} else {
			file_name, extension := split_file_name(file.Name())
			// Only care about files that have the right extension
			if extension_sensitive && extension != s_extension {
				continue
			}
			match := levenshtein_match(file_name, s_file_name)
			if match < 0.75 {
				continue
			}

			file_match := FileMatch{name: file_name, extension: extension, dir: folder, match: match}
			ch <- file_match
		}
	}
}

func get_drives() []string {
	// GetLogicalDrives returns a bitmask representing all drives.
	drives := []string{}
	drivesBitmask, err := windows.GetLogicalDrives()
	if err != nil {
		fmt.Println("Error getting drives:", err)
		return drives
	}

	// Check each bit in the bitmask to see if a drive exists for that letter.
	for i := 0; i < 26; i++ { // There are 26 possible drive letters (A to Z)
		if drivesBitmask&(1<<i) != 0 {
			driveLetter := fmt.Sprintf("%c:\\", 'A'+i)
			drives = append(drives, driveLetter)
		}
	}
	return drives
}
