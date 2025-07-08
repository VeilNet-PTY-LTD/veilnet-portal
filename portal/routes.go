package portal

import (
	"embed"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

//go:embed template.html
var templateFS embed.FS

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func (p *Portal) statusPage(c echo.Context) error {

	currentPath, err := os.Executable()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current path"})
	}
	templatePath := filepath.Join(filepath.Dir(currentPath), "template.html")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		data, err := templateFS.ReadFile("template.html")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read template"})
		}
		if err := os.WriteFile(templatePath, data, 0644); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to write template"})
		}
	}

	name := p.anchor.GetName()
	domain := p.anchor.GetDomain()
	region := p.anchor.GetRegion()
	cidr, err := p.anchor.GetCIDR()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get CIDR"})
	}
	return c.Render(http.StatusOK, "template.html", map[string]interface{}{
		"name":   name,
		"domain": domain,
		"region": region,
		"cidr":   cidr,
	})
}