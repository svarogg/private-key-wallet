package main

import (
	"fmt"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util"
)

type walletUTXO struct {
	Outpoint  *externalapi.DomainOutpoint
	UTXOEntry externalapi.UTXOEntry
}

func getUTXOS(rpcClient *rpcclient.RPCClient, keyPair *secp256k1.SchnorrKeyPair, address string) []*walletUTXO {
	getUTXOsByAddressesResponse, err := rpcClient.GetUTXOsByAddresses([]string{address})
	if err != nil {
		panic(fmt.Sprintf("Error getting UTXOs: %s", err))
	}

	utxos := make([]*walletUTXO, len(getUTXOsByAddressesResponse.Entries))
	for i, entry := range getUTXOsByAddressesResponse.Entries {
		outpoint, err := appmessage.RPCOutpointToDomainOutpoint(entry.Outpoint)
		if err != nil {
			panic(fmt.Sprintf("Error converting RPCUTXOEntryToUTXOEntry: %s", err))
		}

		utxoEntry, err := appmessage.RPCUTXOEntryToUTXOEntry(entry.UTXOEntry)
		if err != nil {
			panic(fmt.Sprintf("Error converting RPCUTXOEntryToUTXOEntry: %s", err))
		}
		utxos[i] = &walletUTXO{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
		}
	}
	return utxos
}

func getAddress(keyPair *secp256k1.SchnorrKeyPair) string {
	pubKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		panic(fmt.Sprintf("Error getting public key: %s", err))
	}
	pubKeySerialized, err := pubKey.Serialize()
	if err != nil {
		panic(fmt.Sprintf("Error serializing public key: %s", err))
	}

	address, err := util.NewAddressPublicKey(pubKeySerialized[:], dagparams.Prefix)

	if err != nil {
		panic(err)
	}

	return address.String()
}
