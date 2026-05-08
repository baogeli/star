package stock_risk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/os/gtime"
	"os"
	"star/internal/dao"
)

type PyFund struct {
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

func GetBeta(ctx context.Context) float64 {
	var allData []*PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合日收益
		"932047", // benchMark 日收益
	}

	// 为每只基金单独查询 31 天数据
	for _, code := range fundCodes {
		var fundData []*PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			//Limit(31).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			//fmt.Printf("查询基金 %s 数据失败：%v\n", code, err)
			continue
		}
		allData = append(allData, fundData...)
	}

	// 按基金代码分组数据
	fundDataMap := make(map[string][]*PyFund)
	for _, item := range allData {
		fundDataMap[item.FundCode] = append(fundDataMap[item.FundCode], item)
	}

	// 打印每只基金的数据量
	//for code, funds := range fundDataMap {
	//	fmt.Printf("基金 %s: %d 条数据\n", code, len(funds))
	//}

	// 计算组合日收益率（4 只基金的平均收益）
	portfolioReturns := calculatePortfolioDailyReturns(fundDataMap, []string{"508008", "508033", "508086", "180201"})

	// 计算基准日收益率
	benchmarkReturns := calculateDailyReturns(fundDataMap["932047"])

	//g.Dump(fundDataMap["932047"])

	//dates := make([]string, 0, len(fundDataMap["932047"]))
	//datePrices := make(map[string]float64)
	//for _, v := range fundDataMap["932047"] {
	//	dates = append(dates, v.CreateTime.String()[:10])
	//	datePrices[v.CreateTime.String()[:10]] = v.CurrentPrice
	//}
	//sort.Strings(dates)
	//for i := 1; i < len(dates); i++ {
	//	currentDate := dates[i]
	//	prevDate := dates[i-1]
	//	currentPrice := datePrices[currentDate]
	//	prevPrice := datePrices[prevDate]
	//	dailyReturn := (currentPrice - prevPrice) / prevPrice
	//	glog.Infof(ctx, "基金:%s 日期:%s 当日收盘价:%.3f 前日收盘价:%.3f 价格变化:%.3f →日收益率:%.10f",
	//		"中证REITs全收益", currentDate, currentPrice, prevPrice, currentPrice-prevPrice, dailyReturn)
	//}

	// 计算 benchmarkReturns 的平均值
	//benchmarkAvgReturn := calculateAverageReturn(benchmarkReturns)
	//fmt.Printf("benchmark avg === %.8f\n", benchmarkAvgReturn)
	//fmt.Println("avg === ", benchmarkAvgReturn)

	// 计算 Beta 系数
	beta := calculateBeta(portfolioReturns, benchmarkReturns)
	//fmt.Printf("组合 Beta: %.4f\n", beta)

	// 导出数据到 JSON 文件供手动比对
	exportCalculationData(fundDataMap, portfolioReturns, benchmarkReturns, beta)

	// 返回 Beta，保留 4 位小数
	return float64(int(beta*10000)) / 10000
}

// calculateDailyReturns 计算单只基金的日收益率
func calculateDailyReturns(prices []*PyFund) []float64 {
	if len(prices) < 2 {
		return []float64{}
	}

	returns := make([]float64, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		// 日收益率 = (今日收盘价 - 昨日收盘价) / 昨日收盘价
		ret := (prices[i-1].CurrentPrice - prices[i].CurrentPrice) / prices[i].CurrentPrice
		returns = append(returns, ret)
	}
	return returns
}

// calculateAverageReturn 计算收益率的平均值
func calculateAverageReturn(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	var sum float64
	for _, ret := range returns {
		sum += ret
		//fmt.Println("sum add === ", sum)
	}
	return sum / float64(len(returns))
}

