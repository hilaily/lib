package swagger

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// path 是 swagger api 的路径，比如 /api/swagger/api.yaml，如果从根 path 开始，开头需要加 /
// filename 是 swagger yaml 文件的路径，比如 ./swagger/api.yaml
func Serve(r gin.IRouter, path, filename string) {
	// 直接提供 swagger yaml 文件
	r.GET(path, func(c *gin.Context) {
		c.File(filename)
	})

	r.GET(filepath.Dir(path), func(c *gin.Context) {
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Swagger UI</title>
	<link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
</head>
<body>
	<div id="swagger-ui"></div>
	<script>
		window.onload = function() {
			SwaggerUIBundle({
				url: "` + path + `",
				dom_id: '#swagger-ui',
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIBundle.SwaggerUIStandalonePreset
				],
			});
		}
	</script>
</body>
</html>`
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	})
}
