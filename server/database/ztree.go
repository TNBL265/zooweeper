package zooweeper

// ZTree maintains the tree data structure. It doesn't have any networking  or client connection code in it
// so that it can be tested in a stand alone way.
//
// The tree maintains two parallel data structures: a hashtable that maps from full paths to DataNodes and a tree of
// ZNodes. All accesses to a path is  through the hashtable. The tree is traversed only when serializing to disk.
type ZTree struct {
}
