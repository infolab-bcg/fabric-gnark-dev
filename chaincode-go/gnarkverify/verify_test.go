package gnarkverify

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/infolab-bcg/fabric-gnark-dev/chaincode-go/gnarkverify/mocks"
	"github.com/oliverustc/gnarkabc/circuits"
	"github.com/oliverustc/gnarkabc/utils"
	"github.com/oliverustc/gnarkabc/wrapper/groth16wrapper"
	"github.com/stretchr/testify/require"
)

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

type GnarkParams struct {
	Vk            string `json:"vk"`
	Proof         string `json:"proof"`
	WitnessPublic string `json:"witnessPublic"`
}

func TestGetContractInfo(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	gnarkVerify := &GnarkVerifyContract{}
	response, err := gnarkVerify.GetContractInfo(transactionContext)
	require.NoError(t, err)
	require.Equal(t, "Gnark Verification Contract - Supports verification of Groth16 proofs for circuit Y = X^3 + X + 5", response)
}

type ZKSNARKParams struct {
	VK            string `json:"vk"`
	Proof         string `json:"proof"`
	WitnessPublic string `json:"witnessPublic"`
}

func groth16Generate(date string, t *testing.T) {
	var circuit circuits.Product
	circuit.PreCompile(nil)
	curve := utils.CurveMap["BN254"]
	zk := groth16wrapper.NewWrapper(&circuit, curve)
	zk.Compile()
	zk.Setup()
	p := utils.RandInt(1, 100)
	q := utils.RandInt(1, 100)
	assignParams := []any{p, q}
	circuit.Assign(assignParams)
	zk.SetAssignment(&circuit)
	zk.Prove()
	zk.Verify()

	t.Logf("Groth16 done")
	vkStr, _ := zk.MarshalVKToStr()
	proofStr, _ := zk.MarshalProofToStr()
	witnessPublicStr, _ := zk.MarshalWitnessToStr(true)
	t.Logf("vk: %s", vkStr)
	t.Logf("proof: %s", proofStr)
	t.Logf("witnessPublic: %s", witnessPublicStr)

	params := ZKSNARKParams{
		VK:            vkStr,
		Proof:         proofStr,
		WitnessPublic: witnessPublicStr,
	}
	// write params into json file
	utils.EnsureDirExists("output")
	jsonFile, err := os.Create("output/groth16_" + date + ".json")
	if err != nil {
		t.Error("failed to create output/groth16_" + date + ".json")
		return
	}
	defer jsonFile.Close()
	json.NewEncoder(jsonFile).Encode(params)
}

func groth16Verify(date string, t *testing.T) {
	circuit := circuits.Product{}
	curve := utils.CurveMap["BN254"]
	zk := groth16wrapper.NewWrapper(&circuit, curve)
	// read params from groth16.json
	jsonFile, err := os.Open("output/groth16_" + date + ".json")
	if err != nil {
		t.Errorf("failed to open output/groth16_" + date + ".json")
		return
	}
	defer jsonFile.Close()
	var params ZKSNARKParams
	json.NewDecoder(jsonFile).Decode(&params)
	zk.UnmarshalVKFromStr(params.VK)
	zk.UnmarshalProofFromStr(params.Proof)
	zk.UnmarshalWitnessFromStr(params.WitnessPublic, true)
	zk.Verify()
	t.Logf("Groth16 verify done")
}

func TestVerifyProof(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	dateStr := time.Now().Format("2006-01-02_15-04-05")

	groth16Generate(dateStr, t)
	groth16Verify(dateStr, t)

	groth16File, err := os.ReadFile("output/groth16_" + dateStr + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var gnarkParams GnarkParams
	if err := json.Unmarshal(groth16File, &gnarkParams); err != nil {
		t.Fatal(err)
	}
	t.Logf("Gnark params: %v", gnarkParams)
	gnarkVerify := &GnarkVerifyContract{}
	response, err := gnarkVerify.VerifyProof(transactionContext, gnarkParams.Proof, gnarkParams.Vk, gnarkParams.WitnessPublic)
	require.NoError(t, err)
	require.Equal(t, true, response.Success)
}
