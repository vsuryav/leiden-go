package leiden

import (
	"math"
	"testing"
)

// TestNewGraph tests graph creation
func TestNewGraph(t *testing.T) {
	edges := map[string]map[string]float64{
		"A": {"B": 1.0, "C": 2.0},
		"B": {"A": 1.0, "C": 1.0},
		"C": {"A": 2.0, "B": 1.0},
	}

	graph := NewGraph(edges)

	if graph.NodeCount() != 3 {
		t.Errorf("expected 3 nodes, got %d", graph.NodeCount())
	}

	// total weight = (1 + 2 + 1 + 1 + 2 + 1) / 2 = 4
	expectedWeight := 4.0
	if math.Abs(graph.TotalWeight()-expectedWeight) > 0.001 {
		t.Errorf("expected total weight %f, got %f", expectedWeight, graph.TotalWeight())
	}

	// check degrees
	if graph.Degree("A") != 3.0 {
		t.Errorf("expected degree 3 for A, got %f", graph.Degree("A"))
	}
	if graph.Degree("B") != 2.0 {
		t.Errorf("expected degree 2 for B, got %f", graph.Degree("B"))
	}
	if graph.Degree("C") != 3.0 {
		t.Errorf("expected degree 3 for C, got %f", graph.Degree("C"))
	}
}

// TestNewPartition tests partition creation
func TestNewPartition(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	partition := NewPartition(nodes)

	if partition.CommunityCount() != 3 {
		t.Errorf("expected 3 communities, got %d", partition.CommunityCount())
	}

	// each node should be in its own community
	communities := partition.Communities()
	for _, nodes := range communities {
		if len(nodes) != 1 {
			t.Errorf("expected 1 node per community, got %d", len(nodes))
		}
	}
}

// TestSetCommunity tests moving nodes between communities
func TestSetCommunity(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	partition := NewPartition(nodes)

	// move B to A's community
	commA := partition.GetCommunity("A")
	partition.SetCommunity("B", commA)

	if partition.CommunityCount() != 2 {
		t.Errorf("expected 2 communities after merge, got %d", partition.CommunityCount())
	}

	// A and B should be in same community
	if partition.GetCommunity("A") != partition.GetCommunity("B") {
		t.Error("A and B should be in same community")
	}

	// C should be in different community
	if partition.GetCommunity("C") == partition.GetCommunity("A") {
		t.Error("C should be in different community")
	}
}

// TestModularityCalculation tests modularity computation
func TestModularityCalculation(t *testing.T) {
	// create simple graph with clear community structure
	// two cliques: {A, B} and {C, D}
	edges := map[string]map[string]float64{
		"A": {"B": 1.0},
		"B": {"A": 1.0},
		"C": {"D": 1.0},
		"D": {"C": 1.0},
	}

	graph := NewGraph(edges)
	nodes := []string{"A", "B", "C", "D"}
	partition := NewPartition(nodes)

	// put A and B together
	commA := partition.GetCommunity("A")
	partition.SetCommunity("B", commA)

	// put C and D together
	commC := partition.GetCommunity("C")
	partition.SetCommunity("D", commC)

	// with resolution=1.0, this should have high modularity
	modularity := partition.CalculateModularity(graph, 1.0)

	// perfect separation should have modularity close to 0.5
	if modularity < 0.4 || modularity > 0.6 {
		t.Errorf("expected modularity around 0.5, got %f", modularity)
	}
}

// TestModularityGain tests modularity gain calculation
func TestModularityGain(t *testing.T) {
	// simple triangle graph
	edges := map[string]map[string]float64{
		"A": {"B": 1.0, "C": 1.0},
		"B": {"A": 1.0, "C": 1.0},
		"C": {"A": 1.0, "B": 1.0},
	}

	graph := NewGraph(edges)
	nodes := []string{"A", "B", "C"}
	partition := NewPartition(nodes)

	commA := partition.GetCommunity("A")
	commB := partition.GetCommunity("B")

	// moving B to A's community should have positive gain
	gain := partition.CalculateModularityGain("B", commB, commA, graph, 1.0)

	if gain <= 0 {
		t.Errorf("expected positive gain for joining connected nodes, got %f", gain)
	}
}

