package reit

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/spf13/cast"
	"io"
	"math/rand"
	"regexp"
	"star/internal/crawler"
	"star/internal/dao"
	"star/internal/model/entity"
	"star/utility/stock_risk"
	"star/utility/stock_risk_return"
	"strings"
	"time"
)

type ReitController struct {
	BaseController
}

type ReitGroupListReq struct {
	g.Meta   `path:"reit/groupList" method:"get"`
	FundName string `json:"fund_name" dc:"模糊查询名称"`
}

type ReitGroupListRes struct {
	Data []*ReitNameCoReitGroupListdeGroupBy `json:"data"`
}

type ReitNameCoReitGroupListdeGroupBy struct {
	FundCode string `json:"fund_code" dc:"代码"`
	FundName string `json:"fund_name" dc:"名称"`
}

func (that *ReitController) ReitGroupList(ctx context.Context, req *ReitGroupListReq) (res *ReitGroupListRes, err error) {
	var data []*ReitNameCoReitGroupListdeGroupBy
	model := dao.PyFund.Ctx(ctx).Fields(dao.PyFund.Columns().FundCode, dao.PyFund.Columns().FundName).
		Group(dao.PyFund.Columns().FundCode, dao.PyFund.Columns().FundName)
	if req.FundName != "" {
		model = model.WhereLike(dao.PyFund.Columns().FundName, "%"+req.FundName+"%")
	}
	err = model.Scan(&data)
	if err != nil {
		return nil, err
	}
	return &ReitGroupListRes{
		Data: data,
	}, nil
}

type ReitListReq struct {
	g.Meta    `path:"reit/list" method:"POST"`
	FundCode  string      `json:"fund_code" dc:"多选func_code 逗号隔开"`
	StartTime *gtime.Time `json:"start_time"`
	EndTime   *gtime.Time `json:"end_time"`
}

type ReitData struct {
	Id                    int         `json:"id"                      orm:"id"                      description:""`        //
	FundCode              string      `json:"fund_code"               orm:"fund_code"               description:"基金代码"`    // 基金代码
	FundName              string      `json:"fund_name"               orm:"fund_name"               description:"基金名称"`    // 基金名称
	Exchange              string      `json:"exchange"                orm:"exchange"                description:"交易所"`     // 交易所
	FundCategory          string      `json:"fund_category"           orm:"fund_category"           description:"资金类别"`    // 资金类别
	OpeningPrice          float64     `json:"opening_price"           orm:"opening_price"           description:"开盘价"`     // 开盘价
	HighestPrice          float64     `json:"highest_price"           orm:"highest_price"           description:"最高价"`     // 最高价
	LowestPrice           float64     `json:"lowest_price"            orm:"lowest_price"            description:"最低价"`     // 最低价
	CurrentPrice          float64     `json:"current_price"           orm:"current_price"           description:"当前价"`     // 当前价
	PreviousClosePrice    float64     `json:"previous_close_price"    orm:"previous_close_price"    description:"上一个收盘价"`  // 上一个收盘价
	PriceChangePercentage float64     `json:"price_change_percentage" orm:"price_change_percentage" description:"价格变化百分比"` // 价格变化百分比
	Volume                int64       `json:"volume"                  orm:"volume"                  description:"成交量"`     // 成交量
	Turnover              float64     `json:"turnover"                orm:"turnover"                description:""`        //
	CreateTime            *gtime.Time `json:"create_time"             orm:"create_time"             description:"创建时间"`    // 创建时间
}

type ReitListRes struct {
	Total int         `json:"total"`
	Data  []*ReitData `json:"data"`
}

func (that *ReitController) ReitList(ctx context.Context, req *ReitListReq) (res *ReitListRes, err error) {
	if req.FundCode == "" {
		return nil, gerror.NewCode(gcode.New(10001, "基金code不能为空", ""))
	}
	if req.StartTime == nil || req.EndTime == nil {
		return nil, gerror.NewCode(gcode.New(10001, "时间不能为空", ""))
	}

	//g.Dump("req === ", req)

	// 实际应用示例：处理基金代码列表
	//codeList := strings.Split(req.FundCode, ",")
	var data []*ReitData
	var total int
	err = dao.PyFund.Ctx(ctx).
		WhereLTE(dao.PyFund.Columns().CreateTime, req.EndTime).
		WhereGTE(dao.PyFund.Columns().CreateTime, req.StartTime).
		Where(dao.PyFund.Columns().FundCode, req.FundCode).
		ScanAndCount(&data, &total, true)
	//g.Dump("total === ", total)
	if err != nil {
		return nil, err
	}
	return &ReitListRes{
		Total: total,
		Data:  data,
	}, nil
}

