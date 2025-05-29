# go-cactus

`go-cactus` is a Go client for interacting with the Cactus digital asset custody service API.

## Installation

`go get github.com/KeiichiKasai/go-cactus`

## Documentation

- [Cactus API Document](https://apidoc.mycactus.com/zh-hans)

## Supported endpoints

- [VerifyAddress](https://apidoc.mycactus.com/zh-hans/addresses/verify_address.html)
- [CreateOrder](https://apidoc.mycactus.com/zh-hans/withdrawal/create_order.html)
- [GetDetails](https://apidoc.mycactus.com/zh-hans/transaction_history/get_details.html)
- [GetSummary](https://apidoc.mycactus.com/zh-hans/transaction_history/get_summary.html)
- [GetAddressList](https://apidoc.mycactus.com/zh-hans/addresses/get_address_list.html)

Feel free to open an issue or PR if you need more endpoints.

## Usage

```go
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

    jsonData, err := json.MarshalIndent(resp, "", "  ")
    if err != nil {
        fmt.Println("Error converting response to JSON:", err)
    return
    }

    fmt.Println(string(jsonData))
}

```