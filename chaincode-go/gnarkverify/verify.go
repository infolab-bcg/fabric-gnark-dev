package gnarkverify

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/oliverustc/gnarkabc/utils"
)

// GnarkVerifyContract 定义智能合约结构
type GnarkVerifyContract struct {
	contractapi.Contract
}

func decodeBase64(name, str string) ([]byte, error) {
	strBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s from base64 %v: %v", name, str, err)
	}
	return strBytes, nil
}

func readGroth16VK(vkStr string, curve ecc.ID) (groth16.VerifyingKey, error) {
	vkStrBytes, err := decodeBase64("vk", vkStr)
	if err != nil {
		return nil, err
	}
	vk := groth16.NewVerifyingKey(curve)
	_, err = vk.ReadFrom(bytes.NewReader(vkStrBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read vk from bytes %v: %v", vkStrBytes, err)
	}
	return vk, nil
}

func readGroth16Proof(proofStr string, curve ecc.ID) (groth16.Proof, error) {
	proofStrBytes, err := decodeBase64("proof", proofStr)
	if err != nil {
		return nil, err
	}
	proof := groth16.NewProof(curve)
	_, err = proof.ReadFrom(bytes.NewReader(proofStrBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read proof from bytes %v: %v", proofStrBytes, err)
	}
	return proof, nil
}

func readPublicWitness(pubWitnessStr string, curve ecc.ID) (witness.Witness, error) {
	pubWitnessStrBytes, err := decodeBase64("publicWitness", pubWitnessStr)
	if err != nil {
		return nil, err
	}
	publicWitness, err := witness.New(curve.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create public witness from bytes %v: %v", pubWitnessStrBytes, err)
	}
	_, err = publicWitness.ReadFrom(bytes.NewReader(pubWitnessStrBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read public witness from bytes %v: %v", pubWitnessStrBytes, err)
	}
	return publicWitness, nil
}

func (c *GnarkVerifyContract) VerifyGroth16Proof(ctx contractapi.TransactionContextInterface, curveName string, proofStr string, vkStr string, pubWitnessStr string) (string, error) {

	curve := utils.CurveMap[curveName]
	proof, err := readGroth16Proof(proofStr, curve)
	if err != nil {
		return "", err
	}

	vk, err := readGroth16VK(vkStr, curve)
	if err != nil {
		return "read groth16 verifyingkey failed", err
	}

	publicWitness, err := readPublicWitness(pubWitnessStr, curve)
	if err != nil {
		return "read groth16 public witness failed", err
	}

	// 验证证明
	if err := groth16.Verify(proof, vk, publicWitness); err != nil {
		return "verify groth16 proof failed", fmt.Errorf("failed to verify proof: %v", err)
	}

	return "verify groth16 proof success", nil
}

func readPlonkVK(vkStr string, curve ecc.ID) (plonk.VerifyingKey, error) {
	vkStrBytes, err := decodeBase64("vk", vkStr)
	if err != nil {
		return nil, err
	}
	vk := plonk.NewVerifyingKey(curve)
	_, err = vk.ReadFrom(bytes.NewReader(vkStrBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read vk from bytes %v: %v", vkStrBytes, err)
	}
	return vk, nil
}

func readPlonkProof(proofStr string, curve ecc.ID) (plonk.Proof, error) {
	proofStrBytes, err := decodeBase64("proof", proofStr)
	if err != nil {
		return nil, err
	}
	proof := plonk.NewProof(curve)
	_, err = proof.ReadFrom(bytes.NewReader(proofStrBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read proof from bytes %v: %v", proofStrBytes, err)
	}
	return proof, nil
}

func (c *GnarkVerifyContract) VerifyPlonkProof(ctx contractapi.TransactionContextInterface, curveName string, proofStr string, vkStr string, pubWitnessStr string) (string, error) {

	curve := utils.CurveMap[curveName]
	vk, err := readPlonkVK(vkStr, curve)
	if err != nil {
		return "read plonk verifyingkey failed", err
	}

	proof, err := readPlonkProof(proofStr, curve)
	if err != nil {
		return "read plonk proof failed", err
	}

	publicWitness, err := readPublicWitness(pubWitnessStr, curve)
	if err != nil {
		return "read plonk public witness failed", err
	}

	// 验证证明
	if err := plonk.Verify(proof, vk, publicWitness); err != nil {
		return "verify plonk proof failed", fmt.Errorf("failed to verify proof: %v", err)
	}

	return "verify plonk proof success", nil
}

// GetContractInfo 获取合约信息
func (c *GnarkVerifyContract) GetContractInfo(ctx contractapi.TransactionContextInterface) (string, error) {
	return "Gnark Verification Contract", nil
}