// 爬虫相关接口

type StockQuoteReq struct {
	g.Meta `path:"stock/quote" method:"get"`
	Symbol string `json:"symbol" dc:"股票代码" v:"required#股票代码不能为空"`
}

type StockQuoteRes struct {
	Symbol    string  `json:"symbol"`     // 股票代码
	Name      string  `json:"name"`       // 股票名称
	Current   float64 `json:"current"`    // 当前价格
	Percent   float64 `json:"percent"`    // 涨跌幅(%)
	Chg       float64 `json:"chg"`        // 涨跌额
	Open      float64 `json:"open"`       // 开盘价
	High      float64 `json:"high"`       // 最高价
	Low       float64 `json:"low"`        // 最低价
	Close     float64 `json:"close"`      // 收盘价
	Volume    int64   `json:"volume"`     // 成交量
	Turnover  float64 `json:"turnover"`   // 成交额
	Timestamp int64   `json:"timestamp"`  // 时间戳
	UpdatedAt string  `json:"updated_at"` // 更新时间格式化
}

type HistoricalDataReq struct {
	g.Meta `path:"stock/historical" method:"get"`
	Symbol string `json:"symbol" dc:"股票代码" v:"required#股票代码不能为空"`
	Date   string `json:"date" dc:"查询日期(YYYY-MM-DD)" v:"required#日期不能为空"`
	Name   string `json:"name" dc:"股票名称" v:"required#股票名字不能为空"`
}

type HistoricalDataRes struct {
	Name         string  `json:"name"`
	Current      float64 `json:"current"`
	Symbol       string  `json:"symbol"`        // 股票代码
	Date         string  `json:"date"`          // 日期
	Open         float64 `json:"open"`          // 开盘价
	High         float64 `json:"high"`          // 最高价
	Low          float64 `json:"low"`           // 最低价
	Close        float64 `json:"close"`         // 收盘价
	Volume       int64   `json:"volume"`        // 成交量
	Amount       float64 `json:"amount"`        // 成交额
	Chg          float64 `json:"chg"`           // 涨跌额
	Percent      float64 `json:"percent"`       // 涨跌幅(%)
	TurnoverRate float64 `json:"turnover_rate"` // 换手率
	Timestamp    int64   `json:"timestamp"`     // 时间戳
	UpdatedAt    string  `json:"updated_at"`
}

// GetStockQuote 获取股票实时报价
func (that *ReitController) GetStockQuote(ctx context.Context, req *StockQuoteReq) (res *StockQuoteRes, err error) {
	if req.Symbol == "" {
		return nil, gerror.NewCode(gcode.New(10001, "股票代码不能为空", ""))
	}

	// 创建爬虫实例
	crawler := crawler.NewXueqiuCrawler()

	// 获取股票数据
	quote, err := crawler.GetStockQuote(req.Symbol)
	if err != nil {
		return nil, gerror.NewCode(gcode.New(50001, fmt.Sprintf("获取股票数据失败: %v", err), ""))
	}

	// 构造响应数据
	res = &StockQuoteRes{
		Symbol:    quote.Symbol,
		Name:      quote.Name,
		Current:   quote.Current,
		Percent:   quote.Percent,
		Chg:       quote.Chg,
		Open:      quote.Open,
		High:      quote.High,
		Low:       quote.Low,
		Close:     quote.Close,
		Volume:    quote.Volume,
		Turnover:  quote.Turnover,
		Timestamp: quote.Timestamp,
		UpdatedAt: quote.FormatTimestamp(),
	}

	return res, nil
}

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

