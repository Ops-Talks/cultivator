package events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EventType represents the type of GitHub event
type EventType string

const (
	EventPullRequest   EventType = "pull_request"
	EventIssueComment  EventType = "issue_comment"
	EventPush          EventType = "push"
)

// Event represents a GitHub event
type Event struct {
	Type          EventType
	Action        string
	PRNumber      int
	Repository    string
	Owner         string
	RepoName      string
	BaseSHA       string
	HeadSHA       string
	Comment       string
	CommentAuthor string
}

// ParseEvent parses a GitHub event from environment variables
func ParseEvent() (*Event, error) {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	repository := os.Getenv("GITHUB_REPOSITORY")

	if eventName == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_NAME not set")
	}

	event := &Event{
		Type:       EventType(eventName),
		Repository: repository,
	}

	// Parse repository owner and name
	parts := strings.Split(repository, "/")
	if len(parts) == 2 {
		event.Owner = parts[0]
		event.RepoName = parts[1]
	}

	// Parse event payload
	if eventPath != "" {
		data, err := os.ReadFile(eventPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read event file: %w", err)
		}

		if err := event.parsePayload(data); err != nil {
			return nil, err
		}
	}

	return event, nil
}

// parsePayload parses the event payload JSON
func (e *Event) parsePayload(data []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse event payload: %w", err)
	}

	switch e.Type {
	case EventPullRequest:
		return e.parsePullRequestEvent(payload)
	case EventIssueComment:
		return e.parseIssueCommentEvent(payload)
	case EventPush:
		return e.parsePushEvent(payload)
	}

	return nil
}

// parsePullRequestEvent parses a pull_request event
func (e *Event) parsePullRequestEvent(payload map[string]interface{}) error {
	if action, ok := payload["action"].(string); ok {
		e.Action = action
	}

	if pr, ok := payload["pull_request"].(map[string]interface{}); ok {
		if number, ok := pr["number"].(float64); ok {
			e.PRNumber = int(number)
		}

		if base, ok := pr["base"].(map[string]interface{}); ok {
			if sha, ok := base["sha"].(string); ok {
				e.BaseSHA = sha
			}
		}

		if head, ok := pr["head"].(map[string]interface{}); ok {
			if sha, ok := head["sha"].(string); ok {
				e.HeadSHA = sha
			}
		}
	}

	return nil
}

// parseIssueCommentEvent parses an issue_comment event
func (e *Event) parseIssueCommentEvent(payload map[string]interface{}) error {
	if action, ok := payload["action"].(string); ok {
		e.Action = action
	}

	if comment, ok := payload["comment"].(map[string]interface{}); ok {
		if body, ok := comment["body"].(string); ok {
			e.Comment = body
		}

		if user, ok := comment["user"].(map[string]interface{}); ok {
			if login, ok := user["login"].(string); ok {
				e.CommentAuthor = login
			}
		}
	}

	if issue, ok := payload["issue"].(map[string]interface{}); ok {
		if number, ok := issue["number"].(float64); ok {
			e.PRNumber = int(number)
		}

		// Get PR info from issue
		if pr, ok := issue["pull_request"].(map[string]interface{}); ok {
			// This is a PR comment
			_ = pr // PR details available if needed
		}
	}

	return nil
}

// parsePushEvent parses a push event
func (e *Event) parsePushEvent(payload map[string]interface{}) error {
	if before, ok := payload["before"].(string); ok {
		e.BaseSHA = before
	}

	if after, ok := payload["after"].(string); ok {
		e.HeadSHA = after
	}

	return nil
}

// IsCommand checks if the event is a command comment
func (e *Event) IsCommand() bool {
	return e.Type == EventIssueComment && strings.HasPrefix(e.Comment, "/cultivator")
}

// ParseCommand extracts the command from a comment
func (e *Event) ParseCommand() (string, []string) {
	if !e.IsCommand() {
		return "", nil
	}

	// Remove /cultivator prefix
	cmd := strings.TrimPrefix(e.Comment, "/cultivator")
	cmd = strings.TrimSpace(cmd)

	// Split into command and args
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", nil
	}

	return parts[0], parts[1:]
}

// ShouldAutoPlan checks if we should automatically run plan
func (e *Event) ShouldAutoPlan() bool {
	if e.Type != EventPullRequest {
		return false
	}

	// Auto plan on PR opened, synchronize (new commits), reopened
	return e.Action == "opened" || 
	       e.Action == "synchronize" || 
	       e.Action == "reopened"
}

// GetPRNumberFromEnv gets PR number from environment if not in event
func GetPRNumberFromEnv() (int, error) {
	// Try GITHUB_REF first (refs/pull/123/merge)
	ref := os.Getenv("GITHUB_REF")
	if strings.HasPrefix(ref, "refs/pull/") {
		parts := strings.Split(ref, "/")
		if len(parts) >= 3 {
			return strconv.Atoi(parts[2])
		}
	}

	return 0, fmt.Errorf("unable to determine PR number")
}
