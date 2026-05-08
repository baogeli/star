package stock_risk_return

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/spf13/cast"
	"math"
	"sort"
	"star/internal/dao"
	"star/internal/model/entity"
	"strings"
)

func SharpeRatio(ctx context.Context, Rp, Rf float64) (sharpeRatio float64) {
	/*
		夏普比率越高 → 风险调整后收益越好
		Sharpe Ratio = (Rp - Rf) / σp
		Rp 投资组合收益率 实际获得的回报率
		Rf 无风险利率 通常用国债收益率（如10年期）
		Rp - Rf 超额收益 超过无风险收益的部分
		σp 组合标准差 衡量总风险（波动率）

		σp 基金年化波动率
		输入数据（30个交易日组合收益率） [0.00477, 0.00280, -0.00128, -0.00088, -0.00221, ...]  # 共30个
		步骤1：计算日均收益
			R̄ = ΣRt / n = 0.000604 = 0.0604%/日
		步骤2：计算日波动率（标准差）
			σ_daily = √[ Σ(Rt - 0.000604)² / 29 ]
		        	= 0.001741 = 0.1741%/日
		步骤3：年化波动率
			σ_annual = 0.001741 × √252
					 = 0.001741 × 15.8745
					 = 0.02764
					 = 2.76%（年化）
	*/

	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合日收益
		//"932047", // benchMark 日收益
	}

	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	//g.Dump(allData)
	//g.Dump("查询到的数据总数:", len(allData))

	// 计算基金组合的年化波动率
	// 假设 4 支基金等权重配置（各 25%）
	portfolioDailyReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioDailyReturns) == 0 {
		return 0
	}

	// 计算组合日收益率的标准差
	portfolioStd := calculateStandardDeviation(portfolioDailyReturns)

	sharpeRatio = (Rp - Rf) / portfolioStd

	//fmt.Println("日夏普比率 === ", sharpeRatio)

	// 年化波动率 = 日波动率 × √252
	//annualizedVolatility := portfolioStd * math.Sqrt(252)

	//g.Dump("组合日收益率序列:", portfolioDailyReturns)
	//g.Dump("组合日波动率:", portfolioStd)
	//g.Dump("组合年化波动率:", annualizedVolatility)

	// 夏普比率 = (Rp - Rf) / σp
	//if annualizedVolatility == 0 {
	//	return 0
	//}
	//sharpeRatio = (Rp - Rf) / annualizedVolatility

	//g.Dump("夏普比率:", sharpeRatio)

	return float64(int(sharpeRatio*10000)) / 10000

	//return sharpeRatio
}

// calculatePortfolioDailyReturns 计算基金组合的每日收益率
// 假设等权重配置
func calculatePortfolioDailyReturns(allData []*entity.PyFund, fundCount int) []float64 {
	if len(allData) == 0 {
		return []float64{}
	}

	// 按日期分组，将每支基金的数据按日期组织
	dateFundMap := make(map[string]map[string]float64)    // date -> fundCode -> dailyReturn
	fundDatePrices := make(map[string]map[string]float64) // fundCode -> date -> price

	// 第一步：先按基金和日期组织价格数据
	for _, data := range allData {
		dateStr := data.CreateTime.String()[:10] // 提取日期部分 YYYY-MM-DD
		if _, ok := fundDatePrices[data.FundCode]; !ok {
			fundDatePrices[data.FundCode] = make(map[string]float64)
		}
		fundDatePrices[data.FundCode][dateStr] = data.CurrentPrice
	}

	//g.Dump("基金价格数据:", fundDatePrices)
	/*
			"888888": {
		    	"2026-01-09": 5.058,
		        "2026-01-06": 5.057,
		        "2025-12-09": 5.152,
				....
				....
		    },
		    "180201": {
		        "2026-03-12": 8.093,
		        "2026-03-10": 8.1,
		        "2026-02-26": 8.112,
				....
				....
	*/

	// 第二步：对每支基金，计算每日的收益率
	for fundCode, datePrices := range fundDatePrices {
		// 获取该基金的所有日期并排序
		dates := make([]string, 0, len(datePrices))
		for date := range datePrices {
			dates = append(dates, date)
		}
		sort.Strings(dates)
		//g.Dump(dates)

		// 计算每日的收益率（从第 2 天开始）
		for i := 1; i < len(dates); i++ {
			currentDate := dates[i]
			prevDate := dates[i-1]
			currentPrice := datePrices[currentDate]
			prevPrice := datePrices[prevDate]

			// 日收益率 = (当日净值 - 前一日净值) / 前一日净值
			dailyReturn := (currentPrice - prevPrice) / prevPrice

			// 存储到日期 - 基金映射中
			if _, ok := dateFundMap[currentDate]; !ok {
				dateFundMap[currentDate] = make(map[string]float64)
			}
			dateFundMap[currentDate][fundCode] = dailyReturn

			// 打印详细日志，保留足够的小数位以便追踪
			//glog.Infof(gctx.GetInitCtx(), "基金:%s 日期:%s 当日收盘价:%.3f 前日收盘价:%.3f 价格变化:%.3f →日收益率:%.10f",
			//	fundCode, currentDate, currentPrice, prevPrice, currentPrice-prevPrice, dailyReturn)

			//glog.Infof(context.Background(), "基金:%s 日期:%s 当日净值:%.3f 前日净值:%.3f 价格变化:%.3f →日收益率:%.10f (%.10e)",
			//	fundCode, currentDate, currentPrice, prevPrice, currentPrice-prevPrice, dailyReturn, dailyReturn)
		}
	}

	// 获取所有日期并排序
	dates := make([]string, 0, len(dateFundMap))
	for date := range dateFundMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// g.Dump("总天数:", len(dates))

	// 计算每日的组合收益率（等权重平均）
	portfolioReturns := make([]float64, 0, len(dates))

	for _, date := range dates {
		fundReturns := dateFundMap[date]
		//g.Dump("第", i+1, "天 - 日期:", date, "基金数量:", len(fundReturns), "收益率:", fundReturns)
		//fmt.Println("日期:", date, "基金数量:", len(fundReturns), "收益率:", fundReturns)
		if len(fundReturns) == fundCount {
			// 所有基金都有数据，计算等权重组合收益
			var sumReturn float64
			for _, ret := range fundReturns {
				sumReturn += ret
			}
			portfolioReturn := sumReturn / float64(fundCount)
			portfolioReturns = append(portfolioReturns, portfolioReturn)
		} else {
			g.Dump("⚠️ 跳过此天（基金数据不完整）")
		}
	}

	return portfolioReturns
}