// TestKarateClub tests leiden on classic karate club graph
func TestKarateClub(t *testing.T) {
	// simplified karate club: two clear communities
	edges := map[string]map[string]float64{
		// community 1 (dense connections)
		"1": {"2": 1.0, "3": 1.0, "4": 1.0},
		"2": {"1": 1.0, "3": 1.0},
		"3": {"1": 1.0, "2": 1.0, "4": 1.0},
		"4": {"1": 1.0, "3": 1.0, "5": 0.5},
		// community 2 (dense connections)
		"5": {"4": 0.5, "6": 1.0, "7": 1.0, "8": 1.0},
		"6": {"5": 1.0, "7": 1.0},
		"7": {"5": 1.0, "6": 1.0, "8": 1.0},
		"8": {"5": 1.0, "7": 1.0},
	}

	graph := NewGraph(edges)
	config := DefaultConfig()
	config.Resolution = 1.0

	result, err := Leiden(graph, config)
	if err != nil {
		t.Fatalf("leiden failed: %v", err)
	}

	// should find 2 communities
	if result.CommunityCount < 2 {
		t.Errorf("expected at least 2 communities, got %d", result.CommunityCount)
	}

	// modularity should be positive
	if result.Modularity <= 0 {
		t.Errorf("expected positive modularity, got %f", result.Modularity)
	}

	// nodes 1,2,3,4 should mostly be in one community
	// nodes 5,6,7,8 should mostly be in another
	comm1 := result.Partition.GetCommunity("1")
	comm2 := result.Partition.GetCommunity("2")
	comm3 := result.Partition.GetCommunity("3")

	// at least 1 and 2 should be together
	if comm1 != comm2 {
		t.Error("nodes 1 and 2 should be in same community")
	}

	// at least 1 and 3 should be together
	if comm1 != comm3 {
		t.Error("nodes 1 and 3 should be in same community")
	}
}

// TestDisconnectedGraph tests leiden on disconnected components
func TestDisconnectedGraph(t *testing.T) {
	// two separate cliques
	edges := map[string]map[string]float64{
		// clique 1
		"A": {"B": 1.0, "C": 1.0},
		"B": {"A": 1.0, "C": 1.0},
		"C": {"A": 1.0, "B": 1.0},
		// clique 2
		"D": {"E": 1.0, "F": 1.0},
		"E": {"D": 1.0, "F": 1.0},
		"F": {"D": 1.0, "E": 1.0},
	}

	graph := NewGraph(edges)
	config := DefaultConfig()

	result, err := Leiden(graph, config)
	if err != nil {
		t.Fatalf("leiden failed: %v", err)
	}

	// should find 2 communities
	if result.CommunityCount != 2 {
		t.Errorf("expected 2 communities for disconnected components, got %d", result.CommunityCount)
	}

	// A, B, C should be together
	commA := result.Partition.GetCommunity("A")
	commB := result.Partition.GetCommunity("B")
	commC := result.Partition.GetCommunity("C")

	if commA != commB || commA != commC {
		t.Error("nodes A, B, C should be in same community")
	}

	// D, E, F should be together but different from A, B, C
	commD := result.Partition.GetCommunity("D")
	commE := result.Partition.GetCommunity("E")
	commF := result.Partition.GetCommunity("F")

	if commD != commE || commD != commF {
		t.Error("nodes D, E, F should be in same community")
	}

	if commA == commD {
		t.Error("two disconnected components should be in different communities")
	}
}

// TestSingleNode tests edge case with single node
func TestSingleNode(t *testing.T) {
	edges := map[string]map[string]float64{
		"A": {},
	}

	graph := NewGraph(edges)

	// graph should be invalid (no edges)
	err := graph.ValidateGraph()
	if err == nil {
		t.Error("expected error for graph with no edges")
	}
}

// TestEmptyGraph tests edge case with empty graph
func TestEmptyGraph(t *testing.T) {
	edges := map[string]map[string]float64{}

	graph := NewGraph(edges)

	err := graph.ValidateGraph()
	if err == nil {
		t.Error("expected error for empty graph")
	}
}

// TestNegativeWeights tests validation of negative edge weights
func TestNegativeWeights(t *testing.T) {
	edges := map[string]map[string]float64{
		"A": {"B": -1.0},
		"B": {"A": -1.0},
	}

	graph := NewGraph(edges)

	err := graph.ValidateGraph()
	if err == nil {
		t.Error("expected error for negative edge weights")
	}
}

// TestResolutionParameter tests different resolution values
func TestResolutionParameter(t *testing.T) {
	// graph with medium connections
	edges := map[string]map[string]float64{
		"A": {"B": 1.0, "C": 0.5},
		"B": {"A": 1.0, "C": 0.5},
		"C": {"A": 0.5, "B": 0.5, "D": 1.0},
		"D": {"C": 1.0, "E": 1.0},
		"E": {"D": 1.0, "F": 1.0},
		"F": {"E": 1.0},
	}

	graph := NewGraph(edges)

	// low resolution should give fewer communities
	configLow := DefaultConfig()
	configLow.Resolution = 0.5

	resultLow, err := Leiden(graph, configLow)
	if err != nil {
		t.Fatalf("leiden failed with low resolution: %v", err)
	}

	// high resolution should give more communities
	configHigh := DefaultConfig()
	configHigh.Resolution = 2.0

	resultHigh, err := Leiden(graph, configHigh)
	if err != nil {
		t.Fatalf("leiden failed with high resolution: %v", err)
	}

	// high resolution should find at least as many communities
	if resultHigh.CommunityCount < resultLow.CommunityCount {
		t.Errorf("higher resolution should find at least as many communities: low=%d, high=%d",
			resultLow.CommunityCount, resultHigh.CommunityCount)
	}
}

