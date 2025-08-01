package controllers

import (
	"io"
	"net/http"
	"shared-charge/config"
	"shared-charge/service"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// UploadImage 上传图片
// @Summary 上传图片
// @Description 用户上传图片文件
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} map[string]interface{}
// @Router /upload/image [post]
func UploadImage(c *gin.Context) {
	utils.InfoCtx(c, "上传图片请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "上传图片未认证")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorCtx(c, "获取文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取文件失败"})
		return
	}

	// 添加文件大小预检查
	maxFileSize := config.GetConfig().App.MaxFileSize
	if file.Size > maxFileSize {
		utils.WarnCtx(c, "文件大小超过限制: %d > %d", file.Size, maxFileSize)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "文件大小超过限制"})
		return
	}

	result, err := service.SaveUploadImage(c, userModel.ID, file)
	if err != nil {
		utils.ErrorCtx(c, "上传图片失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "上传图片失败"})
		return
	}
	utils.InfoCtx(c, "图片上传成功: user_id=%d, filename=%s", userModel.ID, file.Filename)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "图片上传成功", "data": result})
}

// GetImage 读取图片
// @Summary 获取图片
// @Description 通过文件名获取图片内容
// @Tags 文件
// @Produce octet-stream
// @Param filename path string true "图片文件名"
// @Success 200 {file} file
// @Failure 404 {string} string "文件不存在"
// @Router /api/image/{filename} [get]
func GetImage(c *gin.Context) {
	filename := c.Param("filename")
	obj, contentType, err := service.GetImageObject(filename)
	if err != nil {
		c.String(404, "文件不存在")
		return
	}
	defer obj.Close()
	c.Header("Content-Type", contentType)
	c.Status(200)
	_, err = io.Copy(c.Writer, obj)
	if err != nil {
		c.String(500, "文件读取失败")
		return
	}
}
