package server

type BalanceResponse struct {
	FIL  string `json:"fil"`
	IFIL string `json:"ifil"`
}

type SubmitTransactionRequest struct {
	SignedTx string  `json:"signed_tx"`
	Sender   string  `json:"sender"`
	Receiver string  `json:"receiver"`
	Amount   float64 `json:"amount"`
}

type SubmitTransactionResponse struct {
	Hash string `json:"hash"`
}
