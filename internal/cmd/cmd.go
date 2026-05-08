package cmd

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"star/api/investment"
	"star/api/reit"
	"star/internal/logic/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Middleware(ghttp.MiddlewareCORS)
				group.Group("/v1", func(group *ghttp.RouterGroup) {
					group.Bind(
						reit.ReitController{},
						investment.ForecastingController{},
						//users.NewV1(),
						//new(users.ControllerV1),
					)
					group.ALL("show", UploadShow)
					//group.Group("/", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.Auth)
					group.Bind(
					//account.NewV1(),
					//words.NewV1(),
					)
					//})
				})
			})
			s.Run()
			return nil
		},
	}
)

// UploadShow shows uploading simgle file page.
func UploadShow(r *ghttp.Request) {
	r.Response.Write(`
    <html>
    <head>
        <title>GoFrame Upload File Demo</title>
    </head>
        <body>
            <form enctype="multipart/form-data" action="/v1/upload/history" method="post">
                <input type="file" name="file" />
                <input type="submit" value="upload" />
            </form>
        </body>
    </html>
    `)
}
