package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var StoreABI = `[{"inputs":[{"internalType":"uint256","name":"init","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"newCount","type":"uint256"}],"name":"CountIncremented","type":"event"},{"inputs":[],"name":"getCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"increment","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	wssInfuraUrl := os.Getenv("WSS_INFURA_URL")
	address := os.Getenv("ADDRESS")
	client, err := ethclient.Dial(wssInfuraUrl)
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress(address)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}
	contractAbi, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("监听CountIncremented事件中...")
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Println("区块号:", vLog.BlockNumber)
			fmt.Println("交易哈希:", vLog.TxHash.Hex())
			event := struct {
				NewCount *big.Int
			}{}
			err := contractAbi.UnpackIntoInterface(&event, "CountIncremented", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			topics := make([]string, len(vLog.Topics))
			for i, t := range vLog.Topics {
				topics[i] = t.Hex()
			}
			fmt.Println("topics=", topics)
			fmt.Println("NewCount:", event.NewCount.String())
		}
	}
}
