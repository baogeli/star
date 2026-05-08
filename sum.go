package main

import (
	"fmt"
	"math"
	"runtime"
	"sync"
)

// func busy(ch chan bool, i int) {
func busy(i int) {
	fmt.Println("go func ", i, " goroutine count = ", runtime.NumGoroutine())
}

func main() {

	taskNums := math.MaxInt8 // 127
	var wg sync.WaitGroup

	fmt.Println("启动前协程数:", runtime.NumGoroutine())

	for i := 0; i < taskNums; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 这里的数量会接近 128 (127个子协程 + 1个main协程)
			fmt.Printf("任务 %d 执行中, 当前总协程数: %d\n", i, runtime.NumGoroutine())
		}(i)
	}

	wg.Wait()

	// 所有子协程结束后，数量会回落到初始值
	fmt.Println("结束后协程数:", runtime.NumGoroutine())

	//taskNums := math.MaxInt8
	//fmt.Println(taskNums)
	//var wg sync.WaitGroup
	//for i := 0; i < taskNums; i++ {
	//	wg.Add(1)
	//	go func(i int) {
	//		defer wg.Done()
	//		fmt.Println("go func ", i, " goroutine count = ", runtime.NumGoroutine())
	//	}(i)
	//}
	//wg.Wait()
}