// GetHistoricalData 获取股票历史数据
func (that *ReitController) GetHistoricalData(ctx context.Context, req *HistoricalDataReq) (res *HistoricalDataRes, err error) {

	if req.Symbol == "" {
		return nil, gerror.NewCode(gcode.New(10001, "股票代码不能为空", ""))
	}

	if req.Date == "" {
		return nil, gerror.NewCode(gcode.New(10001, "日期不能为空", ""))
	}

	// 验证日期格式
	_, err = time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, gerror.NewCode(gcode.New(10001, "日期格式错误，应为 YYYY-MM-DD 格式", ""))
	}

	re := regexp.MustCompile(`[^0-9]`)
	pureNumber := re.ReplaceAllString(req.Symbol, "")

	var m *PyFund
	dao.PyFund.Ctx(ctx).
		Where(dao.PyFund.Columns().FundCode, pureNumber).
		Scan(&m)
	//fmt.Println("name === ", m.FundName)

	// 创建爬虫实例
	crawler := crawler.NewXueqiuCrawler()

	// 获取历史数据
	klineData, err := crawler.GetHistoricalData(req.Symbol, req.Date)
	if err != nil {
		return nil, gerror.NewCode(gcode.New(50001, fmt.Sprintf("获取历史数据失败: %v", err), ""))
	}

	//g.Dump("klineData === ", klineData)

	// 构造响应数据
	res = &HistoricalDataRes{
		Symbol:       req.Symbol,
		Name:         m.FundName,
		UpdatedAt:    req.Date,
		Current:      klineData.Close,
		Percent:      klineData.Percent,
		Chg:          klineData.Chg,
		Open:         klineData.Open,
		High:         klineData.High,
		Low:          klineData.Low,
		Close:        klineData.Close,
		Volume:       klineData.Volume,
		Amount:       klineData.Amount,
		TurnoverRate: klineData.TurnoverRate,
		Timestamp:    klineData.Timestamp,
	}

	return res, nil
}

// 收益指标
type RiskReturn struct {
	Alpha            float64 `json:"alpha"`
	CumulativeReturn float64 `json:"cumulativeReturn" dc:"累计收益"`
	AnnualizedReturn float64 `json:"annualizedReturn" dc:"年化收益"`
	SharpeRatio      float64 `json:"sharpeRatio" dc:""`
	InfoRatio        float64 `json:"infoRatio"`
	SortinoRatio     float64 `json:"sortinoRatio" dc:"索提诺比率"`
	JensenAlpha      float64 `json:"jensenAlpha"`
	TreynorRatio     float64 `json:"treynorRatio" dc:"特雷诺比率"`
	WinRate          float64 `json:"winRate" dc:"胜率"`
	PositivePeriods  int     `json:"positivePeriods" dc:"正周期数"`
	TotalPeriods     int     `json:"totalPeriods" dc:"总周期数"`
}
type Risk struct {
	Beta              float64 `json:"beta"`
	AnnualizedVol     float64 `json:"annualizedVol" dc:"年化波动率"`
	TrackingError     float64 `json:"trackingError" dc:"跟踪误差"`
	DownsideRisk      float64 `json:"downsideRisk" dc:"下行风险"`
	Var               float64 `json:"var" dc:"在险价值 (VaR)"`
	MaxDrawdown       float64 `json:"maxDrawdown" dc:"最大回撤"`
	DrawdownFormation int     `json:"drawdownFormation" dc:"回撤形成期"`
	DrawdownRecovery  int     `json:"drawdownRecovery" dc:"回撤恢复期"`
	ConsecutiveDrop   int     `json:"consecutiveDrop" dc:"连续下降期数"`
	RSquare           float64 `json:"rSquare" dc:"R方 决定系数，衡量组合与基准的相关性"`
}

type HelloReq struct {
	g.Meta `path:"hello" method:"get"`
}
type HelloRes struct {
	Msg string `json:"msg"`
}

func (that *ReitController) Hello(ctx context.Context, req *HelloReq) (res *HelloRes, err error) {
	return &HelloRes{
		Msg: "hello world",
	}, nil
}

type GetRiskReq struct {
	g.Meta `path:"stock/risk" method:"get"`
}

type GetRiskRes struct {
	RiskReturn RiskReturn `json:"riskReturn"`
	Risk       Risk       `json:"risk"`
}