// calculateStandardDeviation 计算标准差
func calculateStandardDeviation(returns []float64) float64 {
	if len(returns) <= 1 {
		return 0
	}

	// 计算均值
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// 计算方差
	var variance float64
	for _, ret := range returns {
		diff := ret - mean
		variance += diff * diff
	}
	variance = variance / float64(len(returns)-1) // 样本标准差，除以 n-1

	// 标准差
	return math.Sqrt(variance)
}

/* ========================================= */

// GetInformationRatio 获取信息比率
// IR = (Rp - Rb) / σ_tracking_error
// Rp: 投资组合年化收益率
// Rb: 基准年化收益率
// σ_tracking_error: 跟踪误差（超额收益的标准差）
func GetInformationRatio(ctx context.Context, Rp, Rb float64) (informationRatio float64) {
	// 获取组合和基准的日收益率序列，计算跟踪误差
	trackingError := calculateTrackingError(ctx)

	if trackingError == 0 {
		return 0
	}

	// 信息比率 = (组合年化收益 - 基准年化收益) / 跟踪误差
	informationRatio = (Rp - Rb) / trackingError

	//g.Dump("信息比率:", informationRatio, "Rp:", Rp, "Rb:", Rb, "跟踪误差:", trackingError)
	return float64(int(informationRatio*10000)) / 10000

	//return informationRatio
}

// calculateTrackingError 计算跟踪误差（超额收益的年化标准差）
func calculateTrackingError(ctx context.Context) float64 {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
		//"932047", // 基准（中证 REITs 指数）
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	// 计算基准每日收益率
	var benchMarkData []*entity.PyFund
	err := dao.PyFund.Ctx(ctx).
		Where(dao.PyFund.Columns().FundCode, "932047").
		OrderDesc("create_time").Scan(&benchMarkData)
	if err != nil {
		g.Log().Infof(ctx, "获取基准数据失败: %v", err)
	}
	benchmarkReturns := calculateBenchmarkDailyReturns(benchMarkData)

	// 确保两个序列长度一致
	minLen := len(portfolioReturns)
	if len(benchmarkReturns) < minLen {
		minLen = len(benchmarkReturns)
	}

	if minLen <= 1 {
		return 0
	}

	// 计算每日超额收益（组合收益 - 基准收益）
	excessReturns := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		excessReturns[i] = portfolioReturns[i] - benchmarkReturns[i]
	}

	// 计算超额收益的日标准差
	excessStd := calculateStandardDeviation(excessReturns)

	// 年化处理：日标准差 × √252
	trackingError := excessStd * math.Sqrt(252)

	//g.Dump("跟踪误差 (年化):", trackingError, "超额收益日标准差:", excessStd, "样本数:", minLen)

	return trackingError
}

// calculateBenchmarkDailyReturns 计算基准指数的日收益率
func calculateBenchmarkDailyReturns(benchmarkData []*entity.PyFund) []float64 {
	if len(benchmarkData) == 0 {
		return []float64{}
	}

	// 按日期排序
	dates := make([]string, 0, len(benchmarkData))
	datePriceMap := make(map[string]float64)

	for _, data := range benchmarkData {
		dateStr := data.CreateTime.String()[:10]
		dates = append(dates, dateStr)
		datePriceMap[dateStr] = data.CurrentPrice
	}

	sort.Strings(dates)

	// 计算日收益率
	returns := make([]float64, 0, len(dates)-1)
	for i := 1; i < len(dates); i++ {
		currentDate := dates[i]
		prevDate := dates[i-1]
		currentPrice := datePriceMap[currentDate]
		prevPrice := datePriceMap[prevDate]

		if prevPrice == 0 {
			continue
		}

		dailyReturn := (currentPrice - prevPrice) / prevPrice
		returns = append(returns, dailyReturn)
	}
	return returns
}

// GetSortinoRatio 获取索提诺比率
// Sortino Ratio = (Rp - Rf) / σ_downside
// Rp: 投资组合年化收益率
// Rf: 无风险利率
// σ_downside: 下行标准差（只考虑负收益的波动率）
func GetSortinoRatio(ctx context.Context, Rp, Rf float64) (sortinoRatio float64) {
	// 获取组合日收益率序列，计算下行标准差
	downsideDeviation := calculateDownsideDeviation(ctx)
	//g.Log().Infof(ctx, "下行偏差: %v", downsideDeviation)

	if downsideDeviation == 0 {
		return 0
	}

	// 索提诺比率 = (组合年化收益 - 无风险利率) / 下行标准差
	sortinoRatio = (Rp - Rf) / downsideDeviation

	//g.Dump("索提诺比率:", sortinoRatio, "Rp:", Rp, "Rf:", Rf, "下行偏差:", downsideDeviation)

	return float64(int(sortinoRatio*10000)) / 10000

	//return sortinoRatio
}

