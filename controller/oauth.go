package controller

import (
	"encoding/base64"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/dto"
	"github.com/jmdrws/go_gateway/golang_common/lib"
	"github.com/jmdrws/go_gateway/middleware"
	"github.com/jmdrws/go_gateway/public"
	"strings"
	"time"
)

type OAuthController struct {
}

func OAuthRegister(group *gin.RouterGroup) {
	oauth := &OAuthController{}
	group.POST("/tokens", oauth.Tokens)
}

// Tokens godoc
// @Summary 获取TOKEN
// @Description 获取TOKEN
// @Tags OAUTH
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (oauth *OAuthController) Tokens(c *gin.Context) {
	params := &dto.TokensInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	middleware.ResponseSuccess(c, &dto.TokensOutput{})
	//权限
	splits := strings.Split(c.GetHeader("Authorization"), " ")
	if len(splits) != 2 {
		middleware.ResponseError(c, 2001, errors.New("用户名或密码格式错误"))
		return
	}
	//基于base64解密
	appSecret, err := base64.StdEncoding.DecodeString(splits[1])
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//fmt.Println("splits[1]", splits[1])
	//fmt.Println("appSecret", string(appSecret))

	//  取出 app_id secret
	parts := strings.Split(string(appSecret), ":")
	if len(parts) != 2 {
		middleware.ResponseError(c, 2003, errors.New("用户名或密码格式错误"))
		return
	}
	//  生成 app_list
	appList := dao.AppManagerHandler.GetAppList()
	for _, appInfo := range appList {
		//  匹配 app_id
		if appInfo.AppID == parts[0] && appInfo.Secret == parts[1] {
			//  基于 jwt生成token 生成 output
			claims := jwt.StandardClaims{
				Issuer:    appInfo.AppID,
				ExpiresAt: time.Now().Add(public.JwtExpires * time.Second).In(lib.TimeLocation).Unix(),
			}
			token, err := public.JwtEncode(claims)
			if err != nil {
				middleware.ResponseError(c, 2004, err)
				return
			}
			output := &dto.TokensOutput{
				ExpiresIn:   public.JwtExpires,
				TokenType:   "Bearer",
				AccessToken: token,
				Scope:       "read_write",
			}
			middleware.ResponseSuccess(c, output)
			return
		}
	}
	middleware.ResponseError(c, 2005, errors.New("未匹配正确APP信息"))
}
