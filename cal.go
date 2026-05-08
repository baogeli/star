package main

import (
	"context"
	"fmt"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"sort"
	"star/internal/dao"
	"star/internal/model/entity"
)

type d struct {
	FundCode     int     `json:"fund_code"`
	BuyingPrice  float64 `json:"buying_price"`
	CurrentPrice float64 `json:"current_price"`
}

// CRC16 实现 CRC-16/MODBUS 算法，用于 Redis 集群槽位计算
func CRC16(key string) uint16 {
	crc := uint16(0xFFFF)
	poly := uint16(0xA001) // CRC-16/MODBUS 多项式
	//g.Dump(poly)

	for i := 0; i < len(key); i++ {
		crc ^= uint16(key[i])
		for j := 0; j < 8; j++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

type UserInfo struct {
	Name string `json:"name"`
}

func main() {

	// 写个示例帮我理解下 golang 中 make 和 new 关键字的区别

	// ==================== new 关键字 ====================
	// new(T) 返回 *T，指向类型 T 的零值
	// 适用于：基本类型、结构体、数组等

	ptrInt := new(int) // *int，值为 0
	*ptrInt = 100      // 赋值后为 100
	fmt.Println("new(int):", ptrInt, *ptrInt)

	type Person struct {
		Name string
		Age  int
	}
	ptrPerson := new(Person) // *Person，字段为零值
	ptrPerson.Name = "张三"    // 可以直接访问字段
	ptrPerson.Age = 25
	fmt.Println("new(Person):", ptrPerson)

	// ==================== make 关键字 ====================
	// make(T, args...) 返回 T（不是指针），只用于 slice、map、channel
	// 会进行内存分配和初始化

	// 1. 创建 slice
	slice := make([]int, 3, 5) // []int{0, 0, 0}，长度 3，容量 5
	slice[0] = 10
	slice[1] = 20
	fmt.Println("make slice:", slice, "len=", len(slice), "cap=", cap(slice))

	// 2. 创建 map（必须用 make，不能用 new）
	m := make(map[string]int)
	m["age"] = 18
	m["score"] = 95
	fmt.Println("make map:", m)

	// 错误示范：用 new 创建 map
	// nilMap := new(map[string]int)  // 编译通过，但运行时会 panic
	// nilMap["key"] = 1  // panic: assignment to entry in nil map

	// 3. 创建 channel
	ch := make(chan int, 1) // 带缓冲的 channel
	ch <- 42
	val := <-ch
	g.Dump("make channel:", val)

	// ==================== 对比总结 ====================
	/*
		| 特性       | new                          | make                        |
		|------------|------------------------------|-----------------------------|
		| 返回值     | 指针 (*T)                     | 类型本身 (T)                 |
		| 用途       | 基本类型、结构体、数组          | 仅 slice、map、channel      |
		| 初始化     | 返回零值                       | 分配内存并初始化              |
		| 典型场景   | new(int), new(Person)         | make([]int, 3), make(map)  |
	*/

	//u := new(UserInfo)
	//u.Name = "test"
	//pd := &u
	//g.Dump(pd)

	//var p *int

	//userInfo := make(map[string]interface{})
	//userInfo["name"] = 3.14
	//if _, ok := userInfo["name"].(string); ok {
	//	g.Dump("string")
	//} else {
	//	g.Dump("else")
	//}
	return
	//g.Dump(userInfo["name"].(string))
	//return

	// 实现 CRC16
	//crc16 := CRC16("test_key")
	/*
		redis 主从同步
		一。全量同步
			从节点发送 psync ? -1 命令，触发同步
			主节点会判断是全量同步，还是增量同步 这是由伟来的run id和offset来判断的
			主节点bg save 生成rdb 文件, 在这个过程中收到主节点收到新写入的数据时，会把数据先写入 replication buffer
			从节点收到rdb文件，清空自己的数据，加载rdb数据
			主节点补发增量命令,把复制缓冲区中的命令发送给从节点
			从节点会通过持续上报复制偏移量（offset）来反映同步进度
			最后主从节点会保持着长连接，并持续接收主节点的命令
		二。增量同步
			当从节点因为一些网络的不稳定，与主节点断开了连接，这段时间
	*/

	/*

		def redis_cluster_principle():
		    """
		    Redis集群实现原理说明
		    """
		    # 1. 数据分片与哈希槽
		    # Redis集群将数据空间划分为16384个哈希槽（slot），每个键通过CRC16算法计算哈希值后对16384取模确定所属槽位。
		    # 每个节点负责一部分槽位，实现数据的分布式存储。

		    # 2. 节点通信与Gossip协议
		    # 节点之间通过Gossip协议交换信息，包括心跳检测、故障报告、槽分配等。
		    # 每个节点维护集群状态，包括其他节点的IP、端口、状态等信息。

		    # 3. 高可用性与故障转移
		    # 集群中每个主节点可以有多个从节点，当主节点故障时，从节点可自动晋升为主节点。
		    # 故障检测通过半数哨兵确认主节点下线后触发故障转移。

		    # 4. 客户端重定向机制
		    # 客户端向任意节点发送请求，若目标键不在该节点，则返回MOVED或ASK错误，引导客户端访问正确节点。

		    # 5. 扩容与缩容
		    # 可通过添加新节点并迁移槽位实现扩容；删除节点时需将槽位迁移至其他节点。

		    print("Redis集群通过分片、Gossip协议、主从复制和故障转移机制实现高可用和分布式存储。")

		if __name__ == "__main__":
		    redis_cluster_principle()
	*/
	ctx := context.Background()
	fundCodes := []string{
		"508008", "508033", "508086", "180201", // 基金组合日收益
		//"932047", // benchMark 日收益
	}

	var allData []*entity.PyFund
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

	for fundCode, datePrices := range fundDatePrices {

		dates := make([]string, 0, len(datePrices))
		for date := range datePrices {
			dates = append(dates, date)
		}
		//排序日期 [2025-12-08, 2025-12-09, ..., 2026-03-26]
		sort.Strings(dates)

		// 计算每日的收益率（从第 2 天开始）
		for i := 1; i < len(dates); i++ {
			currentDate := dates[i] // 今天
			prevDate := dates[i-1]  // 昨天
			currentPrice := datePrices[currentDate]
			prevPrice := datePrices[prevDate]

			// 日收益率 = (当日净值 - 前一日净值) / 前一日净值
			dailyReturn := (currentPrice - prevPrice) / prevPrice

			// 存储到日期 - 基金映射中
			if _, ok := dateFundMap[currentDate]; !ok {
				dateFundMap[currentDate] = make(map[string]float64)
			}
			dateFundMap[currentDate][fundCode] = dailyReturn
		}
	}

	//jsonData1, err := json.MarshalIndent(dateFundMap, "", "  ")
	//if err != nil {
	//	fmt.Printf("JSON 序列化失败：%v\n", err)
	//	return
	//}
	//// 写入文件
	//filePath1 := "date_fund_map.json"
	//err = os.WriteFile(filePath1, jsonData1, 0644)
	//if err != nil {
	//	fmt.Printf("写入文件失败：%v\n", err)
	//	return
	//}
	/*
		bitmap 16384
	*/
	dates := make([]string, 0, len(dateFundMap))
	for date := range dateFundMap {
		dates = append(dates, date)
		sort.Strings(dates)
	}

	portfolioReturns := make([]float64, 0, len(dates))
	for _, date := range dates {
		fundReturns := dateFundMap[date]
		var ret float64
		for _, v := range fundReturns {
			ret += v
		}
		res := ret / float64(len(fundReturns))
		portfolioReturns = append(portfolioReturns, res)
	}
	g.Dump(portfolioReturns)

	/*
		g.Dump(dateFundMap)   dateFundMap struct
		"2025-12-15": {
			"508033": -0.0061533606816029825,
			"508086": -0.016922471467925897,
			"180201": -0.010523826415986804,
			"508008": -0.001255177607631453,
		},
		...
		...
		"2025-12-16": {
			"508033": -0.0061533606816029825,
			"508086": -0.016922471467925897,
			"180201": -0.010523826415986804,
			"508008": -0.001255177607631453,
		},
	*/

	//a508008 := decimal.NewFromFloat(3000).Mul(decimal.NewFromFloat(8.225))
	//b508033 := decimal.NewFromFloat(3800).Mul(decimal.NewFromFloat(6.495))
	//c508086 := decimal.NewFromFloat(4600).Mul(decimal.NewFromFloat(5.088))
	//d180201 := decimal.NewFromFloat(2800).Mul(decimal.NewFromFloat(8.184))

	//a508008 := decimal.NewFromFloat(3000).Mul(decimal.NewFromFloat(8.0040))
	//b508033 := decimal.NewFromFloat(3800).Mul(decimal.NewFromFloat(6.3970))
	//c508086 := decimal.NewFromFloat(4600).Mul(decimal.NewFromFloat(5.1610))
	//d180201 := decimal.NewFromFloat(2800).Mul(decimal.NewFromFloat(8.4770))

	//sum := a508008.Add(b508033).Add(c508086).Add(d180201)
	//net, _ := sum.Float64()
	//total := net + 4204.59
	//fmt.Println(sum.StringFixed(4)) // 保留 4 位小数
	//fmt.Println(total)
}