// calculateDownsideDeviation 计算下行偏差（下行标准差的年化值）
func calculateDownsideDeviation(ctx context.Context) float64 {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) <= 1 {
		return 0
	}

	// 计算下行偏差（只考虑负收益）
	// 假设目标收益率为 0（即只关心亏损的情况）
	targetReturn := 0.0
	var sumSquaredDownside float64
	downsideCount := 0

	for _, ret := range portfolioReturns {
		if ret < targetReturn {
			// 只有当收益率低于目标时，才计入下行偏差
			downsideDiff := ret - targetReturn
			sumSquaredDownside += downsideDiff * downsideDiff
			downsideCount++
		}
	}

	if downsideCount == 0 {
		// 如果没有负收益，下行偏差为 0
		return 0
	}

	// 计算下行标准差（日度）
	// 注意：这里除以的是总样本数 n，而不是 n-1，因为这是半方差
	downsideStdDaily := math.Sqrt(sumSquaredDownside / float64(len(portfolioReturns)))

	// 年化处理：日下行标准差 × √252
	downsideDeviation := downsideStdDaily * math.Sqrt(252)

	//g.Dump("下行偏差 (年化):", downsideDeviation, "下行标准差 (日):", downsideStdDaily,
	//	"负收益天数:", downsideCount, "总天数:", len(portfolioReturns))
	return float64(int(downsideDeviation*10000)) / 10000
	//return downsideDeviation
}

// GetTreynorRatio 获取特雷诺比率
// Treynor Ratio = (Rp - Rf) / β
// Rp: 投资组合年化收益率
// Rf: 无风险利率
// β: 贝塔系数（系统性风险）
func GetTreynorRatio(ctx context.Context, Rp, Rf, Beta float64) (treynorRatio float64) {
	// 避免除以零
	if Beta == 0 {
		return 0
	}

	// 特雷诺比率 = (组合年化收益 - 无风险利率) / 贝塔系数
	treynorRatio = (Rp - Rf) / Beta

	//g.Dump("特雷诺比率:", treynorRatio, "Rp:", Rp, "Rf:", Rf, "Beta:", Beta)

	// 保留 4 位小数
	return float64(int(treynorRatio*10000)) / 10000
}

// GetWinRate 获取胜率
// 胜率 = 正收益周期数 / 总周期数
func GetWinRate(ctx context.Context) (winRate float64) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) == 0 {
		return 0
	}

	// 统计正收益天数
	positiveDays := 0
	for _, ret := range portfolioReturns {
		if ret > 0 {
			positiveDays++
		}
	}

	// 计算胜率
	winRate = float64(positiveDays) / float64(len(portfolioReturns))

	//g.Dump("胜率:", winRate, "正收益天数:", positiveDays, "总天数:", len(portfolioReturns))

	// 保留 4 位小数
	return float64(int(winRate*10000)) / 10000
}

// GetPositivePeriods 获取正收益周期数
func GetPositivePeriods(ctx context.Context) (positivePeriods int) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) == 0 {
		return 0
	}

	// 统计正收益天数
	positivePeriods = 0
	for _, ret := range portfolioReturns {
		if ret > 0 {
			positivePeriods++
		}
	}

	//g.Dump("正收益周期数:", positivePeriods, "总周期数:", len(portfolioReturns))

	return positivePeriods
}

// GetTotalPeriods 获取总周期数
func GetTotalPeriods(ctx context.Context) (totalPeriods int) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	// 总周期数就是收益率序列的长度
	totalPeriods = len(portfolioReturns)

	//g.Dump("总周期数:", totalPeriods)

	return totalPeriods
}

// GetAnnualizedVol 获取年化波动率
// σ_annual = σ_daily × √252
func GetAnnualizedVol(ctx context.Context) (annualizedVol float64) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) <= 1 {
		return 0
	}

	// 计算日收益率的标准差
	dailyStd := calculateStandardDeviation(portfolioReturns)

	//g.Dump("样本数:", len(portfolioReturns))
	//g.Dump("日波动率 (标准差):", dailyStd)
	//g.Dump("日波动率百分比:", fmt.Sprintf("%.6f%%", dailyStd*100))

	// 年化波动率 = 日标准差 × √252
	annualizedVol = dailyStd * math.Sqrt(252)

	//g.Dump("=== GetAnnualizedVol ===")
	//g.Dump("样本数:", len(portfolioReturns))
	//g.Dump("日波动率:", dailyStd, "=>", fmt.Sprintf("%.6f%%", dailyStd*100))
	//g.Dump("√252:", math.Sqrt(252))
	//g.Dump("年化波动率 (√252):", annualizedVol, "=>", fmt.Sprintf("%.6f%%", annualizedVol*100))
	//g.Dump("年化波动率 (保留 4 位后):", float64(int(annualizedVol*10000))/10000, "=>", fmt.Sprintf("%.6f%%", float64(int(annualizedVol*10000))/10000*100))

	// 保留 4 位小数
	return float64(int(annualizedVol*10000)) / 10000
}

// GetTrackingError 获取跟踪误差
// TE = Stdev(Rp - Rb)
// Rp: 组合日收益率序列
// Rb: 基准日收益率序列
// 跟踪误差是超额收益的标准差（年化后）
func GetTrackingError(ctx context.Context) (trackingError float64) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
		//"932047", // 基准（中证 REITs 指数）
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	var benchMarkData []*entity.PyFund
	err := dao.PyFund.Ctx(ctx).
		Where(dao.PyFund.Columns().FundCode, "932047").
		OrderDesc("create_time").Scan(&benchMarkData)
	if err != nil {
		g.Log().Infof(ctx, "查询基准数据失败: %v", err)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	// 计算基准每日收益率
	benchmarkReturns := calculateBenchmarkDailyReturns(benchMarkData)

	// 确保两个序列长度一致
	minLen := len(portfolioReturns)
	if len(benchmarkReturns) < minLen {
		minLen = len(benchmarkReturns)
	}

	if minLen <= 1 {
		return 0
	}

	// 计算每日超额收益（组合收益 - 基准收益）
	excessReturns := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		excessReturns[i] = portfolioReturns[i] - benchmarkReturns[i]
	}

	// 计算超额收益的日标准差
	excessStd := calculateStandardDeviation(excessReturns)

	// 年化处理：日标准差 × √252
	trackingError = excessStd * math.Sqrt(252)

	//g.Dump("跟踪误差:", trackingError, "超额收益日标准差:", excessStd, "样本数:", minLen)

	// 保留 4 位小数
	return float64(int(trackingError*10000)) / 10000
}

