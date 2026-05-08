package note

/*

一、程序启动总览
	操作系统加载程序
		↓
	rt0_go (汇编入口)
			文件: /usr/local/go/src/runtime/asm_amd64.s
			1. 操作系统加载可执行文件
			2. 跳转到 rt0_go (运行时入口)
			3. 调用 osinit - 获取CPU核数等系统信息
			4. 调用 schedinit - 初始化调度器
		↓
	schedinit (调度器初始化)
			文件: proc.go 第798行
			1. 初始化各种锁
			lockInit(&sched.lock, ...)
			2. 初始化内存分配器
			mallocinit()
			3. 初始化当前M (m0，即主线程)
			mcommoninit(gp.m, -1)
			4. 确定P的数量（默认=CPU核数）
			procs := ncpu
			if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok {
				procs = n  // 可通过环境变量设置
			}
			5. 创建并初始化P（处理器）
			procresize(procs)  // ← 关键！创建P数组
			6. 初始化全局变量
			sched.maxmcount = 10000  // 最多1万个线程
			结论：
				此时创建了：
				m0: 第一个M（主线程），全局变量
				g0: m0的系统栈，用于调度
				P数组: 数量为GOMAXPROCS，初始都空闲
		↓
	创建main goroutine
			文件: proc.go 的 newproc() → newproc1()
			在 rt0_go 中 newproc(main_main)  // 创建运行main函数的goroutine
			newproc1 内部做：
			1. 从缓存或堆分配一个G结构体
			gp := mallocg()
			2. 设置G的状态
			gp.status = _Grunnable  // 可运行状态
			gp.fn = fn              // 要执行的函数
			gp.goid = goidgen++     // 分配ID（main是G1）
			3. 分配栈空间（默认2KB）
			stackalloc()
			4. 放入当前P的本地运行队列
			runqput(pp, gp, false)
			结论：
				此时有了：
				G1: main goroutine，在P的运行队列中等待执行
		↓
	mstart (启动M0)
			文件: asm_amd64.s 和 proc.go
			rt0_go (汇编)
				↓
			调用 mstart()  ← 启动m0
				↓
			mstart1()
				↓
			schedule()  ← 进入调度循环
				↓
			findRunnable()  ← 找到G1（main goroutine）
				↓
			execute(G1)  ← 切换到G1的栈执行
				↓
			gogo(&G1.sched)  ← 汇编切换上下文
				↓
			开始执行 runtime.main
		↓
	runtime.main (Go主函数)
			文件: proc.go 第147行
			1. 允许创建新的M
			mainStarted = true

			2. 启动系统监控线程（sysmon）
			if haveSysmon {
				newm(sysmon, nil, -1)  // 创建后台监控M
			}

			3. 锁定到主线程
			lockOSThread()

			4. 初始化包（执行所有init函数）
			doInit(runtime_inittasks)

			5. 启动GC
			gcenable()

			6. 执行用户的main函数
			main_main()  // ← 你的代码从这里开始！

			7. main返回后退出
			exit(0)
		↓
	用户main函数
		↓
	开始调度循环

*/
