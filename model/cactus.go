package model

import "github.com/shopspring/decimal"

type CheckAddressReq struct {
	Addresses []string `json:"addresses"`
	CoinName  string   `json:"coin_name"`
}
type CheckAddressResp struct {
	Code       int      `json:"code"`
	Message    string   `json:"message"`
	Successful bool     `json:"successful"`
	Data       []string `json:"data"`
}

type CreateOrderReq struct {
	FromAddress         *string           `json:"from_address,omitempty"`
	FromWalletCode      string            `json:"from_wallet_code"`
	CoinName            string            `json:"coin_name"`
	OrderNo             string            `json:"order_no"`
	DestAddressItemList []DestAddressItem `json:"dest_address_item_list"`
	Description         *string           `json:"description,omitempty"`
	FeeRateLevel        float64           `json:"fee_rate_level,omitempty"`
	FeeRate             float64           `json:"fee_rate,omitempty"`
}
type DestAddressItem struct {
	MemoType         *string         `json:"memo_type,omitempty"`
	Memo             *string         `json:"memo,omitempty"`
	DestAddress      string          `json:"dest_address"`
	Amount           decimal.Decimal `json:"amount"`
	IsAllWithdrawal  bool            `json:"is_all_withdrawal"`
	Remark           *string         `json:"remark,omitempty"`
	ContractTransfer *bool           `json:"contract_transfer,omitempty"`
	ContractAggre    *bool           `json:"contract_aggre,omitempty"`
}

type CreateOrderResp struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Successful bool   `json:"successful"`
	Data       struct {
		OrderNo string `json:"order_no"`
	} `json:"data"`
}

type TxSummaryReq struct {
	CoinName        string   `json:"coin_name"`
	TxTypes         []string `json:"tx_types,omitempty"`
	Addresses       []string `json:"addresses,omitempty"`
	Offset          *int     `json:"offset,omitempty"`
	Limit           *int     `json:"limit,omitempty"`
	CreateTimeOrder *int     `json:"create_time_order,omitempty"`
	StartTime       *int64   `json:"start_time,omitempty"`
	EndTime         *int64   `json:"end_time,omitempty"`
}

type TxSummaryResp struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Successful bool   `json:"successful"` // 使用指针处理可能为null的情况
	Data       struct {
		Total  int `json:"total"`
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
		List   []struct {
			WalletCode      string  `json:"wallet_code"`
			Chain           string  `json:"chain,omitempty"` // 根据文档补充
			WalletType      string  `json:"wallet_type"`     // MIXED_ADDRESS/SEGREGATED_ADDRESS
			CoinName        string  `json:"coin_name"`
			OrderNo         string  `json:"order_no"`
			BlockHeight     int64   `json:"block_height"`
			TxID            string  `json:"tx_id"`
			TxType          string  `json:"tx_type"`        // 枚举值参考文档
			Amount          float64 `json:"amount"`         // 使用string处理大数/精度
			WalletBalance   float64 `json:"wallet_balance"` // 同上
			RemarkDetail    string  `json:"remark_detail"`
			TxTimeStamp     int64   `json:"tx_time_stamp"` // 时间戳用int64
			CreateTimeStamp int64   `json:"create_time_stamp"`
		} `json:"list"`
	} `json:"data"`
}

type TxDetailReq struct {
	BID             string   `json:"-"`                   //业务线ID
	WalletCode      string   `json:"-"`                   //钱包地址
	CoinName        string   `json:"coin_name,omitempty"` //币种名称
	TxTypes         []string `json:"tx_types,omitempty"`  // 可选
	Addresses       []string `json:"addresses,omitempty"` // 可选
	ID              int64    `json:"id,omitempty"`        // 指针处理可选整型
	TxID            *string  `json:"tx_id,omitempty"`
	OrderNo         string   `json:"order_no,omitempty"`
	Offset          *int     `json:"offset,omitempty"`            // 默认0
	Limit           *int     `json:"limit,omitempty"`             // 默认10
	CreateTimeOrder *int     `json:"create_time_order,omitempty"` // 0=降序 1=升序
	StartTime       *int64   `json:"start_time,omitempty"`        // 时间戳用int64
	EndTime         *int64   `json:"end_time,omitempty"`
}

type TxDetailResp struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Successful bool   `json:"successful"` // 处理可能为null的布尔值
	Data       struct {
		Total  int `json:"total"`
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
		List   []struct {
			ID              int             `json:"id"`
			DomainID        string          `json:"domain_id"`
			WalletCode      string          `json:"wallet_code"`
			WalletType      string          `json:"wallet_type"` // MIXED_ADDRESS/SEGREGATED_ADDRESS
			CoinName        string          `json:"coin_name"`
			OrderNo         string          `json:"order_no,omitempty"`
			BlockHeight     int64           `json:"block_height"`
			ConfirmRatio    string          `json:"confirm_ratio,omitempty"`
			TxID            string          `json:"tx_id"`
			TxSize          int64           `json:"tx_size"`
			TxType          string          `json:"tx_type"` // 枚举值参考文档
			WithdrawAmount  decimal.Decimal `json:"withdraw_amount,omitempty"`
			GasPrice        *string         `json:"gas_price,omitempty"`
			GasLimit        *string         `json:"gas_limit,omitempty"`
			TxFee           decimal.Decimal `json:"tx_fee"`
			MinerReward     *string         `json:"miner_reward,omitempty"`
			DepositAmount   decimal.Decimal `json:"deposit_amount"`
			WalletBalance   float64         `json:"wallet_balance"`
			TxStatus        string          `json:"tx_status"` // 状态枚举
			RemarkDetail    *string         `json:"remark_detail,omitempty"`
			TxTimeStamp     int64           `json:"tx_time_stamp"` // 时间戳用int64
			CreateTimeStamp int64           `json:"create_time_stamp"`
			Vins            []Vin           `json:"vins"`
			Vouts           []Vout          `json:"vouts"`
		} `json:"list"`
	} `json:"data"`
}

