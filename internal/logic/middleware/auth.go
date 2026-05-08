package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
	"star/internal/consts"
)

type jwtClaims struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Auth(r *ghttp.Request) {
	var tokenString = r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(consts.JwtKey), nil
	})
	if err != nil || !token.Valid {
		result := ghttp.DefaultHandlerResponse{
			Code:    403,
			Message: "token不合法",
			Data:    "",
		}
		r.Response.WriteJson(result)
		//r.Response.WriteStatus(http.StatusForbidden)
		r.Exit()
	}

	//tokenString := g.RequestFromCtx(ctx).Request.Header.Get("Authorization")
	tokenClaims, _ := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(consts.JwtKey), nil
	})

	if claims, ok := tokenClaims.Claims.(*jwtClaims); ok && tokenClaims.Valid {
		r.SetCtxVar("userId", claims.Id)
		r.SetCtxVar("username", claims.Username)
	}

	r.Middleware.Next()
}