func (that *ReitController) GetRisk(ctx context.Context, req *GetRiskReq) (res *GetRiskRes, err error) {
	//fundCrawler(ctx)
	//return
	/*
		fundPortfolio := stock_risk_return.IntervalReturnOfFundPortfolio()
		jsonData1, err := json.MarshalIndent(fundPortfolio, "", "  ")
		if err != nil {
			fmt.Printf("JSON 序列化失败：%v\n", err)
			return
		}
		// 写入文件
		filePath1 := "fundPortfolio.json"
		err = os.WriteFile(filePath1, jsonData1, 0644)
		if err != nil {
			fmt.Printf("写入文件失败：%v\n", err)
			return
		}
		//stock_risk_return.IntervalReturnOfFundPortfolio()
		benchmark := stock_risk_return.IntervalReturnOfBenchmark()
		jsonData2, err := json.MarshalIndent(benchmark, "", "  ")
		if err != nil {
			fmt.Printf("JSON 序列化失败：%v\n", err)
			return
		}
		// 写入文件
		filePath := "benchmark.json"
		err = os.WriteFile(filePath, jsonData2, 0644)
		if err != nil {
			fmt.Printf("写入文件失败：%v\n", err)
			return
		}
	*/

	var riskReturn RiskReturn
	var risk Risk
	beta := stock_risk.GetBeta(ctx)
	risk.Beta = beta
	//g.Dump(beta)
	//return
	benchMarkAnnualized, err := benchMarkPrice()
	if err != nil {
		return
	}
	riskReturn.CumulativeReturn = stock_risk_return.GetCumulativeReturn(ctx)
	riskReturn.AnnualizedReturn = stock_risk_return.GetAnnualizedReturn(ctx)
	//g.Log().Infof(ctx, "基金组合年化收益率 : %f ", riskReturn.AnnualizedReturn)
	riskReturn.Alpha = stock_risk_return.Alpha(ctx, riskReturn.AnnualizedReturn, 0.02, benchMarkAnnualized, risk.Beta)
	riskReturn.SharpeRatio = stock_risk_return.SharpeRatio(ctx, riskReturn.AnnualizedReturn, 0.02)
	//g.Dump(riskReturn.SharpeRatio)
	riskReturn.InfoRatio = stock_risk_return.GetInformationRatio(ctx, riskReturn.AnnualizedReturn, benchMarkAnnualized)
	riskReturn.SortinoRatio = stock_risk_return.GetSortinoRatio(ctx, riskReturn.AnnualizedReturn, 0.02)
	//g.Dump(riskReturn.SortinoRatio)
	//return

	riskReturn.JensenAlpha = riskReturn.Alpha

	riskReturn.TreynorRatio = stock_risk_return.GetTreynorRatio(ctx, riskReturn.AnnualizedReturn, 0.02, risk.Beta)
	//g.Log().Infof(ctx, " TreynorRatio : %f ", riskReturn.TreynorRatio)
	//return
	riskReturn.WinRate = stock_risk_return.GetWinRate(ctx)
	//g.Log().Infof(ctx, " WinRate : %f ", riskReturn.WinRate)
	riskReturn.PositivePeriods = stock_risk_return.GetPositivePeriods(ctx)
	//g.Log().Infof(ctx, " PositivePeriods : %d ", riskReturn.PositivePeriods)
	riskReturn.TotalPeriods = stock_risk_return.GetTotalPeriods(ctx)

	/*  stock_risk  */
	risk.AnnualizedVol = stock_risk_return.GetAnnualizedVol(ctx)
	//g.Log().Infof(ctx, " AnnualizedVol : %f ", risk.AnnualizedVol)

	risk.TrackingError = stock_risk_return.GetTrackingError(ctx)
	//g.Log().Infof(ctx, " TrackingError : %f ", risk.TrackingError)

	risk.DownsideRisk = stock_risk_return.GetDownsideRisk(ctx)

	//g.Log().Infof(ctx, " DownsideRisk : %f ", risk.DownsideRisk)

	risk.Var = stock_risk_return.GetVaR(ctx, 0.95) // confidenceLevel：置信水平 0-1 之间
	//g.Log().Infof(ctx, " Var : %f ", risk.Var)

	risk.MaxDrawdown = stock_risk_return.GetMaxDrawdown(ctx)
	//g.Log().Infof(ctx, " MaxDrawdown : %f ", risk.MaxDrawdown)

	risk.DrawdownFormation = stock_risk_return.GetDrawdownFormation(ctx)
	risk.DrawdownRecovery = stock_risk_return.GetDrawdownRecovery(ctx)
	risk.ConsecutiveDrop = stock_risk_return.GetConsecutiveDrop(ctx)
	risk.RSquare = stock_risk_return.GetRSquare(ctx)
	//g.Dump(benchMarkAnnualized)
	//g.Dump(last)

	return &GetRiskRes{
		RiskReturn: riskReturn,
		Risk:       risk,
	}, nil
}

