package users

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"
	"star/internal/consts"
	"star/internal/dao"
	"star/internal/model/entity"
	"time"
)

type jwtClaims struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (u *Users) Login(ctx context.Context, username, password string) (tokenString string, err error) {
	var user *entity.Users
	err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Username, username).
		Where(dao.Users.Columns().Password, u.encryptPassword(password)).
		Scan(&user)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", gerror.New("用户名或密码错误")
	}
	uc := &jwtClaims{
		Id:       user.Id,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	return token.SignedString([]byte(consts.JwtKey))
}

func (u *Users) Info(ctx context.Context) (user *entity.Users, err error) {
	userId := ctx.Value("userId")
	err = dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	return
}
