package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/TeslaMode1X/snippetbox/pkg/models"
	"github.com/TeslaMode1X/snippetbox/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type contextKey string

var contextKeyUser = contextKey("user")

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	users    interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
	}
	session  *sessions.Session
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
}

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000",
	// and a short help text explaining what the flag controls. The value of this
	// flag will be stored in the 'addr' variable at runtime.
	//
	// If you want to run the program with a specific port using the flag -addr="PORT",
	// you should run it from the root directory of the project (e.g., C:\Users\anuar\code\snippetbox).
	// The reason is that the program uses relative paths (starting with "./") to access
	// files like templates. If you run the program from any other directory, these
	// relative paths won't resolve correctly, and you'll get an error because the
	// program won't be able to find the required files.
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data")

	// Define a new command-line flag for the session secret (a random key which
	// will be used to encrypt and authenticate session cookies). It should be 32
	// bytes long.
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret")

	// Importantly, we use the flag.Parse() function to parse the command-line
	// arguments. This reads in the command-line flag value and assigns it to the 'addr'
	// variable. You need to call this *before* you use the 'addr' variable
	// otherwise, it will always contain the default value of ":4000". If any error
	// is encountered during parsing, the application will terminate.
	flag.Parse()

	// Create a logger for writing informational messages. This uses the destination
	// (os.Stdout), a prefix for the message ("INFO" followed by a tab), and flags
	// to include the local date and time in the log entries.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for writing error messages. This uses stderr as the destination
	// and the log.Llongfile flag to include the file name and line number in the log entries.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	// To keep the main() function tidy I've put the code for creating a connec
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Use the sessions.New() function to initialize a new session manager,
	// passing in the secret key as the parameter. Then we configure it so
	// sessions always expires after 12 hours.
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true // Set the Secure flag on our session cookies
	session.SameSite = http.SameSiteStrictMode

	// Create an instance of the application struct, which will hold the loggers.
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		users:         &mysql.UserModel{DB: db},
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we want
	// the server to use
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize a new http.Server struct. We set the Addr field to the network address
	// specified by the command-line flag, the Handler field to the ServeMux, and the
	// ErrorLog field to use the custom error logger.
	//
	// The http.Server struct provides a way to configure and run an HTTP server with
	// specific settings, including network address, request handling, and logging.
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Log the starting address of the server.
	infoLog.Printf("Starting server on %s", *addr)

	// Start the HTTP server by calling ListenAndServe on the http.Server struct.
	// This method will block and run until an error occurs or the server is stopped.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	// If ListenAndServe returns an error, log it and terminate the application.
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