// calculatePortfolioDailyReturns 计算组合的日收益率（多只基金的平均收益）
func calculatePortfolioDailyReturns(fundDataMap map[string][]*PyFund, fundCodes []string) []float64 {
	if len(fundCodes) == 0 {
		return []float64{}
	}

	// 获取第一只基金的日收益作为基础
	firstFundReturns := calculateDailyReturns(fundDataMap[fundCodes[0]])
	if len(firstFundReturns) == 0 {
		return []float64{}
	}

	// 组合日收益 = 各基金日收益的平均值
	portfolioReturns := make([]float64, len(firstFundReturns))
	for i := range portfolioReturns {
		sum := 0.0
		count := 0
		for _, code := range fundCodes {
			fundReturns := calculateDailyReturns(fundDataMap[code])
			if i < len(fundReturns) {
				sum += fundReturns[i]
				count++
			}
		}
		if count > 0 {
			portfolioReturns[i] = sum / float64(count)
		}
	}

	return portfolioReturns
}

// exportCalculationData 导出计算数据到 JSON 文件
func exportCalculationData(fundDataMap map[string][]*PyFund, portfolioReturns, benchmarkReturns []float64, beta float64) {
	// 准备导出数据结构
	type FundDailyData struct {
		Date        string  `json:"date"`
		FundCode    string  `json:"fundCode"`
		Price       float64 `json:"price"`
		DailyReturn float64 `json:"dailyReturn,omitempty"`
	}

	type CalculationData struct {
		PortfolioFundCodes []string          `json:"portfolioFundCodes"` // 组合基金代码
		BenchmarkCode      string            `json:"benchmarkCode"`      // 基准代码
		FundDetails        [][]FundDailyData `json:"fundDetails"`        // 每只基金的每日数据
		PortfolioReturns   []float64         `json:"portfolioReturns"`   // 组合日收益
		BenchmarkReturns   []float64         `json:"benchmarkReturns"`   // 基准日收益
		Beta               float64           `json:"beta"`               // Beta 系数
		TotalDays          int               `json:"totalDays"`          // 总天数
	}

	data := CalculationData{
		PortfolioFundCodes: []string{"508008", "508033", "508086", "180201"},
		BenchmarkCode:      "932047",
		PortfolioReturns:   portfolioReturns,
		BenchmarkReturns:   benchmarkReturns,
		Beta:               beta,
	}

	// 获取日期数量（以组合收益天数为准）
	totalDays := len(portfolioReturns)
	if totalDays > len(benchmarkReturns) {
		totalDays = len(benchmarkReturns)
	}
	data.TotalDays = totalDays

	// 收集每只基金的详细数据（倒序，最新的在前）
	portfolioCodes := []string{"508008", "508033", "508086", "180201"}
	for _, code := range portfolioCodes {
		funds := fundDataMap[code]
		var fundDetail []FundDailyData
		for i := 0; i < len(funds) && i <= totalDays; i++ {
			var dailyReturn float64
			if i > 0 {
				// 计算日收益率
				dailyReturn = (funds[i-1].CurrentPrice - funds[i].CurrentPrice) / funds[i].CurrentPrice
			}
			// 调试：打印 CreateTime 原始值
			//if i < 3 {
			//	fmt.Printf("基金 %s[%d] - CreateTime: %v, IsNil: %v\n",
			//		code, i, funds[i].CreateTime, funds[i].CreateTime == nil)
			//	// 如果 CreateTime 为 nil，尝试从 UpdateTime 提取日期部分
			//	if funds[i].CreateTime == nil {
			//		fmt.Printf("  ⚠️ CreateTime 为空，使用 UpdateTime: %v\n", funds[i].UpdateTime)
			//	}
			//}
			// 安全格式化日期
			dateStr := formatDate(funds[i].CreateTime)
			//if i < 3 {
			//	fmt.Printf("  ✓ 格式化后的日期：%s\n", dateStr)
			//}
			fundDetail = append(fundDetail, FundDailyData{
				Date:        dateStr,
				FundCode:    funds[i].FundCode,
				Price:       funds[i].CurrentPrice,
				DailyReturn: dailyReturn,
			})
		}
		data.FundDetails = append(data.FundDetails, fundDetail)
	}

	// 添加基准数据
	benchmarkFunds := fundDataMap["932047"]
	var benchmarkDetail []FundDailyData
	for i := 0; i < len(benchmarkFunds) && i <= totalDays; i++ {
		var dailyReturn float64
		if i > 0 {
			dailyReturn = (benchmarkFunds[i-1].CurrentPrice - benchmarkFunds[i].CurrentPrice) / benchmarkFunds[i].CurrentPrice
		}
		benchmarkDetail = append(benchmarkDetail, FundDailyData{
			Date:        formatDate(benchmarkFunds[i].CreateTime),
			FundCode:    benchmarkFunds[i].FundCode,
			Price:       benchmarkFunds[i].CurrentPrice,
			DailyReturn: dailyReturn,
		})
	}
	data.FundDetails = append(data.FundDetails, benchmarkDetail)

	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("JSON 序列化失败：%v\n", err)
		return
	}

	// 写入文件
	filePath := "beta_calculation_data.json"
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("写入文件失败：%v\n", err)
		return
	}

	//fmt.Printf("计算数据已导出到：%s\n", filePath)
	//fmt.Printf("总天数：%d 天，Beta: %.4f\n", totalDays, beta)
}

