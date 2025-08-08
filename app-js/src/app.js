/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

const grpc = require('@grpc/grpc-js');
const { connect, hash, signers } = require('@hyperledger/fabric-gateway');
const crypto = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');
const { exec } = require('node:child_process');
const { TextDecoder } = require('node:util');

const channelName = envOrDefault('CHANNEL_NAME', 'mychannel');
const chaincodeName = envOrDefault('CHAINCODE_NAME', 'gnarkverify');
const mspId = envOrDefault('MSP_ID', 'Org1MSP');

// Path to crypto materials.
const cryptoPath = envOrDefault(
    'CRYPTO_PATH',
    path.resolve(
        __dirname,
        '..',
        '..',
        '..',
        'test-network',
        'organizations',
        'peerOrganizations',
        'org1.example.com'
    )
);

// Path to user private key directory.
const keyDirectoryPath = envOrDefault(
    'KEY_DIRECTORY_PATH',
    path.resolve(
        cryptoPath,
        'users',
        'User1@org1.example.com',
        'msp',
        'keystore'
    )
);

// Path to user certificate directory.
const certDirectoryPath = envOrDefault(
    'CERT_DIRECTORY_PATH',
    path.resolve(
        cryptoPath,
        'users',
        'User1@org1.example.com',
        'msp',
        'signcerts'
    )
);

// Path to peer tls certificate.
const tlsCertPath = envOrDefault(
    'TLS_CERT_PATH',
    path.resolve(cryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt')
);

// Gateway peer endpoint.
const peerEndpoint = envOrDefault('PEER_ENDPOINT', 'localhost:7051');

// Gateway peer SSL host name override.
const peerHostAlias = envOrDefault('PEER_HOST_ALIAS', 'peer0.org1.example.com');

const utf8Decoder = new TextDecoder();
const assetId = `asset${String(Date.now())}`;

async function main() {
    displayInputParameters();

    // The gRPC client connection should be shared by all Gateway connections to this endpoint.
    const client = await newGrpcConnection();

    const gateway = connect({
        client,
        identity: await newIdentity(),
        signer: await newSigner(),
        hash: hash.sha256,
        // Default timeouts for different gRPC calls
        evaluateOptions: () => {
            return { deadline: Date.now() + 5000 }; // 5 seconds
        },
        endorseOptions: () => {
            return { deadline: Date.now() + 15000 }; // 15 seconds
        },
        submitOptions: () => {
            return { deadline: Date.now() + 5000 }; // 5 seconds
        },
        commitStatusOptions: () => {
            return { deadline: Date.now() + 60000 }; // 1 minute
        },
    });

    try {
        console.log('Connect to gateway....');
        // Get a network instance representing the channel where the smart contract is deployed.
        const network = gateway.getNetwork(channelName);

        console.log('Get the smart contract from the network....');
        // Get the smart contract from the network.
        const contract = network.getContract(chaincodeName);

        console.log('Read contract info....');
        await readContractInfo(contract);
        // 
        await generateJson();
        await verify(contract);


    } finally {
        gateway.close();
        client.close();
    }
}

main().catch((error) => {
    console.error('******** FAILED to run the application:', error);
    process.exitCode = 1;
});

async function newGrpcConnection() {
    const tlsRootCert = await fs.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
    return new grpc.Client(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': peerHostAlias,
    });
}

async function newIdentity() {
    const certPath = await getFirstDirFileName(certDirectoryPath);
    const credentials = await fs.readFile(certPath);
    return { mspId, credentials };
}

async function getFirstDirFileName(dirPath) {
    const files = await fs.readdir(dirPath);
    const file = files[0];
    if (!file) {
        throw new Error(`No files in directory: ${dirPath}`);
    }
    return path.join(dirPath, file);
}

async function newSigner() {
    const keyPath = await getFirstDirFileName(keyDirectoryPath);
    const privateKeyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}


async function readContractInfo(contract) {
    console.log(
        '\n--\u003e Evaluate Transaction: ReadContractInfo, function returns contract info'
    );

    const resultBytes = await contract.evaluateTransaction('GetContractInfo');
    // resultBytes convert to string
    const resultString = utf8Decoder.decode(resultBytes);
    console.log('*** Result:', resultString);
}

async function generateJson() {
    // 切换目录到../../verify-on-chain
    const verifyOnChainPath = path.resolve(__dirname, '..', '..', 'verify-on-chain');
    process.chdir(verifyOnChainPath);
    
    // 使用 Promise 包装 exec 以支持 async/await
    return new Promise((resolve, reject) => {
        exec('./run.sh generate', (error, stdout, stderr) => {
            if (error) {
                console.error(`执行错误: ${error}`);
                process.chdir(__dirname); // 确保目录切换回来
                reject(error);
                return;
            }
            console.log(`标准输出: ${stdout}`);
            if (stderr) console.error(`标准错误: ${stderr}`);
            process.chdir(__dirname); // 确保目录切换回来
            resolve(stdout);
        });
    });
}

