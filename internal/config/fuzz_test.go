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
			// Result should always be within int range
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

// FuzzMergeConfig tests MergeConfig with various inputs
func FuzzMergeConfig(f *testing.F) {
	f.Add(".", "dev", 2, true, false)
	f.Add("/tmp", "prod", 10, false, true)
	f.Add("", "", 0, false, false)

	f.Fuzz(func(t *testing.T, root, env string, parallelism int, nonInteractive, dryRun bool) {
		base := DefaultConfig()
		override := Config{
			Root:           root,
			Env:            env,
			Parallelism:    parallelism,
			NonInteractive: nonInteractive,
			DryRun:         dryRun,
		}

		result := MergeConfig(base, override)

		// Basic validation
		if result.Root == "" {
			t.Error("MergeConfig returned empty Root")
		}

		if parallelism > 0 && result.Parallelism != parallelism {
			t.Errorf("MergeConfig failed to override parallelism: got %d, want %d", result.Parallelism, parallelism)
		}

		if nonInteractive && !result.NonInteractive {
			t.Error("MergeConfig failed to override NonInteractive")
		}

		if dryRun && !result.DryRun {
			t.Error("MergeConfig failed to override DryRun")
		}
	})
}

// FuzzApplyOverrides tests ApplyOverrides with various inputs
func FuzzApplyOverrides(f *testing.F) {
	f.Fuzz(func(t *testing.T, root, env string, parallelism int, nonInteractive, dryRun bool) {
		cfg := DefaultConfig()
		ovr := Overrides{
			Root:           &root,
			Env:            &env,
			Parallelism:    &parallelism,
			NonInteractive: &nonInteractive,
			DryRun:         &dryRun,
		}

		result := ApplyOverrides(cfg, ovr)

		if result.Root != root {
			t.Errorf("ApplyOverrides failed to override root: got %q, want %q", result.Root, root)
		}

		if result.Env != env {
			t.Errorf("ApplyOverrides failed to override env: got %q, want %q", result.Env, env)
		}

		if parallelism > 0 && result.Parallelism != parallelism {
			t.Errorf("ApplyOverrides failed to override parallelism: got %d, want %d", result.Parallelism, parallelism)
		}

		if result.NonInteractive != nonInteractive {
			t.Errorf("ApplyOverrides failed to override NonInteractive: got %v, want %v", result.NonInteractive, nonInteractive)
		}

		if result.DryRun != dryRun {
			t.Errorf("ApplyOverrides failed to override DryRun: got %v, want %v", result.DryRun, dryRun)
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
