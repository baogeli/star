package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type LogEntry struct {
	Timestamp   string
	FundCode    string
	Date        string
	CurrPrice   float64
	PrevPrice   float64
	PriceChange float64
	DailyReturn float64
}

func main() {
	logFile := "/home/boa/golang/star/log/2026-04-03.log"
	outputFile := "/home/boa/golang/star/log/2026-04-03.xlsx"

	// 读取日志文件
	content, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("读取日志文件失败：%v\n", err)
		return
	}

	// 解析日志
	entries := parseLogContent(string(content))

	// 创建 Excel 文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// 创建工作表
	_, err = f.NewSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 设置表头
	headers := []string{
		"时间戳",
		"基金代码",
		"日期",
		"当日收盘价",
		"前日收盘价",
		"价格变化",
		"日收益率",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}

	// 设置列宽
	f.SetColWidth("Sheet1", "A", "G", 15)

	// 填充数据
	for i, entry := range entries {
		row := i + 2 // 从第 2 行开始（第 1 行是表头）
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), entry.Timestamp)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), entry.FundCode)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), entry.Date)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), entry.CurrPrice)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), entry.PrevPrice)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), entry.PriceChange)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), entry.DailyReturn)
	}

	// 删除默认的工作表
	f.DeleteSheet(f.GetSheetName(0))

	// 保存文件
	if err := f.SaveAs(outputFile); err != nil {
		fmt.Printf("保存 Excel 文件失败：%v\n", err)
		return
	}

	fmt.Printf("成功将 %d 条记录写入到 %s\n", len(entries), outputFile)
}

func parseLogContent(content string) []LogEntry {
	var entries []LogEntry

	// 正则表达式匹配日志格式
	pattern := regexp.MustCompile(`^\S+\s+\[INFO\]\s+基金:(\d+)\s+日期:(\d+-\d+-\d+)\s+当日收盘价:([\d.]+)\s+前日收盘价:([-?\d.]+)\s+价格变化:([-?\d.]+)\s+→日收益率:([-?\d.]+)`)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := pattern.FindStringSubmatch(line)
		if len(matches) == 7 {
			currPrice, _ := strconv.ParseFloat(matches[3], 64)
			prevPrice, _ := strconv.ParseFloat(matches[4], 64)
			priceChange, _ := strconv.ParseFloat(matches[5], 64)
			dailyReturn, _ := strconv.ParseFloat(matches[6], 64)

			// 提取时间戳（去掉时区部分）
			timestampParts := strings.Split(line, " ")
			timestamp := timestampParts[0]

			entry := LogEntry{
				Timestamp:   timestamp,
				FundCode:    matches[1],
				Date:        matches[2],
				CurrPrice:   currPrice,
				PrevPrice:   prevPrice,
				PriceChange: priceChange,
				DailyReturn: dailyReturn,
			}
			entries = append(entries, entry)
		}
	}

	return entries
}
