package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success writes a standard 200 success response.
func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// SuccessWithStatus writes a success response with a custom HTTP status code.
func SuccessWithStatus(ctx *gin.Context, status int, data interface{}) {
	ctx.JSON(status, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// PaginatedSuccess writes a paginated list success response.
func PaginatedSuccess(ctx *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list": list,
			"pagination": gin.H{
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": totalPages,
				"has_next":    page < totalPages,
				"has_prev":    page > 1,
			},
		},
	})
}