// benchMarkPrice 计算基准年化收益率
func benchMarkPrice() (benchMarkAnnualized float64, err error) {

	ctx := context.Background()
	var latest *PyFund
	var last *PyFund
	err = dao.PyFund.Ctx(ctx).Where(dao.PyFund.Columns().FundCode, "932047").
		OrderDesc("create_time").Scan(&latest)
	if err != nil {
		return
	}
	err = dao.PyFund.Ctx(ctx).Where(dao.PyFund.Columns().FundCode, "932047").
		OrderAsc("create_time").Scan(&last)
	if err != nil {
		return
	}

	//g.Log().Infof(ctx, "初始价格：%f,最新价格：%f", last.CurrentPrice, latest.CurrentPrice)

	//R_period = (期末价值 - 期初价值) / 期初价值

	periodReturn := (latest.CurrentPrice - last.CurrentPrice) / last.CurrentPrice

	//fmt.Println("periodReturn:", periodReturn)

	// 计算持有天数（通过时间戳）
	// Timestamp 返回秒级时间戳，相减后除以 86400(24*3600) 得到天数
	holdingDays := float64(latest.CreateTime.Timestamp()-last.CreateTime.Timestamp()) / (24 * 3600)
	//g.Log().Infof(ctx, "持有天数：%f", holdingDays)
	benchMarkAnnualized = stock_risk_return.CalculateAnnualizedReturn(periodReturn, cast.ToInt(holdingDays))
	return
	// 创建爬虫实例
	//currentYear := gtime.Now().Format("Y-m-d")
	//lastYear := gtime.Now().AddDate(-1, 0, 0).Format("Y-m-d")
	//crawler := crawler.NewXueqiuCrawler()
	// 获取历史数据
	//currentData, err := crawler.GetHistoricalData("CSI932047", currentYear)
	//if err != nil {
	//	return
	//}
	//lastData, err := crawler.GetHistoricalData("CSI932047", lastYear)
	//if err != nil {
	//	return
	//}
	//currentPrice := currentData.Close
	//lastPrice := lastData.Close
	// 基准 年收益率 = (期末指数点位 - 期初指数点位) / 期初指数点位
	//benchMarkAnnualized = (currentPrice - lastPrice) / lastPrice
	//benchMarkAnnualized = float64(int(benchMarkAnnualized*10000)) / 10000
	//return
}

// fundCrawler 爬取基金数据
func fundCrawler(ctx context.Context) {
	// 创建爬虫实例
	crawler := crawler.NewXueqiuCrawler()

	// 初始日期
	initialDate := "2025-01-31"
	// 结束日期
	endDate := "2026-04-10"

	// 1. 解析字符串为 time.Time
	startDate, err := time.Parse("2006-01-02", initialDate)
	if err != nil {
		fmt.Printf("解析日期失败：%v\n", err)
		return
	}

	endDateTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		fmt.Printf("解析结束日期失败：%v\n", err)
		return
	}

	// 基金代码列表
	//fundCodes := []string{
	//	"508008", "508033", "508086", "180201", // 基金组合
	//	"932047", // 基准（中证 REITs 指数）
	//}

	// 循环遍历每个日期
	for currentDate := startDate; !currentDate.After(endDateTime); currentDate = currentDate.AddDate(0, 0, 1) {
		ms := rand.Intn(51) + 100
		time.Sleep(time.Duration(ms) * time.Millisecond)
		dateStr := currentDate.Format("2006-01-02")
		fmt.Printf("正在处理日期：%s\n", dateStr)
		fundCode := "000001"
		fundName := "上证指数"
		fundExchange := "SH"
		// 遍历每个基金代码
		//for _, fundCode := range fundCodes {
		//	fmt.Printf("  正在获取基金 %s 的数据...\n", fundCode)
		//
		//	// 获取历史数据
		k, err := crawler.GetHistoricalData(fundExchange+fundCode, dateStr)

		if err != nil {
			fmt.Printf("获取基金 %s 在 %s 的数据失败：%v\n", fundCode, dateStr, err)
			continue
		}
		_, err = dao.PyT.Ctx(ctx).Data(g.Map{
			dao.PyT.Columns().FundCode:     fundCode,
			dao.PyT.Columns().FundName:     fundName,
			dao.PyT.Columns().Exchange:     fundExchange,
			dao.PyT.Columns().FundCategory: "REITs",
			dao.PyT.Columns().OpeningPrice: k.Open,
			dao.PyT.Columns().HighestPrice: k.High,
			dao.PyT.Columns().LowestPrice:  k.Low,
			dao.PyT.Columns().CurrentPrice: k.Close,
			dao.PyT.Columns().CreateTime:   currentDate,
			dao.PyT.Columns().UpdateTime:   gtime.Now(),
		}).Insert()
		if err != nil {
			fmt.Printf("插入数据失败：%v\n", err)
		}
		//
		//	// 在这里可以保存数据到数据库或进行其他处理
		//	fmt.Printf("    成功获取：%s, 收盘价=%.2f, 涨跌幅=%.2f%%\n", fundCode, klineData.Close, klineData.Percent)
		//}

		fmt.Println("-----------------------------------")
	}

	fmt.Println("所有日期数据处理完成！")
}

