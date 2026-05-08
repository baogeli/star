// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PyHt is the golang structure for table py_ht.
type PyHt struct {
	Id              int         `json:"id"              orm:"id"                description:""`                       //
	TradeTime       *gtime.Time `json:"tradeTime"       orm:"trade_time"        description:"交易日期"`                   // 交易日期
	SecurityName    string      `json:"securityName"    orm:"security_name"     description:"基金名  / sɪˈkjʊrəti /"`    // 基金名  / sɪˈkjʊrəti /
	SecurityId      int         `json:"securityId"      orm:"security_id"       description:"基金代码"`                   // 基金代码
	EntrustTypeName string      `json:"entrustTypeName" orm:"entrust_type_name" description:"委托类型名 / ɪnˈtrʌst /"`     // 委托类型名 / ɪnˈtrʌst /
	EntrustType     int         `json:"entrustType"     orm:"entrust_type"      description:"委托类型 1买入 2卖出 3红利 4手工冻结"` // 委托类型 1买入 2卖出 3红利 4手工冻结
	Volume          int         `json:"volume"          orm:"volume"            description:"总量"`                     // 总量
	Price           float64     `json:"price"           orm:"price"             description:"单价"`                     // 单价
	Turnover        float64     `json:"turnover"        orm:"turnover"          description:"交易额   / tɜːrnoʊvər /"`   // 交易额   / tɜːrnoʊvər /
	TradeId         string      `json:"tradeId"         orm:"trade_id"          description:"交易id"`                   // 交易id
	ExchangeName    string      `json:"exchangeName"    orm:"exchange_name"     description:"交易所名称"`                  // 交易所名称
	OrderSerialId   string      `json:"orderSerialId"   orm:"order_serial_id"   description:"订单序列号id  / ˈsɪriəl /"`   // 订单序列号id  / ˈsɪriəl /
	ShareholderId   string      `json:"shareholderId"   orm:"shareholder_id"    description:"股东id"`                   // 股东id
	UpdateTime      *gtime.Time `json:"updateTime"      orm:"update_time"       description:"更新时间"`                   // 更新时间
	CreateTime      *gtime.Time `json:"createTime"      orm:"create_time"       description:"创建时间"`                   // 创建时间
}
