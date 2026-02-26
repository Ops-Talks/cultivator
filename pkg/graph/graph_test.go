package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph_AddNode(t *testing.T) {
	g := NewGraph()
	g.AddNode("module1")

	node, exists := g.GetNode("module1")
	assert.True(t, exists)
	assert.Equal(t, "module1", node.Path)
}

func TestGraph_AddDependency(t *testing.T) {
	g := NewGraph()
	g.AddDependency("app", "vpc")

	app, _ := g.GetNode("app")
	vpc, _ := g.GetNode("vpc")

	assert.Contains(t, app.Dependencies, "vpc")
	assert.Contains(t, vpc.Dependents, "app")
}

func TestGraph_TopologicalSort(t *testing.T) {
	g := NewGraph()

	// vpc <- database <- app
	g.AddDependency("database", "vpc")
	g.AddDependency("app", "database")

	sorted, err := g.TopologicalSort()
	assert.NoError(t, err)

	// vpc should come before database, database before app
	assert.Equal(t, []string{"vpc", "database", "app"}, sorted)
}

func TestGraph_CircularDependency(t *testing.T) {
	g := NewGraph()

	// Create circular dependency
	g.AddDependency("module1", "module2")
	g.AddDependency("module2", "module3")
	g.AddDependency("module3", "module1")

	_, err := g.TopologicalSort()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestGraph_GetAffectedModules(t *testing.T) {
	g := NewGraph()

	// vpc <- database <- app
	g.AddDependency("database", "vpc")
	g.AddDependency("app", "database")

	// Changing vpc should affect database and app
	affected := g.GetAffectedModules([]string{"vpc"})
	assert.ElementsMatch(t, []string{"vpc", "database", "app"}, affected)

	// Changing database should affect only database and app
	affected = g.GetAffectedModules([]string{"database"})
	assert.ElementsMatch(t, []string{"database", "app"}, affected)

	// Changing app should affect only app
	affected = g.GetAffectedModules([]string{"app"})
	assert.ElementsMatch(t, []string{"app"}, affected)
}

func TestGraph_GetExecutionGroups(t *testing.T) {
	g := NewGraph()

	// Create a graph:
	// vpc
	// ├── database
	// └── cache
	// app (depends on database and cache)
	g.AddDependency("database", "vpc")
	g.AddDependency("cache", "vpc")
	g.AddDependency("app", "database")
	g.AddDependency("app", "cache")

	groups, err := g.GetExecutionGroups([]string{"vpc", "database", "cache", "app"})
	assert.NoError(t, err)
	assert.Len(t, groups, 3)

	// Group 0: vpc (no dependencies)
	assert.ElementsMatch(t, []string{"vpc"}, groups[0])

	// Group 1: database and cache (both depend only on vpc, can run in parallel)
	assert.ElementsMatch(t, []string{"database", "cache"}, groups[1])

	// Group 2: app (depends on database and cache)
	assert.ElementsMatch(t, []string{"app"}, groups[2])
}

func TestGraph_Subgraph(t *testing.T) {
	g := NewGraph()

	g.AddDependency("app", "database")
	g.AddDependency("database", "vpc")
	g.AddDependency("cache", "vpc")

	// Create subgraph with only app and database
	subgraph := g.Subgraph([]string{"app", "database"})

	// Should have 2 nodes
	assert.Len(t, subgraph.GetAllPaths(), 2)

	// Should maintain dependency between app and database
	app, _ := subgraph.GetNode("app")
	assert.Contains(t, app.Dependencies, "database")

	// Should not have vpc or cache
	_, exists := subgraph.GetNode("vpc")
	assert.False(t, exists)
}
