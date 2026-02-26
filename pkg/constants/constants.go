// Package constants provides centralized definitions for configuration values and constants.
package constants

// HCL Block Types
const (
	BlockTypeDependency = "dependency"
	BlockTypeTerraform  = "terraform"
	BlockTypeInclude    = "include"
	BlockTypeInputs     = "inputs"
)

// HCL Attribute Names
const (
	AttrConfigPath = "config_path"
	AttrSource     = "source"
	AttrPath       = "path"
	AttrRef        = "ref"
)

// Source Types
const (
	SourceTypeGit  = "git"
	SourceTypeHTTP = "http"
)

// Source Prefixes
const (
	GitPrefix   = "git::"
	HTTPPrefix  = "http://"
	HTTPSPrefix = "https://"
	FilePrefix  = "file://"
)

// Terragrunt File Names
const (
	TerragruntConfigFile = "terragrunt.hcl"
)

// Default Values
const (
	DefaultLockTimeout = "10m"
	DefaultMaxParallel = 5
	DefaultConfigPath  = "cultivator.yml"
	DefaultWorkingDir  = "."
	DefaultVersion     = 1
)

// Query Parameters
const (
	QueryParamRef = "ref"
)

// RelevantExtensions lists file extensions relevant to Terragrunt configurations.
var RelevantExtensions = [...]string{".hcl", ".tf", ".tfvars"}

// Subpath Separator
const (
	SubpathSeparator = "//"
)

// Default Git Reference
const (
	DefaultGitRef = "HEAD"
)

// GitHub Env Variables
const (
	EnvGitHubEventName  = "GITHUB_EVENT_NAME"
	EnvGitHubEventPath  = "GITHUB_EVENT_PATH"
	EnvGitHubRepository = "GITHUB_REPOSITORY"
	EnvGitHubToken      = "GITHUB_TOKEN"
)

// Executor Flags
const (
	FlagNoColor            = "-no-color"
	FlagAutoApprove        = "-auto-approve"
	FlagTerragruntNonInter = "--terragrunt-non-interactive"
)

// Terraform Actions
const (
	ActionPlan     = "plan"
	ActionApply    = "apply"
	ActionValidate = "validate"
	ActionInit     = "init"
	ActionRunAll   = "run-all"
)

// Formatting Constants
const (
	MaxModulesDisplay   = 5
	ChangeSymbolAdd     = "+"
	ChangeSymbolChange  = "~"
	ChangeSymbolDestroy = "-"
	RedactedPlaceholder = "[REDACTED]"
)

// Error Messages Templates
const (
	ErrFailedToRead  = "failed to read %s: %w"
	ErrFailedToParse = "failed to parse %s: %w"
	ErrInvalidFormat = "invalid %s format: %s"
	ErrNotFound      = "%s not found: %s"
	ErrAlreadyExists = "%s already exists: %s"
)

// Log Message Templates
const (
	MsgProcessingModule = "Processing module: %s"
	MsgDetectedChanges  = "Detected %d changed modules"
)
