package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractGlobalConfigArg(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantArgs   []string
		wantConfig string
		wantErr    bool
	}{
		{
			name:       "short flag before command",
			args:       []string{"-f", "custom.yml", "test", "-scenario", "basic_chat"},
			wantArgs:   []string{"test", "-scenario", "basic_chat"},
			wantConfig: "custom.yml",
		},
		{
			name:       "long flag equals syntax",
			args:       []string{"--config=custom.yml", "validate", "-scenario", "demo.yaml"},
			wantArgs:   []string{"validate", "-scenario", "demo.yaml"},
			wantConfig: "custom.yml",
		},
		{
			name:       "flag without command",
			args:       []string{"-f", "custom.yml"},
			wantArgs:   []string{},
			wantConfig: "custom.yml",
		},
		{
			name:    "missing flag value",
			args:    []string{"-f"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, gotConfig, err := extractGlobalConfigArg(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("extractGlobalConfigArg() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Fatalf("extractGlobalConfigArg() args = %v, want %v", gotArgs, tt.wantArgs)
			}
			if gotConfig != tt.wantConfig {
				t.Fatalf("extractGlobalConfigArg() config = %q, want %q", gotConfig, tt.wantConfig)
			}
		})
	}
}

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
