package server

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// ConfigData holds configuration information for the template
type ConfigData struct {
	WebSocketURL string
	DeviceCount  int
}

// TemplateData holds the data passed to the HTML template
type TemplateData struct {
	Version   string
	Commit    string
	BuildDate string
	Status    string
	Metrics   []MetricData
	Config    ConfigData
}

// MetricData represents a metric for template rendering
type MetricData struct {
	Name         string
	Help         string
	Labels       []string
	ExampleValue string
}

var mainTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))
