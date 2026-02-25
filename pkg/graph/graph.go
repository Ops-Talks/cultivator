package graph

import (
	"fmt"
	"sort"
)

// Node represents a module in the dependency graph
type Node struct {
	Path         string
	Dependencies []string
	Dependents   []string
}

// Graph represents a dependency graph of Terragrunt modules
type Graph struct {
	nodes map[string]*Node
}

// NewGraph creates a new dependency graph
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(path string) {
	if _, exists := g.nodes[path]; !exists {
		g.nodes[path] = &Node{
			Path:         path,
			Dependencies: []string{},
			Dependents:   []string{},
		}
	}
}

// AddDependency adds a dependency relationship
func (g *Graph) AddDependency(from, to string) {
	g.AddNode(from)
	g.AddNode(to)

	// Add to dependencies
	if !contains(g.nodes[from].Dependencies, to) {
		g.nodes[from].Dependencies = append(g.nodes[from].Dependencies, to)
	}

	// Add to dependents (reverse)
	if !contains(g.nodes[to].Dependents, from) {
		g.nodes[to].Dependents = append(g.nodes[to].Dependents, from)
	}
}

// GetNode returns a node by path
func (g *Graph) GetNode(path string) (*Node, bool) {
	node, exists := g.nodes[path]
	return node, exists
}

// TopologicalSort returns modules in topological order (dependencies first)
func (g *Graph) TopologicalSort() ([]string, error) {
	visited := make(map[string]bool)
	tempMarked := make(map[string]bool)
	result := []string{}

	var visit func(string) error
	visit = func(path string) error {
		if tempMarked[path] {
			return fmt.Errorf("circular dependency detected involving %s", path)
		}
		if visited[path] {
			return nil
		}

		tempMarked[path] = true
		node := g.nodes[path]

		// Visit dependencies first
		for _, dep := range node.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		tempMarked[path] = false
		visited[path] = true
		result = append(result, path)

		return nil
	}

	// Visit all nodes
	paths := g.GetAllPaths()
	for _, path := range paths {
		if !visited[path] {
			if err := visit(path); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// GetAllPaths returns all module paths in the graph
func (g *Graph) GetAllPaths() []string {
	paths := make([]string, 0, len(g.nodes))
	for path := range g.nodes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// GetAffectedModules returns all modules affected by changes to the given modules
func (g *Graph) GetAffectedModules(changedModules []string) []string {
	affected := make(map[string]bool)

	var addAffected func(string)
	addAffected = func(path string) {
		if affected[path] {
			return
		}
		affected[path] = true

		node, exists := g.nodes[path]
		if !exists {
			return
		}

		// Add all dependents recursively
		for _, dependent := range node.Dependents {
			addAffected(dependent)
		}
	}

	// Add all changed modules and their dependents
	for _, module := range changedModules {
		addAffected(module)
	}

	// Convert to slice
	result := make([]string, 0, len(affected))
	for path := range affected {
		result = append(result, path)
	}

	sort.Strings(result)
	return result
}

// GetExecutionGroups returns groups of modules that can be executed in parallel
func (g *Graph) GetExecutionGroups(modules []string) ([][]string, error) {
	// Get subgraph with only specified modules
	subgraph := g.Subgraph(modules)

	// Topologically sort
	sorted, err := subgraph.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Group modules by level (modules at same level can run in parallel)
	levels := make(map[int][]string)
	moduleLevel := make(map[string]int)

	for _, path := range sorted {
		maxDepLevel := -1
		node := subgraph.nodes[path]

		for _, dep := range node.Dependencies {
			if level, exists := moduleLevel[dep]; exists {
				if level > maxDepLevel {
					maxDepLevel = level
				}
			}
		}

		level := maxDepLevel + 1
		moduleLevel[path] = level
		levels[level] = append(levels[level], path)
	}

	// Convert to ordered groups
	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	groups := make([][]string, maxLevel+1)
	for level := 0; level <= maxLevel; level++ {
		groups[level] = levels[level]
		sort.Strings(groups[level])
	}

	return groups, nil
}

// Subgraph creates a subgraph containing only specified modules
func (g *Graph) Subgraph(modules []string) *Graph {
	subgraph := NewGraph()
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	for _, module := range modules {
		if node, exists := g.nodes[module]; exists {
			subgraph.AddNode(module)

			// Add only dependencies that are in the module set
			for _, dep := range node.Dependencies {
				if moduleSet[dep] {
					subgraph.AddDependency(module, dep)
				}
			}
		}
	}

	return subgraph
}

// Validate checks if the graph is valid (no circular dependencies)
func (g *Graph) Validate() error {
	_, err := g.TopologicalSort()
	return err
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
