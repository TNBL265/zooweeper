package zooweeper

import (
	"net/http"
)

type ReadOps struct {
	ab *AtomicBroadcast
}

func (ro *ReadOps) GetAllMetadata(w http.ResponseWriter, r *http.Request) {
	// connect to the ztree.
	results, err := ro.ab.ZTree.AllMetadata()
	// return results
	err = ro.ab.writeJSON(w, http.StatusOK, results)
	if err != nil {
		ro.ab.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}
