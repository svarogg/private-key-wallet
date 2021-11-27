package main

import (
	"fmt"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func send(rpcClient *rpcclient.RPCClient, keyPair *secp256k1.SchnorrKeyPair, utxos []*walletUTXO,
	fromAddressStr string) {

	toAddress, err := util.DecodeAddress(sendTo, dagparams.Prefix)
	if err != nil {
		panic(fmt.Sprintf("Error decoding to address: %s", err))
	}
	fromAddress, err := util.DecodeAddress(fromAddressStr, dagparams.Prefix)
	if err != nil {
		panic(fmt.Sprintf("Error decoding to address: %s", err))
	}

	const feePerInput = 10000

	selectedUTXOs, changeSompi, err := selectUTXOs(rpcClient, utxos, sendAmount, feePerInput)
	if err != nil {
		panic(fmt.Sprintf("Error selecting utxos: %s", err))
	}

	transaction := createUnsignedTransaction(selectedUTXOs, changeSompi, toAddress, fromAddress)
	signTransaction(keyPair, transaction)

	submitTransactionResponse, err := rpcClient.SubmitTransaction(
		appmessage.DomainTransactionToRPCTransaction(transaction), false)
	if err != nil {
		panic(fmt.Sprintf("error submitting transaction: %s", err))
	}

	fmt.Printf("Transaction submitted!\nTransaction ID: %s\n", submitTransactionResponse.TransactionID)
}

func signTransaction(keyPair *secp256k1.SchnorrKeyPair, transaction *externalapi.DomainTransaction) {
	sighashReusedValues := &consensushashing.SighashReusedValues{}

	for i, input := range transaction.Inputs {
		signature, err := txscript.SignatureScript(transaction, i, consensushashing.SigHashAll, keyPair, sighashReusedValues)
		if err != nil {
			panic(fmt.Errorf("error creating signature: %s", err))
		}
		input.SignatureScript = signature
	}
}

func createUnsignedTransaction(selectedUTXOs []*walletUTXO, changeSompi uint64, toAddress util.Address,
	fromAddress util.Address) *externalapi.DomainTransaction {

	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *utxo.Outpoint,
			UTXOEntry:        utxo.UTXOEntry,
			SigOpCount:       1,
		}
	}

	toScriptPubKey, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		panic(fmt.Sprintf("Error creating pay to adressee script: %s", err))
	}
	changeScriptPubKey, err := txscript.PayToAddrScript(fromAddress)
	if err != nil {
		panic(fmt.Sprintf("Error creating pay to change script: %s", err))
	}

	outputs := []*externalapi.DomainTransactionOutput{
		{
			Value:           sendAmount,
			ScriptPublicKey: toScriptPubKey,
		},
		{
			Value:           changeSompi,
			ScriptPublicKey: changeScriptPubKey,
		},
	}

	return &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Payload:      nil,
	}
}

func selectUTXOs(rpcClient *rpcclient.RPCClient, utxos []*walletUTXO, spendAmount uint64, feePerInput uint64) (
	selectedUTXOs []*walletUTXO, changeSompi uint64, err error) {

	totalValue := uint64(0)

	dagInfo, err := rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, 0, err
	}

	for _, utxo := range utxos {
		if !isUTXOSpendable(utxo, dagInfo.VirtualDAAScore, dagparams.BlockCoinbaseMaturity) {
			continue
		}

		selectedUTXOs = append(selectedUTXOs, utxo)
		totalValue += utxo.UTXOEntry.Amount()

		fee := feePerInput * uint64(len(selectedUTXOs))
		totalSpend := spendAmount + fee
		if totalValue >= totalSpend {
			break
		}
	}

	fee := feePerInput * uint64(len(selectedUTXOs))
	totalSpend := spendAmount + fee
	if totalValue < totalSpend {
		return nil, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalSpend)/constants.SompiPerKaspa, float64(totalValue)/constants.SompiPerKaspa)
	}

	return selectedUTXOs, totalValue - totalSpend, nil
}

func isUTXOSpendable(entry *walletUTXO, virtualDAAScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase() {
		return true
	}
	return entry.UTXOEntry.BlockDAAScore()+coinbaseMaturity < virtualDAAScore
}
