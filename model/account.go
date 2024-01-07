package model

type Account struct {
	Address      string `json:"address"`
	TotalPaidFee uint64 `json:"totalPaidFee"`
	Height       int64
	TxIndex      int
}