async function invokeContract(contract,protocol,curveName,proofStr,vkStr,pubWitnessStr) {
    console.log(
        '\n--\u003e Submit Transaction: VerifyProof, function commits a new proof'
    );
    if (protocol== "groth16") {
        funcName = "VerifyGroth16Proof"
    } else if (protocol== "plonk") {
        funcName = "VerifyPlonkProof"
    }
    const resultBytes = await contract.submitTransaction(funcName, curveName, proofStr, vkStr, pubWitnessStr);
    // resultBytes convert to string
    const resultString = utf8Decoder.decode(resultBytes);
    console.log('*** Result:', resultString);
}

const verifyOnChainPath = path.resolve(
    __dirname,
    "..",
    "..",
    "chaincode-go",
    "gnarkverify",
    "output"
  );
  process.chdir(verifyOnChainPath);
  const protocolList = ["groth16", "plonk"];
  const groth16CurveNameList = [
    "BN254",
    "BLS12-377",
    "BLS12-381",
    "BLS24-315",
    "BLS24-317",
    "BW6-633",
    "BW6-761",
  ];
  const plonkCurveNameList = [
    "BN254",
    "BLS12-377",
    "BLS12-381",
    "BLS24-315",
    "BLS24-317",
  ];

async function verify(contract) {

  for (let protocol of protocolList) {
    if (protocol === "groth16") {
      for (let curveName of groth16CurveNameList) {
        // 匹配 groth16_BLS12-377_2025-08-08_13-52-52.json 类似命名的文件
        const jsonFilePrefix = `${protocol}_${curveName}_`;
        const jsonFiles = await fs.readdir(verifyOnChainPath);
        const jsonFile = jsonFiles.find(
          (file) => file.startsWith(jsonFilePrefix) && file.endsWith(".json")
        );
        if (!jsonFile) {
          console.log(`未找到 ${protocol}_${curveName} 的 json 文件`);
          continue;
        }
        const jsonFilePath = path.resolve(verifyOnChainPath, jsonFile);
        // 读取 json 文件
        const jsonFileContent = await fs.readFile(jsonFilePath, "utf-8");
        const jsonFileContentObj = JSON.parse(jsonFileContent);
        const proofStr = jsonFileContentObj.proof;
        const vkStr = jsonFileContentObj.vk;
        const pubWitnessStr = jsonFileContentObj.witnessPublic;
        console.log(`invoke contract to verify ${protocol}_${curveName}`)
        await invokeContract(contract,protocol,curveName,proofStr,vkStr,pubWitnessStr);
      }
    } else if (protocol === "plonk") {
      for (let curveName of plonkCurveNameList) {
        // 匹配 plonk_BLS12-377_2025-08-08_13-52-52.json 类似命名的文件
        const jsonFilePrefix = `${protocol}_${curveName}_`;
        const jsonFiles = await fs.readdir(verifyOnChainPath);
        const jsonFile = jsonFiles.find(
          (file) => file.startsWith(jsonFilePrefix) && file.endsWith(".json")
        );
        if (!jsonFile) {
          console.log(`未找到 ${protocol}_${curveName} 的 json 文件`);
          continue;
        }
        const jsonFilePath = path.resolve(verifyOnChainPath, jsonFile);
        // 读取 json 文件
        const jsonFileContent = await fs.readFile(jsonFilePath, "utf-8");
        const jsonFileContentObj = JSON.parse(jsonFileContent);
        const proofStr = jsonFileContentObj.proof;
        const vkStr = jsonFileContentObj.vk;
        const pubWitnessStr = jsonFileContentObj.witnessPublic;
        console.log(`invoke contract to verify ${protocol}_${curveName}`)
        await invokeContract(contract,protocol,curveName,proofStr,vkStr,pubWitnessStr);
      }
    }
  }
}



/**
 * envOrDefault() will return the value of an environment variable, or a default value if the variable is undefined.
 */
function envOrDefault(key, defaultValue) {
    return process.env[key] || defaultValue;
}

/**
 * displayInputParameters() will print the global scope parameters used by the main driver routine.
 */
function displayInputParameters() {
    console.log(`channelName:       ${channelName}`);
    console.log(`chaincodeName:     ${chaincodeName}`);
    console.log(`mspId:             ${mspId}`);
    console.log(`cryptoPath:        ${cryptoPath}`);
    console.log(`keyDirectoryPath:  ${keyDirectoryPath}`);
    console.log(`certDirectoryPath: ${certDirectoryPath}`);
    console.log(`tlsCertPath:       ${tlsCertPath}`);
    console.log(`peerEndpoint:      ${peerEndpoint}`);
    console.log(`peerHostAlias:     ${peerHostAlias}`);
}
