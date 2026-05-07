package relay

import (
	"net/http"

	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// DoImageResponseWithEarlyFlush 直接调用 adaptor.DoResponse。
// 实际的 early flush 逻辑已移至 service.IOCopyBytesGracefully 中，
// 在 WriteHeader 后立即 Flush，避免 NAT 超时导致的 broken pipe。
func DoImageResponseWithEarlyFlush(c *gin.Context, adaptor channel.Adaptor, resp *http.Response, info *relaycommon.RelayInfo) (any, *types.NewAPIError) {
	return adaptor.DoResponse(c, resp, info)
}
