package main

import (
	"fmt"
	"time"
)

func main() {
	// ==================== 1. 基础协程 ====================
	//fmt.Println("=== 示例 1: 基础协程 ===")
	//go sayHello("张三") // 启动协程（异步执行）
	//sayHello("李四")    // 主协程继续执行
	//time.Sleep(100 * time.Millisecond) // 等待协程完成

	// ==================== 2. 带缓冲的 channel ====================
	//fmt.Println("\n=== 示例 2: 带缓冲 channel ===")
	//ch := make(chan string, 2) // 创建容量为 2 的缓冲 channel
	//ch <- "任务 1"              // 发送数据（不会阻塞）
	//ch <- "任务 2"
	//fmt.Println(<-ch) // 接收：任务 1
	//fmt.Println(<-ch) // 接收：任务 2

	// ==================== 3. 多协程通信 ====================
	//fmt.Println("\n=== 示例 3: 多协程通信 ===")
	//msgCh := make(chan string)
	//go func() {
	//	msgCh <- "你好，这是来自协程的消息"
	//}()
	//// 主协程接收消息（会阻塞直到有数据）
	//fmt.Println(<-msgCh)

	// ==================== 4. Worker Pool 模式 ====================
	fmt.Println("\n=== 示例 4: Worker Pool ===")
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	// 启动 3 个 worker
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// 发送 5 个任务
	for j := 1; j <= 5; j++ {
		jobs <- j
	}
	close(jobs)  // ⭐ 关键：发送完所有任务后立即关闭 channel

	// 收集结果
	for r := 1; r <= 5; r++ {
		fmt.Printf("Worker 处理结果：%d\n", <-results)
	}

	// ==================== 5. select 多路复用 ====================
	fmt.Println("\n=== 示例 5: select 多路复用 ===")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "来自 ch1 的消息"
		close(ch1)  // 发送完成后关闭
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "来自 ch2 的消息"
		close(ch2)  // 发送完成后关闭
	}()

	// 等待任意一个 channel 的消息
	for i := 0; i < 2; i++ {
		select {
		case msg1, ok := <-ch1:
			if !ok {
				fmt.Println("ch1 已关闭")
			} else {
				fmt.Println(msg1)
			}
		case msg2, ok := <-ch2:
			if !ok {
				fmt.Println("ch2 已关闭")
			} else {
				fmt.Println(msg2)
			}
		}
	}

	// ==================== 6. channel 关闭的正确时机 ====================
	fmt.Println("\n=== 示例 6: channel 关闭时机演示 ===")
	demoCloseChannel()
}

func sayHello(name string) {
	fmt.Printf("你好，%s!\n", name)
}

// worker 处理任务
func worker(id int, jobs <-chan int, results chan<- int) {
	for job := range jobs {  // ⭐ 自动检测 channel 关闭
		fmt.Printf("Worker %d 开始处理任务 %d\n", id, job)
		time.Sleep(50 * time.Millisecond) // 模拟耗时操作
		results <- job * 2                // 返回结果
		fmt.Printf("Worker %d 完成任务 %d\n", id, job)
	}
	fmt.Printf("Worker %d: jobs channel 已关闭，退出循环\n", id)
}

// demoCloseChannel 演示 channel 关闭的正确时机
func demoCloseChannel() {
	// ✅ 正确做法：发送方关闭 channel
	ch := make(chan int, 5)
	
	// 生产者：发送数据并关闭
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i
			fmt.Printf("发送数据：%d\n", i)
		}
		close(ch)  // ⭐ 发送完成后立即关闭
		fmt.Println("生产者：channel 已关闭")
	}()
	
	// 消费者：接收数据
	for data := range ch {  // ⭐ 自动检测关闭
		fmt.Printf("消费者收到：%d\n", data)
	}
	fmt.Println("消费者：channel 已关闭，退出循环")
	
	// ❌ 错误做法 1：重复关闭会 panic
	// close(ch)  // panic: close of closed channel
	
	// ❌ 错误做法 2：接收方关闭（多生产者时会 panic）
	// 如果有多个生产者，接收方关闭会导致某个生产者 panic
}
