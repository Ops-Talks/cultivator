package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/cultivator-dev/cultivator/pkg/formatter"
	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

// Client is a wrapper around the GitHub API client
type Client struct {
	client *github.Client
	owner  string
	repo   string
}

// NewClient creates a new GitHub client
func NewClient(token, owner, repo string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		owner:  owner,
		repo:   repo,
	}
}

// CommentOnPR posts a comment on a pull request
func (c *Client) CommentOnPR(ctx context.Context, prNumber int, comment string) error {
	_, _, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, prNumber, &github.IssueComment{
		Body: github.String(comment),
	})
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

// GetPRFiles gets the list of files changed in a PR
func (c *Client) GetPRFiles(ctx context.Context, prNumber int) ([]string, error) {
	opts := &github.ListOptions{PerPage: 100}
	var allFiles []string

	for {
		files, resp, err := c.client.PullRequests.ListFiles(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list PR files: %w", err)
		}

		for _, file := range files {
			if file.Filename != nil {
				allFiles = append(allFiles, *file.Filename)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFiles, nil
}

// GetPR gets pull request information
func (c *Client) GetPR(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}
	return pr, nil
}

// UpdateCommitStatus updates the commit status
func (c *Client) UpdateCommitStatus(ctx context.Context, sha, state, description, targetURL string) error {
	status := &github.RepoStatus{
		State:       github.String(state),
		Description: github.String(description),
		Context:     github.String("cultivator"),
		TargetURL:   github.String(targetURL),
	}

	_, _, err := c.client.Repositories.CreateStatus(ctx, c.owner, c.repo, sha, status)
	if err != nil {
		return fmt.Errorf("failed to update commit status: %w", err)
	}
	return nil
}

// OutputFormat defines parameters for formatting output sections
type OutputFormat struct {
	Title      string
	Emoji      string
	CodeLang   string
	HasChanges bool
}

// formatOutputSection encapsulates common output formatting logic (DRY principle)
func formatOutputSection(modulePath, output string, format OutputFormat) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %s %s\n\n", format.Emoji, format.Title))
	sb.WriteString(fmt.Sprintf("**Module:** `%s`\n\n", modulePath))

	if format.HasChanges {
		if format.Title == "Plan Results" {
			summary := formatter.ParsePlanOutput(output)
			if summary.HasChanges {
				sb.WriteString(fmt.Sprintf("**%s**\n\n", summary.String()))
			} else {
				sb.WriteString("**No changes**\n\n")
			}
		} else {
			sb.WriteString("**Applied successfully**\n\n")
		}
	} else {
		sb.WriteString("**Operation failed**\n\n")
	}

	cleanOutput := formatter.CleanTerraformOutput(output)
	truncatedOutput := formatter.TruncateOutput(cleanOutput, 100)

	sb.WriteString("<details>\n")
	sb.WriteString(fmt.Sprintf("<summary>Show %s</summary>\n\n", format.Title))
	sb.WriteString(fmt.Sprintf("```%s\n", format.CodeLang))
	sb.WriteString(truncatedOutput)
	sb.WriteString("\n```\n\n")
	sb.WriteString("</details>\n")

	return sb.String()
}

// FormatPlanOutput formats the plan output for PR comment
func FormatPlanOutput(modulePath string, planOutput string, hasChanges bool) string {
	return formatOutputSection(modulePath, planOutput, OutputFormat{
		Title:      "Plan Results",
		Emoji:      "",
		CodeLang:   "terraform",
		HasChanges: hasChanges,
	})
}

// FormatApplyOutput formats the apply output for PR comment
func FormatApplyOutput(modulePath string, applyOutput string, success bool) string {
	return formatOutputSection(modulePath, applyOutput, OutputFormat{
		Title:      "Apply Results",
		Emoji:      "",
		CodeLang:   "",
		HasChanges: success,
	})
}
