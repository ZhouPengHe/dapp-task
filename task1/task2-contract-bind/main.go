package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"task2-contract-bind/counters"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	infuraUrl := os.Getenv("INFURA_URL")
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	client, err := ethclient.Dial(infuraUrl)
	if err != nil {
		log.Fatalf("Failed to connect to Sepolia:", err)
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key:", err)
	}
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)
	// 获取 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get pending nonce:", err)
	}
	// 获取sepolia链ID
	chainId, _ := client.NetworkID(context.Background())
	// 构造交易授权器
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatalf("Failed to create transactor:", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	// gas 上限
	auth.GasLimit = uint64(300000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("Failed to get gas price:", err)
	}
	auth.GasPrice = gasPrice
	// 部署合约
	address, tx, instance, err := counters.DeployCounter(auth, client, big.NewInt(1))
	if err != nil {
		log.Fatalf("Failed to deploy counter:", err)
	}
	fmt.Println("合约部署中, 地址:", address.Hex())
	fmt.Println("合约部署中, 交易哈希:", tx.Hash().Hex())
	// 获取最新的 nonce
	newNonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get new nonce: %v", err)
	}
	auth.Nonce = big.NewInt(int64(newNonce))
	// 调用 increment
	txTwo, err := instance.Increment(auth)
	if err != nil {
		log.Fatalf("Failed to increment counter:", err)
	}
	fmt.Println("调用increment交易哈希:", txTwo.Hash().Hex())
	// 等待交易确认
	receipt, err := bind.WaitMined(context.Background(), client, txTwo)
	if err != nil {
		log.Fatalf("Failed to wait for tx:", err)
	}
	if receipt.Status == 1 {
		fmt.Println("Increment 交易成功")
	} else {
		log.Fatal("Increment 交易失败")
	}
	// 查询 count
	count, err := instance.GetCount(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatalf("Failed to get count:", err)
	}
	fmt.Println("当前计数:", count)
}
