package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"math/big"
	"time"
)

type Erc20filter struct {
	Topics []string `json:"topics"`
}
type Erc20Log struct {
	BlockNum        uint64
	BlockHash       string
	TransHash       string
	ContractAddress string
	From            string
	To              string
	Value           string
}

var (
	topiChan = make(chan string, 50)
	bus      = EventBus.New()
)

func main() {
	fmt.Println("你好，我的个人简历在doc目录")
	logChan := make(chan []byte, 50)
	startERC20Listen(logChan)

	go startStreamClientMointor()
	/*
		server := EventBus.NewServer(":3344", "/_erc20_transfer_", bus)
		err = server.Start()
		if err != nil {
			panic(err)
		}
	*/
	for {
		select {
		case buf := <-logChan:
			var log ethgo.Log
			err := json.Unmarshal(buf, &log)
			if err != nil {
				fmt.Println(err)
			}
			if len(log.Topics) == 3 {
				v, b := new(big.Int).SetString(hex.EncodeToString(log.Data), 16)
				if !b {
					fmt.Println(err)
				}
				erc20Log := Erc20Log{
					log.BlockNumber,
					log.BlockHash.String(),
					log.TransactionHash.String(),
					log.Address.String(),
					ethgo.HexToAddress(log.Topics[1].String()).String(),
					ethgo.HexToAddress(log.Topics[1].String()).String(),
					v.String(),
				}
				publicTops, err := json.MarshalIndent(erc20Log,"","\t")
				if err != nil {
					fmt.Println(err)
				}
				if bus.HasCallback("erc20:transfer") {
					bus.Publish("erc20:transfer", string(publicTops))
				}
			}
		case t := <-time.After(30 * time.Second):
			fmt.Println(t, "server stream heartbeat...")
		}
	}
}
func startERC20Listen(logChan chan []byte) {
	client, err := jsonrpc.NewClient("wss://mainnet.infura.io/ws/v3/e5aa3d782eff42508c31debf527987ff")
	if err != nil {
		fmt.Println(err)
		return
	}
	ss := []interface{}{
		"logs",
	}
	transferSigHash := ethgo.BytesToHash(ethgo.Keccak256([]byte("Transfer(address,address,uint256)"))).String()
	ss = append(ss, Erc20filter{
		Topics: []string{transferSigHash},
	})
	_, err = client.Subscribe(ss, func(b []byte) {
		logChan <- b
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}
func startStreamClientMointor() {
	err := bus.Subscribe("erc20:transfer", func(msg string) {
		topiChan <- msg
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		select {
		case topic := <-topiChan:
			fmt.Println(time.Now(), topic)
		case t := <-time.After(time.Second * 30):
			fmt.Println(t, "Monitor hearbeat ...")
		}
	}
}
