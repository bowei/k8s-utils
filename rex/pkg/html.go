package pkg

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"io"
)

//go:embed template.html
var htmlTemplate string

//go:embed app.js
var appJS string

// GenerateHTML writes the self-contained HTML to `w`.
func GenerateHTML(types map[string]TypeInfo, w io.Writer, startType string) error {
	tmpl, err := template.New("godoc").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	typeData, err := json.Marshal(types)
	if err != nil {
		return err
	}

	templateData := struct {
		TypeData   template.JS
		StartTypes string
		JS         template.JS
		CSS        template.CSS
	}{
		TypeData:   template.JS(typeData),
		StartTypes: startType,
		JS:         template.JS(appJS),
	}

	return tmpl.Execute(w, templateData)
}