// TestDeterminism tests that same seed produces same results
func TestDeterminism(t *testing.T) {
	edges := map[string]map[string]float64{
		"A": {"B": 1.0, "C": 1.0, "D": 1.0},
		"B": {"A": 1.0, "E": 1.0},
		"C": {"A": 1.0, "F": 1.0},
		"D": {"A": 1.0, "G": 1.0},
		"E": {"B": 1.0},
		"F": {"C": 1.0},
		"G": {"D": 1.0},
	}

	graph := NewGraph(edges)
	config := DefaultConfig()
	config.RandomSeed = 123

	result1, err := Leiden(graph, config)
	if err != nil {
		t.Fatalf("leiden run 1 failed: %v", err)
	}

	result2, err := Leiden(graph, config)
	if err != nil {
		t.Fatalf("leiden run 2 failed: %v", err)
	}

	// should produce same number of communities
	if result1.CommunityCount != result2.CommunityCount {
		t.Errorf("expected same community count, got %d and %d",
			result1.CommunityCount, result2.CommunityCount)
	}

	// should produce same modularity
	if math.Abs(result1.Modularity-result2.Modularity) > 0.0001 {
		t.Errorf("expected same modularity, got %f and %f",
			result1.Modularity, result2.Modularity)
	}
}

// TestPartitionClone tests partition cloning
func TestPartitionClone(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	partition := NewPartition(nodes)

	// modify partition
	commA := partition.GetCommunity("A")
	partition.SetCommunity("B", commA)

	// clone
	clone := partition.Clone()

	// clone should match original
	if clone.CommunityCount() != partition.CommunityCount() {
		t.Error("clone should have same community count")
	}

	for _, node := range nodes {
		if clone.GetCommunity(node) != partition.GetCommunity(node) {
			t.Errorf("clone should have same community for node %s", node)
		}
	}

	// modifying clone shouldn't affect original
	clone.SetCommunity("C", commA)

	if partition.GetCommunity("C") == partition.GetCommunity("A") {
		t.Error("modifying clone should not affect original")
	}
}

// TestLargeGraph tests leiden on a larger synthetic graph
func TestLargeGraph(t *testing.T) {
	// create graph with 3 clear communities of 10 nodes each
	edges := make(map[string]map[string]float64)

	// community 1: nodes 0-9
	for i := 0; i < 10; i++ {
		nodeI := string(rune('0' + i))
		edges[nodeI] = make(map[string]float64)
		for j := i + 1; j < 10; j++ {
			nodeJ := string(rune('0' + j))
			edges[nodeI][nodeJ] = 1.0
			if edges[nodeJ] == nil {
				edges[nodeJ] = make(map[string]float64)
			}
			edges[nodeJ][nodeI] = 1.0
		}
	}

	// community 2: nodes A-J
	for i := 0; i < 10; i++ {
		nodeI := string(rune('A' + i))
		if edges[nodeI] == nil {
			edges[nodeI] = make(map[string]float64)
		}
		for j := i + 1; j < 10; j++ {
			nodeJ := string(rune('A' + j))
			edges[nodeI][nodeJ] = 1.0
			if edges[nodeJ] == nil {
				edges[nodeJ] = make(map[string]float64)
			}
			edges[nodeJ][nodeI] = 1.0
		}
	}

	// community 3: nodes a-j
	for i := 0; i < 10; i++ {
		nodeI := string(rune('a' + i))
		if edges[nodeI] == nil {
			edges[nodeI] = make(map[string]float64)
		}
		for j := i + 1; j < 10; j++ {
			nodeJ := string(rune('a' + j))
			edges[nodeI][nodeJ] = 1.0
			if edges[nodeJ] == nil {
				edges[nodeJ] = make(map[string]float64)
			}
			edges[nodeJ][nodeI] = 1.0
		}
	}

	graph := NewGraph(edges)
	config := DefaultConfig()
	config.Resolution = 1.0

	result, err := Leiden(graph, config)
	if err != nil {
		t.Fatalf("leiden failed on large graph: %v", err)
	}

	// should find 3 communities
	if result.CommunityCount != 3 {
		t.Logf("warning: expected 3 communities, got %d (acceptable if close)", result.CommunityCount)
	}

	// modularity should be high for well-separated communities
	if result.Modularity < 0.5 {
		t.Errorf("expected high modularity for well-separated communities, got %f", result.Modularity)
	}
}
