// Package orchestrator coordinates the execution of Terragrunt operations across modules.
package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cultivator-dev/cultivator/pkg/config"
	"github.com/cultivator-dev/cultivator/pkg/detector"
	custErrors "github.com/cultivator-dev/cultivator/pkg/errors"
	"github.com/cultivator-dev/cultivator/pkg/events"
	"github.com/cultivator-dev/cultivator/pkg/executor"
	"github.com/cultivator-dev/cultivator/pkg/github"
	"github.com/cultivator-dev/cultivator/pkg/graph"
	"github.com/cultivator-dev/cultivator/pkg/lock"
	"github.com/cultivator-dev/cultivator/pkg/module"
	"github.com/cultivator-dev/cultivator/pkg/parser"
)

// Orchestrator coordinates the entire Cultivator workflow
type Orchestrator struct {
	config              *config.Config
	ghClient            *github.Client
	parser              *parser.Parser
	detector            *detector.ChangeDetector
	executor            *executor.Executor
	lockMgr             *lock.Manager
	stdout              io.Writer
	stderr              io.Writer
	lastDetectedModules []*detector.Module // Cache detected modules for external module extraction
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(cfg *config.Config, ghToken string, stdout, stderr io.Writer) (*Orchestrator, error) {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	// Parse repository from environment
	event, err := events.ParseEvent()
	if err != nil {
		return nil, custErrors.NewParseError("GitHub event", err)
	}

	return &Orchestrator{
		config:              cfg,
		ghClient:            github.NewClient(ghToken, event.Owner, event.RepoName),
		parser:              parser.NewParser(),
		executor:            executor.NewExecutor(".", stdout, stderr),
		lockMgr:             lock.NewManager(cfg.Settings.LockTimeout),
		stdout:              stdout,
		stderr:              stderr,
		lastDetectedModules: nil,
	}, nil
}

// Run executes the main Cultivator logic based on the GitHub event
func (o *Orchestrator) Run(ctx context.Context) error {
	event, err := events.ParseEvent()
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	_, _ = fmt.Fprintf(o.stdout, "Event: %s (action: %s)\n", event.Type, event.Action)

	// Handle different event types
	if event.IsCommand() {
		return o.handleCommand(ctx, event)
	}

	if event.ShouldAutoPlan() && o.config.Settings.AutoPlan {
		return o.handleAutoPlan(ctx, event)
	}

	_, _ = fmt.Fprintf(o.stdout, "No action needed for this event\n")
	return nil
}

// logMessage logs a formatted message to stdout
func (o *Orchestrator) logMessage(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, format+"\n", args...)
}

// logError logs an error message to stderr
func (o *Orchestrator) logError(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stderr, format+"\n", args...)
}

// getModulesToProcess returns modules to process based on 'all' flag (DRY principle)
func (o *Orchestrator) getModulesToProcess(_ context.Context, event *events.Event, all bool) ([]string, error) {
	if all {
		return o.parser.GetAllModules(".")
	}

	// Detect changed modules
	o.detector = detector.NewChangeDetector(event.BaseSHA, event.HeadSHA, ".")
	changedModules, err := o.detector.DetectChangedModules()
	if err != nil {
		return nil, err
	}

	// Store detected modules for later access to external modules
	// Convert []detector.Module to []*detector.Module
	o.lastDetectedModules = make([]*detector.Module, len(changedModules))
	for i := range changedModules {
		o.lastDetectedModules[i] = &changedModules[i]
	}

	modules := make([]string, 0, len(changedModules))
	for _, m := range changedModules {
		modules = append(modules, m.Path)
	}

	return modules, nil
}

// buildDependencyGraph builds the dependency graph for modules (DRY principle)
func (o *Orchestrator) buildDependencyGraph(modules []string) (*graph.Graph, error) {
	depGraph := graph.NewGraph()

	for _, modulePath := range modules {
		deps, err := o.parser.FindDependencies(modulePath)
		if err != nil {
			o.logError("Failed to parse dependencies for %s: %v", modulePath, err)
			continue
		}

		depGraph.AddNode(modulePath)
		for _, dep := range deps {
			depGraph.AddDependency(modulePath, dep)
		}
	}

	return depGraph, nil
}