// GetDownsideRisk 获取下行风险
// 下行风险 = 下行偏差 × √252
// 与索提诺比率中的下行偏差计算类似，但这里只返回下行风险值
func GetDownsideRisk(ctx context.Context) (downsideRisk float64) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) <= 1 {
		return 0
	}

	// 计算下行偏差（只考虑负收益）
	// 假设目标收益率为 0（即只关心亏损的情况）
	targetReturn := 0.0
	var sumSquaredDownside float64
	downsideCount := 0

	for _, ret := range portfolioReturns {
		if ret < targetReturn {
			// 只有当收益率低于目标时，才计入下行偏差
			downsideDiff := ret - targetReturn
			sumSquaredDownside += downsideDiff * downsideDiff
			downsideCount++
		}
	}

	if downsideCount == 0 {
		// 如果没有负收益，下行风险为 0
		return 0
	}

	// 计算下行标准差（日度）
	// 注意：这里除以的是总样本数 n，而不是 n-1，因为这是半方差
	downsideStdDaily := math.Sqrt(sumSquaredDownside / float64(len(portfolioReturns)))

	// 年化处理：日下行标准差 × √252
	downsideRisk = downsideStdDaily * math.Sqrt(252)

	//g.Dump("下行风险:", downsideRisk, "下行标准差 (日):", downsideStdDaily,
	//	"负收益天数:", downsideCount, "总天数:", len(portfolioReturns))

	// 保留 4 位小数
	return float64(int(downsideRisk*10000)) / 10000
}

// GetVaR 在险价值
// 历史模拟法（Historical Simulation）
// 置信水平默认 95%，即在 95% 的把握下，最大损失不会超过 VaR 值
func GetVaR(ctx context.Context, confidenceLevel float64) (varValue float64) {
	// 默认置信水平为 95%
	if confidenceLevel <= 0 || confidenceLevel >= 1 {
		confidenceLevel = 0.95
	}

	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) == 0 {
		return 0
	}

	// 历史模拟法步骤：
	// 1. 将收益率从小到大排序
	// 2. 找到对应置信水平的分位数
	// 3. VaR = -分位数（取负数，因为 VaR 表示为正数）

	// 排序
	sort.Float64s(portfolioReturns)

	// 计算分位数位置
	// 例如 95% 置信度，找第 5 百分位（0.05 分位数）
	percentile := 1 - confidenceLevel
	index := int(percentile * float64(len(portfolioReturns)))

	// 确保索引不越界
	if index < 0 {
		index = 0
	}
	if index >= len(portfolioReturns) {
		index = len(portfolioReturns) - 1
	}

	// 获取 VaR 值（取负数，转换为正数表示）
	varValue = -portfolioReturns[index]

	// 如果 VaR 是负数，说明在最好的情况下也是赚钱的，此时 VaR 设为 0
	if varValue < 0 {
		varValue = 0
	}

	//g.Dump("VaR:", varValue, "置信水平:", confidenceLevel*100, "%",
	//	"分位数位置:", index, "总样本数:", len(portfolioReturns),
	//	"对应收益率:", portfolioReturns[index])

	// 保留 4 位小数
	return float64(int(varValue*10000)) / 10000
}

// GetMaxDrawdown 最大回撤
// MDD = min((Pt - Ppeak) / Ppeak)
// Pt: 当前净值
// Ppeak: 之前的最高净值
// 最大回撤表示从历史最高点下跌的最大幅度
func GetMaxDrawdown(ctx context.Context) (maxDrawdown float64) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	if len(portfolioReturns) == 0 {
		return 0
	}

	// 方法 2：通过累计收益计算净值曲线，然后计算最大回撤
	// 假设初始净值为 1
	initialNav := 1.0
	nav := initialNav
	peakNav := initialNav
	maxDrawdown = 0.0

	for _, ret := range portfolioReturns {
		// 更新当前净值
		nav = nav * (1 + ret)

		// 更新历史最高净值
		if nav > peakNav {
			peakNav = nav
		}

		// 计算当前回撤
		currentDrawdown := (nav - peakNav) / peakNav

		// 更新最大回撤
		if currentDrawdown < maxDrawdown {
			maxDrawdown = currentDrawdown
		}
	}

	// 转换为正数表示（通常回撤用正数表示）
	maxDrawdown = -maxDrawdown
	//g.Dump("最大回撤:", maxDrawdown, "最终净值:", nav, "历史最高净值:", peakNav)
	// 保留 4 位小数
	return float64(int(maxDrawdown*10000)) / 10000
}

// GetDrawdownFormation 回撤形成期（Drawdown Formation Period）
// 指从最高点下跌到最大回撤位置所经历的天数
// 用于衡量回撤形成的时间长度
func GetDrawdownFormation(ctx context.Context) (formationDays int) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	//g.Dump("查询到的数据总数:", len(allData), "收益率序列长度:", len(portfolioReturns))

	if len(portfolioReturns) == 0 {
		//g.Dump("⚠️ 收益率序列为空，返回 0")
		return 0
	}

	// 假设初始净值为 1
	initialNav := 1.0
	nav := initialNav
	//peakNav := initialNav
	maxDrawdown := 0.0

	// 记录最大回撤发生时的相关信息
	var maxDrawdownPeakDay int // 最大回撤的起点（最高点）
	var troughDay int          // 最大回撤的最低点
	currentDay := 0

	// 临时变量，用于追踪当前回撤
	currentPeakNav := initialNav
	currentPeakDay := 1

	for i, ret := range portfolioReturns {
		currentDay = i + 1
		// 更新当前净值
		nav = nav * (1 + ret)

		// 如果创出新高，更新高点
		if nav > currentPeakNav {
			currentPeakNav = nav
			currentPeakDay = currentDay
		}

		// 计算当前回撤（相对于最近的高点）
		currentDrawdown := (nav - currentPeakNav) / currentPeakNav

		// 如果创出最大回撤，记录完整的回撤信息
		if currentDrawdown < maxDrawdown {
			maxDrawdown = currentDrawdown
			maxDrawdownPeakDay = currentPeakDay // 这次回撤的起点
			troughDay = currentDay              // 这次回撤的终点
			//g.Dump("🔴 刷新最大回撤 - 第", currentDay, "天，回撤:", currentDrawdown,
			//	"起点：第", currentPeakDay, "天，形成期:", troughDay-maxDrawdownPeakDay, "天")
		}
	}

	// 回撤形成期 = 最大回撤的起点到终点的天数
	formationDays = troughDay - maxDrawdownPeakDay

	// 确保不为负数
	if formationDays < 0 {
		formationDays = 0
	}

	//g.Dump("回撤形成期:", formationDays, "最大回撤起点:", maxDrawdownPeakDay,
	//	"最低点位置:", troughDay, "最大回撤:", maxDrawdown)

	return formationDays
}

