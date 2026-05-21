package ci

import (
	"os"
	"testing"
)

// setEnv sets environment variables for the duration of a test, restoring the
// original values when the test completes.
func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for key, val := range pairs {
		orig, existed := os.LookupEnv(key)
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("os.Setenv(%q): %v", key, err)
		}
		if existed {
			t.Cleanup(func() { os.Setenv(key, orig) }) //nolint:errcheck
		} else {
			t.Cleanup(func() { os.Unsetenv(key) }) //nolint:errcheck
		}
	}
}

func TestDetect_Unknown(t *testing.T) {
	// Ensure none of the CI sentinel variables are set.
	for _, key := range []string{"BITBUCKET_BUILD_NUMBER", "GITHUB_ACTIONS", "GITLAB_CI"} {
		os.Unsetenv(key) //nolint:errcheck
	}

	env := Detect()

	if env.Provider != ProviderUnknown {
		t.Errorf("Provider = %q, want %q", env.Provider, ProviderUnknown)
	}
	if env.IsPR {
		t.Error("IsPR should be false in unknown environment")
	}
	if env.BaseRef != "" {
		t.Errorf("BaseRef = %q, want empty", env.BaseRef)
	}
}

func TestDetect_Bitbucket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		envVars   map[string]string
		wantIsPR  bool
		wantBase  string
		wantHead  string
		wantSHA   string
	}{
		{
			name: "PR pipeline",
			envVars: map[string]string{
				"BITBUCKET_BUILD_NUMBER":          "42",
				"BITBUCKET_PR_ID":                 "7",
				"BITBUCKET_PR_DESTINATION_BRANCH": "main",
				"BITBUCKET_BRANCH":                "feat/my-feature",
				"BITBUCKET_COMMIT":                "abc123",
			},
			wantIsPR: true,
			wantBase: "main",
			wantHead: "feat/my-feature",
			wantSHA:  "abc123",
		},
		{
			name: "branch pipeline (no PR)",
			envVars: map[string]string{
				"BITBUCKET_BUILD_NUMBER": "10",
				"BITBUCKET_BRANCH":       "main",
				"BITBUCKET_COMMIT":       "def456",
			},
			wantIsPR: false,
			wantBase: "",
			wantHead: "main",
			wantSHA:  "def456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Clear competing providers first.
			for _, key := range []string{"GITHUB_ACTIONS", "GITLAB_CI"} {
				os.Unsetenv(key) //nolint:errcheck
			}
			setEnv(t, tc.envVars)

			env := Detect()

			if env.Provider != ProviderBitbucket {
				t.Errorf("Provider = %q, want %q", env.Provider, ProviderBitbucket)
			}
			if env.IsPR != tc.wantIsPR {
				t.Errorf("IsPR = %v, want %v", env.IsPR, tc.wantIsPR)
			}
			if env.BaseRef != tc.wantBase {
				t.Errorf("BaseRef = %q, want %q", env.BaseRef, tc.wantBase)
			}
			if env.HeadRef != tc.wantHead {
				t.Errorf("HeadRef = %q, want %q", env.HeadRef, tc.wantHead)
			}
			if env.CommitSHA != tc.wantSHA {
				t.Errorf("CommitSHA = %q, want %q", env.CommitSHA, tc.wantSHA)
			}
		})
	}
}

