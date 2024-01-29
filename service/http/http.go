package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chain-crawler/service"

	"chain-crawler/db"
	"chain-crawler/model"
	"chain-crawler/utils"

	"github.com/gorilla/mux"
)

type httpService struct {
	db   db.DB[model.Account]
	port uint16
	log  *utils.ZapLogger
}

func New(db db.DB[model.Account], log *utils.ZapLogger, port uint16) service.Service {
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
	r.HandleFunc("/v2/doc", h.serveDoc).Methods("GET")

	// h.port to string
	portString := strconv.FormatUint(uint64(h.port), 10)

	r.Host("localhost:" + portString)
	err = http.ListenAndServe(":"+portString, r)
	if err != nil {
		h.log.Errorw(err.Error())
	}

	return
}

func (h *httpService) serveDoc(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./openapi/generated/doc.html")
}

func (h *httpService) totalPaidFeeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := h.db.Get(vars["address"])
	if err != nil && !h.db.IsNotFoundError(err) {
		h.log.Errorw(err.Error())
	} else if err != nil && h.db.IsNotFoundError(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
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
	records := make(chan db.DBItem[model.Account], 10)
	err := error(nil)
	go func() {
		err = h.db.Records(nil, nil, records)
	}()

	minTransactionHeight := int64(1e10)
	var firstAccount model.Account

	for {
		item, ok := <-records
		if ok {
			if item.Value.FirstHeight < minTransactionHeight && item.Value.FirstHeight != 0 {
				minTransactionHeight = item.Value.FirstHeight
				firstAccount = item.Value
			}
		} else {
			break
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
