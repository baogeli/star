package stock_risk_return

import (
	"context"
	"github.com/spf13/cast"
	"star/internal/dao"
	"star/internal/model/entity"
)

//type PyHad struct {
//	Id          int     `json:"id"          orm:"id"           description:""` //
//	Stocks      int     `json:"stocks"      orm:"stocks"       description:""` //
//	FundName    string  `json:"fundName"    orm:"fund_name"    description:""` //
//	FundCode    string  `json:"fundCode"    orm:"fund_code"    description:""` //
//	BuyingPrice float64 `json:"buyingPrice" orm:"buying_price" description:""` //
//	Exchange    string  `json:"exchange"    orm:"exchange"     description:""` //
//}

var abc entity.PyHad

// GetCumulativeReturn 累计收益
func GetCumulativeReturn(ctx context.Context) float64 {
	var hadFund []*entity.PyHad

	err := dao.PyHad.Ctx(ctx).Scan(&hadFund)
	if err != nil {
		return 0
	}

	var totalMarketValue float64
	for _, h := range hadFund {
		var fundData *entity.PyFund
		dao.PyFund.Ctx(ctx).Where("fund_code", h.FundCode).OrderDesc("create_time").Limit().Scan(&fundData)
		//g.Dump(fundData)
		currentPrice := fundData.CurrentPrice * cast.ToFloat64(h.Stocks)
		//fmt.Printf("日期：%s, 基金代码:%s, 持有数:%d,当天价格:%f, 总价:%f \n", gtime.Now().Format("Y-m-d"), h.FundCode, h.Stocks, fundData.CurrentPrice, currentPrice)
		totalMarketValue += currentPrice
		//g.Dump(totalMarketValue)

	}
	remainingCash := 4204.59
	// 计算净资产
	netAssets := remainingCash + totalMarketValue

	//fmt.Println("netAssets === ", netAssets)

	// 计算当前净值
	netValue := netAssets / 100000.0
	// 累计收益 = ( 当前净值 - 期初净值 ) / 期初净值
	cumulativeReturn := (netValue - 1.0) / 1.0
	return float64(int(cumulativeReturn*1000000)) / 1000000
}
