package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// Config holds the plugin configuration
type Config struct {
	Enabled         bool   `yaml:"enabled"`          // Enable/disable the plugin
	Filename        string `yaml:"filename"`         // Path to the HTML file
	TriggerFilename string `yaml:"triggerFilename"`  // Path to the trigger file
	HTTPResponseCode int    `yaml:"httpResponseCode"` // HTTP status code (e.g., 503)
	HTTPContentType  string `yaml:"httpContentType"`  // Content type (e.g., text/html)
	ImageFile       string `yaml:"ImageFile"`        // Path to the image file
}

// CreateConfig initializes the default configuration
func CreateConfig() *Config {
	return &Config{
		Enabled:         true,
		Filename:        "/path/to/maintenance.html",
		TriggerFilename: "/path/to/maintenance.trigger",
		HTTPResponseCode: 503,
		HTTPContentType:  "text/html; charset=utf-8",
		ImageFile:       "/path/to/maintenance-image.png",
	}
}

// MaintenancePlugin is the plugin struct
type MaintenancePlugin struct {
	next            http.Handler
	name            string
	enabled         bool
	filename        string
	triggerFilename string
	httpResponseCode int
	httpContentType  string
	imageFile       string
	htmlContent     string
}

// New creates a new instance of the plugin
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// Read the HTML file at startup
	htmlContent, err := ioutil.ReadFile(config.Filename)
	if err != nil {
		return nil, err
	}

	return &MaintenancePlugin{
		next:            next,
		name:            name,
		enabled:         config.Enabled,
		filename:        config.Filename,
		triggerFilename: config.TriggerFilename,
		httpResponseCode: config.HTTPResponseCode,
		httpContentType:  config.HTTPContentType,
		imageFile:       config.ImageFile,
		htmlContent:     string(htmlContent),
	}, nil
}

// ServeHTTP handles the request
func (m *MaintenancePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Check if plugin is enabled and maintenance mode is active
	if m.enabled && fileExists(m.triggerFilename) {
		// Serve the image if requested
		if req.URL.Path == "/maintenance-image.png" { // Fixed path for simplicity; can be made dynamic
			m.serveImage(rw, req)
			return
		}
		// Serve the HTML
		rw.Header().Set("Content-Type", m.httpContentType)
		rw.WriteHeader(m.httpResponseCode)
		rw.Write([]byte(m.htmlContent))
		return
	}
	// Pass to the next handler if not in maintenance mode
	m.next.ServeHTTP(rw, req)
}

// serveImage serves the image file
func (m *MaintenancePlugin) serveImage(rw http.ResponseWriter, req *http.Request) {
	imgContent, err := ioutil.ReadFile(m.imageFile)
	if err != nil {
		http.Error(rw, "Image not found", http.StatusNotFound)
		return
	}

	// Set content type based on file extension (assuming PNG here)
	rw.Header().Set("Content-Type", "image/png")
	rw.WriteHeader(http.StatusOK)
	rw.Write(imgContent)
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}