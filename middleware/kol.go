package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

// KolAuth checks that the authenticated user belongs to the "kol" group.
// Must be used after UserAuth().
func KolAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		group, exists := c.Get("group")
		if !exists || group == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "无权访问",
			})
			c.Abort()
			return
		}
		groupStr, ok := group.(string)
		if !ok || groupStr != "kol" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "仅达人可访问",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ReviewerOrRootAuth allows access only to users in the "reviewer" group or with role >= RoleRootUser.
// Must be used after UserAuth().
func ReviewerOrRootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if r, ok := role.(int); ok && r >= common.RoleRootUser {
			c.Next()
			return
		}
		group, _ := c.Get("group")
		if g, ok := group.(string); ok && g == "reviewer" {
			c.Next()
			return
		}
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权访问",
		})
		c.Abort()
	}
}
