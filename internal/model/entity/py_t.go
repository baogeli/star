// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PyT is the golang structure for table py_t.
type PyT struct {
	Id                    int         `json:"id"                    orm:"id"                      description:""`        //
	FundCode              string      `json:"fundCode"              orm:"fund_code"               description:"基金代码"`    // 基金代码
	FundName              string      `json:"fundName"              orm:"fund_name"               description:"基金名称"`    // 基金名称
	Exchange              string      `json:"exchange"              orm:"exchange"                description:"交易所"`     // 交易所
	FundCategory          string      `json:"fundCategory"          orm:"fund_category"           description:"资金类别"`    // 资金类别
	OpeningPrice          float64     `json:"openingPrice"          orm:"opening_price"           description:"开盘价"`     // 开盘价
	HighestPrice          float64     `json:"highestPrice"          orm:"highest_price"           description:"最高价"`     // 最高价
	LowestPrice           float64     `json:"lowestPrice"           orm:"lowest_price"            description:"最低价"`     // 最低价
	CurrentPrice          float64     `json:"currentPrice"          orm:"current_price"           description:"当前价"`     // 当前价
	PreviousClosePrice    float64     `json:"previousClosePrice"    orm:"previous_close_price"    description:"上一个收盘价"`  // 上一个收盘价
	PriceChangePercentage float64     `json:"priceChangePercentage" orm:"price_change_percentage" description:"价格变化百分比"` // 价格变化百分比
	Volume                int64       `json:"volume"                orm:"volume"                  description:"成交量"`     // 成交量
	Turnover              float64     `json:"turnover"              orm:"turnover"                description:""`        //
	UpdateTime            *gtime.Time `json:"updateTime"            orm:"update_time"             description:"更新时间"`    // 更新时间
	CreateTime            *gtime.Time `json:"createTime"            orm:"create_time"             description:"创建时间"`    // 创建时间
}