func TestDetect_GitHub(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envVars  map[string]string
		wantIsPR bool
		wantBase string
		wantHead string
		wantSHA  string
	}{
		{
			name: "PR event",
			envVars: map[string]string{
				"GITHUB_ACTIONS":  "true",
				"GITHUB_BASE_REF": "main",
				"GITHUB_HEAD_REF": "feat/new-feature",
				"GITHUB_SHA":      "sha111",
			},
			wantIsPR: true,
			wantBase: "main",
			wantHead: "feat/new-feature",
			wantSHA:  "sha111",
		},
		{
			name: "push event (no PR)",
			envVars: map[string]string{
				"GITHUB_ACTIONS":  "true",
				"GITHUB_HEAD_REF": "",
				"GITHUB_SHA":      "sha222",
			},
			wantIsPR: false,
			wantBase: "",
			wantHead: "",
			wantSHA:  "sha222",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, key := range []string{"BITBUCKET_BUILD_NUMBER", "GITLAB_CI"} {
				os.Unsetenv(key) //nolint:errcheck
			}
			setEnv(t, tc.envVars)

			env := Detect()

			if env.Provider != ProviderGitHub {
				t.Errorf("Provider = %q, want %q", env.Provider, ProviderGitHub)
			}
			if env.IsPR != tc.wantIsPR {
				t.Errorf("IsPR = %v, want %v", env.IsPR, tc.wantIsPR)
			}
			if env.BaseRef != tc.wantBase {
				t.Errorf("BaseRef = %q, want %q", env.BaseRef, tc.wantBase)
			}
			if env.HeadRef != tc.wantHead {
				t.Errorf("HeadRef = %q, want %q", env.HeadRef, tc.wantHead)
			}
			if env.CommitSHA != tc.wantSHA {
				t.Errorf("CommitSHA = %q, want %q", env.CommitSHA, tc.wantSHA)
			}
		})
	}
}

func TestDetect_GitLab(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envVars  map[string]string
		wantIsPR bool
		wantBase string
		wantHead string
		wantSHA  string
	}{
		{
			name: "merge request pipeline",
			envVars: map[string]string{
				"GITLAB_CI":                              "true",
				"CI_MERGE_REQUEST_IID":                   "12",
				"CI_MERGE_REQUEST_TARGET_BRANCH_NAME":    "main",
				"CI_COMMIT_REF_NAME":                     "feat/something",
				"CI_COMMIT_SHA":                          "sha333",
			},
			wantIsPR: true,
			wantBase: "main",
			wantHead: "feat/something",
			wantSHA:  "sha333",
		},
		{
			name: "branch pipeline (no MR)",
			envVars: map[string]string{
				"GITLAB_CI":          "true",
				"CI_COMMIT_REF_NAME": "main",
				"CI_COMMIT_SHA":      "sha444",
			},
			wantIsPR: false,
			wantBase: "",
			wantHead: "main",
			wantSHA:  "sha444",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, key := range []string{"BITBUCKET_BUILD_NUMBER", "GITHUB_ACTIONS"} {
				os.Unsetenv(key) //nolint:errcheck
			}
			setEnv(t, tc.envVars)

			env := Detect()

			if env.Provider != ProviderGitLab {
				t.Errorf("Provider = %q, want %q", env.Provider, ProviderGitLab)
			}
			if env.IsPR != tc.wantIsPR {
				t.Errorf("IsPR = %v, want %v", env.IsPR, tc.wantIsPR)
			}
			if env.BaseRef != tc.wantBase {
				t.Errorf("BaseRef = %q, want %q", env.BaseRef, tc.wantBase)
			}
			if env.HeadRef != tc.wantHead {
				t.Errorf("HeadRef = %q, want %q", env.HeadRef, tc.wantHead)
			}
			if env.CommitSHA != tc.wantSHA {
				t.Errorf("CommitSHA = %q, want %q", env.CommitSHA, tc.wantSHA)
			}
		})
	}
}

func TestDetect_PriorityBitbucketOverGitHub(t *testing.T) {
	// When multiple CI env vars are set simultaneously, Bitbucket takes priority.
	for _, key := range []string{"GITHUB_ACTIONS", "GITLAB_CI", "BITBUCKET_BUILD_NUMBER"} {
		os.Unsetenv(key) //nolint:errcheck
	}

	setEnv(t, map[string]string{
		"BITBUCKET_BUILD_NUMBER": "1",
		"GITHUB_ACTIONS":         "true",
	})

	env := Detect()
	if env.Provider != ProviderBitbucket {
		t.Errorf("Provider = %q, want %q (Bitbucket must take priority)", env.Provider, ProviderBitbucket)
	}
}
