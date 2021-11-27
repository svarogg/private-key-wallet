package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
)

func balance(utxos []*walletUTXO) {
	totalBalance := uint64(0)
	for _, utxo := range utxos{
		totalBalance += utxo.UTXOEntry.Amount()
	}

	fmt.Printf("Your balance is: %f\n", float64(totalBalance)/constants.SompiPerKaspa)
}