// GetDrawdownRecovery 回撤恢复期（Drawdown Recovery Period）
// 方法 1：从最低点回到最高点的时间 ⭐ 最常用
// 恢复期 = 恢复到前期高点的日期 - 最大回撤最低点的日期
func GetDrawdownRecovery(ctx context.Context) (recoveryDays int) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	//g.Dump("查询到的数据总数:", len(allData), "收益率序列长度:", len(portfolioReturns))

	if len(portfolioReturns) == 0 {
		g.Dump("⚠️ 收益率序列为空，返回 0")
		return 0
	}

	// 假设初始净值为 1
	initialNav := 1.0
	nav := initialNav
	maxDrawdown := 0.0

	// 记录最大回撤发生时的相关信息
	var maxDrawdownPeakDay int // 最大回撤的起点（最高点）
	var troughDay int          // 最大回撤的最低点
	var troughNav float64      // 最低点的净值
	var peakNav float64        // 前期高点的净值
	currentDay := 0

	// 临时变量，用于追踪当前回撤
	currentPeakNav := initialNav
	currentPeakDay := 1

	// 第一步：找到最大回撤的起点、终点和净值
	for i, ret := range portfolioReturns {
		currentDay = i + 1
		// 更新当前净值
		nav = nav * (1 + ret)

		// 如果创出新高，更新高点
		if nav > currentPeakNav {
			currentPeakNav = nav
			currentPeakDay = currentDay
		}

		// 计算当前回撤（相对于最近的高点）
		currentDrawdown := (nav - currentPeakNav) / currentPeakNav

		// 如果创出最大回撤，记录完整的回撤信息
		if currentDrawdown < maxDrawdown {
			maxDrawdown = currentDrawdown
			maxDrawdownPeakDay = currentPeakDay // 这次回撤的起点
			troughDay = currentDay              // 这次回撤的终点
			troughNav = nav                     // 最低点的净值
			peakNav = currentPeakNav            // 需要恢复到的前期高点
		}
	}

	g.Dump("最大回撤信息 - 起点:", maxDrawdownPeakDay, "最低点:", troughDay,
		"最低点净值:", troughNav, "前期高点净值:", peakNav)

	// 如果没有发生回撤（一直涨），返回 0
	if troughDay == 0 {
		//g.Dump("⚠️ 未发生回撤，返回 0")
		return 0
	}

	// 第二步：从最低点之后开始找，直到净值恢复或超过前期高点
	recoveryDay := 0
	for i := troughDay; i <= len(portfolioReturns); i++ {
		// 重新计算第 i 天的净值
		nempNav := initialNav
		for j := 0; j < i; j++ {
			nempNav = nempNav * (1 + portfolioReturns[j])
		}

		// 检查是否已经恢复
		if nempNav >= peakNav {
			recoveryDay = i
			//g.Dump("✅ 已恢复 - 第", recoveryDay, "天，净值:", nempNav)
			break
		}
	}

	// 如果还没恢复
	if recoveryDay == 0 {
		//g.Dump("⚠️ 至今未恢复，最后净值:", nav)
		return 0
	}

	// 恢复期 = 恢复日 - 最低点日
	recoveryDays = recoveryDay - troughDay

	//g.Dump("回撤恢复期:", recoveryDays, "最低点:", troughDay, "恢复日:", recoveryDay,
	//	"形成期:", troughDay-maxDrawdownPeakDay, "完整周期:", recoveryDay-maxDrawdownPeakDay)

	return recoveryDays
}

// GetConsecutiveDrop 连续下降期数（Consecutive Drop）
// 连续下降期数是指资产净值连续下跌的最大交易日数量
func GetConsecutiveDrop(ctx context.Context) (maxConsecutiveDrop int) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	//g.Dump("查询到的数据总数:", len(allData), "收益率序列长度:", len(portfolioReturns))

	if len(portfolioReturns) == 0 {
		g.Dump("⚠️ 收益率序列为空，返回 0")
		return 0
	}

	// 统计连续下跌期数
	currentConsecutiveDrop := 0 // 当前连续下跌天数
	maxConsecutiveDrop = 0      // 最大连续下跌天数
	var maxDropStartDay int     // 最大连续下跌的起始日
	var maxDropEndDay int       // 最大连续下跌的结束日
	currentDropStartDay := 0    // 当前连续下跌的起始日

	for i, ret := range portfolioReturns {
		currentDay := i + 1

		if ret < 0 {
			// 收益率为负，表示下跌
			if currentConsecutiveDrop == 0 {
				// 刚开始下跌，记录起始日
				currentDropStartDay = currentDay
			}
			currentConsecutiveDrop++

			// 更新最大连续下跌天数
			if currentConsecutiveDrop > maxConsecutiveDrop {
				maxConsecutiveDrop = currentConsecutiveDrop
				maxDropStartDay = currentDropStartDay
				maxDropEndDay = currentDay
			}
		} else {
			// 收益率为正或零，下跌中断
			currentConsecutiveDrop = 0
		}
	}

	g.Dump("最大连续下降期数:", maxConsecutiveDrop,
		"起始日:", maxDropStartDay,
		"结束日:", maxDropEndDay)

	return maxConsecutiveDrop
}

