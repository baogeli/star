// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PyT is the golang structure of table py_t for DAO operations like Where/Data.
type PyT struct {
	g.Meta                `orm:"table:py_t, do:true"`
	Id                    interface{} //
	FundCode              interface{} // 基金代码
	FundName              interface{} // 基金名称
	Exchange              interface{} // 交易所
	FundCategory          interface{} // 资金类别
	OpeningPrice          interface{} // 开盘价
	HighestPrice          interface{} // 最高价
	LowestPrice           interface{} // 最低价
	CurrentPrice          interface{} // 当前价
	PreviousClosePrice    interface{} // 上一个收盘价
	PriceChangePercentage interface{} // 价格变化百分比
	Volume                interface{} // 成交量
	Turnover              interface{} //
	UpdateTime            *gtime.Time // 更新时间
	CreateTime            *gtime.Time // 创建时间
}
