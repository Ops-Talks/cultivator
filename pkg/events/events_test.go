package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent_ParseCommand(t *testing.T) {
	tests := []struct {
		name            string
		comment         string
		expectedCommand string
		expectedArgs    []string
	}{
		{
			name:            "plan command",
			comment:         "/cultivator plan",
			expectedCommand: "plan",
			expectedArgs:    []string{},
		},
		{
			name:            "apply command",
			comment:         "/cultivator apply",
			expectedCommand: "apply",
			expectedArgs:    []string{},
		},
		{
			name:            "plan with module",
			comment:         "/cultivator plan -m vpc",
			expectedCommand: "plan",
			expectedArgs:    []string{"-m", "vpc"},
		},
		{
			name:            "plan-all command",
			comment:         "/cultivator plan-all",
			expectedCommand: "plan-all",
			expectedArgs:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{
				Type:    EventIssueComment,
				Comment: tt.comment,
			}

			command, args := event.ParseCommand()
			assert.Equal(t, tt.expectedCommand, command)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestEvent_IsCommand(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected bool
	}{
		{
			name: "cultivator command",
			event: Event{
				Type:    EventIssueComment,
				Comment: "/cultivator plan",
			},
			expected: true,
		},
		{
			name: "not a command",
			event: Event{
				Type:    EventIssueComment,
				Comment: "This is just a comment",
			},
			expected: false,
		},
		{
			name: "PR event",
			event: Event{
				Type: EventPullRequest,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.IsCommand()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvent_ShouldAutoPlan(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected bool
	}{
		{
			name: "PR opened",
			event: Event{
				Type:   EventPullRequest,
				Action: "opened",
			},
			expected: true,
		},
		{
			name: "PR synchronize",
			event: Event{
				Type:   EventPullRequest,
				Action: "synchronize",
			},
			expected: true,
		},
		{
			name: "PR closed",
			event: Event{
				Type:   EventPullRequest,
				Action: "closed",
			},
			expected: false,
		},
		{
			name: "Issue comment",
			event: Event{
				Type: EventIssueComment,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.ShouldAutoPlan()
			assert.Equal(t, tt.expected, result)
		})
	}
}
