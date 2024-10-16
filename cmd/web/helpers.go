package main

import (
	"bytes"
	"fmt"
	"github.com/TeslaMode1X/snippetbox/pkg/models"
	"github.com/justinas/nosurf"
	"net/http"
	"runtime/debug"
	"time"
)

// render renders a specific HTML template with the provided data.
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	// Retrieve the appropriate template set from the cache based on the page name (like 'home.page.tmpl').
	// If no entry exists in the cache with the provided name, call the serverError helper method.
	ts, ok := app.templateCache[name]
	if !ok {
		// Template does not exist in the cache, respond with an error message.
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError helper and
	// return.
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write the contents of the buffer to the http.ResponseWriter. Again, this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	buf.WriteTo(w)
}

// serverError logs the error with a stack trace and sends a 500 Internal Server Error response to the client.
func (app *application) serverError(w http.ResponseWriter, err error) {
	// Capture the error message and stack trace for detailed logging.
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// Log the error message and stack trace using the application's error logger.
	// Output(2, trace) ensures the log includes the caller information (i.e., where the logging was called).
	app.errorLog.Output(2, trace)
	// Send a 500 Internal Server Error response to the client with a generic message.
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError sends an HTTP error response to the client with the specified status code.
func (app *application) clientError(w http.ResponseWriter, status int) {
	// Send an error response with the specified status code and message.
	http.Error(w, http.StatusText(status), status)
}

// notFound sends a 404 Not Found error response to the client.
func (app *application) notFound(w http.ResponseWriter) {
	// Use clientError to send a 404 Not Found response.
	app.clientError(w, http.StatusNotFound)
}

// Create an addDefaultData helper. This takes a pointer to a templateData
// struct, adds the current year to the CurrentYear field, and then returns
// the pointer. Again, we're not using the *http.Request parameter at the
// moment, but we will do later in the book.
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	// If td is nil, create a new templateData instance.
	if td == nil {
		td = &templateData{}
	}

	// Add the current year to the CurrentYear field.
	td.CurrentYear = time.Now().Year()

	// Retrieve the flash message from the session, if it exists.
	// If no flash message is found, an empty string is returned.
	td.Flash = app.session.PopString(r, "flash")
	td.AuthenticatedUser = app.authenticatedUser(r)
	td.CSRFToken = nosurf.Token(r)
	// Return the updated templateData instance.
	return td
}

func (app *application) authenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}
