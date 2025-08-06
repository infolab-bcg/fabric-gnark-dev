package gnarkverify

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleCircuit 定义一个简单的电路结构
type SimpleCircuit struct {
	X frontend.Variable `gnark:"x"`       // 输入变量 X
	Y frontend.Variable `gnark:",public"` // 公共输出变量 Y
}

// Define 定义电路的逻辑：Y = X^3 + X + 5
func (sc *SimpleCircuit) Define(api frontend.API) error {
	// 计算 X 的立方
	x3 := api.Mul(sc.X, sc.X, sc.X)
	// 计算 Y 的值
	res := api.Add(x3, sc.X, 5)
	// 断言 Y 等于计算结果
	api.AssertIsEqual(sc.Y, res)
	return nil
}

// GnarkVerifyContract 定义智能合约结构
type GnarkVerifyContract struct {
	contractapi.Contract
}

// VerifyProofResponse 定义验证结果响应结构
type VerifyProofResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VerifyProof 主要验证函数
// 接受一个长度为3的字符串数组，分别对应 vk, proof, publicWitness 的序列化数据
func (c *GnarkVerifyContract) VerifyProof(ctx contractapi.TransactionContextInterface, proofStr string, vkStr string, pubWitnessStr string) (*VerifyProofResponse, error) {

	// 反序列化证明 (proof) from base64
	proofStrBytes, err := base64.StdEncoding.DecodeString(proofStr)
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to decode proof from base64 %v: %v", proofStr, err),
		}, nil
	}
	proof := groth16.NewProof(ecc.BN254)
	_, err = proof.ReadFrom(bytes.NewReader(proofStrBytes))
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read proof from bytes %v: %v", proofStrBytes, err),
		}, nil
	}

	// 反序列化验证密钥 (vk) from base64
	vkStrBytes, err := base64.StdEncoding.DecodeString(vkStr)
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to decode vk from base64 %v: %v", vkStr, err),
		}, nil
	}
	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err = vk.ReadFrom(bytes.NewReader(vkStrBytes))
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read vk from bytes %v: %v", vkStrBytes, err),
		}, nil
	}

	// 反序列化公共见证 (publicWitness) from base64
	pubWitnessStrBytes, err := base64.StdEncoding.DecodeString(pubWitnessStr)
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to decode public witness from base64 %v: %v", pubWitnessStr, err),
		}, nil
	}
	publicWitness, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create public witness from bytes %v: %v", pubWitnessStrBytes, err),
		}, nil
	}
	_, err = publicWitness.ReadFrom(bytes.NewReader(pubWitnessStrBytes))
	if err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read public witness from bytes %v: %v", pubWitnessStrBytes, err),
		}, nil
	}

	// 验证证明
	if err := groth16.Verify(proof, vk, publicWitness); err != nil {
		return &VerifyProofResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to verify proof: %v", err),
		}, nil
	}

	return &VerifyProofResponse{
		Success: true,
		Message: "Proof verified successfully",
	}, nil
}

// GetContractInfo 获取合约信息
func (c *GnarkVerifyContract) GetContractInfo(ctx contractapi.TransactionContextInterface) (string, error) {
	return "Gnark Verification Contract - Supports verification of Groth16 proofs for circuit Y = X^3 + X + 5", nil
}
