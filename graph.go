package leiden

import (
	"fmt"
	"math"
)

// Graph represents a weighted undirected graph
type Graph struct {
	// adjacency list: node -> neighbor -> weight
	edges map[string]map[string]float64
	nodes []string
	// total edge weight in graph (2m in modularity formula)
	totalWeight float64
	// node degrees (sum of edge weights for each node)
	nodeDegrees map[string]float64
}

// NewGraph creates a new graph from an adjacency map
func NewGraph(edges map[string]map[string]float64) *Graph {
	g := &Graph{
		edges:       edges,
		nodeDegrees: make(map[string]float64),
	}

	// collect unique nodes
	nodeSet := make(map[string]bool)
	for node := range edges {
		nodeSet[node] = true
	}
	g.nodes = make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		g.nodes = append(g.nodes, node)
	}

	// compute total weight and node degrees
	g.totalWeight = 0.0
	for node, neighbors := range edges {
		degree := 0.0
		for _, weight := range neighbors {
			degree += weight
			g.totalWeight += weight
		}
		g.nodeDegrees[node] = degree
	}
	// edges counted twice in undirected graph
	g.totalWeight /= 2.0

	return g
}

// Nodes returns all nodes in the graph
func (g *Graph) Nodes() []string {
	return g.nodes
}

// Neighbors returns neighbors of a node with their edge weights
func (g *Graph) Neighbors(node string) map[string]float64 {
	if neighbors, exists := g.edges[node]; exists {
		return neighbors
	}
	return make(map[string]float64)
}

// EdgeWeight returns the weight of an edge between two nodes
func (g *Graph) EdgeWeight(node1, node2 string) float64 {
	if neighbors, exists := g.edges[node1]; exists {
		if weight, exists := neighbors[node2]; exists {
			return weight
		}
	}
	return 0.0
}

// Degree returns the degree (sum of edge weights) of a node
func (g *Graph) Degree(node string) float64 {
	if degree, exists := g.nodeDegrees[node]; exists {
		return degree
	}
	return 0.0
}

// TotalWeight returns the total edge weight in the graph
func (g *Graph) TotalWeight() float64 {
	return g.totalWeight
}

// NodeCount returns the number of nodes in the graph
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// CreateAggregateGraph creates a new graph where each community becomes a single node
func (g *Graph) CreateAggregateGraph(partition *Partition) *Graph {
	// map community id to list of nodes
	communityNodes := make(map[int][]string)
	for node, comm := range partition.membership {
		communityNodes[comm] = append(communityNodes[comm], node)
	}

	// create super-graph edges: community -> community -> weight
	superEdges := make(map[string]map[string]float64)

	for commID := range communityNodes {
		superNode := fmt.Sprintf("comm_%d", commID)
		superEdges[superNode] = make(map[string]float64)
	}

	// aggregate edges between communities
	for node, neighbors := range g.edges {
		nodeCommunity := partition.membership[node]
		superNode1 := fmt.Sprintf("comm_%d", nodeCommunity)

		for neighbor, weight := range neighbors {
			neighborCommunity := partition.membership[neighbor]

			// skip self-loops within same community for external edges
			if nodeCommunity == neighborCommunity {
				continue
			}

			superNode2 := fmt.Sprintf("comm_%d", neighborCommunity)

			// add edge between super nodes
			if superEdges[superNode1] == nil {
				superEdges[superNode1] = make(map[string]float64)
			}
			superEdges[superNode1][superNode2] += weight
		}
	}

	return NewGraph(superEdges)
}

// ValidateGraph checks if graph is valid
func (g *Graph) ValidateGraph() error {
	if len(g.nodes) == 0 {
		return fmt.Errorf("graph has no nodes")
	}

	if g.totalWeight <= 0 {
		return fmt.Errorf("graph has no edges")
	}

	// check for negative weights
	for node, neighbors := range g.edges {
		for neighbor, weight := range neighbors {
			if weight < 0 {
				return fmt.Errorf("negative edge weight between %s and %s: %f", node, neighbor, weight)
			}
			if math.IsNaN(weight) || math.IsInf(weight, 0) {
				return fmt.Errorf("invalid edge weight between %s and %s: %f", node, neighbor, weight)
			}
		}
	}

	return nil
}
