// Package util provides common utility functions for string manipulation.
package util

import "strings"

// Contains checks if a slice contains a string value
// Following DRY principle - reusable utility function
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate entries from a string slice
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Filter returns a new slice containing only elements that pass the predicate
func Filter(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map applies a function to each element and returns the results
func Map(slice []string, fn func(string) string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// TrimPrefixIfExists removes prefix from string if it exists, returns original if not
func TrimPrefixIfExists(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return strings.TrimPrefix(s, prefix)
	}
	return s
}

// SplitAndTrimSpace splits by delimiter and trims whitespace from each part
func SplitAndTrimSpace(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
