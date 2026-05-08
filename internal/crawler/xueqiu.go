package crawler

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"star/internal/utils"
	"time"
)

// XueqiuQuote 雪球股票报价数据结构
type XueqiuQuote struct {
	Data struct {
		Quote QuoteInfo `json:"quote"`
	} `json:"data"`
}

// QuoteInfo 股票报价详细信息
type QuoteInfo struct {
	Symbol    string  `json:"symbol"`    // 股票代码
	Name      string  `json:"name"`      // 股票名称
	Current   float64 `json:"current"`   // 当前价格
	Percent   float64 `json:"percent"`   // 涨跌幅(%)
	Chg       float64 `json:"chg"`       // 涨跌额
	Open      float64 `json:"open"`      // 开盘价
	High      float64 `json:"high"`      // 最高价
	Low       float64 `json:"low"`       // 最低价
	Close     float64 `json:"close"`     // 收盘价
	Volume    int64   `json:"volume"`    // 成交量
	Turnover  float64 `json:"turnover"`  // 成交额
	Timestamp int64   `json:"timestamp"` // 时间戳
}

// KlineData K线数据结构
type KlineData struct {
	Timestamp    int64   `json:"timestamp"`    // 时间戳(ms)
	Volume       int64   `json:"volume"`       // 成交量
	Open         float64 `json:"open"`         // 开盘价
	High         float64 `json:"high"`         // 最高价
	Low          float64 `json:"low"`          // 最低价
	Close        float64 `json:"close"`        // 收盘价
	Chg          float64 `json:"chg"`          // 涨跌额
	Percent      float64 `json:"percent"`      // 涨跌幅(%)
	TurnoverRate float64 `json:"turnoverrate"` // 换手率
	Amount       float64 `json:"amount"`       // 成交额
	VolumePost   *int64  `json:"volume_post"`  // 盘后成交量
	AmountPost   *int64  `json:"amount_post"`  // 盘后成交额
}

