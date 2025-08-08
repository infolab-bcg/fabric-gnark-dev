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
	"github.com/oliverustc/gnarkabc/logger"
	"github.com/oliverustc/gnarkabc/utils"
	"github.com/oliverustc/gnarkabc/wrapper/groth16wrapper"
	"github.com/oliverustc/gnarkabc/wrapper/plonkwrapper"
	"github.com/stretchr/testify/require"
)

// 运行 go generate ./... 生成 mock 文件, 运行一次即可
//
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
	require.Equal(t, "Gnark Verification Contract", response)
}

type ZKSNARKParams struct {
	VK            string `json:"vk"`
	Proof         string `json:"proof"`
	WitnessPublic string `json:"witnessPublic"`
}

func groth16Generate(date string, curveName string) error {
	var circuit circuits.Product
	circuit.PreCompile(nil)
	curve := utils.CurveMap[curveName]
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

	logger.Debug("Groth16 done")
	vkStr, _ := zk.MarshalVKToStr()
	proofStr, _ := zk.MarshalProofToStr()
	witnessPublicStr, _ := zk.MarshalWitnessToStr(true)
	logger.Debug("vk: %s", vkStr)
	logger.Debug("proof: %s", proofStr)
	logger.Debug("witnessPublic: %s", witnessPublicStr)

	params := ZKSNARKParams{
		VK:            vkStr,
		Proof:         proofStr,
		WitnessPublic: witnessPublicStr,
	}
	// write params into json file
	utils.EnsureDirExists("output")
	jsonFile, err := os.Create("output/groth16_" + curveName + "_" + date + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	json.NewEncoder(jsonFile).Encode(params)
	return nil
}

func groth16Verify(date string, curveName string) error {
	circuit := circuits.Product{}
	curve := utils.CurveMap[curveName]
	zk := groth16wrapper.NewWrapper(&circuit, curve)
	// read params from groth16.json
	jsonFile, err := os.Open("output/groth16_" + curveName + "_" + date + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	var params ZKSNARKParams
	json.NewDecoder(jsonFile).Decode(&params)
	zk.UnmarshalVKFromStr(params.VK)
	zk.UnmarshalProofFromStr(params.Proof)
	zk.UnmarshalWitnessFromStr(params.WitnessPublic, true)
	zk.Verify()
	logger.Debug("Groth16 verify done")
	return nil
}

func TestGroth16(t *testing.T) {
	// 使用UTC+8时区(北京时间)
	loc := time.FixedZone("CST", 8*3600) // 东八区，偏移量为8小时
	dateStr := time.Now().In(loc).Format("2006-01-02_15-04-05")
	for _, curveName := range utils.CurveNameList {
		t.Logf("generating groth16 proof... curve: [%s]", curveName)
		err := groth16Generate(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("generate groth16 proof done, curve: [%s]", curveName)
	}
	for _, curveName := range utils.CurveNameList {
		t.Logf("verifying groth16 proof... curve: [%s]", curveName)
		err := groth16Verify(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("verify groth16 proof done, curve: [%s]", curveName)
	}
}

func TestVerifyGroth16Proof(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	loc := time.FixedZone("CST", 8*3600) // 东八区，偏移量为8小时
	dateStr := time.Now().In(loc).Format("2006-01-02_15-04-05")

	for _, curveName := range utils.CurveNameList {
		t.Logf("verifying groth16 proof on chaincode... curve: [%s]", curveName)
		err := groth16Generate(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		err = groth16Verify(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		groth16File, err := os.ReadFile("output/groth16_" + dateStr + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var gnarkParams GnarkParams
		if err := json.Unmarshal(groth16File, &gnarkParams); err != nil {
			t.Fatal(err)
		}
		logger.Debug("Gnark params: %v", gnarkParams)
		gnarkVerify := &GnarkVerifyContract{}
		_, err = gnarkVerify.VerifyGroth16Proof(transactionContext, curveName, gnarkParams.Proof, gnarkParams.Vk, gnarkParams.WitnessPublic)
		require.NoError(t, err)
		t.Logf("verify groth16 proof on chaincode done, curve: [%s]", curveName)
	}

}

func plonkGenerate(date string, curveName string) error {
	var circuit circuits.Product
	circuit.PreCompile(nil)
	curve := utils.CurveMap[curveName]
	zk := plonkwrapper.NewWrapper(&circuit, curve)
	zk.Compile()
	zk.Setup()
	p := utils.RandInt(1, 100)
	q := utils.RandInt(1, 100)
	assignParams := []any{p, q}
	circuit.Assign(assignParams)
	zk.SetAssignment(&circuit)
	zk.Prove()
	zk.Verify()

	logger.Debug("Plonk done")
	vkStr, _ := zk.MarshalVKToStr()
	proofStr, _ := zk.MarshalProofToStr()
	witnessPublicStr, _ := zk.MarshalWitnessToStr(true)
	logger.Debug("vk: %s", vkStr)
	logger.Debug("proof: %s", proofStr)
	logger.Debug("witnessPublic: %s", witnessPublicStr)

	params := ZKSNARKParams{
		VK:            vkStr,
		Proof:         proofStr,
		WitnessPublic: witnessPublicStr,
	}
	// write params into json file
	utils.EnsureDirExists("output")
	jsonFile, err := os.Create("output/plonk_" + curveName + "_" + date + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	json.NewEncoder(jsonFile).Encode(params)
	return nil
}

func plonkVerify(date string, curveName string) error {
	circuit := circuits.Product{}
	curve := utils.CurveMap[curveName]
	zk := plonkwrapper.NewWrapper(&circuit, curve)
	// read params from plonk.json
	jsonFile, err := os.Open("output/plonk_" + curveName + "_" + date + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	var params ZKSNARKParams
	json.NewDecoder(jsonFile).Decode(&params)
	zk.UnmarshalVKFromStr(params.VK)
	zk.UnmarshalProofFromStr(params.Proof)
	zk.UnmarshalWitnessFromStr(params.WitnessPublic, true)
	zk.Verify()
	logger.Debug("Plonk verify done")
	return nil
}

func TestPlonk(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600) // 东八区，偏移量为8小时
	dateStr := time.Now().In(loc).Format("2006-01-02_15-04-05")
	for _, curveName := range utils.CurveNameList {
		t.Logf("generating plonk proof... curve: [%s]", curveName)
		err := plonkGenerate(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("generate plonk proof done, curve: [%s]", curveName)
	}
	for _, curveName := range utils.CurveNameList {
		t.Logf("verifying plonk proof... curve: [%s]", curveName)
		err := plonkVerify(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("verify plonk proof done, curve: [%s]", curveName)
	}
}

func TestVerifyPlonkProof(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	gnarkVerify := &GnarkVerifyContract{}
	loc := time.FixedZone("CST", 8*3600) // 东八区，偏移量为8小时
	dateStr := time.Now().In(loc).Format("2006-01-02_15-04-05")
	for _, curveName := range utils.CurveNameList {
		t.Logf("verifying plonk proof on chaincode... curve: [%s]", curveName)
		err := plonkGenerate(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		err = plonkVerify(dateStr, curveName)
		if err != nil {
			t.Fatal(err)
		}
		plonkFile, err := os.ReadFile("output/plonk_" + dateStr + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var gnarkParams GnarkParams
		if err := json.Unmarshal(plonkFile, &gnarkParams); err != nil {
			t.Fatal(err)
		}
		logger.Debug("Gnark params: %v", gnarkParams)
		_, err = gnarkVerify.VerifyPlonkProof(transactionContext, curveName, gnarkParams.Proof, gnarkParams.Vk, gnarkParams.WitnessPublic)
		require.NoError(t, err)
		t.Logf("verify plonk proof on chaincode done, curve: [%s]", curveName)
	}

}