// formatDate 安全格式化日期，处理 nil 情况
func formatDate(t *gtime.Time) string {
	if t == nil {
		return ""
	}
	// gtime.Time 的 String() 方法会返回 "2006-01-02 15:04:05" 格式
	// 我们只需要日期部分
	timeStr := t.String()
	if len(timeStr) >= 10 {
		return timeStr[:10]
	}
	return timeStr
}

// returns: 组合日收益率数组
// benchmarkReturns: 基准日收益率数组
func calculateBeta(returns, benchmarkReturns []float64) float64 {
	if len(returns) != len(benchmarkReturns) || len(returns) < 2 {
		return 0
	}

	// 取最小长度，防止数组长度不一致
	length := len(returns)
	if len(benchmarkReturns) < length {
		length = len(benchmarkReturns)
	}

	// 计算平均收益率
	var sumRp, sumRm float64
	for i := 0; i < length; i++ {
		sumRp += returns[i]
		sumRm += benchmarkReturns[i]
	}

	meanRp := sumRp / float64(length)
	meanRm := sumRm / float64(length)

	//fmt.Printf("mean rp avg === %.8f\n", meanRp)
	//fmt.Printf("mean rm avg === %.8f\n", meanRm)

	// 计算协方差和方差
	var cov, varRm float64
	for i := 0; i < length; i++ {
		cov += (returns[i] - meanRp) * (benchmarkReturns[i] - meanRm)
		varRm += (benchmarkReturns[i] - meanRm) * (benchmarkReturns[i] - meanRm)
	}

	// 防止除以零
	if varRm < 1e-8 {
		return 0
	}

	//fmt.Printf("mean rp avg === %.8f\n", meanRp)
	//fmt.Printf("mean rm avg === %.8f\n", meanRm)
	//fmt.Printf("beta === %.8f\n", cov/varRm)

	// Alpha根据单指数模型
	//alphaDay := meanRp - cov/varRm*meanRm
	//alphaYear := alphaDay * 250

	//fmt.Printf("res === %.8f\n", alphaDay)
	//fmt.Printf("res year  === %.8f\n", alphaYear)

	//Step 1: 计算 β·X̄
	//= 0.729480 × (-0.00004978)
	//= -0.000036314
	//
	//Step 2: 代入公式
	//α_日度 = 0.00001366 - (-0.000036314)
	//= 0.00001366 + 0.000036314
	//= ✅ 0.000049974
	//
	//Step 3: 转回百分比形式（×100%）
	//α_日度 = 0.000049974 × 100%
	//= ✅ +0.0049974% ≈ +0.004998%

	return cov / varRm
}
