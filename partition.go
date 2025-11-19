package leiden

import (
	"math"
)

// Partition represents a community assignment for all nodes
type Partition struct {
	// membership maps node to community id
	membership map[string]int
	// communitySizes tracks number of nodes in each community
	communitySizes map[int]int
	// communityWeights tracks total internal weight of each community
	communityWeights map[int]float64
}

// NewPartition creates a new partition where each node is in its own community
func NewPartition(nodes []string) *Partition {
	p := &Partition{
		membership:       make(map[string]int),
		communitySizes:   make(map[int]int),
		communityWeights: make(map[int]float64),
	}

	for i, node := range nodes {
		p.membership[node] = i
		p.communitySizes[i] = 1
		p.communityWeights[i] = 0.0
	}

	return p
}

// NewPartitionFromMembership creates a partition from existing membership map
func NewPartitionFromMembership(membership map[string]int) *Partition {
	p := &Partition{
		membership:       make(map[string]int),
		communitySizes:   make(map[int]int),
		communityWeights: make(map[int]float64),
	}

	for node, comm := range membership {
		p.membership[node] = comm
		p.communitySizes[comm]++
	}

	return p
}

// GetCommunity returns the community id for a node
func (p *Partition) GetCommunity(node string) int {
	if comm, exists := p.membership[node]; exists {
		return comm
	}
	return -1
}

// SetCommunity moves a node to a new community
func (p *Partition) SetCommunity(node string, newCommunity int) {
	oldCommunity := p.membership[node]

	if oldCommunity == newCommunity {
		return
	}

	// update membership
	p.membership[node] = newCommunity

	// update sizes
	p.communitySizes[oldCommunity]--
	p.communitySizes[newCommunity]++

	// if old community is empty, clean up
	if p.communitySizes[oldCommunity] == 0 {
		delete(p.communitySizes, oldCommunity)
		delete(p.communityWeights, oldCommunity)
	}
}

// Communities returns map of community id to list of nodes
func (p *Partition) Communities() map[int][]string {
	communities := make(map[int][]string)
	for node, comm := range p.membership {
		communities[comm] = append(communities[comm], node)
	}
	return communities
}

// CommunityCount returns the number of communities
func (p *Partition) CommunityCount() int {
	return len(p.communitySizes)
}

// Membership returns the membership map
func (p *Partition) Membership() map[string]int {
	return p.membership
}

// CalculateModularity computes the modularity of the partition
func (p *Partition) CalculateModularity(graph *Graph, resolution float64) float64 {
	if graph.TotalWeight() == 0 {
		return 0.0
	}

	modularity := 0.0
	communities := p.Communities()

	for _, nodes := range communities {
		if len(nodes) == 0 {
			continue
		}

		// compute edges within community
		edgesInside := 0.0
		totalDegree := 0.0

		for _, node := range nodes {
			totalDegree += graph.Degree(node)

			// count edges to other nodes in same community
			neighbors := graph.Neighbors(node)
			for neighbor, weight := range neighbors {
				if p.membership[neighbor] == p.membership[node] {
					edgesInside += weight
				}
			}
		}

		// edges counted twice
		edgesInside /= 2.0

		// modularity formula: (edges_inside / total_weight) - resolution * (total_degree / (2 * total_weight))^2
		m := graph.TotalWeight()
		modularity += (edgesInside / m) - resolution*math.Pow(totalDegree/(2.0*m), 2)
	}

	return modularity
}

// CalculateModularityGain computes the change in modularity from moving a node to a new community
func (p *Partition) CalculateModularityGain(
	node string,
	oldCommunity int,
	newCommunity int,
	graph *Graph,
	resolution float64,
) float64 {
	if oldCommunity == newCommunity {
		return 0.0
	}

	m := graph.TotalWeight()
	if m == 0 {
		return 0.0
	}

	// calculate edge weights to old and new communities
	weightToOld := 0.0
	weightToNew := 0.0

	neighbors := graph.Neighbors(node)
	for neighbor, weight := range neighbors {
		neighborComm := p.membership[neighbor]
		if neighborComm == oldCommunity {
			weightToOld += weight
		} else if neighborComm == newCommunity {
			weightToNew += weight
		}
	}

	// node degree
	nodeDegree := graph.Degree(node)

	// compute community degrees (total degree of nodes in each community)
	oldCommDegree := 0.0
	newCommDegree := 0.0
	for n, comm := range p.membership {
		if comm == oldCommunity {
			oldCommDegree += graph.Degree(n)
		} else if comm == newCommunity {
			newCommDegree += graph.Degree(n)
		}
	}

	// modularity gain formula
	// gain from joining new community
	gainNew := weightToNew - resolution*nodeDegree*newCommDegree/(2.0*m)
	// loss from leaving old community
	lossOld := weightToOld - resolution*nodeDegree*(oldCommDegree-nodeDegree)/(2.0*m)

	return gainNew - lossOld
}

// Clone creates a deep copy of the partition
func (p *Partition) Clone() *Partition {
	newPartition := &Partition{
		membership:       make(map[string]int),
		communitySizes:   make(map[int]int),
		communityWeights: make(map[int]float64),
	}

	for node, comm := range p.membership {
		newPartition.membership[node] = comm
	}

	for comm, size := range p.communitySizes {
		newPartition.communitySizes[comm] = size
	}

	for comm, weight := range p.communityWeights {
		newPartition.communityWeights[comm] = weight
	}

	return newPartition
}
