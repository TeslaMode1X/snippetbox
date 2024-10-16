package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	// Create a new router using chi. Chi is a lightweight and efficient
	// router for Go that supports pattern-based routing.
	r := chi.NewRouter()

	// Use the alice package to create a middleware chain.
	// Here, we are chaining `recoverPanic`, `logRequest`, and `secureHeaders` middleware.
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes. For now, this chain will only contain
	// the session middleware, but we'll add more to it later.
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	r.Get("/", dynamicMiddleware.ThenFunc(app.home).ServeHTTP)

	r.Route("/snippet", func(r chi.Router) {
		r.Get("/{id}", dynamicMiddleware.ThenFunc(app.showSnippet).ServeHTTP)
		// New snippet
		r.Get("/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm).ServeHTTP)
		r.Post("/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet).ServeHTTP) // Use Post for resource creation
	})

	r.Route("/user", func(r chi.Router) {
		// User creating
		r.Get("/signup", dynamicMiddleware.ThenFunc(app.signupUserForm).ServeHTTP)
		r.Post("/signup", dynamicMiddleware.ThenFunc(app.signupUser).ServeHTTP)
		// User logging
		r.Get("/login", dynamicMiddleware.ThenFunc(app.loginUserForm).ServeHTTP)
		r.Post("/login", dynamicMiddleware.ThenFunc(app.loginUser).ServeHTTP)
		// User exit
		r.Post("/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser).ServeHTTP)
	})

	// testing purposes
	r.Get("/ping", ping)

	// Create a file server to serve static files from the "./ui/static" directory.
	// Chi supports serving static files with http.StripPrefix to remove "/static" from the request URL.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// Register the file server to handle all paths starting with "/static".
	// We use http.StripPrefix to remove the "/static" prefix before passing the request to the file server.
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Return the middleware chain combined with the chi router.
	// The `Then` method of alice combines the middleware chain with the final handler (`r`).
	return standardMiddleware.Then(r)
}