type ImportHistoryTradeReq struct {
	g.Meta `path:"upload/history" method:"post"`
}

type ImportHistoryTradeRes struct {
	Data    []map[string]interface{} `json:"data" dc:"读取的 Excel 数据"`
	Success int                      `json:"success" dc:"成功导入的数据条数"`
}

// ImportHistoryTrade POST 接受上传 excel 文件，读取文件内容并导入数据库
func (that *ReitController) ImportHistoryTrade(ctx context.Context, req *ImportHistoryTradeReq) (res *ImportHistoryTradeRes, err error) {
	// 从请求中获取上传的文件
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.New(50001, "获取请求上下文失败", ""))
	}

	// 获取上传的文件
	file, header, err := r.Request.FormFile("file")
	if err != nil {
		return nil, gerror.NewCode(gcode.New(10001, "请上传 CSV 文件", ""))
	}
	defer file.Close()

	// 读取文件内容到内存
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, gerror.NewCode(gcode.New(50001, fmt.Sprintf("读取文件失败：%v", err), ""))
	}

	// 检测文件类型（通过文件扩展名）
	filename := strings.ToLower(header.Filename)
	var rows [][]string

	if strings.HasSuffix(filename, ".csv") {
		// 解析 CSV 文件
		reader := csv.NewReader(bytes.NewReader(content))
		// 自动处理常见的 CSV 格式问题
		reader.LazyQuotes = true
		reader.TrimLeadingSpace = true
		reader.FieldsPerRecord = -1 // 允许每行字段数不同
		rows, err = reader.ReadAll()
		if err != nil {
			return nil, gerror.NewCode(gcode.New(50001, fmt.Sprintf("解析 CSV 文件失败：%v", err), ""))
		}
	} else {
		return nil, gerror.NewCode(gcode.New(10001, "仅支持 CSV 文件格式", ""))
	}

	if len(rows) < 2 {
		return nil, gerror.NewCode(gcode.New(10001, "CSV 文件为空或没有数据行", ""))
	}

	// 第一行是表头
	headers := rows[0]
	fmt.Printf("表头：%v\n", headers)

	// 找到 TradeDate 和 TradeTime 列的索引
	tradeDateIndex := -1
	tradeTimeIndex := -1
	for i, h := range headers {
		if h == "TradeDate" || h == "trade_date" || h == "交易日期" {
			tradeDateIndex = i
		}
		if h == "TradeTime" || h == "trade_time" || h == "交易时间" {
			tradeTimeIndex = i
		}
	}

	if tradeDateIndex == -1 {
		return nil, gerror.NewCode(gcode.New(10001, "CSV 中缺少 TradeDate 列", ""))
	}

	// 解析数据行并导入数据库
	successCount := 0
	var dataList []map[string]interface{}
	for i, row := range rows[1:] {
		rowData := make(map[string]interface{})
		for j, header := range headers {
			if j < len(row) {
				rowData[header] = row[j]
			} else {
				rowData[header] = ""
			}
		}

		entrustTypeName := getCellValue(row, headers, "EntrustTypeName")
		if entrustTypeName == "手工冻结" {
			continue
		}

		// 合并 TradeDate 和 TradeTime
		var tradeTimeStr string
		if tradeTimeIndex != -1 && tradeTimeIndex < len(row) && strings.TrimSpace(row[tradeTimeIndex]) != "" {
			// 有时间字段，合并日期和时间
			// 清理制表符和空格
			dateStr := strings.TrimSpace(strings.ReplaceAll(row[tradeDateIndex], "\t", ""))
			timeStr := strings.TrimSpace(strings.ReplaceAll(row[tradeTimeIndex], "\t", ""))

			// 如果日期是 YYYYMMDD 格式，转换为 YYYY-MM-DD
			if len(dateStr) == 8 {
				dateStr = fmt.Sprintf("%s-%s-%s", dateStr[:4], dateStr[4:6], dateStr[6:])
			}

			tradeTimeStr = fmt.Sprintf("%s %s", dateStr, timeStr)
		} else {
			// 只有日期字段
			dateStr := strings.TrimSpace(strings.ReplaceAll(row[tradeDateIndex], "\t", ""))
			// 如果日期是 YYYYMMDD 格式，转换为 YYYY-MM-DD
			if len(dateStr) == 8 {
				dateStr = fmt.Sprintf("%s-%s-%s", dateStr[:4], dateStr[4:6], dateStr[6:])
			}
			tradeTimeStr = fmt.Sprintf("%s 00:00:00", dateStr)
		}

		// 解析为 time.Time
		tradeTime, err := time.Parse("2006-01-02 15:04:05", tradeTimeStr)
		if err != nil {
			fmt.Printf("第 %d 行日期解析失败：%s, 错误：%v\n", i+2, tradeTimeStr, err)
			continue
		}

		// 构建 PyHt 对象
		pyHt := &entity.PyHt{
			TradeTime:       gtime.New(tradeTime),
			SecurityName:    getCellValue(row, headers, "SecurityName"),
			SecurityId:      cast.ToInt(getCellValue(row, headers, "SecurityID")),
			EntrustTypeName: getCellValue(row, headers, "EntrustTypeName"),
			EntrustType:     cast.ToInt(getCellValue(row, headers, "EntrustType")),
			Volume:          cast.ToInt(getCellValue(row, headers, "Volume")),
			Price:           cast.ToFloat64(getCellValue(row, headers, "Price")),
			Turnover:        cast.ToFloat64(getCellValue(row, headers, "Turnover")),
			TradeId:         getCellValue(row, headers, "TradeID"),
			ExchangeName:    getCellValue(row, headers, "ExchangeName"),
			OrderSerialId:   getCellValue(row, headers, "OrderSerialID"),
			ShareholderId:   getCellValue(row, headers, "ShareholderID"),
			UpdateTime:      gtime.Now(),
			CreateTime:      gtime.Now(),
		}

		switch getCellValue(row, headers, "EntrustTypeName") {
		case "买入":
			pyHt.EntrustType = 1
		case "卖出":
			pyHt.EntrustType = 2
		case "红利":
			pyHt.EntrustType = 3
		case "手工冻结":
			pyHt.EntrustType = 4
		case "指定":
			pyHt.EntrustType = 5
		default:
			return nil, gerror.NewCode(gcode.New(10001, "CSV 中缺少 EntrustTypeName 列", ""))
		}

		// 验证数据
		count, err := dao.PyHt.Ctx(ctx).
			Where(dao.PyHt.Columns().TradeTime, pyHt.TradeTime).
			Where(dao.PyHt.Columns().SecurityName, pyHt.SecurityName).
			Where(dao.PyHt.Columns().SecurityId, pyHt.SecurityId).Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			fmt.Printf("第 %d 行数据已存在，跳过插入\n", i+2)
			continue
		}

		// 插入数据库
		_, err = dao.PyHt.Ctx(ctx).Data(pyHt).Insert()
		if err != nil {
			fmt.Printf("第 %d 行插入数据库失败：%v\n", i+2, err)
			continue
		}

		successCount++
		dataList = append(dataList, rowData)
	}

	fmt.Printf("成功导入 %d 行数据到数据库\n", successCount)
	fmt.Printf("文件名：%s\n", header.Filename)

	return &ImportHistoryTradeRes{
		Data:    dataList,
		Success: successCount,
	}, nil
}

