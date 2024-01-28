package model

type Account struct {
	Address      string `json:"address"`
	TotalPaidFee uint64 `json:"totalPaidFee"`
	LastHeight   int64  `json:"lastHeight"`
	TxIndex      int    `json:"txIndex"`
	FirstHeight  int64  `json:"firstHeight"`
	IsContract   bool   `json:"isContract"`
}

//type OldAccount struct {
//	Address      string `json:"address"`
//	TotalPaidFee uint64 `json:"totalPaidFee"`
//	LastHeight   int64
//	TxIndex      int
//	FirstHeight  int64
//}
