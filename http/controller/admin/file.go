package admin

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/lib/upload"
	"os"
	"time"
)

type File struct {
}

// allowedExtensions defines safe file extensions for upload
var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".txt": true, ".csv": true, ".zip": true,
}

// allowedMIMETypes defines valid MIME types for upload
var allowedMIMETypes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"text/plain": true, "text/csv": true,
	"application/zip": true,
}

// sanitizeFilename removes path components and validates extension
func sanitizeFilename(filename string) (string, error) {
	// Extract base name only (removes any path traversal attempts)
	base := filepath.Base(filename)

	// Reject if still contains path separators (shouldn't happen after Base)
	if strings.ContainsAny(base, `/\`) {
		return "", fmt.Errorf("invalid filename")
	}

	// Reject hidden files
	if strings.HasPrefix(base, ".") {
		return "", fmt.Errorf("hidden files not allowed")
	}

	// Check extension against allowlist
	ext := strings.ToLower(filepath.Ext(base))
	if ext == "" || !allowedExtensions[ext] {
		return "", fmt.Errorf("file extension not allowed: %s", ext)
	}

	return base, nil
}

// OssToken 文件
// @Tags 文件
// @Summary 获取ossToken
// @Description 获取ossToken
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/file/oss_token [get]
// @Security token
func (f *File) OssToken(c *gin.Context) {
	token := global.Oss.GetPolicyToken("")
	response.Success(c, token)
}

type FileBack struct {
	upload.CallbackBaseForm
	Url string `json:"url"`
}

// Notify 上传成功后回调
func (f *File) Notify(c *gin.Context) {

	res := global.Oss.Verify(c.Request)
	if !res {
		response.Fail(c, 101, response.TranslateMsg(c, "NoAccess"))
		return
	}
	fm := &FileBack{}
	if err := c.ShouldBind(fm); err != nil {
		fmt.Println(err)
	}
	fm.Url = global.Config.Oss.Host + "/" + fm.Filename
	response.Success(c, fm)

}

// Upload 上传文件到本地
// @Tags 文件
// @Summary 上传文件到本地
// @Description 上传文件到本地
// @Accept  multipart/form-data
// @Produce  json
// @Param file formData file true "上传文件示例"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/file/upload [post]
// @Security token
func (f *File) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "No file uploaded")
		return
	}

	// Security: Sanitize filename and validate extension
	safeFilename, err := sanitizeFilename(file.Filename)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	// Security: Validate MIME type by reading file header
	src, err := file.Open()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "Failed to read file")
		return
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, _ := src.Read(buffer)
	contentType := http.DetectContentType(buffer[:n])

	// Check against allowed MIME types
	if !allowedMIMETypes[contentType] {
		response.Fail(c, http.StatusBadRequest, fmt.Sprintf("File type not allowed: %s", contentType))
		return
	}

	// Security: Limit file size (10MB max)
	if file.Size > 10*1024*1024 {
		response.Fail(c, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	timePath := time.Now().Format("20060102") + "/"
	webPath := "/upload/" + timePath
	path := global.Config.Gin.ResourcesPath + webPath
	dst := path + safeFilename

	err = os.MkdirAll(path, 0750)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "Failed to create directory")
		return
	}

	// 上传文件至指定目录
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// 返回文件web地址
	response.Success(c, gin.H{
		"url": webPath + safeFilename,
	})
}
