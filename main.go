package main

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
)

const (
	ACTION_BALANCE = iota
	ACTION_SEND
)

var (
	privateKey = ""
	action     = ACTION_BALANCE

	sendTo            = ""
	sendAmount uint64 = 100 * constants.SompiPerKaspa
)

var dagparams = dagconfig.MainnetParams

func main() {
	keyPair := parsePrivateKey()

	rpcClient, err := rpcclient.NewRPCClient("localhost:16110")
	if err != nil {
		panic(fmt.Sprintf("Error connecting to RPC: %s", err))
	}

	address := getAddress(keyPair)
	fmt.Printf("Your address is %s\n", address)

	utxos := getUTXOS(rpcClient, keyPair, address)
	switch action {
	case ACTION_BALANCE:
		balance(utxos)
	case ACTION_SEND:
		send(rpcClient, keyPair, utxos, address)
	}
}
