package users

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"star/api/users/v1"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	token, err := c.users.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{
		Token: token,
	}, nil
}

type GetMessageReq struct {
	g.Meta `path:"/get_message" method:"get" sm:"获取一条消息" tags:"用户"`
	//Username string `json:"username"`
	//Password string `json:"password"`
	//Email    string `json:"email"`
	//Username string `json:"username" v:"required|length:3,12"`
	//Password string `json:"password" v:"required|length:6,16"`
	//Email    string `json:"email" v:"required|email"`
}

type GetMessageRes struct {
	Message string `json:"message"`
}

func (c *ControllerV1) GetMessage(ctx context.Context, req *GetMessageReq) (res *GetMessageRes, err error) {
	return &GetMessageRes{
		Message: "Hello World",
	}, nil
}
