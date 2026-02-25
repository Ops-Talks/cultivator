package module

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// HTTPModuleSource handles Terraform sources from HTTP/HTTPS URLs
// Single Responsibility: handles only http(s):// sources
type HTTPModuleSource struct {
	client *http.Client
}

// NewHTTPModuleSource creates a new HTTP module source handler
func NewHTTPModuleSource() *HTTPModuleSource {
	return &HTTPModuleSource{
		client: &http.Client{},
	}
}

// Type returns the source type identifier
func (h *HTTPModuleSource) Type() string {
	return "http"
}

// Parse extracts repository information from an HTTP(S) source
// Example: https://github.com/org/repo/archive/refs/tags/v1.0.0.tar.gz
// Example: https://github.com/org/repo/archive/refs/heads/main.zip
func (h *HTTPModuleSource) Parse(source string) (*SourceInfo, error) {
	// Validate that it's a proper HTTPS/HTTP URL
	if err := isValidURL(source); err != nil {
		return nil, fmt.Errorf("invalid HTTP URL: %w", err)
	}

	if !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://") {
		return nil, fmt.Errorf("http source must start with http:// or https://")
	}

	// Extract subpath if present (after #)
	subPath := ""
	if idx := strings.Index(source, "#"); idx != -1 {
		subPath = source[idx+1:]
		source = source[:idx]
	}

	return &SourceInfo{
		Type:      "http",
		URL:       source,
		SubPath:   subPath,
		Ref:       h.extractRefFromURL(source),
		RawSource: source,
	}, nil
}

// FetchVersion retrieves the HTTP response headers to validate URL availability
// For HTTP sources, we use the URL as the "version" (immutable URLs)
func (h *HTTPModuleSource) FetchVersion(ctx context.Context, source string) (string, error) {
	info, err := h.Parse(source)
	if err != nil {
		return nil, err
	}

	// Make a HEAD request to validate the URL is accessible
	req, err := http.NewRequestWithContext(ctx, "HEAD", info.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// For HTTP sources, use the ETag header as version if available, otherwise use the URL
	if etag := resp.Header.Get("ETag"); etag != "" {
		return etag, nil
	}

	// If no ETag, return a hash of the URL as a consistent version
	return h.hashURL(info.URL), nil
}

// Checkout downloads and extracts the HTTP archive to the specified directory
func (h *HTTPModuleSource) Checkout(ctx context.Context, source string, workdir string) error {
	info, err := h.Parse(source)
	if err != nil {
		return err
	}

	// Download the file
	tempFile, err := h.downloadFile(ctx, info.URL)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile)

	// Create working directory
	if err := os.MkdirAll(workdir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Extract based on file extension
	if strings.HasSuffix(info.URL, ".tar.gz") || strings.HasSuffix(info.URL, ".tgz") {
		return h.extractTarGz(tempFile, workdir, info.SubPath)
	} else if strings.HasSuffix(info.URL, ".zip") {
		return h.extractZip(tempFile, workdir, info.SubPath)
	}

	return fmt.Errorf("unsupported archive format: %s", info.URL)
}

// downloadFile downloads a file from HTTP(S) and returns the path to the temp file
// Extracted method to follow DRY principle
func (h *HTTPModuleSource) downloadFile(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "cultivator-module-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy response body to temp file
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	return tempFile.Name(), nil
}

// extractTarGz extracts a tar.gz archive to the working directory
// Extracted method to keep Checkout() clean (DRY + Single Responsibility)
func (h *HTTPModuleSource) extractTarGz(archivePath, workdir, subPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Skip if file is outside the target subpath
		if !h.shouldExtractFile(header.Name, subPath) {
			continue
		}

		// Extract file to working directory
		targetPath := h.getTargetPath(header.Name, subPath, workdir)
		if err := h.extractFile(targetPath, tr, header.FileInfo().IsDir()); err != nil {
			return err
		}
	}

	return nil
}

// extractZip extracts a zip archive to the working directory
// Extracted method to keep Checkout() clean (DRY + Single Responsibility)
func (h *HTTPModuleSource) extractZip(archivePath, workdir, subPath string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		// Skip if file is outside the target subpath
		if !h.shouldExtractFile(file.Name, subPath) {
			continue
		}

		// Extract file to working directory
		targetPath := h.getTargetPath(file.Name, subPath, workdir)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		if err := h.writeFile(targetPath, src); err != nil {
			src.Close()
			return err
		}
		src.Close()
	}

	return nil
}

// shouldExtractFile determines if a file should be extracted based on subPath
// DRY: single location for extraction logic
func (h *HTTPModuleSource) shouldExtractFile(filePath, subPath string) bool {
	if subPath == "" {
		return true
	}
	return strings.Contains(filePath, subPath)
}

// getTargetPath calculates the target path for extracted files
// DRY: single location for path calculation
func (h *HTTPModuleSource) getTargetPath(filePath, subPath, workdir string) string {
	// Remove the root directory name (e.g., "repo-main/")
	parts := strings.Split(filePath, "/")
	if len(parts) > 1 {
		filePath = filepath.Join(parts[1:]...)
	}

	// Remove subpath prefix if present
	if subPath != "" && strings.HasPrefix(filePath, subPath) {
		filePath = strings.TrimPrefix(filePath, subPath)
		filePath = strings.TrimPrefix(filePath, "/")
	}

	return filepath.Join(workdir, filePath)
}

// extractFile writes a file from an archive to disk
// Extracted method for DRY principle
func (h *HTTPModuleSource) extractFile(targetPath string, src io.Reader, isDir bool) error {
	if isDir {
		return os.MkdirAll(targetPath, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return h.writeFile(targetPath, src)
}

// writeFile writes content from a reader to a file
// DRY: single location for file writing logic
func (h *HTTPModuleSource) writeFile(targetPath string, src io.Reader) error {
	dst, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// extractRefFromURL tries to extract version/tag from the URL
// For GitHub, this would be something like "v1.0.0" or "main"
// DRY: single location for URL parsing
func (h *HTTPModuleSource) extractRefFromURL(url string) string {
	// Look for common patterns in archive URLs
	patterns := []string{
		"/archive/refs/tags/",
		"/archive/refs/heads/",
		"/releases/download/",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(url, pattern); idx != -1 {
			remainder := url[idx+len(pattern):]
			// Take everything up to the file extension
			if dotIdx := strings.LastIndex(remainder, "."); dotIdx != -1 {
				return remainder[:dotIdx]
			}
			return remainder
		}
	}

	return "HEAD"
}

// hashURL creates a simple hash of the URL for version comparison
func (h *HTTPModuleSource) hashURL(url string) string {
	return fmt.Sprintf("http-%d", len(url))
}
