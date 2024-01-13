package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ethereum-crawler/db"
	"ethereum-crawler/model"
	"ethereum-crawler/utils"
	"github.com/gorilla/mux"
)

// Create new http server with the given db, port
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
	r.HandleFunc("/totalPaidFee/{address}", h.totalPaidFeeHandler).Methods("GET")
	r.HandleFunc("/status", h.statusHandler).Methods("GET")
	r.HandleFunc("/firstTransaction", h.firstTransactionHandler).Methods("GET")

	// h.port to string
	portString := strconv.FormatUint(uint64(h.port), 10)

	r.Host("localhost:" + portString)
	err = http.ListenAndServe(":"+portString, r)
	err = http.ListenAndServe("localhost:"+strconv.FormatUint(uint64(h.port), 10), r)

	return
}

func (h *httpService) totalPaidFeeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	res, err := h.db.Get(vars["address"])
	if err != nil && !h.db.IsNotFoundError(err) {
		h.log.Errorw(err.Error())
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil && !h.db.IsNotFoundError(err) {
		h.log.Errorw(err.Error())
	}
}

func (h *httpService) statusHandler(w http.ResponseWriter, r *http.Request) {
	res, err := h.db.Get(db.LastHeightKey)
	w.WriteHeader(http.StatusOK)
	if err != nil && !h.db.IsNotFoundError(err) {
		h.log.Errorw(err.Error())
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		h.log.Errorw(err.Error())
	}
}

// get firstTransaction
func (h *httpService) firstTransactionHandler(w http.ResponseWriter, r *http.Request) {
	records, err := h.db.Records(nil, nil)
	minTransactionHeight := int64(1e10)
	var firstAccount model.Account
	for _, v := range records {
		//Find the minimum value
		if v.FirstHeight < minTransactionHeight && v.FirstHeight != 0 {
			minTransactionHeight = v.FirstHeight
			firstAccount = v
		}
	}
	w.WriteHeader(http.StatusOK)
	if err != nil && !h.db.IsNotFoundError(err) {
		h.log.Errorw(err.Error())
	}
	err = json.NewEncoder(w).Encode(firstAccount)
	if err != nil {
		h.log.Errorw(err.Error())
	}
}
