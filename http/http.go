package http

import (
	"encoding/json"
	"ethereum-crawler/db"
	"ethereum-crawler/model"
	"ethereum-crawler/sync"
	"ethereum-crawler/utils"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// Create new http server with the given db, port
// New Get method for /totalpaidfee/{address} endpoint => json response with totalpaidfee, address , height
// New Get method for /status endpoint => json latest height and tx id
type httpService struct {
	db   db.DB[model.Account]
	port uint16
	log  *utils.ZapLogger
}

func New(db db.DB[model.Account], log *utils.ZapLogger, port uint16) *httpService {
	return &httpService{
		db:   db,
		port: port,
		log:  log,
	}
}

func (h *httpService) Run() (err error) {

	r := mux.NewRouter()
	r.HandleFunc("/totalpaidfee/{address}", h.totalPaidFeeHandler).Methods("GET")
	r.HandleFunc("/status", h.statusHandler).Methods("GET")

	//h.port to string
	portString := strconv.FormatUint(uint64(h.port), 10)

	r.Host("localhost:" + portString)
	err = http.ListenAndServe(":"+portString, r)
	return
}

func (h *httpService) totalPaidFeeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	res, err := h.db.Get(vars["address"])
	if err != nil && !h.db.IsNotFoundError(err) {
		fmt.Fprintf(w, err.Error())
	}
	json.NewEncoder(w).Encode(res)

}
func (h *httpService) statusHandler(w http.ResponseWriter, r *http.Request) {
	res, err := h.db.Get(sync.LastHeightKey)
	w.WriteHeader(http.StatusOK)
	if err != nil && !h.db.IsNotFoundError(err) {
		fmt.Fprintf(w, err.Error())
	}
	json.NewEncoder(w).Encode(res)

}
