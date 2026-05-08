// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PyHt is the golang structure of table py_ht for DAO operations like Where/Data.
type PyHt struct {
	g.Meta          `orm:"table:py_ht, do:true"`
	Id              interface{} //
	TradeTime       *gtime.Time // 交易日期
	SecurityName    interface{} // 基金名  / sɪˈkjʊrəti /
	SecurityId      interface{} // 基金代码
	EntrustTypeName interface{} // 委托类型名 / ɪnˈtrʌst /
	EntrustType     interface{} // 委托类型 1买入 2卖出 3红利 4手工冻结
	Volume          interface{} // 总量
	Price           interface{} // 单价
	Turnover        interface{} // 交易额   / tɜːrnoʊvər /
	TradeId         interface{} // 交易id
	ExchangeName    interface{} // 交易所名称
	OrderSerialId   interface{} // 订单序列号id  / ˈsɪriəl /
	ShareholderId   interface{} // 股东id
	UpdateTime      *gtime.Time // 更新时间
	CreateTime      *gtime.Time // 创建时间
}