// getCellValue 获取单元格值，支持多种列名格式
func getCellValue(row []string, headers []string, columnName string) string {
	for i, h := range headers {
		if h == columnName || h == toSnakeCase(columnName) || lowerFirst(h) == lowerFirst(columnName) {
			if i < len(row) {
				// 清理制表符和空格
				return strings.TrimSpace(strings.ReplaceAll(row[i], "\t", ""))
			}
		}
	}
	return ""
}

func toSnakeCase(s string) string {
	result := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return result
}

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

type TradeHistoryListReq struct {
	g.Meta `path:"tradeHistoryList" method:"get"`
	Page   int `json:"page" v:"min:1" dc:"页码，默认1"`
	Size   int `json:"size" v:"between:1,100" dc:"每页数量，默认10"`
}

type TradeData struct {
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
}

type TradeHistoryListRes struct {
	Data  []*TradeData `json:"data"`
	Total int          `json:"total"`
}

func (that *ReitController) TradeHistoryList(ctx context.Context, req *TradeHistoryListReq) (res *TradeHistoryListRes, err error) {
	var data []*TradeData
	var total int

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Size == 0 {
		req.Size = 10
	}

	err = dao.PyHt.Ctx(ctx).Page(req.Page, req.Size).OrderDesc(dao.PyHt.Columns().TradeTime).ScanAndCount(&data, &total, true)
	if err != nil {
		return nil, err
	}
	return &TradeHistoryListRes{
		Data:  data,
		Total: total,
	}, nil
}

