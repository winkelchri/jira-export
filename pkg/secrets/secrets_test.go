package secrets

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func CreateTestSecretFile(filename string) {
	// Extract the parent directory from the filename
	dir := filepath.Dir(filename)

	// Create all parent directories if they don't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(fmt.Errorf("error creating directory: %v", err))
		}
	}

	// Create a test secrets file
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Write the test secrets to the file
	_, err = file.WriteString(`{
		"username": "testuser",
		"token": "testtoken",
		"url": "https://testurl.atlassian.net"
	}`)

	if err != nil {
		log.Fatal(err)
	}
}

func RemoveTestSecretsFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
}

func TestReadFromFile(t *testing.T) {
	s := Secrets{}
	f := "test/secrets.json"
	CreateTestSecretFile(f)
	err := s.FromFile(f)
	if err != nil {
		t.Errorf("error reading secrets file: %v", err)
	}

	if s.Username != "testuser" {
		t.Errorf("expected username to be testuser, got %s", s.Username)
	}

	if s.Token != "testtoken" {
		t.Errorf("expected token to be testtoken, got %s", s.Token)
	}

	if s.URL != "https://testurl.atlassian.net" {
		t.Errorf("expected url to be https://testurl.atlassian.net, got %s", s.URL)
	}

	RemoveTestSecretsFile(f)
}
