# leiden-go

high-quality community detection for go using the leiden algorithm.

## what is leiden?

leiden is a community detection algorithm that improves upon the popular louvain method. it guarantees well-connected communities and produces higher quality partitions.

**why leiden over louvain:**
- guarantees communities are internally well-connected
- produces more stable results across runs
- higher quality partitions with better modularity
- faster convergence in practice

published in [Scientific Reports (2019)](https://www.nature.com/articles/s41598-019-41695-z)

## installation

```bash
go get github.com/villagelabsco/leiden-go
```

## usage

```go
package main

import (
    "fmt"
    "github.com/villagelabsco/leiden-go"
)

func main() {
    // create a graph using adjacency map
    edges := map[string]map[string]float64{
        "A": {"B": 1.0, "C": 1.0},
        "B": {"A": 1.0, "C": 1.0},
        "C": {"A": 1.0, "B": 1.0},
        "D": {"E": 1.0, "F": 1.0},
        "E": {"D": 1.0, "F": 1.0},
        "F": {"D": 1.0, "E": 1.0},
    }

    graph := leiden.NewGraph(edges)

    // configure leiden
    config := &leiden.Config{
        Resolution:        1.0,
        MaxIterations:     100,
        MinModularityGain: 0.0001,
        RandomSeed:        42,
    }

    // run leiden algorithm
    result, err := leiden.Leiden(graph, config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("found %d communities\n", result.CommunityCount)
    fmt.Printf("modularity: %.4f\n", result.Modularity)
    fmt.Printf("iterations: %d\n", result.Iterations)

    // get community assignments
    communities := result.Partition.Communities()
    for commID, nodes := range communities {
        fmt.Printf("community %d: %v\n", commID, nodes)
    }
}
```

## configuration

```go
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
```

**defaults:**
- resolution: 1.0
- max iterations: 100
- min modularity gain: 0.0001
- random seed: 42

**tuning resolution:**
- `0.5-0.8`: fewer, larger communities
- `1.0`: balanced (default)
- `1.5-2.0`: more, smaller communities

## graph format

graphs are represented as weighted, undirected adjacency maps:

```go
edges := map[string]map[string]float64{
    "node1": {
        "node2": 1.0,  // edge weight
        "node3": 2.0,
    },
    "node2": {
        "node1": 1.0,
        "node3": 0.5,
    },
    "node3": {
        "node1": 2.0,
        "node2": 0.5,
    },
}
```

**important:**
- edges must be symmetric (undirected)
- weights must be positive
- self-loops are allowed but not recommended

## performance

benchmarks on apple m4 pro:

```
BenchmarkLeidenSmall-14   (30 nodes)    151 μs/op
BenchmarkLeiden-14        (100 nodes)   1.8 ms/op
BenchmarkLeidenLarge-14   (500 nodes)   103 ms/op
```

## algorithm details

leiden consists of three phases:

1. **move nodes**: iteratively move nodes to neighboring communities to optimize modularity (like louvain)

2. **refine partition**: split poorly connected communities using local moves (leiden's innovation)

3. **aggregate**: create super-graph where communities become nodes

phases repeat until convergence.

**key difference from louvain:**
the refinement phase ensures communities are well-connected. louvain can create "islands" where nodes are connected only through external nodes. leiden guarantees internal connectivity.

## testing

```bash
# run tests
go test -v

# run benchmarks
go test -bench=. -benchtime=3s

# test coverage
go test -cover
```

all tests include:
- basic graph operations
- partition management
- modularity calculations
- known community structures
- edge cases (disconnected graphs, single nodes)
- determinism verification
- resolution parameter effects

## validation

the implementation is validated against:
- **karate club graph**: classic community detection benchmark
- **disconnected components**: should find separate communities
- **synthetic graphs**: known community structures
- **modularity scores**: compared against expected values

## use cases

- social network analysis
- document clustering
- protein interaction networks
- recommendation systems
- project discovery (original use case in thor)

## comparison with other algorithms

| algorithm | quality | speed | guarantees |
|-----------|---------|-------|------------|
| louvain   | good    | fast  | none       |
| leiden    | better  | fast  | well-connected |
| label propagation | variable | very fast | none |
| infomap   | good    | slow  | none       |

leiden provides the best balance of quality and performance.

## license

mit

## references

- traag, v.a., waltman, l. & van eck, n.j. "from louvain to leiden: guaranteeing well-connected communities." sci rep 9, 5233 (2019).
- blondel, v. d., guillaume, j. l., lambiotte, r., & lefebvre, e. "fast unfolding of communities in large networks." j stat mech 2008, p10008 (2008).

## contributing

contributions welcome! please:
- add tests for new features
- ensure all tests pass
- follow existing code style
- update documentation

## author

surya v (vsuryav@gmail.com)
