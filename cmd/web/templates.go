package main

import (
	"github.com/TeslaMode1X/snippetbox/pkg/forms"
	"github.com/TeslaMode1X/snippetbox/pkg/models"
	"html/template"
	"path/filepath"
	"time"
)

type templateData struct {
	CurrentYear       int
	CSRFToken         string
	Flash             string
	Form              *forms.Form
	AuthenticatedUser *models.User
	Snippet           *models.Snippet
	Snippets          []*models.Snippet
}

func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value.
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Get all filepaths with the extension '.page.tmpl' using filepath.Glob.
	// This gives us all the 'page' templates in the specified directory.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// Loop through each page template.
	for _, page := range pages {
		// Extract the base filename (e.g., 'home.page.tmpl') from the full path.
		name := filepath.Base(page)

		// Parse the page template file into a template set.
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Add any 'layout' templates to the template set (e.g., 'base.layout.tmpl').
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// Add any 'partial' templates to the template set (e.g., 'footer.partial.tmpl').
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// Store the template set in the cache, with the page filename as the key.
		cache[name] = ts
	}

	// Return the completed cache.
	return cache, nil
}
