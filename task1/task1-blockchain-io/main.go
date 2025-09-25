package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// 1.加载环境变量
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	infuraUrl := os.Getenv("INFURA_URL")
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	receiverHex := os.Getenv("RECEIVER")
	// 2.连接 sepolia
	client, err := ethclient.Dial(infuraUrl)
	if err != nil {
		log.Fatal("Failed to connect to Sepolia:", err)
	}
	defer client.Close()

	// 3.查询区块
	blockNumber := big.NewInt(9144476)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal("Failed to get block number:", err)
	}
	fmt.Println("区块号:", block.Number().Uint64())
	fmt.Println("区块哈希:", block.Hash().Hex())
	ts := time.Unix(int64(block.Time()), 0)
	fmt.Println("UTC时间:", ts.UTC())
	fmt.Println("本地时间:", ts.Local())
	fmt.Println("交易数量:", len(block.Transactions()))

	// 4.发送交易
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("Failed to parse private key:", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Failed to get pending nonce:", err)
	}
	// 建议小费
	gasTipCap, _ := client.SuggestGasTipCap(context.Background())
	// 建议gas价格
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	// sepolia链ID
	chainId, _ := client.ChainID(context.Background())
	receiver := common.HexToAddress(receiverHex)
	// 0.001 ETH
	amount := big.NewInt(1000000000000000)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasPrice,
		Gas:       22000,
		To:        &receiver,
		Value:     amount,
		Data:      nil,
	})
	// 获取签名器
	signer := types.LatestSignerForChainID(chainId)
	// 交易签名
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}
	// 发送交易到网络
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal("Failed to send transaction:", err)
	}
	fmt.Println("交易已发送，哈希:", signedTx.Hash().Hex())
}