// Vin 付款方地址详情
type Vin struct {
	Address  string          `json:"address"`
	Index    int             `json:"idx"`
	Tag      *string         `json:"tag,omitempty"`
	Amount   decimal.Decimal `json:"amount,omitempty"`
	Balance  float64         `json:"balance,omitempty"`
	IsChange int             `json:"is_change"`
	Desc     *string         `json:"desc,omitempty"`
}

// Vout 收款方地址详情
type Vout struct {
	Address  string          `json:"address"`
	Index    int             `json:"idx"`
	Tag      *string         `json:"tag,omitempty"`
	Amount   decimal.Decimal `json:"amount"`
	Balance  float64         `json:"balance"`
	IsChange int             `json:"is_change"`
	Desc     *string         `json:"desc,omitempty"`
}

type GetAddressesReq struct {
	// 查询参数（使用指针和 omitempty 处理可选性）
	CoinName            string  `json:"coin_name"`                       // 币种名称
	HideNoCoinAddress   *string `json:"hide_no_coin_address,omitempty"`  // 是否隐藏无币地址（"true"/"false"）
	KeyWord             *string `json:"key_word,omitempty"`              // 关键字搜索
	Offset              *int    `json:"offset,omitempty"`                // 分页偏移量
	Limit               *int    `json:"limit,omitempty"`                 // 每页数量
	SortByBalance       *string `json:"sort_by_balance,omitempty"`       // 排序方式（"DESC"/"ASC"）
	MinBalance          *int64  `json:"min_balance,omitempty"`           // 最小余额（单位：最小数币单位）
	MaxBalance          *int64  `json:"max_balance,omitempty"`           // 最大余额（单位：最小数币单位）
	ManageWalletAddress *bool   `json:"manage_wallet_address,omitempty"` // 是否查询ETH管理地址
}

type GetAddressesResp struct {
	Code       int    `json:"code"`       // 状态码（0=成功）
	Message    string `json:"message"`    // 错误信息
	Successful bool   `json:"successful"` // 请求状态（可能为 null）
	Data       struct {
		Offset int `json:"offset"` // 当前偏移量
		Limit  int `json:"limit"`  // 每页限制
		Total  int `json:"total"`  // 总数
		List   []struct {
			DomainID         string  `json:"domain_id"`          // 企业 Domain ID
			BID              string  `json:"b_id"`               // 业务线 ID
			WalletCode       string  `json:"wallet_code"`        // 钱包编号
			WalletType       string  `json:"wallet_type"`        // 钱包类型（MIXED/SEGREGATED）
			Address          string  `json:"address"`            // 地址字符串
			AddressType      string  `json:"address_type"`       // 地址类型（NORMAL_ADDRESS）
			AddressStorage   string  `json:"address_storage"`    // 存储类型（COLD/HOT）
			CoinName         string  `json:"coin_name"`          // 币种名称（如 BTC）
			BCHAddressFormat *string `json:"bch_address_format"` // BCH 格式（CashAddr/Legacy）
			Description      string  `json:"description"`        // 地址描述
			FreezeAmount     float64 `json:"freeze_amount"`      // 冻结金额（string 避免精度丢失）
			TotalAmount      float64 `json:"total_amount"`       // 总金额（string 避免精度丢失）
			AvailableAmount  float64 `json:"available_amount"`   // 可用金额（可选字段）
		} `json:"list"` // 地址列表
	} `json:"data"` // 数据主体
}

type AddressList struct {
	Offset int           `json:"offset"` // 当前偏移量
	Limit  int           `json:"limit"`  // 每页限制
	Total  int           `json:"total"`  // 总数
	List   []AddressInfo `json:"list"`   // 地址列表
}

type AddressInfo struct {
	DomainID         string  `json:"domain_id"`          // 企业 Domain ID
	BID              string  `json:"b_id"`               // 业务线 ID
	WalletCode       string  `json:"wallet_code"`        // 钱包编号
	WalletType       string  `json:"wallet_type"`        // 钱包类型（MIXED/SEGREGATED）
	Address          string  `json:"address"`            // 地址字符串
	AddressType      string  `json:"address_type"`       // 地址类型（NORMAL_ADDRESS）
	AddressStorage   string  `json:"address_storage"`    // 存储类型（COLD/HOT）
	CoinName         string  `json:"coin_name"`          // 币种名称（如 BTC）
	BCHAddressFormat *string `json:"bch_address_format"` // BCH 格式（CashAddr/Legacy）
	Description      string  `json:"description"`        // 地址描述
	FreezeAmount     float64 `json:"freeze_amount"`      // 冻结金额（string 避免精度丢失）
	TotalAmount      float64 `json:"total_amount"`       // 总金额（string 避免精度丢失）
	AvailableAmount  float64 `json:"available_amount"`   // 可用金额（可选字段）
}
