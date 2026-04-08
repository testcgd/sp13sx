package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateScenarioParsesAndValidatesYAML(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "valid.yaml")
	invalidPath := filepath.Join(tempDir, "invalid.yaml")

	if err := os.WriteFile(validPath, []byte(`
name: valid
turns:
  - response_id: resp_1
    events:
      - type: message
        content: ok
`), 0644); err != nil {
		t.Fatalf("write valid scenario: %v", err)
	}

	if err := os.WriteFile(invalidPath, []byte(`
name: invalid
turns:
  - response_id: resp_1
    events:
      - content: missing-type
`), 0644); err != nil {
		t.Fatalf("write invalid scenario: %v", err)
	}

	if err := validateScenario(validPath); err != nil {
		t.Fatalf("expected valid scenario to pass, got %v", err)
	}

	if err := validateScenario(invalidPath); err == nil {
		t.Fatal("expected invalid scenario to fail validation")
	}
}
