const path = require("node:path");
const { exec, spawn } = require("child_process");
const fs = require("node:fs/promises");

async function generateJson() {
  // 切换目录到../../verify-on-chain
  const verifyOnChainPath = path.resolve(
    __dirname,
    "..",
    "..",
    "verify-on-chain"
  );
  process.chdir(verifyOnChainPath);
  // shell 命令 ./run.sh generate
  exec("./run.sh generate", (error, stdout, stderr) => {
    if (error) {
      console.error(`执行错误: ${error}`);
      return;
    }
    console.log(`标准输出: ${stdout}`);
    if (stderr) console.error(`标准错误: ${stderr}`);
  });
  process.chdir(__dirname);
}

async function main() {
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
        // console.log(`proofStr: ${proofStr}`);
        // console.log(`vkStr: ${vkStr}`);
        console.log(`${protocol}_${curveName}`);
        console.log(`pubWitnessStr: ${pubWitnessStr}`);
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
        // console.log(`proofStr: ${proofStr}`);
        // console.log(`vkStr: ${vkStr}`);
        console.log(`${protocol}_${curveName}`);
        console.log(`pubWitnessStr: ${pubWitnessStr}`);
      }
    }
  }
}

// 运行generateJson 函数
// generateJson();
main();
