// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PyHad is the golang structure for table py_had.
type PyHad struct {
	Id          int     `json:"id"          orm:"id"           description:""` //
	Stocks      int     `json:"stocks"      orm:"stocks"       description:""` //
	FundName    string  `json:"fundName"    orm:"fund_name"    description:""` //
	FundCode    string  `json:"fundCode"    orm:"fund_code"    description:""` //
	BuyingPrice float64 `json:"buyingPrice" orm:"buying_price" description:""` //
	Exchange    string  `json:"exchange"    orm:"exchange"     description:""` //
}
