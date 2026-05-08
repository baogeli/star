// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PyHtDao is the data access object for the table py_ht.
type PyHtDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PyHtColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PyHtColumns defines and stores column names for the table py_ht.
type PyHtColumns struct {
	Id              string //
	TradeTime       string // 交易日期
	SecurityName    string // 基金名  / sɪˈkjʊrəti /
	SecurityId      string // 基金代码
	EntrustTypeName string // 委托类型名 / ɪnˈtrʌst /
	EntrustType     string // 委托类型 1买入 2卖出 3红利 4手工冻结
	Volume          string // 总量
	Price           string // 单价
	Turnover        string // 交易额   / tɜːrnoʊvər /
	TradeId         string // 交易id
	ExchangeName    string // 交易所名称
	OrderSerialId   string // 订单序列号id  / ˈsɪriəl /
	ShareholderId   string // 股东id
	UpdateTime      string // 更新时间
	CreateTime      string // 创建时间
}

// pyHtColumns holds the columns for the table py_ht.
var pyHtColumns = PyHtColumns{
	Id:              "id",
	TradeTime:       "trade_time",
	SecurityName:    "security_name",
	SecurityId:      "security_id",
	EntrustTypeName: "entrust_type_name",
	EntrustType:     "entrust_type",
	Volume:          "volume",
	Price:           "price",
	Turnover:        "turnover",
	TradeId:         "trade_id",
	ExchangeName:    "exchange_name",
	OrderSerialId:   "order_serial_id",
	ShareholderId:   "shareholder_id",
	UpdateTime:      "update_time",
	CreateTime:      "create_time",
}

// NewPyHtDao creates and returns a new DAO object for table data access.
func NewPyHtDao(handlers ...gdb.ModelHandler) *PyHtDao {
	return &PyHtDao{
		group:    "default",
		table:    "py_ht",
		columns:  pyHtColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PyHtDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PyHtDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PyHtDao) Columns() PyHtColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PyHtDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PyHtDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *PyHtDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
