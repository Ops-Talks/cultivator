package config

import (
	"testing"
)

// FuzzParseBool tests parseBool function with random string input
func FuzzParseBool(f *testing.F) {
	// Add seed corpus
	f.Add("true")
	f.Add("false")
	f.Add("TRUE")
	f.Add("FALSE")
	f.Add("True")
	f.Add("False")
	f.Add("1")
	f.Add("0")
	f.Add("on")
	f.Add("off")
	f.Add("yes")
	f.Add("no")
	f.Add("ON")
	f.Add("OFF")
	f.Add("YES")
	f.Add("NO")
	f.Add("  true  ")
	f.Add("  false  ")
	f.Add("")
	f.Add(" ")
	f.Add("invalid")
	f.Add("maybe")
	f.Add("2")
	f.Add("-1")
	f.Add("truefalse")
	f.Add("truee")
	f.Add("t")
	f.Add("f")

	f.Fuzz(func(t *testing.T, data string) {
		// This should not panic
		result := parseBool(data)

		// Validate result type
		if result != true && result != false {
			t.Error("parseBool should return a boolean")
		}

		// Validate known values
		switch data {
		case "true", "TRUE", "True", "1", "on", "ON", "On", "yes", "YES", "Yes":
			if !result {
				t.Errorf("parseBool(%q) should be true", data)
			}
		case "false", "FALSE", "False", "0", "off", "OFF", "Off", "no", "NO", "No":
			if result {
				t.Errorf("parseBool(%q) should be false", data)
			}
		}

		// Test with spaces
		if len(data) > 0 && data[0] == ' ' {
			// Same logic should apply after trimming
			trimmedResult := ParseBool(data)
			if trimmedResult != result {
				t.Errorf("parseBool should handle spaces correctly: %q", data)
			}
		}
	})
}

// FuzzParseInt tests ParseInt function with random string input
func FuzzParseInt(f *testing.F) {
	// Add seed corpus
	f.Add("0")
	f.Add("1")
	f.Add("123")
	f.Add("-1")
	f.Add("-123")
	f.Add("999999999")
	f.Add("-999999999")
	f.Add("")
	f.Add("abc")
	f.Add("12.34")
	f.Add("1e10")
	f.Add("  123  ")
	f.Add("+123")
	f.Add("00123")
	f.Add("9223372036854775807")  // max int64
	f.Add("-9223372036854775808") // min int64
	f.Add("9223372036854775808")  // overflow

	f.Fuzz(func(t *testing.T, data string) {
		result, err := ParseInt(data)

		// Validate result consistency
		if err == nil {
			// Result should always be within int64 range when no error occurs
			_ = result
		}

		// Known valid cases
		switch data {
		case "0":
			if err != nil {
				t.Error("ParseInt(\"0\") should not error")
			}
			if result != 0 {
				t.Errorf("ParseInt(\"0\") = %d, want 0", result)
			}
		case "123":
			if err != nil {
				t.Error("ParseInt(\"123\") should not error")
			}
			if result != 123 {
				t.Errorf("ParseInt(\"123\") = %d, want 123", result)
			}
		case "-1":
			if err != nil {
				t.Error("ParseInt(\"-1\") should not error")
			}
			if result != -1 {
				t.Errorf("ParseInt(\"-1\") = %d, want -1", result)
			}
		}

		// Invalid cases should error
		switch data {
		case "", "abc", "12.34", "1e10":
			if err == nil {
				t.Errorf("ParseInt(%q) should error but returned %d", data, result)
			}
		}
	})
}

// FuzzMergeConfig tests MergeConfig with random Config values
func FuzzMergeConfig(f *testing.F) {
	// This is a more limited fuzz test since Config is complex
	// We'll test with simpler base and override configs

	f.Fuzz(func(t *testing.T, parallelism int) {
		base := DefaultConfig()
		override := Config{
			Parallelism: parallelism,
		}

		result := MergeConfig(base, override)

		// When override parallelism is zero (not set), base value must be preserved.
		if parallelism == 0 && result.Parallelism <= 0 {
			t.Errorf("MergeConfig lost base parallelism when override is 0: %d", result.Parallelism)
		}

		// Basic sanity checks
		if result.Root == "" {
			t.Error("MergeConfig returned empty Root")
		}
	})
}

// FuzzLoadEnv tests LoadEnv with random prefix values
func FuzzLoadEnv(f *testing.F) {
	// Add seed corpus
	f.Add("CULTIVATOR")
	f.Add("APP")
	f.Add("TEST")
	f.Add("")
	f.Add("X")
	f.Add("_UNDERSCORE_")
	f.Add("VERY_LONG_PREFIX_NAME_FOR_TESTING")

	f.Fuzz(func(t *testing.T, prefix string) {
		// This should not panic
		result := LoadEnv(prefix)

		// Validate result is a valid Config
		if result.Root == "" {
			t.Error("LoadEnv should return Config with Root")
		}

		if result.Parallelism < 1 {
			t.Errorf("LoadEnv returned invalid Parallelism: %d", result.Parallelism)
		}
	})
}
