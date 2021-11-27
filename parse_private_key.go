package main

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/go-secp256k1"
)

func parsePrivateKey() *secp256k1.SchnorrKeyPair {
	privateKeyBytes := make([]byte, hex.DecodedLen(len(privateKey)))
	_,err := hex.Decode(privateKeyBytes, []byte(privateKey))
	if err !=nil{
		panic(fmt.Sprintf("Error decoding private key: %s", err))
	}
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err !=nil{
		panic(fmt.Sprintf("Error deserializing private key: %s", err))
	}
	return keyPair
}
