package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-cactus/cactus"
	"go-cactus/model"
)

// 示例
func main() {
	ctx := context.Background()
	client := cactus.NewClient()
	resp, err := client.CheckAddress(ctx, &model.CheckAddressReq{
		Addresses: []string{"3SYQn32YG7XowiCzXKuXqnqBWtFvQDp3WeK36eE8rTEi"},
		CoinName:  "USDT_SOL",
	})
	if err != nil {
		fmt.Println(err)
	}
	// 将结构体转换为 JSON 格式
	jsonData, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Println("Error converting response to JSON:", err)
		return
	}

	// 打印 JSON 数据
	fmt.Println(string(jsonData))
}