// GetRSquare R 方（R Square，决定系数）
// 方法 1：相关系数的平方
// R² = (Correlation)² = ρ²
// 其中 ρ = 组合与基准的相关系数
// R 方表示组合收益的变化中有多少比例可以由基准解释
func GetRSquare(ctx context.Context) (rSquare float64) {
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
		//"932047", // 基准（中证 REITs 指数）
	}

	// 查询所有数据
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderDesc("create_time").Scan(&fundData)
		if err != nil {
			continue
		}
		allData = append(allData, fundData...)
	}

	var benchMarkData []*entity.PyFund
	err := dao.PyFund.Ctx(ctx).
		Where(dao.PyFund.Columns().FundCode, "932047").
		OrderDesc("create_time").Scan(&benchMarkData)
	if err != nil {
		g.Log().Infof(ctx, "⚠️ 查询基准数据失败: %v", err)
	}

	// 计算组合每日收益率
	portfolioReturns := calculatePortfolioDailyReturns(allData, 4)

	// 计算基准每日收益率
	benchmarkReturns := calculateBenchmarkDailyReturns(benchMarkData)

	// 确保两个序列长度一致
	minLen := len(portfolioReturns)
	if len(benchmarkReturns) < minLen {
		minLen = len(benchmarkReturns)
	}

	if minLen <= 1 {
		return 0
	}

	// 截取相同长度的序列
	portfolioReturns = portfolioReturns[:minLen]
	benchmarkReturns = benchmarkReturns[:minLen]

	// 计算相关系数 ρ
	correlation := calculateCorrelation(portfolioReturns, benchmarkReturns)

	// R 方 = 相关系数的平方
	rSquare = correlation * correlation

	//g.Dump("R 方:", rSquare, "相关系数:", correlation, "样本数:", minLen)

	// 保留 4 位小数
	return float64(int(rSquare*10000)) / 10000
}

// calculateCorrelation 计算两个收益率序列的相关系数
func calculateCorrelation(returns1, returns2 []float64) float64 {
	n := len(returns1)
	if n != len(returns2) || n <= 1 {
		return 0
	}

	// 计算均值
	var sum1, sum2 float64
	for i := 0; i < n; i++ {
		sum1 += returns1[i]
		sum2 += returns2[i]
	}
	mean1 := sum1 / float64(n)
	mean2 := sum2 / float64(n)

	// 计算协方差和标准差
	var cov, var1, var2 float64
	for i := 0; i < n; i++ {
		diff1 := returns1[i] - mean1
		diff2 := returns2[i] - mean2
		cov += diff1 * diff2
		var1 += diff1 * diff1
		var2 += diff2 * diff2
	}

	// 相关系数 = 协方差 / (标准差 1 × 标准差 2)
	if var1 == 0 || var2 == 0 {
		return 0
	}

	correlation := cov / math.Sqrt(var1*var2)

	return correlation
}

type Fund struct {
	FundName    string  `json:"fund_name" dc:"基金名"`
	FundCode    string  `json:"fund_code" dc:"基金代码"`
	BuyingPrice float64 `json:"buying_price" dc:"买入价格"`
}