// KlineResponse K线响应结构
type KlineResponse struct {
	Data struct {
		Symbol string          `json:"symbol"`
		Column []string        `json:"column"` // 字段名列表
		Items  [][]interface{} `json:"item"`   // 二维数组数据
	} `json:"data"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

// XueqiuCrawler 雪球爬虫
type XueqiuCrawler struct {
	httpClient *utils.HTTPClient
}

// NewXueqiuCrawler 创建新的雪球爬虫实例
func NewXueqiuCrawler() *XueqiuCrawler {
	return &XueqiuCrawler{
		httpClient: utils.NewHTTPClient(),
	}
}

// GetStockQuote 获取股票报价
func (xc *XueqiuCrawler) GetStockQuote(symbol string) (*QuoteInfo, error) {
	url := "https://stock.xueqiu.com/v5/stock/quote.json"

	params := fmt.Sprintf("?symbol=%s&extend=detail", symbol)
	fullURL := url + params

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Origin":          "https://xueqiu.com",
		"Referer":         fmt.Sprintf("https://xueqiu.com/S/%s", symbol),
	}

	// 从配置文件读取cookies
	cookies := map[string]string{
		"device_id":  g.Cfg().MustGet(nil, "xueqiu.cookies.device_id").String(),
		"xq_a_token": g.Cfg().MustGet(nil, "xueqiu.cookies.xq_a_token").String(),
		"u":          g.Cfg().MustGet(nil, "xueqiu.cookies.u").String(),
	}

	//g.Dump(cookies)

	responseBody, err := xc.httpClient.Get(fullURL, headers, cookies)
	if err != nil {
		return nil, fmt.Errorf("获取股票数据失败: %w", err)
	}

	var quoteData XueqiuQuote
	err = utils.ParseJSON(responseBody, &quoteData)
	if err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &quoteData.Data.Quote, nil
}

// GetHistoricalData 获取历史K线数据
func (xc *XueqiuCrawler) GetHistoricalData(symbol string, date string) (*KlineData, error) {
	// 解析日期
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("日期格式错误，应为 YYYY-MM-DD 格式: %w", err)
	}

	// 转换为毫秒级时间戳（13位）
	timestamp := t.Unix()*1000 + 8*3600*1000 // 加上8小时时区偏移

	url := "https://stock.xueqiu.com/v5/stock/chart/kline.json"
	params := fmt.Sprintf("?symbol=%s&begin=%d&period=day&type=before&count=-1&indicator=kline", symbol, timestamp)
	fullURL := url + params

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"Origin":          "https://xueqiu.com",
		"Referer":         fmt.Sprintf("https://xueqiu.com/S/%s", symbol),
	}

	// 从配置文件读取cookies
	cookies := map[string]string{
		"device_id":  g.Cfg().MustGet(nil, "xueqiu.cookies.device_id").String(),
		"xq_a_token": g.Cfg().MustGet(nil, "xueqiu.cookies.xq_a_token").String(),
		"u":          g.Cfg().MustGet(nil, "xueqiu.cookies.u").String(),
	}

	//g.Log().Infof(nil, "请求历史数据URL: %s", fullURL)
	g.Log().Infof(nil, "请求时间戳: %d (对应日期: %s)", timestamp, date)

	responseBody, err := xc.httpClient.Get(fullURL, headers, cookies)
	if err != nil {
		return nil, fmt.Errorf("获取历史数据失败: %w", err)
	}

	var klineResp KlineResponse
	err = utils.ParseJSON(responseBody, &klineResp)
	if err != nil {
		return nil, fmt.Errorf("解析历史数据响应失败: %w", err)
	}

	// 检查是否有错误
	if klineResp.ErrorCode != 0 {
		return nil, fmt.Errorf("API返回错误: %s (错误码: %d)", klineResp.ErrorDescription, klineResp.ErrorCode)
	}

	//g.Dump(klineResp)

	// 解析二维数组数据
	var klineDatas []KlineData
	for _, item := range klineResp.Data.Items {
		if len(item) >= 10 { // 确保有足够的字段
			klineData := KlineData{}

			// 按照column顺序解析数据
			for i, columnName := range klineResp.Data.Column {
				if i >= len(item) {
					break
				}

				switch columnName {
				case "timestamp":
					if val, ok := item[i].(float64); ok {
						klineData.Timestamp = int64(val)
					}
				case "volume":
					if val, ok := item[i].(float64); ok {
						klineData.Volume = int64(val)
					}
				case "open":
					if val, ok := item[i].(float64); ok {
						klineData.Open = val
					}
				case "high":
					if val, ok := item[i].(float64); ok {
						klineData.High = val
					}
				case "low":
					if val, ok := item[i].(float64); ok {
						klineData.Low = val
					}
				case "close":
					if val, ok := item[i].(float64); ok {
						klineData.Close = val
					}
				case "chg":
					if val, ok := item[i].(float64); ok {
						klineData.Chg = val
					}
				case "percent":
					if val, ok := item[i].(float64); ok {
						klineData.Percent = val
					}
				case "turnoverrate":
					if val, ok := item[i].(float64); ok {
						klineData.TurnoverRate = val
					}
				case "amount":
					if val, ok := item[i].(float64); ok {
						klineData.Amount = val
					}
				case "volume_post":
					// 盘后数据可能为null
				case "amount_post":
					// 盘后数据可能为null
				}
			}
			klineDatas = append(klineDatas, klineData)
		}
	}

	//g.Dump(klineDatas)

	// 查找指定日期的数据
	targetDate := t.Format("2006-01-02")
	for _, klineData := range klineDatas {
		itemDate := time.Unix(klineData.Timestamp/1000, 0).Format("2006-01-02")
		//g.Dump()
		if itemDate == targetDate {
			g.Log().Infof(nil, "找到匹配日期数据: %s", itemDate)
			return &klineData, nil
		}
	}
	return nil, fmt.Errorf("未找到指定日期 %s 的数据", date)
}

// FormatTimestamp 格式化时间戳
func (qi *QuoteInfo) FormatTimestamp() string {
	if qi.Timestamp == 0 {
		return ""
	}
	t := time.Unix(qi.Timestamp/1000, 0)
	return t.Format("2006-01-02 15:04:05")
}
