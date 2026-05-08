package stock_risk_return

import (
	"context"
	"github.com/spf13/cast"
	"math"
	"star/internal/dao"
	"star/internal/model/entity"
	"time"
)

// GetAnnualizedReturn 年化收益
func GetAnnualizedReturn(ctx context.Context) float64 {

	// 计算年化收益，
	// 1. 先算:期间收益率
	var hadFund []*entity.PyHad
	err := dao.PyHad.Ctx(ctx).Scan(&hadFund)
	if err != nil {
		return 0
	}
	var fundReturn float64
	for _, h := range hadFund {
		var fundData *entity.PyFund
		dao.PyFund.Ctx(ctx).Where("fund_code", h.FundCode).OrderDesc("create_time").Limit().Scan(&fundData)
		d := (fundData.CurrentPrice - h.BuyingPrice) / h.BuyingPrice
		fundReturn += d
	}
	avgFundReturn := fundReturn / float64(len(hadFund))
	// 计算持有天数
	had := getHoldingDays(ctx)
	// 年化收益率 = (1 + 期间收益率)^(365/持有天数) - 1
	annualizedReturn := CalculateAnnualizedReturn(avgFundReturn, had)

	//g.Log().Infof(ctx, "持有天数：%d, 年化收益率：%.4f", had, annualizedReturn)

	return annualizedReturn
}

// getHoldingDays 计算持有天数（从固定成立日期到现在）
func getHoldingDays(ctx context.Context) int {
	// 固定成立日期：2025-12-09
	setupDate := time.Date(2025, 12, 9, 0, 0, 0, 0, time.Local)
	now := time.Now()

	// 计算从成立到现在的天数
	days := int(now.Sub(setupDate).Hours() / 24)

	// 至少为 1 天
	if days < 1 {
		days = 1
	}

	return days
}

// calculateAnnualizedReturn 计算年化收益率
func CalculateAnnualizedReturn(periodReturn float64, holdingDays int) float64 {
	if holdingDays <= 0 {
		return 0
	}

	// 年化收益率公式：(1 + R_period)^(365/days) - 1
	annualizedReturn := 1.0
	exponent := float64(365) / float64(holdingDays)
	//fmt.Println("指数：", exponent)

	// 使用指数和对数计算幂
	// (1 + periodReturn)^exponent = e^(exponent * ln(1 + periodReturn))
	if periodReturn > -1 { // 确保 1+periodReturn > 0
		annualizedReturn = cast.ToFloat64(cast.ToString(float64(1) + periodReturn))
		// 简单处理：如果持有期收益率为正，直接计算
		annualizedReturn = math.Pow(1.0+periodReturn, exponent) - 1.0
	}

	// 保留 4 位小数
	return float64(int(annualizedReturn*1000000)) / 1000000
}

// pow 计算幂运算
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	// 处理小数部分（简化处理）
	if exp-float64(int(exp)) > 0 {
		// 使用近似计算
		frac := exp - float64(int(exp))
		result *= 1.0 + frac*(base-1.0)
	}
	return result
}
