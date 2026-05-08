package note

/*


// One round of scheduler: find a runnable goroutine and execute it.
// 一轮调度：找到一个可运行的goroutine并执行它。
// Never returns.
// 永不返回
func schedule() {
 	-- mp 全称 M Pointer
	mp := getg().m

	if mp.locks != 0 {
		throw("schedule: holding locks")
	}

	if mp.lockedg != 0 {
		-- 当前M（线程）被某个特定的G（goroutine）锁定
		-- 这个goroutine要求必须在这个特定的操作系统线程上运行

		stoplockedm()
		-- 解除当前M与P的关联
    	-- 因为锁定的goroutine可能需要特殊的处理
		-- 作用：处理M被G锁定的情况
		-- 场景：当 mp.lockedg != 0 时调用
		-- 功能：解除当前M与P的关联，因为锁定的goroutine需要特殊处理
		-- 位置：在 schedule() 函数开头，检查lockedg之后
		-- 执行流程
			释放P
			调用handoffp移交P
			阻塞等待
			恢复时获取新P
		execute(mp.lockedg.ptr(), false) // Never returns.
	}

	// We should not schedule away from a g that is executing a cgo call,
	// 我们不应该从一个正在执行 cgo 调用的 g（goroutine）上调度离开，
	// since the cgo call is using the m's g0 stack.
	// 因为该 cgo 调用正在使用 m 的 g0 栈。
	if mp.incgo {
		throw("schedule: in cgo")
	}

top:
	pp := mp.p.ptr()
	pp.preempt = false

	// Safety check: if we are spinning, the run queue should be empty.
	// Check this before calling checkTimers, as that might call
	// goready to put a ready goroutine on the local run queue.
	if mp.spinning && (pp.runnext != 0 || pp.runqhead != pp.runqtail) {
		throw("schedule: spinning with local work")
	}

	gp, inheritTime, tryWakeP := findRunnable() // blocks until work is available

	if debug.dontfreezetheworld > 0 && freezing.Load() {
		// See comment in freezetheworld. We don't want to perturb
		// scheduler state, so we didn't gcstopm in findRunnable, but
		// also don't want to allow new goroutines to run.
		//
		// Deadlock here rather than in the findRunnable loop so if
		// findRunnable is stuck in a loop we don't perturb that
		// either.
		lock(&deadlock)
		lock(&deadlock)
	}

	// This thread is going to run a goroutine and is not spinning anymore,
	// so if it was marked as spinning we need to reset it now and potentially
	// start a new spinning M.
	if mp.spinning {
		resetspinning()
	}

	if sched.disable.user && !schedEnabled(gp) {
		// Scheduling of this goroutine is disabled. Put it on
		// the list of pending runnable goroutines for when we
		// re-enable user scheduling and look again.
		lock(&sched.lock)
		if schedEnabled(gp) {
			// Something re-enabled scheduling while we
			// were acquiring the lock.
			unlock(&sched.lock)
		} else {
			sched.disable.runnable.pushBack(gp)
			sched.disable.n++
			unlock(&sched.lock)
			goto top
		}
	}

	// If about to schedule a not-normal goroutine (a GCworker or tracereader),
	// wake a P if there is one.
	if tryWakeP {
		wakep()
	}
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}

	execute(gp, inheritTime)
}



*/

/*
一、两个函数的关系
	mstart() 和 mstart1() 详解
	一、两个函数的关系
	mstart() (汇编)
		↓ 调用
	mstart0() (Go代码)
		↓ 初始化栈信息
	mstart1() (Go代码)
		↓ 绑定P（如果不是m0）
	schedule() (进入调度循环)


二、mstart() - 汇编入口
	文件: asm_amd64.s
	TEXT runtime·mstart(SB), NOSPLIT|TOPFRAME|NOFRAME, $0
    CALL    runtime·mstart0(SB)  // 直接调用mstart0
    RET  // 永远不会执行到这里
	跳板函数，从汇编代码切换到Go代码

	mstart0() (Go代码)
		gp := getg()  // 获取当前g（此时是g0）
		// 1. 初始化栈边界（如果是操作系统分配的栈）
		osStack := gp.stack.lo == 0
		if osStack {
			size := gp.stack.hi
			if size == 0 {
				size = 16384 * sys.StackGuardMultiplier
			}
			gp.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
			gp.stack.lo = gp.stack.hi - size + 1024
		}
		// 2. 设置栈保护标志（用于栈溢出检测）
		gp.stackguard0 = gp.stack.lo + stackGuard
		gp.stackguard1 = gp.stackguard0

		// 3. 调用mstart1继续初始化
		mstart1()

		// 4. mstart1返回后，退出线程（正常情况下不会到这里）
		mexit(osStack)

	其中mstart1()函数：

		gp := getg()

		// 1. 安全检查：必须在g0上运行
		if gp != gp.m.g0 {
			throw("bad runtime·mstart")
		}

		// 2. 保存调度上下文（用于后续goexit返回）
		gp.sched.g = guintptr(unsafe.Pointer(gp))
		gp.sched.pc = sys.GetCallerPC()
		gp.sched.sp = sys.GetCallerSP()

		// 3. 汇编层面的初始化
		asminit()

		// 4. 最小化初始化（信号处理等）
		minit()

		// 5. 如果是m0（主线程），做一些特殊设置
		if gp.m == &m0 {
			mstartm0()  // 创建额外的M用于CGO回调等
		}

		// 6. 执行M的启动函数（如果有）
		if fn := gp.m.mstartfn; fn != nil {
			fn()
		}

		// 7. ★★★ 关键：绑定P（非m0的情况）★★★
		if gp.m != &m0 {
			acquirep(gp.m.nextp.ptr())  // ← 这里绑定P！
			gp.m.nextp = 0
		}

		// 8. 进入调度循环
		schedule()

		-- 流程结束，看下第七条绑定P	acquirep(gp.m.nextp.ptr())  // ← 这里绑定P！

		// 第5876行  acquire / əˈkwaɪər / 获取
		func acquirep(pp *p) {
			wirep(pp)  // 真正的绑定逻辑
			// ...
		}

		// 第5899行
		func wirep(pp *p) {
			gp := getg()

			// ★★★ 双向绑定 ★★★
			gp.m.p.set(pp)   // M的p字段指向P
			pp.m.set(gp.m)   // P的m字段指向M
			pp.status = _Prunning  // P状态设为运行中
	}

*/
