// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PyHadDao is the data access object for the table py_had.
type PyHadDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PyHadColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PyHadColumns defines and stores column names for the table py_had.
type PyHadColumns struct {
	Id          string //
	Stocks      string //
	FundName    string //
	FundCode    string //
	BuyingPrice string //
	Exchange    string //
}

// pyHadColumns holds the columns for the table py_had.
var pyHadColumns = PyHadColumns{
	Id:          "id",
	Stocks:      "stocks",
	FundName:    "fund_name",
	FundCode:    "fund_code",
	BuyingPrice: "buying_price",
	Exchange:    "exchange",
}

// NewPyHadDao creates and returns a new DAO object for table data access.
func NewPyHadDao(handlers ...gdb.ModelHandler) *PyHadDao {
	return &PyHadDao{
		group:    "default",
		table:    "py_had",
		columns:  pyHadColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PyHadDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PyHadDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PyHadDao) Columns() PyHadColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PyHadDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PyHadDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PyHadDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
