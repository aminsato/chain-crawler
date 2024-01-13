package model

type Account struct {
	Address      string `json:"address"`
	TotalPaidFee uint64 `json:"totalPaidFee"`
	LastHeight   int64
	TxIndex      int
	FirstHeight  int64
}