type FundVolumeListReq struct {
	g.Meta `path:"fundVolumeList" method:"get"`
}

type FundVolumeListRes struct {
	Data []*TradeData `json:"data"`
}

func (that *ReitController) FundVolumeList(ctx context.Context, req *FundVolumeListReq) (res *FundVolumeListRes, err error) {
	var securityId []int
	all, err := dao.PyHt.Ctx(ctx).Group(dao.PyHt.Columns().SecurityId).
		Having(dao.PyHt.Columns().SecurityId + "!= 799999").
		Fields(dao.PyHt.Columns().SecurityId).
		All()
	if err != nil {
		return nil, err
	}
	for _, v := range all {
		securityId = append(securityId, v["security_id"].Int())
	}
	var tradeData []*TradeData
	for _, v := range securityId {
		//fmt.Println("securityId === ", v)
		var ht []*TradeData
		err := dao.PyHt.Ctx(ctx).
			Where(dao.PyHt.Columns().SecurityId, v).
			OrderAsc(dao.PyHt.Columns().TradeTime).
			Scan(&ht)
		if err != nil {
			continue
		}

		if len(ht) > 0 {
			var volume int
			var sub *TradeData
			for _, sv := range ht {

				// EntrustType not in (1,2) continue
				if sv.EntrustType != 1 && sv.EntrustType != 2 {
					continue
				}

				switch sv.EntrustType {
				case 1:
					volume += sv.Volume
				case 2:
					volume -= sv.Volume
				}
				sub = sv
			}
			if volume == 0 || sub == nil {
				continue
			}
			sub.Volume = volume
			tradeData = append(tradeData, sub)
		}
	}
	return &FundVolumeListRes{
		Data: tradeData,
	}, nil
}

type OriginFundBuyListReq struct {
	g.Meta `path:"originFundBuyList" method:"get"`
}

type OriginFundBuyData struct {
	Id          int     `json:"id"          orm:"id"           description:""` //
	Stocks      int     `json:"stocks"      orm:"stocks"       description:""` //
	FundName    string  `json:"fundName"    orm:"fund_name"    description:""` //
	FundCode    string  `json:"fundCode"    orm:"fund_code"    description:""` //
	BuyingPrice float64 `json:"buyingPrice" orm:"buying_price" description:""` //
	Exchange    string  `json:"exchange"    orm:"exchange"     description:""` //
}

type OriginFundBuyListRes struct {
	Data []*OriginFundBuyData `json:"data"`
}

func (that *ReitController) OriginFundBuyList(ctx context.Context, req *OriginFundBuyListReq) (res *OriginFundBuyListRes, err error) {
	var data []*OriginFundBuyData
	err = dao.PyHad.Ctx(ctx).Scan(&data)
	if err != nil {
		return nil, err
	}
	return &OriginFundBuyListRes{
		Data: data,
	}, nil
}