// CalculateDailyPortfolioReturn 计算基金组合每日区间收益
// 投资组合收益率公式详解
/*
        期末净值 - 期初净值 + 期间分红
Rp = ───────────────────────────────── × 100%
            期初净值
*/
func CalculateDailyPortfolioReturn(Rp, Rf float64) float64 {
	ctx := context.Background()
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	/*
		区间收益
		2025-12-09日成立时，当时的购买价来做一个基础点
		再按成立日开始到现在每一个交易日的收盘价格来和12月9号买入价计算区间收益
		再算出基金组合的区间收益 (根据持有数量 这4个基金是否会用到加权平均)
	*/

	// 查询所有数据（按期初时间排序）
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderAsc("create_time").Scan(&fundData)
		if err != nil {
			fmt.Printf("查询基金 %s 数据失败：%v\n", code, err)
			continue
		}
		allData = append(allData, fundData...)
	}

	// 期初的买入价格
	originFunds := getFunds()

	// 构建基金代码到买入价格的映射
	buyingPriceMap := make(map[string]float64)
	for _, fund := range originFunds {
		buyingPriceMap[fund.FundCode] = fund.BuyingPrice
	}

	// 按日期分组，将每支基金的数据按日期组织
	dateFundPrices := make(map[string]map[string]float64) // date -> fundCode -> price

	for _, data := range allData {
		dateStr := data.CreateTime.String()[:10] // 提取日期部分 YYYY-MM-DD
		if _, ok := dateFundPrices[dateStr]; !ok {
			dateFundPrices[dateStr] = make(map[string]float64)
		}
		dateFundPrices[dateStr][data.FundCode] = data.CurrentPrice
	}

	// 获取所有日期并排序
	dates := make([]string, 0, len(dateFundPrices))
	for date := range dateFundPrices {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	fmt.Printf("\n=== 基金组合每日区间收益计算 ===\n")
	fmt.Printf("期初买入价格:\n")
	for _, fund := range originFunds {
		fmt.Printf("  %s (%s): %.3f\n", fund.FundName, fund.FundCode, fund.BuyingPrice)
	}
	fmt.Printf("\n总交易日数：%d\n\n", len(dates))

	// 计算每日的组合区间收益
	portfolioDailyReturns := make([]float64, 0, len(dates))

	for i, date := range dates {
		fundPrices := dateFundPrices[date]

		// 检查是否 4 支基金都有数据
		if len(fundPrices) != 4 {
			fmt.Printf("⚠️  第 %d 天 (%s): 基金数据不完整 (只有 %d 支)\n", i+1, date, len(fundPrices))
			continue
		}

		// 计算每支基金的当日区间收益率
		var fundReturns []float64
		fmt.Printf("第 %d 天 (%s):\n", i+1, date)

		for _, fund := range originFunds {
			currentPrice := fundPrices[fund.FundCode]
			beginningNAV := fund.BuyingPrice

			// 计算该基金从期初到当日的收益率
			// Rp = (当前价格 - 期初价格) / 期初价格 × 100%
			periodReturn := (currentPrice - beginningNAV) / beginningNAV * 100
			fundReturns = append(fundReturns, periodReturn)

			fmt.Printf("  %s: %.3f → %.3f, 收益率：%.2f%%\n",
				fund.FundName, beginningNAV, currentPrice, periodReturn)
		}

		// 计算等权重组合收益（4 支基金各占 25%）
		var sumReturn float64
		for _, ret := range fundReturns {
			sumReturn += ret
		}
		portfolioReturn := sumReturn / 4.0
		portfolioDailyReturns = append(portfolioDailyReturns, portfolioReturn)

		fmt.Printf("  ➤ 组合收益率：%.2f%%\n", portfolioReturn)
		fmt.Println(strings.Repeat("-", 60))
	}

	// 统计信息
	if len(portfolioDailyReturns) > 0 {
		fmt.Printf("\n=== 统计摘要 ===\n")
		fmt.Printf("总交易日数：%d\n", len(portfolioDailyReturns))

		// 计算平均收益
		var sumReturn float64
		for _, ret := range portfolioDailyReturns {
			sumReturn += ret
		}
		avgReturn := sumReturn / float64(len(portfolioDailyReturns))
		fmt.Printf("平均组合收益率：%.2f%%\n", avgReturn)

		// 找到最大和最小收益
		maxReturn := portfolioDailyReturns[0]
		minReturn := portfolioDailyReturns[0]
		maxDay := 1
		minDay := 1

		for i, ret := range portfolioDailyReturns {
			if ret > maxReturn {
				maxReturn = ret
				maxDay = i + 1
			}
			if ret < minReturn {
				minReturn = ret
				minDay = i + 1
			}
		}

		fmt.Printf("最高收益：%.2f%% (第 %d 天)\n", maxReturn, maxDay)
		fmt.Printf("最低收益：%.2f%% (第 %d 天)\n", minReturn, minDay)
		fmt.Printf("最新收益：%.2f%% (第 %d 天)\n", portfolioDailyReturns[len(portfolioDailyReturns)-1], len(portfolioDailyReturns))

		// ===== 计算年化波动率 =====
		// 步骤 1: 将百分比收益率转换为小数形式
		portfolioDecimalReturns := make([]float64, len(portfolioDailyReturns))
		for i, ret := range portfolioDailyReturns {
			portfolioDecimalReturns[i] = ret / 100.0
		}

		// 步骤 2: 计算日波动率（标准差）
		portfolioDailyStd := calculateStandardDeviation(portfolioDecimalReturns)

		// 步骤 3: 年化波动率 = 日波动率 × √252
		annualizedVolatility := portfolioDailyStd * math.Sqrt(252)

		fmt.Printf("\n=== 年化波动率计算 ===\n")
		fmt.Printf("使用的组合收益率序列长度：%d\n", len(portfolioDecimalReturns))
		fmt.Printf("前 5 个收益率数据：%.6f%%, %.6f%%, %.6f%%, %.6f%%, %.6f%%\n",
			portfolioDecimalReturns[0]*100, portfolioDecimalReturns[1]*100,
			portfolioDecimalReturns[2]*100, portfolioDecimalReturns[3]*100,
			portfolioDecimalReturns[4]*100)
		fmt.Printf("日波动率 (标准差): %.6f%%\n", portfolioDailyStd*100)
		fmt.Printf("年化因子：√252 = %.4f\n", math.Sqrt(252))
		fmt.Printf("年化波动率：%.2f%%\n", annualizedVolatility*100)

		// ===== 计算夏普比率 =====
		// Sharpe Ratio = (Rp - Rf) / σp
		var sharpeRatio float64
		if annualizedVolatility > 0 {
			sharpeRatio = (Rp - Rf) / annualizedVolatility
		}

		fmt.Printf("\n=== 夏普比率计算 ===\n")
		fmt.Printf("组合年化收益率 (Rp): %.2f%%\n", Rp*100)
		fmt.Printf("无风险利率 (Rf): %.2f%%\n", Rf*100)
		fmt.Printf("超额收益 (Rp - Rf): %.2f%%\n", (Rp-Rf)*100)
		fmt.Printf("年化波动率 (σp): %.2f%%\n", annualizedVolatility*100)
		fmt.Printf("夏普比率：%.4f\n", sharpeRatio)
		if sharpeRatio > 0 {
			fmt.Printf("✓ 夏普比率 > 0，风险调整后收益为正\n")
		} else {
			fmt.Printf("⚠ 夏普比率 ≤ 0，风险调整后收益为负\n")
		}

		// 打印详细计算过程，便于验证
		var sumBoaReturn float64
		for _, ret := range portfolioDecimalReturns {
			sumBoaReturn += ret
		}
		meanReturn := sumBoaReturn / float64(len(portfolioDecimalReturns))
		fmt.Printf("\n平均日收益率：%.6f%%\n", meanReturn*100)

		// 计算方差，显示中间步骤
		var variance float64
		for _, ret := range portfolioDecimalReturns {
			diff := ret - meanReturn
			variance += diff * diff
		}
		variance = variance / float64(len(portfolioDecimalReturns)-1)
		fmt.Printf("方差：%.10f\n", variance)
		fmt.Printf("日标准差：%.10f\n", math.Sqrt(variance))

		return sharpeRatio
	}

	// 如果没有数据，返回 0
	return 0
}
func getFunds() []*Fund {
	return []*Fund{
		{
			FundName:    "国金中国铁建REIT",
			FundCode:    "508008",
			BuyingPrice: 8.004,
		},
		{
			FundName:    "易方綦达深高速爬IT",
			FundCode:    "508033",
			BuyingPrice: 6.397,
		},
		{
			FundName:    "工银河北高速REIT",
			FundCode:    "508086",
			BuyingPrice: 5.161,
		},
		{
			FundName:    "平安广州广河REIT",
			FundCode:    "180201",
			BuyingPrice: 8.477,
		},
	}
}

