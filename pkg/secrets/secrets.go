package secrets

import (
	"encoding/json"
	"fmt"
	"os"
)

// Secrets contains the Jira API credentials
type Secrets struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	URL      string `json:"url"`
}

// FromFile reads the JSON file and populates the Secrets struct
func (s *Secrets) FromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening secrets file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(s)
}