// updateStatus updates commit status with error handling (DRY principle)
func (o *Orchestrator) updateStatus(ctx context.Context, sha, state, description string) {
	if err := o.ghClient.UpdateCommitStatus(ctx, sha, state, description, ""); err != nil {
		o.logError("Failed to update status: %v", err)
	}
}
func (o *Orchestrator) handleCommand(ctx context.Context, event *events.Event) error {
	command, args := event.ParseCommand()

	_, _ = fmt.Fprintf(o.stdout, "Command: %s %v\n", command, args)

	switch command {
	case "plan":
		return o.runPlan(ctx, event, false)
	case "apply":
		return o.runApply(ctx, event, false)
	case "plan-all":
		return o.runPlan(ctx, event, true)
	case "apply-all":
		return o.runApply(ctx, event, true)
	case "unlock":
		return o.runUnlock(ctx, event)
	default:
		return o.postComment(ctx, event.PRNumber, fmt.Sprintf("Unknown command: %s\n\nAvailable commands:\n- '/cultivator plan'\n- '/cultivator apply'\n- '/cultivator plan-all'\n- '/cultivator apply-all'\n- '/cultivator unlock'", command))
	}
}

// handleAutoPlan automatically runs plan when PR is opened/updated
func (o *Orchestrator) handleAutoPlan(ctx context.Context, event *events.Event) error {
	_, _ = fmt.Fprintf(o.stdout, "Auto-running plan...\n")
	return o.runPlan(ctx, event, false)
}

// runPlan executes terragrunt plan
func (o *Orchestrator) runPlan(ctx context.Context, event *events.Event, all bool) error {
	o.updateStatus(ctx, event.HeadSHA, "pending", "Running plan...")

	modules, err := o.getModulesToProcess(ctx, event, all)
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	if len(modules) == 0 {
		msg := "No Terragrunt modules changed in this PR"
		o.updateStatus(ctx, event.HeadSHA, "success", msg)
		return o.postComment(ctx, event.PRNumber, msg)
	}

	o.logMessage("Modules to plan: %d", len(modules))

	// Build dependency graph
	depGraph, err := o.buildDependencyGraph(modules)
	if err != nil {
		return err
	}

	sorted, err := depGraph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort modules: %w", err)
	}

	// Collect and prepare external modules
	externalModules := o.getExternalModulesFromDetected()
	if len(externalModules) > 0 {
		o.logMessage("Preparing external modules: %d module(s) with external dependencies", len(externalModules))
		if err := o.executor.PrepareExternalModules(ctx, externalModules); err != nil {
			o.updateStatus(ctx, event.HeadSHA, "failure", "Failed to prepare external modules")
			return fmt.Errorf("failed to prepare external modules: %w", err)
		}
	}

	// Run plan for each module
	results, planErr := o.executePlans(ctx, sorted, event)

	if planErr != nil {
		o.updateStatus(ctx, event.HeadSHA, "failure", "Plan failed")
		return planErr
	}

	// Post results
	comment := "### Plan Successful\n\n"
	comment += fmt.Sprintf("Planned %d module(s):\n\n", len(sorted))
	for _, r := range results {
		comment += r + "\n\n"
	}

	o.updateStatus(ctx, event.HeadSHA, "success", "Plan successful")
	return o.postComment(ctx, event.PRNumber, comment)
}

// executePlans runs plan for each module and returns formatted results (extracted method)
func (o *Orchestrator) executePlans(ctx context.Context, modules []string, _ *events.Event) ([]string, error) {
	var results []string

	for _, modulePath := range modules {
		o.logMessage("Planning %s...", modulePath)

		result, err := o.executor.Plan(ctx, modulePath)
		if err != nil || result.ExitCode != 0 {
			errMsg := fmt.Sprintf("Plan failed for %s", modulePath)
			output := github.FormatPlanOutput(modulePath, result.Stdout+"\n"+result.Stderr, false)
			return nil, custErrors.NewExternalError("terraform plan", fmt.Errorf("%s", errMsg)).
				WithContext("module", modulePath).
				WithContext("output", output)
		}

		results = append(results, github.FormatPlanOutput(modulePath, result.Stdout, true))
	}

	return results, nil
}

