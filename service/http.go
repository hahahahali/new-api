package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"github.com/gin-gonic/gin"
)

func CloseResponseBodyGracefully(httpResponse *http.Response) {
	if httpResponse == nil || httpResponse.Body == nil {
		return
	}
	err := httpResponse.Body.Close()
	if err != nil {
		common.SysError("failed to close response body: " + err.Error())
	}
}

func IOCopyBytesGracefully(c *gin.Context, src *http.Response, data []byte) {
	if c.Writer == nil {
		return
	}

	body := io.NopCloser(bytes.NewBuffer(data))

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	if src != nil {
		for k, v := range src.Header {
			// avoid setting Content-Length
			if k == "Content-Length" {
				continue
			}
			c.Writer.Header().Set(k, v[0])
		}
	}

	// set Content-Length header manually BEFORE calling WriteHeader
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Write header with status code (this sends the headers)
	statusCode := http.StatusOK
	if src != nil {
		statusCode = src.StatusCode
		c.Writer.WriteHeader(src.StatusCode)
	} else {
		c.Writer.WriteHeader(http.StatusOK)
	}

	// Flush immediately after WriteHeader to keep TCP alive during long io.Copy
	flushed := false
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
		flushed = true
	}
	logger.LogInfo(c, fmt.Sprintf("IOCopyBytesGracefully: WriteHeader(%d) + Flush(success=%v), body_size=%d", statusCode, flushed, len(data)))

	_, err := io.Copy(c.Writer, body)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("failed to copy response body: %s", err.Error()))
	} else {
		logger.LogInfo(c, fmt.Sprintf("IOCopyBytesGracefully: io.Copy completed successfully, body_size=%d", len(data)))
	}
	c.Writer.Flush()
}
