package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func StoreJSON(reader io.ReadCloser, filename string) error {
	// Extract the parent directory from the filename
	dir := filepath.Dir(filename)

	// Create all parent directories if they don't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	// Write the data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func WriteToFile(filename string, data string) error {
	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.WriteString(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func ListFiles(dir string, glob string) ([]string, error) {
	// Get all files in the directory
	files, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		return nil, fmt.Errorf("error listing files: %v", err)
	}

	return files, nil
}
