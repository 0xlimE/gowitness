package web

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"strconv"

	"github.com/sensepost/gowitness/web/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/web/api"
)

// Server is a web server
type Server struct {
	Host           string
	Port           int
	DbUri          string
	ScreenshotPath string
	Password       string
}

// NewServer returns a new server intance
func NewServer(host string, port int, dburi string, screenshotpath string, password string) *Server {
	return &Server{
		Host:           host,
		Port:           port,
		DbUri:          dburi,
		ScreenshotPath: screenshotpath,
		Password:       password,
	}
}

// isJSON sets the Content-Type header to application/json
func isJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// hashPassword creates a SHA256 hash of the password
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// getBasePath extracts the base path from X-Forwarded-Prefix header or returns "/"
func getBasePath(r *http.Request) string {
	prefix := r.Header.Get("X-Forwarded-Prefix")
	if prefix == "" {
		return "/"
	}
	// Ensure prefix ends with /
	if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}
	return prefix
}

// passwordAuthMiddleware checks if password authentication is required and valid
func (s *Server) passwordAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If no password is set, proceed without authentication
		if s.Password == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for password cookie
		cookie, err := r.Cookie("gowitness_auth")
		if err != nil || cookie.Value != hashPassword(s.Password) {
			// Get the base path for proper redirection
			basePath := getBasePath(r)
			// Redirect to login page
			http.Redirect(w, r, basePath+"login", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loginHandler serves the login page and processes login requests
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	basePath := getBasePath(r)

	if r.Method == "POST" {
		// Process login form
		password := r.FormValue("password")
		if password == s.Password {
			// Set authentication cookie with the correct path
			cookiePath := basePath
			if basePath != "/" {
				cookiePath = basePath[:len(basePath)-1] // Remove trailing slash for non-root paths
			}

			cookie := &http.Cookie{
				Name:     "gowitness_auth",
				Value:    hashPassword(s.Password),
				Path:     cookiePath,
				HttpOnly: true,
				Secure:   r.TLS != nil,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, basePath, http.StatusTemporaryRedirect)
			return
		}
		// Invalid password - show error
		s.renderLoginPage(w, "Invalid password", basePath)
		return
	}

	// Show login page
	s.renderLoginPage(w, "", basePath)
}

// renderLoginPage renders the login form
func (s *Server) renderLoginPage(w http.ResponseWriter, errorMsg string, basePath string) {
	loginTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Defend Denmark ASM - Login Required</title>
    <style>
        body {
            font-family: system-ui, -apple-system, sans-serif;
            background: #f5f5f5;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .login-container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            width: 100%;
            max-width: 400px;
        }
        .logo {
            text-align: center;
            margin-bottom: 2rem;
        }
        .logo h1 {
            color: #333;
            margin: 0;
            font-size: 2rem;
        }
        .form-group {
            margin-bottom: 1rem;
        }
        label {
            display: block;
            margin-bottom: 0.5rem;
            color: #555;
            font-weight: 500;
        }
        input[type="password"] {
            width: 100%;
            padding: 0.75rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 1rem;
            box-sizing: border-box;
        }
        input[type="password"]:focus {
            outline: none;
            border-color: #007bff;
        }
        .btn {
            background: #007bff;
            color: white;
            padding: 0.75rem 1.5rem;
            border: none;
            border-radius: 4px;
            font-size: 1rem;
            cursor: pointer;
            width: 100%;
        }
        .btn:hover {
            background: #0056b3;
        }
        .error {
            color: #dc3545;
            margin-bottom: 1rem;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">
            <h1>Defend Denmark ASM</h1>
            <p>Authentication Required</p>
        </div>
        {{if .Error}}
        <div class="error">{{.Error}}</div>
        {{end}}
        <form method="POST" action="{{.BasePath}}login">
            <div class="form-group">
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required autofocus>
            </div>
            <button type="submit" class="btn">Login</button>
        </form>
    </div>
</body>
</html>`

	tmpl := template.Must(template.New("login").Parse(loginTemplate))
	data := struct {
		Error    string
		BasePath string
	}{
		Error:    errorMsg,
		BasePath: basePath,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

// Run a server
func (s *Server) Run() {

	// configure our swagger docs
	docs.SwaggerInfo.Title = "gowitness v3 api"
	docs.SwaggerInfo.Description = "API documentation for gowitness v3"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api"

	// get the router ready
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	apih, err := api.NewApiHandler(s.DbUri, s.ScreenshotPath)
	if err != nil {
		log.Error("could not get api handler up", "err", err)
		return
	}

	// Add login route (not protected by auth middleware)
	if s.Password != "" {
		r.HandleFunc("/login", s.loginHandler)
	}

	// Apply authentication middleware to all routes except login
	r.Route("/", func(r chi.Router) {
		r.Use(s.passwordAuthMiddleware)

		r.Route("/api", func(r chi.Router) {
			r.Use(isJSON)
			r.Use(cors.Handler(cors.Options{
				AllowedOrigins: []string{"*"}, // TODO: flag this
			}))

			r.Get("/ping", apih.PingHandler)
			r.Get("/statistics", apih.StatisticsHandler)
			r.Get("/scan-sessions", apih.ScanSessionsHandler)
			r.Get("/wappalyzer", apih.WappalyzerHandler)
			r.Get("/security/status", apih.SecurityStatusHandler)
			r.Get("/ip/{ip}", apih.IPInfoHandler)
			r.Get("/logo", apih.LogoHandler)
			r.Post("/search", apih.SearchHandler)
			r.Post("/submit", apih.SubmitHandler)
			r.Post("/submit/single", apih.SubmitSingleHandler)

			r.Get("/results/gallery", apih.GalleryHandler)
			r.Get("/results/list", apih.ListHandler)
			r.Get("/results/detail/{id}", apih.DetailHandler)
			r.Post("/results/delete", apih.DeleteResultHandler)
			r.Get("/results/technology", apih.TechnologyListHandler)
		})

		// screenshot files
		r.Mount("/screenshots", http.StripPrefix("/screenshots/", http.FileServer(http.Dir(s.ScreenshotPath))))

		// swagger documentation
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

		// the spa
		r.Handle("/*", SpaHandler())
	})

	log.Info("starting web server", "host", s.Host, "port", s.Port)
	if s.Password != "" {
		log.Info("password protection enabled")
	}
	if err := http.ListenAndServe(s.Host+":"+strconv.Itoa(s.Port), r); err != nil {
		log.Error("server listen error", "err", err)
	}
}