type FundPortfolio struct {
	Date           string
	IntervalReturn float64
}

type IntervalReturn struct {
	FundType string
	FundData []FundPortfolio
}

// IntervalReturnOfFundPortfolio 基金组合区间收益
func IntervalReturnOfFundPortfolio() (intervalReturn IntervalReturn) {

	ctx := context.Background()
	var allData []*entity.PyFund
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合
	}

	// 查询所有数据（按期初时间排序）
	for _, code := range fundCodes {
		var fundData []*entity.PyFund
		err := dao.PyFund.Ctx(ctx).
			Where(dao.PyFund.Columns().FundCode, code).
			OrderAsc("create_time").Scan(&fundData)
		if err != nil {
			fmt.Printf("查询基金 %s 数据失败：%v\n", code, err)
			continue
		}
		allData = append(allData, fundData...)
	}

	//g.Dump(allData)

	//期初的买入价格
	originFunds := getFunds()

	//g.Dump(originFunds)
	//return

	// 构建基金代码到买入价格的映射
	buyingPriceMap := make(map[string]float64)
	for _, fund := range originFunds {
		buyingPriceMap[fund.FundCode] = fund.BuyingPrice
	}

	//g.Dump(buyingPriceMap)

	//var fp []FundPortfolio

	// 按日期分组，将每支基金的数据按日期组织
	dateFundPrices := make(map[string]map[string]float64) // date -> fundCode -> price

	for _, data := range allData {
		dateStr := data.CreateTime.String()[:10] // 提取日期部分 YYYY-MM-DD
		if _, ok := dateFundPrices[dateStr]; !ok {
			dateFundPrices[dateStr] = make(map[string]float64)
		}
		dateFundPrices[dateStr][data.FundCode] = data.CurrentPrice
	}

	// 获取所有日期并排序
	dates := make([]string, 0, len(dateFundPrices))
	for date := range dateFundPrices {
		//fmt.Println("date === ", date)
		dates = append(dates, date)
	}
	sort.Strings(dates)

	//g.Dump(dates)
	// 计算每日的组合区间收益
	//portfolioDailyReturns = make([]float64, 0, len(dates))

	var fp []FundPortfolio

	for i, date := range dates {
		fundPrices := dateFundPrices[date]

		// 检查是否 4 支基金都有数据
		if len(fundPrices) != 4 {
			fmt.Printf("⚠️  第 %d 天 (%s): 基金数据不完整 (只有 %d 支)\n", i+1, date, len(fundPrices))
			continue
		}

		// 计算每支基金的当日区间收益率
		var fundReturns []float64
		fmt.Printf("第 %d 天 (%s):\n", i+1, date)

		for _, fund := range originFunds {
			currentPrice := fundPrices[fund.FundCode]
			beginningNAV := fund.BuyingPrice

			// 计算该基金从期初到当日的收益率
			// Rp = (当前价格 - 期初价格) / 期初价格 × 100%
			periodReturn := (currentPrice - beginningNAV) / beginningNAV * 100
			fundReturns = append(fundReturns, periodReturn)
			fmt.Printf("  %s: %.3f → %.3f, 收益率：%.2f%%\n", fund.FundName, beginningNAV, currentPrice, periodReturn)
		}

		// 计算等权重组合收益（4 支基金各占 25%）
		var sumReturn float64
		for _, ret := range fundReturns {
			sumReturn += ret
		}

		var f FundPortfolio

		portfolioReturn := sumReturn / 4.0

		f.Date = date
		f.IntervalReturn = portfolioReturn

		fp = append(fp, f)

		//portfolioDailyReturns = append(portfolioDailyReturns, portfolioReturn)

		fmt.Printf("  ➤ 组合收益率：%.2f%%\n", portfolioReturn)
		fmt.Println(strings.Repeat("-", 60))
	}
	intervalReturn.FundType = "基金组合区间收益"
	intervalReturn.FundData = fp
	//g.Dump(intervalReturn)
	return
}

// IntervalReturnOfBenchmark 基准组合区间收益
func IntervalReturnOfBenchmark() (intervalReturn IntervalReturn) {

	ctx := context.Background()
	var fundData []*entity.PyFund
	err := dao.PyFund.Ctx(ctx).
		Where(dao.PyFund.Columns().FundCode, "932047").
		OrderAsc("create_time").Scan(&fundData)
	if err != nil {
		fmt.Printf("查询基金 932047 数据失败：%v\n", err)
		return
	}

	// 按日期分组，将每支基金的数据按日期组织
	dateFundPrices := make(map[string]map[string]float64) // date -> fundCode -> price

	for _, data := range fundData {
		dateStr := data.CreateTime.String()[:10] // 提取日期部分 YYYY-MM-DD
		if _, ok := dateFundPrices[dateStr]; !ok {
			dateFundPrices[dateStr] = make(map[string]float64)
		}
		dateFundPrices[dateStr][data.FundCode] = data.CurrentPrice
	}
	// 获取所有日期并排序
	dates := make([]string, 0, len(dateFundPrices))
	for date := range dateFundPrices {
		//fmt.Println("date === ", date)
		dates = append(dates, date)
	}
	sort.Strings(dates)
	benchmarkOriginPrice := 1027.838
	var fp []FundPortfolio
	for _, date := range dates {
		var f FundPortfolio
		fundPrices := dateFundPrices[date]
		currentPrice := cast.ToFloat64(fundPrices["932047"])
		// Rp = (当前价格 - 期初价格) / 期初价格 × 100%
		periodReturn := (currentPrice - benchmarkOriginPrice) / benchmarkOriginPrice * 100
		f.Date = date
		f.IntervalReturn = periodReturn
		fp = append(fp, f)
	}

	intervalReturn.FundType = "基准组合区间收益"
	intervalReturn.FundData = fp
	return
}
