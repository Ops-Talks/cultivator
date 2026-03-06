package dag

import (
	"reflect"
	"strings"
	"testing"
)

func TestGraph_TopologicalSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		edges   [][2]string
		want    []string
		wantErr bool
	}{
		{
			name: "linear dependency",
			edges: [][2]string{
				{"B", "A"}, // B depends on A
				{"C", "B"}, // C depends on B
			},
			want: []string{"A", "B", "C"},
		},
		{
			name: "diamond dependency",
			edges: [][2]string{
				{"B", "A"},
				{"C", "A"},
				{"D", "B"},
				{"D", "C"},
			},
			want: []string{"A", "B", "C", "D"}, // B and C order is deterministic due to sorting
		},
		{
			name: "disconnected components",
			edges: [][2]string{
				{"B", "A"},
				{"D", "C"},
			},
			want: []string{"A", "B", "C", "D"},
		},
		{
			name: "simple cycle",
			edges: [][2]string{
				{"A", "B"},
				{"B", "A"},
			},
			wantErr: true,
		},
		{
			name: "complex cycle",
			edges: [][2]string{
				{"A", "B"},
				{"B", "C"},
				{"C", "A"},
			},
			wantErr: true,
		},
		{
			name:  "no dependencies",
			edges: [][2]string{},
			want:  nil, // Graph is empty
		},
		{
			name: "single node with multiple dependencies",
			edges: [][2]string{
				{"App", "DB"},
				{"App", "VPC"},
				{"App", "S3"},
			},
			want: []string{"DB", "S3", "VPC", "App"}, // Sorted: DB, S3, VPC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			for _, edge := range tt.edges {
				g.AddEdge(edge[0], edge[1])
			}

			got, err := g.TopologicalSort()
			if (err != nil) != tt.wantErr {
				t.Errorf("TopologicalSort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), "cycle detected") {
					t.Errorf("expected cycle error, got %v", err)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TopologicalSort() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraph_AddNode(t *testing.T) {
	t.Parallel()
	g := New()
	g.AddNode("A")
	g.AddNode("B")

	nodes := g.GetNodes()
	want := []string{"A", "B"}
	if !reflect.DeepEqual(nodes, want) {
		t.Errorf("GetNodes() = %v, want %v", nodes, want)
	}
}

func TestGraph_GetDependencies(t *testing.T) {
	t.Parallel()
	g := New()
	g.AddEdge("App", "DB")
	g.AddEdge("App", "VPC")

	deps := g.GetDependencies("App")
	want := []string{"DB", "VPC"}
	if !reflect.DeepEqual(deps, want) {
		t.Errorf("GetDependencies() = %v, want %v", deps, want)
	}
}
