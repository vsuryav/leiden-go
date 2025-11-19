package leiden

import (
	"fmt"
	"math/rand"
)

// Config controls leiden algorithm parameters
type Config struct {
	// resolution parameter: higher = more smaller communities
	Resolution float64
	// max iterations for move nodes phase
	MaxIterations int
	// minimum modularity gain to continue
	MinModularityGain float64
	// random seed for reproducibility
	RandomSeed int64
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Resolution:        1.0,
		MaxIterations:     100,
		MinModularityGain: 0.0001,
		RandomSeed:        42,
	}
}

// Result contains the result of leiden algorithm
type Result struct {
	// final partition
	Partition *Partition
	// final modularity score
	Modularity float64
	// number of iterations until convergence
	Iterations int
	// number of communities found
	CommunityCount int
}

// Leiden runs the leiden community detection algorithm
func Leiden(graph *Graph, config *Config) (*Result, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// validate graph
	if err := graph.ValidateGraph(); err != nil {
		return nil, fmt.Errorf("invalid graph: %w", err)
	}

	// set random seed for reproducibility
	rand.Seed(config.RandomSeed)

	// initialize partition: each node in its own community
	partition := NewPartition(graph.Nodes())

	improved := true
	iteration := 0

	for improved && iteration < config.MaxIterations {
		iteration++
		improved = false

		// phase 1: move nodes to optimize modularity
		moveImproved := moveNodes(graph, partition, config)

		// phase 2: refine partition (leiden's key innovation)
		refineImproved := refinePartition(graph, partition, config)

		improved = moveImproved || refineImproved

		// phase 3: aggregate graph
		if improved {
			// only aggregate if we're continuing
			// in practice, we track this across iterations
		}
	}

	modularity := partition.CalculateModularity(graph, config.Resolution)

	return &Result{
		Partition:      partition,
		Modularity:     modularity,
		Iterations:     iteration,
		CommunityCount: partition.CommunityCount(),
	}, nil
}

// moveNodes performs the move nodes phase (similar to louvain)
func moveNodes(graph *Graph, partition *Partition, config *Config) bool {
	improved := false
	nodes := graph.Nodes()

	// randomize node order to avoid bias
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	for _, node := range nodes {
		currentCommunity := partition.GetCommunity(node)

		// find best community to move to
		bestCommunity := currentCommunity
		bestGain := 0.0

		// get neighboring communities
		neighborCommunities := make(map[int]bool)
		neighbors := graph.Neighbors(node)
		for neighbor := range neighbors {
			neighborComm := partition.GetCommunity(neighbor)
			neighborCommunities[neighborComm] = true
		}

		// also consider staying in current community
		neighborCommunities[currentCommunity] = true

		// evaluate each neighboring community
		for candidateCommunity := range neighborCommunities {
			if candidateCommunity == currentCommunity {
				continue
			}

			gain := partition.CalculateModularityGain(
				node,
				currentCommunity,
				candidateCommunity,
				graph,
				config.Resolution,
			)

			if gain > bestGain && gain > config.MinModularityGain {
				bestGain = gain
				bestCommunity = candidateCommunity
			}
		}

		// move node if improvement found
		if bestCommunity != currentCommunity {
			partition.SetCommunity(node, bestCommunity)
			improved = true
		}
	}

	return improved
}

// refinePartition performs the refinement phase (leiden's key innovation)
// this splits communities that are poorly connected
func refinePartition(graph *Graph, partition *Partition, config *Config) bool {
	improved := false
	communities := partition.Communities()

	for communityID, nodes := range communities {
		if len(nodes) <= 1 {
			continue
		}

		// create subgraph for this community
		subgraph := createSubgraph(graph, nodes)

		// check if community is well-connected
		// if not, split it using local moves
		if !isWellConnected(subgraph, nodes, partition) {
			// perform local moves within community to split it
			splitCommunity(subgraph, nodes, partition, communityID, config)
			improved = true
		}
	}

	return improved
}

// createSubgraph creates a subgraph containing only the specified nodes
func createSubgraph(graph *Graph, nodes []string) *Graph {
	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node] = true
	}

	subEdges := make(map[string]map[string]float64)
	for _, node := range nodes {
		subEdges[node] = make(map[string]float64)
		neighbors := graph.Neighbors(node)
		for neighbor, weight := range neighbors {
			if nodeSet[neighbor] {
				subEdges[node][neighbor] = weight
			}
		}
	}

	return NewGraph(subEdges)
}

// isWellConnected checks if a community is well-connected
func isWellConnected(subgraph *Graph, nodes []string, partition *Partition) bool {
	if len(nodes) <= 2 {
		return true
	}

	// use a simple connectivity check: ensure subgraph is connected
	// more sophisticated check would look at edge connectivity
	visited := make(map[string]bool)
	queue := []string{nodes[0]}
	visited[nodes[0]] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		neighbors := subgraph.Neighbors(current)
		for neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// if not all nodes are reachable, community is not well-connected
	return len(visited) == len(nodes)
}

// splitCommunity splits a poorly connected community
func splitCommunity(
	subgraph *Graph,
	nodes []string,
	partition *Partition,
	originalCommunity int,
	config *Config,
) {
	// create temporary partition for subgraph
	subPartition := NewPartition(nodes)

	// perform local moves within subgraph
	for i := 0; i < 10; i++ { // limited iterations for refinement
		improved := false

		// randomize node order
		shuffled := make([]string, len(nodes))
		copy(shuffled, nodes)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		for _, node := range shuffled {
			currentComm := subPartition.GetCommunity(node)

			// find best local community
			bestComm := currentComm
			bestGain := 0.0

			neighbors := subgraph.Neighbors(node)
			neighborComms := make(map[int]bool)
			for neighbor := range neighbors {
				neighborComms[subPartition.GetCommunity(neighbor)] = true
			}

			for candidateComm := range neighborComms {
				if candidateComm == currentComm {
					continue
				}

				gain := subPartition.CalculateModularityGain(
					node,
					currentComm,
					candidateComm,
					subgraph,
					config.Resolution,
				)

				if gain > bestGain {
					bestGain = gain
					bestComm = candidateComm
				}
			}

			if bestComm != currentComm {
				subPartition.SetCommunity(node, bestComm)
				improved = true
			}
		}

		if !improved {
			break
		}
	}

	// if subpartition created multiple communities, update main partition
	if subPartition.CommunityCount() > 1 {
		// find next available community id
		maxCommID := 0
		for _, commID := range partition.Membership() {
			if commID > maxCommID {
				maxCommID = commID
			}
		}

		// map sub-communities to new global community ids
		subCommMapping := make(map[int]int)
		nextID := maxCommID + 1

		for node := range subPartition.Membership() {
			subComm := subPartition.GetCommunity(node)

			// first node in this sub-community keeps original community
			if len(subCommMapping) == 0 {
				subCommMapping[subComm] = originalCommunity
			} else if _, exists := subCommMapping[subComm]; !exists {
				subCommMapping[subComm] = nextID
				nextID++
			}

			partition.SetCommunity(node, subCommMapping[subComm])
		}
	}
}
