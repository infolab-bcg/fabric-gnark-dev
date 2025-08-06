package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/infolab-bcg/fabric-gnark-dev/chaincode-go/gnarkverify"
)

func main() {
	gnarkVerifyCode, err := contractapi.NewChaincode(&gnarkverify.GnarkVerifyContract{})
	if err != nil {
		log.Panicf("Error creating ZK proof chaincode: %v", err)
	}

	if err := gnarkVerifyCode.Start(); err != nil {
		log.Panicf("Error starting ZK proof chaincode: %v", err)
	}
}
