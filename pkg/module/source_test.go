package module

import (
	"testing"
)

func TestSourceParser_Parse_GitSource(t *testing.T) {
	parser := NewSourceParser()

	tests := []struct {
		name        string
		source      string
		expectedURL string
		expectedRef string
		wantErr     bool
	}{
		{
			name:        "git ssh with ref",
			source:      "git::ssh://git@github.com/org/repo.git//vpc?ref=v1.0.0",
			expectedURL: "ssh://git@github.com/org/repo.git",
			expectedRef: "v1.0.0",
			wantErr:     false,
		},
		{
			name:        "git https with ref",
			source:      "git::https://github.com/org/repo//database?ref=main",
			expectedURL: "https://github.com/org/repo",
			expectedRef: "main",
			wantErr:     false,
		},
		{
			name:        "git without ref defaults to HEAD",
			source:      "git::https://github.com/org/repo//app",
			expectedURL: "https://github.com/org/repo",
			expectedRef: "HEAD",
			wantErr:     false,
		},
		{
			name:    "invalid git source without prefix",
			source:  "https://github.com/org/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parser.Parse(tt.source)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if info.URL != tt.expectedURL {
					t.Errorf("URL mismatch: got %q, want %q", info.URL, tt.expectedURL)
				}
				if info.Ref != tt.expectedRef {
					t.Errorf("Ref mismatch: got %q, want %q", info.Ref, tt.expectedRef)
				}
				if info.Type != "git" {
					t.Errorf("Type mismatch: got %q, want %q", info.Type, "git")
				}
			}
		})
	}
}

func TestSourceParser_Parse_HTTPSource(t *testing.T) {
	parser := NewSourceParser()

	tests := []struct {
		name        string
		source      string
		expectedURL string
		wantErr     bool
	}{
		{
			name:        "https github archive",
			source:      "https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz",
			expectedURL: "https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz",
			wantErr:     false,
		},
		{
			name:        "https github releases",
			source:      "https://github.com/org/repo/releases/download/v1.0.0/module.zip",
			expectedURL: "https://github.com/org/repo/releases/download/v1.0.0/module.zip",
			wantErr:     false,
		},
		{
			name:    "invalid http url",
			source:  "http://invalid url with spaces",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parser.Parse(tt.source)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if info.URL != tt.expectedURL {
					t.Errorf("URL mismatch: got %q, want %q", info.URL, tt.expectedURL)
				}
				if info.Type != "http" {
					t.Errorf("Type mismatch: got %q, want %q", info.Type, "http")
				}
			}
		})
	}
}

func TestSourceParser_DetectSourceType(t *testing.T) {
	parser := NewSourceParser()

	tests := []struct {
		source       string
		expectedType string
	}{
		{"git::https://github.com/org/repo", "git"},
		{"git::ssh://git@github.com/org/repo.git", "git"},
		{"https://example.com/module.tar.gz", "http"},
		{"http://example.com/module.zip", "http"},
		{"file://local/path", ""},
		{"/local/path", ""},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			sourceType := parser.detectSourceType(tt.source)
			if sourceType != tt.expectedType {
				t.Errorf("detectSourceType(%q) = %q, want %q", tt.source, sourceType, tt.expectedType)
			}
		})
	}
}

func TestGitModuleSource_Type(t *testing.T) {
	git := NewGitModuleSource()
	if git.Type() != "git" {
		t.Errorf("Type() = %q, want %q", git.Type(), "git")
	}
}

func TestGitModuleSource_Parse(t *testing.T) {
	git := NewGitModuleSource()

	tests := []struct {
		name        string
		source      string
		expectedURL string
		expectedRef string
		expectedSub string
		wantErr     bool
	}{
		{
			name:        "valid git source with subpath and ref",
			source:      "git::https://github.com/org/repo//vpc?ref=v1.0.0",
			expectedURL: "https://github.com/org/repo",
			expectedRef: "v1.0.0",
			expectedSub: "/vpc",
			wantErr:     false,
		},
		{
			name:        "git source without subpath",
			source:      "git::https://github.com/org/repo?ref=main",
			expectedURL: "https://github.com/org/repo",
			expectedRef: "main",
			expectedSub: "",
			wantErr:     false,
		},
		{
			name:    "invalid source without git:: prefix",
			source:  "https://github.com/org/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := git.Parse(tt.source)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if info.URL != tt.expectedURL {
					t.Errorf("URL mismatch: got %q, want %q", info.URL, tt.expectedURL)
				}
				if info.Ref != tt.expectedRef {
					t.Errorf("Ref mismatch: got %q, want %q", info.Ref, tt.expectedRef)
				}
				if info.SubPath != tt.expectedSub {
					t.Errorf("SubPath mismatch: got %q, want %q", info.SubPath, tt.expectedSub)
				}
			}
		})
	}
}

func TestHTTPModuleSource_Type(t *testing.T) {
	http := NewHTTPModuleSource()
	if http.Type() != "http" {
		t.Errorf("Type() = %q, want %q", http.Type(), "http")
	}
}

func TestHTTPModuleSource_Parse(t *testing.T) {
	http := NewHTTPModuleSource()

	tests := []struct {
		name        string
		source      string
		expectedURL string
		wantErr     bool
	}{
		{
			name:        "valid https URL",
			source:      "https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz",
			expectedURL: "https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz",
			wantErr:     false,
		},
		{
			name:    "invalid URL",
			source:  "https://invalid url with spaces",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := http.Parse(tt.source)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr && info.URL != tt.expectedURL {
				t.Errorf("URL mismatch: got %q, want %q", info.URL, tt.expectedURL)
			}
		})
	}
}

func TestExtractSubPath(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"git::https://github.com/org/repo//vpc", "/vpc"},
		{"git::https://github.com/org/repo//modules/database", "/modules/database"},
		{"git::https://github.com/org/repo//vpc?ref=v1.0.0", "/vpc"},
		{"git::https://github.com/org/repo", ""},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := extractSubPath(tt.source)
			if result != tt.expected {
				t.Errorf("extractSubPath(%q) = %q, want %q", tt.source, result, tt.expected)
			}
		})
	}
}

func TestExtractQueryParam(t *testing.T) {
	tests := []struct {
		source    string
		param     string
		expected  string
	}{
		{"git::https://github.com/org/repo?ref=v1.0.0", "ref", "v1.0.0"},
		{"git::https://github.com/org/repo?ref=main&depth=1", "ref", "main"},
		{"git::https://github.com/org/repo?depth=1&ref=v2.0.0", "ref", "v2.0.0"},
		{"git::https://github.com/org/repo", "ref", ""},
		{"git::https://github.com/org/repo?tag=v1.0.0", "ref", ""},
	}

	for _, tt := range tests {
		t.Run(tt.source+"+"+tt.param, func(t *testing.T) {
			result := extractQueryParam(tt.source, tt.param)
			if result != tt.expected {
				t.Errorf("extractQueryParam(%q, %q) = %q, want %q", tt.source, tt.param, result, tt.expected)
			}
		})
	}
}
