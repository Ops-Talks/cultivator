// Package dag provides a Directed Acyclic Graph implementation for
// managing module execution order based on dependencies.
package dag

import (
	"fmt"
	"sort"
)

// Graph represents a directed graph.
type Graph struct {
	nodes        map[string]bool
	dependencies map[string]map[string]bool // map[node]set[dependencies]
}

// New returns a new empty Graph.
func New() *Graph {
	return &Graph{
		nodes:        make(map[string]bool),
		dependencies: make(map[string]map[string]bool),
	}
}

// AddNode adds a node to the graph if it doesn't already exist.
func (g *Graph) AddNode(id string) {
	g.nodes[id] = true
}

// AddEdge adds a directed edge between two nodes.
// 'from' depends on 'to'.
func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	if g.dependencies[from] == nil {
		g.dependencies[from] = make(map[string]bool)
	}
	g.dependencies[from][to] = true
}

// TopologicalSort returns the nodes in an order that respects dependencies.
// If a cycle is detected, it returns an error.
// The result is ordered such that independent nodes come first.
func (g *Graph) TopologicalSort() ([]string, error) {
	var result []string
	visited := make(map[string]bool)
	onStack := make(map[string]bool)

	// Sort nodes for deterministic output
	var keys []string
	for k := range g.nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var visit func(string) error
	visit = func(n string) error {
		if onStack[n] {
			return fmt.Errorf("cycle detected involving node: %s", n)
		}
		if visited[n] {
			return nil
		}

		onStack[n] = true

		// Sort dependencies for deterministic output
		var deps []string
		for dep := range g.dependencies[n] {
			deps = append(deps, dep)
		}
		sort.Strings(deps)

		for _, dep := range deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		onStack[n] = false
		visited[n] = true
		result = append(result, n)
		return nil
	}

	for _, n := range keys {
		if !visited[n] {
			if err := visit(n); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// GetDependencies returns the set of nodes that the given node depends on.
func (g *Graph) GetDependencies(id string) []string {
	var deps []string
	for dep := range g.dependencies[id] {
		deps = append(deps, dep)
	}
	sort.Strings(deps)
	return deps
}

// GetNodes returns all nodes in the graph.
func (g *Graph) GetNodes() []string {
	var nodes []string
	for n := range g.nodes {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)
	return nodes
}
