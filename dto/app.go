package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/public"
	"time"
)

type APPListInput struct {
	Info     string `json:"info" form:"info" comment:"查找信息" validate:""`
	PageSize int    `json:"page_size" form:"page_size" comment:"页数" validate:"required,min=1,max=999"`
	PageNo   int    `json:"page_no" form:"page_no" comment:"页码" validate:"required,min=1,max=999"`
}

func (params *APPListInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type APPListItemOutput struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配		"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	RealQpd   int64     `json:"real_qpd" description:"日请求量限制"`
	RealQps   int64     `json:"real_qps" description:"每秒请求量限制"`
	UpdatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	CreatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

type APPListOutput struct {
	List  []APPListItemOutput `json:"list" form:"list" comment:"租户列表"`
	Total int64               `json:"total" form:"total" comment:"租户总数"`
}

type APPDetailInput struct {
	ID int64 `json:"id" form:"id" comment:"租户ID" validate:"required"`
}

func (params *APPDetailInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}

type APPDeleteInput struct {
	ID int64 `json:"id" form:"id" comment:"租户ID" validate:"required"`
}

func (params *APPDeleteInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}