// runApply executes terragrunt apply
func (o *Orchestrator) runApply(ctx context.Context, event *events.Event, all bool) error {
	o.logMessage("Running apply...")

	// Check apply requirements
	_, err := o.ghClient.GetPR(ctx, event.PRNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	// Check if PR is approved (simplified check)
	if o.config.Settings.RequireApproval {
		// TODO: Implement proper approval check
		o.logMessage("Skipping approval check (not implemented yet)")
	}

	modules, err := o.getModulesToProcess(ctx, event, all)
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	o.logMessage("Modules to apply: %d", len(modules))

	// Build graph and get execution order
	depGraph, err := o.buildDependencyGraph(modules)
	if err != nil {
		return err
	}

	sorted, err := depGraph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort modules: %w", err)
	}

	// Collect and prepare external modules
	externalModules := o.getExternalModulesFromDetected()
	if len(externalModules) > 0 {
		o.logMessage("Preparing external modules: %d module(s) with external dependencies", len(externalModules))
		if err := o.executor.PrepareExternalModules(ctx, externalModules); err != nil {
			o.updateStatus(ctx, event.HeadSHA, "failure", "Failed to prepare external modules")
			return fmt.Errorf("failed to prepare external modules: %w", err)
		}
	}

	// Apply each module
	results, applyErr := o.executeApplies(ctx, sorted, event)

	if applyErr != nil {
		o.updateStatus(ctx, event.HeadSHA, "failure", "Apply failed")
		return applyErr
	}

	// Post results
	comment := "### Apply Successful\n\n"
	comment += fmt.Sprintf("Applied %d module(s):\n\n", len(sorted))
	for _, r := range results {
		comment += r + "\n\n"
	}

	o.updateStatus(ctx, event.HeadSHA, "success", "Apply successful")
	return o.postComment(ctx, event.PRNumber, comment)
}

// executeApplies runs apply for each module with locking (extracted method)
func (o *Orchestrator) executeApplies(ctx context.Context, modules []string, event *events.Event) ([]string, error) {
	var results []string

	for _, modulePath := range modules {
		o.logMessage("Applying %s...", modulePath)

		// Try to acquire lock
		if err := o.lockMgr.Acquire(ctx, modulePath, event.CommentAuthor, event.PRNumber); err != nil {
			return nil, custErrors.NewExternalError("acquire module lock", err).
				WithContext("module", modulePath).
				WithContext("pr_number", event.PRNumber).
				WithContext("author", event.CommentAuthor)
		}
		defer func() {
			_ = o.lockMgr.Release(modulePath)
		}()

		result, err := o.executor.Apply(ctx, modulePath)
		if err != nil || result.ExitCode != 0 {
			errMsg := fmt.Sprintf("Apply failed for %s", modulePath)
			output := github.FormatApplyOutput(modulePath, result.Stdout+"\n"+result.Stderr, false)
			return nil, custErrors.NewExternalError("terraform apply", fmt.Errorf("%s", errMsg)).
				WithContext("module", modulePath).
				WithContext("output", output)
		}

		results = append(results, github.FormatApplyOutput(modulePath, result.Stdout, true))
	}

	return results, nil
}

// runUnlock removes locks
func (o *Orchestrator) runUnlock(ctx context.Context, event *events.Event) error {
	// TODO: Implement unlock logic
	return o.postComment(ctx, event.PRNumber, "Unlock functionality not yet implemented")
}

// getExternalModulesFromDetected extracts and deduplicates external modules from detected modules
// Implements Single Responsibility: extracts external module handling logic into dedicated method
func (o *Orchestrator) getExternalModulesFromDetected() []module.SourceInfo {
	if o.lastDetectedModules == nil {
		return nil
	}

	// Use a map to deduplicate external modules by their raw source
	// This prevents fetching the same module multiple times
	externalModuleMap := make(map[string]module.SourceInfo)

	for _, m := range o.lastDetectedModules {
		for _, extModule := range m.ExternalModules {
			// Skip if already processed
			if _, exists := externalModuleMap[extModule.RawSource]; !exists {
				externalModuleMap[extModule.RawSource] = extModule
			}
		}
	}

	// Convert map to slice
	result := make([]module.SourceInfo, 0, len(externalModuleMap))
	for _, extModule := range externalModuleMap {
		result = append(result, extModule)
	}

	return result
}

// postComment posts a comment to the PR
func (o *Orchestrator) postComment(ctx context.Context, prNumber int, comment string) error {
	return o.ghClient.CommentOnPR(ctx, prNumber, comment)
}
