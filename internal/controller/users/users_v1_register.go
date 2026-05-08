package users

import (
	"context"
	"star/api/users/v1"
)

func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	err = c.users.Register(ctx, req.Username, req.Password, req.Email)
	if err != nil {
		return nil, err
	}
	return res, nil
}
