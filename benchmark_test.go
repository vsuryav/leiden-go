package leiden

import (
	"fmt"
	"testing"
)

// BenchmarkLeiden benchmarks leiden algorithm
func BenchmarkLeiden(b *testing.B) {
	// create a medium-sized graph
	edges := createBenchmarkGraph(100, 3)
	graph := NewGraph(edges)
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Leiden(graph, config)
		if err != nil {
			b.Fatalf("leiden failed: %v", err)
		}
	}
}

// BenchmarkLeidenSmall benchmarks leiden on small graph
func BenchmarkLeidenSmall(b *testing.B) {
	edges := createBenchmarkGraph(30, 2)
	graph := NewGraph(edges)
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Leiden(graph, config)
		if err != nil {
			b.Fatalf("leiden failed: %v", err)
		}
	}
}

// BenchmarkLeidenLarge benchmarks leiden on large graph
func BenchmarkLeidenLarge(b *testing.B) {
	edges := createBenchmarkGraph(500, 5)
	graph := NewGraph(edges)
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Leiden(graph, config)
		if err != nil {
			b.Fatalf("leiden failed: %v", err)
		}
	}
}

// BenchmarkModularityCalculation benchmarks modularity calculation
func BenchmarkModularityCalculation(b *testing.B) {
	edges := createBenchmarkGraph(100, 3)
	graph := NewGraph(edges)
	partition := NewPartition(graph.Nodes())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		partition.CalculateModularity(graph, 1.0)
	}
}

// createBenchmarkGraph creates a graph with multiple communities
// numCommunities: number of communities to create
// nodesPerCommunity: nodes in each community
func createBenchmarkGraph(numCommunities, nodesPerCommunity int) map[string]map[string]float64 {
	edges := make(map[string]map[string]float64)

	for comm := 0; comm < numCommunities; comm++ {
		// create nodes for this community
		nodes := make([]string, nodesPerCommunity)
		for i := 0; i < nodesPerCommunity; i++ {
			nodes[i] = fmt.Sprintf("c%d_n%d", comm, i)
		}

		// create dense connections within community
		for i := 0; i < len(nodes); i++ {
			if edges[nodes[i]] == nil {
				edges[nodes[i]] = make(map[string]float64)
			}
			for j := i + 1; j < len(nodes); j++ {
				// high weight within community
				edges[nodes[i]][nodes[j]] = 1.0
				if edges[nodes[j]] == nil {
					edges[nodes[j]] = make(map[string]float64)
				}
				edges[nodes[j]][nodes[i]] = 1.0
			}
		}

		// add weak connections to next community
		if comm < numCommunities-1 {
			nextCommFirstNode := fmt.Sprintf("c%d_n0", comm+1)
			currentCommLastNode := nodes[len(nodes)-1]
			if edges[currentCommLastNode] == nil {
				edges[currentCommLastNode] = make(map[string]float64)
			}
			edges[currentCommLastNode][nextCommFirstNode] = 0.2
			if edges[nextCommFirstNode] == nil {
				edges[nextCommFirstNode] = make(map[string]float64)
			}
			edges[nextCommFirstNode][currentCommLastNode] = 0.2
		}
	}

	return edges
}
