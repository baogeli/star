package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"star/internal/cmd"
	_ "star/internal/packed"
)

func main() {
	// 全局设置 i18n
	g.I18n().SetLanguage("zh-CN")
	
	// 初始化日志配置（从配置文件读取）
	glog.SetPath("log")
	glog.SetFile("{Y-m-d}.log")
	
	cmd.Main.Run(gctx.GetInitCtx())
}
