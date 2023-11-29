package zab

import (
	"net/http"
)

// ReadOps for Read Request
type ReadOps struct {
	ab *AtomicBroadcast
}

// GetAllMetadata returns all ZNode from the ZTree as a list of Metadata
func (ro *ReadOps) GetAllMetadata(w http.ResponseWriter, r *http.Request) {
	results, _ := ro.ab.ZTree.AllMetadata()
	ro.ab.writeJSON(w, http.StatusOK, results)
}
