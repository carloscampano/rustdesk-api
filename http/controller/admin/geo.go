package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/utils"
)

type Geo struct{}

type geoLookupForm struct {
	IPs []string `json:"ips"`
}

func (g *Geo) Lookup(c *gin.Context) {
	f := &geoLookupForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	response.Success(c, utils.LookupGeoBatch(f.IPs))
}
