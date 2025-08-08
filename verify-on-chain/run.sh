#!/bin/bash

function networkUp() {
    pushd ../../test-network
    ./network.sh up createChannel -ca
    popd    
}

function networkDown() {
    pushd ../../test-network
    ./network.sh down
    popd
}

function deployCC() {
    pushd ../../test-network
    ./network.sh deployCC -ccn gnarkverify -ccp ../fabric-gnark-dev/chaincode-go -ccl go
    popd
}

#  use setEnv after pushd
function setEnv() {
    export PATH=${PWD}/../bin:$PATH
    export FABRIC_CFG_PATH=$PWD/../config/
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_LOCALMSPID="Org1MSP"
    
    export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=localhost:7051

    echo $PATH
    echo $FABRIC_CFG_PATH
    echo $CORE_PEER_TLS_ENABLED
    echo $CORE_PEER_LOCALMSPID
    echo $CORE_PEER_TLS_ROOTCERT_FILE
    echo $CORE_PEER_MSPCONFIGPATH
    echo $CORE_PEER_ADDRESS
}

function queryChainCode() {
    pushd ../../test-network
    setEnv
    peer chaincode query -C mychannel -n gnarkverify -c '{"Args":["GetContractInfo"]}'
    popd
}

function invokeChainCode() {
    funcName=$1
    curveName=$2
    proofStr=$3
    vkStr=$4
    pubWitnessStr=$5
    pushd ../../test-network
    setEnv
    peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n gnarkverify --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"'${funcName}'","Args":["'${curveName}'","'${proofStr}'","'${vkStr}'","'${pubWitnessStr}'"]}'
    popd
}

function generateJson() {
    pushd ../chaincode-go
    mkdir -p gnarkverify/output/.backup
    mv gnarkverify/output/*.json gnarkverify/output/.backup
    go test -v -run TestGroth16$ github.com/infolab-bcg/fabric-gnark-dev/chaincode-go/gnarkverify
    go test -v -run TestPlonk$ github.com/infolab-bcg/fabric-gnark-dev/chaincode-go/gnarkverify
    popd
}

function groth16() {
    curveNameList=("BN254" "BLS12-381" "BLS12-377" "BLS24-315" "BLS24-317" "BW6-633" "BW6-761")
    for curveName in ${curveNameList[@]}; do
        echo "============ groth16_${curveName} ============"
        pushd ../chaincode-go/gnarkverify/output
        jsonFile=$(ls | grep groth16_${curveName})
        proofStr=$(cat ${jsonFile} | jq -r '.proof')
        vkStr=$(cat ${jsonFile} | jq -r '.vk')
        pubWitnessStr=$(cat ${jsonFile} | jq -r '.witnessPublic')
        popd
        invokeChainCode "VerifyGroth16Proof" ${curveName} ${proofStr} ${vkStr} ${pubWitnessStr} 2>&1 | grep "chaincodeInvokeOrQuery"
    done
}

function plonk() {
    curveNameList=("BN254" "BLS12-381" "BLS12-377" "BLS24-315" "BLS24-317")
    for curveName in ${curveNameList[@]}; do
        echo "============ plonk_${curveName} ============"
        pushd ../chaincode-go/gnarkverify/output
        jsonFile=$(ls | grep plonk_${curveName})
        proofStr=$(cat ${jsonFile} | jq -r '.proof')
        vkStr=$(cat ${jsonFile} | jq -r '.vk')
        pubWitnessStr=$(cat ${jsonFile} | jq -r '.witnessPublic')
        popd
        invokeChainCode "VerifyPlonkProof" ${curveName} ${proofStr} ${vkStr} ${pubWitnessStr} 2>&1 | grep "chaincodeInvokeOrQuery"
    done
}


function main() {
    groth16
    plonk
}


args=$1
if [ "$args" == "up" ]; then
    networkUp
elif [ "$args" == "down" ]; then
    networkDown
elif [ "$args" == "deploy" ]; then
    deployCC
elif [ "$args" == "setenv" ]; then
    setEnv
elif [ "$args" == "query" ]; then
    queryChainCode
elif [ "$args" == "generate" ]; then
    generateJson
elif [ "$args" == "verify" ]; then
    main
else
    echo "Usage: ./run.sh [up|down|deploy|setenv|query]"
    exit 1
fi