package users

import (
	"context"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"star/internal/dao"
	"star/internal/model/do"
)

func (u *Users) Register(ctx context.Context, username, password, email string) (err error) {
	err = u.CheckUsername(ctx, username)
	if err != nil {
		return
	}
	_, err = dao.Users.Ctx(ctx).Data(do.Users{
		Username: username,
		Password: u.encryptPassword(password),
		Email:    email,
	}).Insert()
	if err != nil {
		return
	}
	return
}

func (u *Users) CheckUsername(ctx context.Context, username string) error {

	count, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Username, username).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("用户名已存在")
	}
	return nil
}

func (u *Users) encryptPassword(password string) string {
	return gmd5.MustEncryptString(password)
}
