package zooweeper

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type ReadOps struct {
	ab *AtomicBroadcast
}

func (ro *ReadOps) GetAllMetadata(w http.ResponseWriter, r *http.Request) {
	// connect to the database.
	results, err := ro.ab.ZTree.AllMetadata()
	// return results
	err = ro.ab.writeJSON(w, http.StatusOK, results)
	if err != nil {
		ro.ab.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}

func (ro *ReadOps) DoesScoreExist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leader")

	result, err := ro.ab.ZTree.CheckMetadataExist(id)
	if err != nil {
		ro.ab.errorJSON(w, err)
		return
	}

	err = ro.ab.writeJSON(w, http.StatusOK, strconv.FormatBool(result))
	if err != nil {
		ro.ab.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

}
