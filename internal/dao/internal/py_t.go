// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PyTDao is the data access object for the table py_t.
type PyTDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PyTColumns         // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PyTColumns defines and stores column names for the table py_t.
type PyTColumns struct {
	Id                    string //
	FundCode              string // 基金代码
	FundName              string // 基金名称
	Exchange              string // 交易所
	FundCategory          string // 资金类别
	OpeningPrice          string // 开盘价
	HighestPrice          string // 最高价
	LowestPrice           string // 最低价
	CurrentPrice          string // 当前价
	PreviousClosePrice    string // 上一个收盘价
	PriceChangePercentage string // 价格变化百分比
	Volume                string // 成交量
	Turnover              string //
	UpdateTime            string // 更新时间
	CreateTime            string // 创建时间
}

// pyTColumns holds the columns for the table py_t.
var pyTColumns = PyTColumns{
	Id:                    "id",
	FundCode:              "fund_code",
	FundName:              "fund_name",
	Exchange:              "exchange",
	FundCategory:          "fund_category",
	OpeningPrice:          "opening_price",
	HighestPrice:          "highest_price",
	LowestPrice:           "lowest_price",
	CurrentPrice:          "current_price",
	PreviousClosePrice:    "previous_close_price",
	PriceChangePercentage: "price_change_percentage",
	Volume:                "volume",
	Turnover:              "turnover",
	UpdateTime:            "update_time",
	CreateTime:            "create_time",
}

// NewPyTDao creates and returns a new DAO object for table data access.
func NewPyTDao(handlers ...gdb.ModelHandler) *PyTDao {
	return &PyTDao{
		group:    "default",
		table:    "py_t",
		columns:  pyTColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PyTDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PyTDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PyTDao) Columns() PyTColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PyTDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PyTDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PyTDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
