package relay

import (
	"net/http"

	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// DoImageResponseWithEarlyFlush 在调用 adaptor.DoResponse 前提前写出 HTTP 200
// 响应头并 Flush，使 TCP 连接在上游 io.ReadAll 等待期间保持活跃，
// 避免中间层（NAT/防火墙）因长时间无数据而关闭连接（broken pipe）。
func DoImageResponseWithEarlyFlush(c *gin.Context, adaptor channel.Adaptor, resp *http.Response, info *relaycommon.RelayInfo) (any, *types.NewAPIError) {
	c.Writer.WriteHeader(http.StatusOK)
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}
	return adaptor.DoResponse(c, resp, info)
}
