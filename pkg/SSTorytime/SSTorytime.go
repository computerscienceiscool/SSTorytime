package SSTorytime

type PoSST struct{}

// Stub implementations to make example run
func Open(load bool) PoSST                          { return PoSST{} }
func Close(ctx PoSST)                               {}
func Vertex(ctx PoSST, name, chap string) string    { return name }
func Edge(ctx PoSST, src, label, dst string, context []string, weight float32) {}

type NodePtr string
type Link struct {
    Dst  NodePtr
    Wgt  float32
    Ctx  []string
}
type Node struct {
    S string
}
func GetDBNodePtrMatchingName(ctx PoSST, name, chap string) []NodePtr { return []NodePtr{"n1"} }
func GetDBArrowsWithArrowName(ctx PoSST, name string) ([]string, string) { return nil, "arrowType" }
func GetFwdPathsAsLinks(ctx PoSST, start NodePtr, arrowType string, depth int) ([][]Link, error) {
    return [][]Link{
        {
            {Dst: "n1", Wgt: 1.0, Ctx: []string{"ctx"}},
            {Dst: "n2", Wgt: 1.0, Ctx: []string{"ctx"}},
        },
    }, nil
}
func GetDBNodeByNodePtr(ctx PoSST, ptr NodePtr) Node { return Node{S: string(ptr)} }
