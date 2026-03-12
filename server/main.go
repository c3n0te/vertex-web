package main

import (
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
	"vertex/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var tokenAuth *jwtauth.JWTAuth

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file: ", "error", err)
		return
	}

	key := os.Getenv("SERVER_SECRET_KEY")
	if key == "" {
		slog.Error("Error loading Server Key env var.")
		return
	}

	tokenAuth = jwtauth.New("HS256", []byte(key), nil)
	db, err := sqlx.Connect("sqlite3", "./vertex.db")
	if err != nil {
		slog.Error("Failed to open SQLite Database: ", "error", err)
		return
	}

	defer db.Close()
	Migrate(db)
	stns, err := ReadStns(db)
	if err != nil {
		slog.Error("Failed to read stations: ", "error", err)
		return
	}

	sats, err := ReadSats(db)
	if err != nil {
		slog.Error("Failed to read satellitess: ", "error", err)
		return
	}

	if len(stns) == 0 {
		InsertStns(db)
	}

	if len(sats) == 0 {
		InsertSats(db)
	}

	router := NewRouter()
	ServeAssets(router)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		isLoggedIn := false
		jwtStr, err := r.Cookie("jwt")
		if err != nil {
			slog.Error("Failed to parse JWT: ", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		isLoggedIn = ParseToken(jwtStr.Value, []byte(key))
		if !isLoggedIn {
			slog.Error("Invalid JWT: ", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		home := pages.HomePage(isLoggedIn)
		templ.Handler(home).ServeHTTP(w, r)
	})
	router.Get("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		login := pages.LoginPage()
		templ.Handler(login).ServeHTTP(w, r)
	})

	router.Get("/auth/sign-up", func(w http.ResponseWriter, r *http.Request) {
		signup := pages.SignUpPage()
		templ.Handler(signup).ServeHTTP(w, r)
	})

	router.Post("/auth/sign-up", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("Failed to parse login form: ", "error", err)
		}

		user, err := ReadUser(db, r.Form)
		if err != nil || user.Email != "" {
			slog.Error("Failed to authenticate or user has account: ", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		err = InsertUser(db, r.Form)
		if err != nil {
			slog.Error("Failed to insert user: ", "error", err)
			http.Redirect(w, r, "/auth/sign-up", http.StatusSeeOther)
			return
		}

		_, jwtStr, err := tokenAuth.Encode(map[string]interface{}{})
		if err != nil {
			slog.Error("Failed to create JWT: ", "error", err)
			http.Redirect(w, r, "/auth/sign-up", http.StatusSeeOther)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    jwtStr,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(1 * time.Hour),
			Path:     "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	router.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("Failed to parse login form: ", "error", err)
		}

		user, err := ReadUser(db, r.Form)
		if err != nil || user.Email == "" {
			slog.Error("Failed to authenticate user or user does not exist: ", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		_, jwtStr, err := tokenAuth.Encode(map[string]interface{}{})
		if err != nil {
			slog.Error("Failed to create JWT: ", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    jwtStr,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(1 * time.Hour),
			Path:     "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			isLoggedIn := false
			jwtStr, err := r.Cookie("jwt")
			if err != nil {
				slog.Error("Failed to parse JWT: ", "error", err)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			isLoggedIn = ParseToken(jwtStr.Value, []byte(key))
			if !isLoggedIn {
				slog.Error("Invalid JWT: ", "error", err)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			stns, err := ReadStns(db)
			if err != nil {
				slog.Error("Failed to read stations: ", "error", err)
			}

			sats, err := ReadSats(db)
			if err != nil {
				slog.Error("Failed to read satellites: ", "error", err)
			}

			dashboard := pages.DashboardPage(stns, sats, isLoggedIn)
			templ.Handler(dashboard).ServeHTTP(w, r)
		})

		r.Get("/plan", func(w http.ResponseWriter, r *http.Request) {
			isLoggedIn := false
			jwtStr, err := r.Cookie("jwt")
			if err != nil {
				slog.Error("Failed to parse JWT: ", "error", err)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			isLoggedIn = ParseToken(jwtStr.Value, []byte(key))
			if !isLoggedIn {
				slog.Error("Invalid JWT: ", "error", err)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			sats, err := ReadSats(db)
			if err != nil {
				slog.Error("Failed to read satellites: ", "error", err)
			}

			plan := pages.PlanPage(sats, isLoggedIn)
			templ.Handler(plan).ServeHTTP(w, r)
		})

		r.Post("/plan/submit", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				slog.Error("Failed to parse plan form: ", "error", err)
			}

			err = InsertTask(db, r.Form)
			if err != nil {
				slog.Error("Failed to insert task: ", "error", err)
			}

			http.Redirect(w, r, "/plan", http.StatusSeeOther)
		})

		r.Get("/schedule", func(w http.ResponseWriter, r *http.Request) {
			isLoggedIn := false
			cookie, err := r.Cookie("jwt")
			if err != nil {
				slog.Error("Failed to parse JWT: ", "error", err)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			isLoggedIn = ParseToken(cookie.Value, []byte(key))
			if !isLoggedIn {
				slog.Error("Invalid JWT: ", "error", err)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			pageStr := r.URL.Query().Get("page")
			if pageStr == "" {
				pageStr = "1"
			}

			page, err := strconv.Atoi(pageStr)
			if err != nil {
				slog.Error("Failed to convert page to str: ", "error", err)
				http.Redirect(w, r, "/schedule", http.StatusSeeOther)
				return
			}

			if page < 1 {
				page = 1
			}

			pageSize := 10
			offset := (page - 1) * pageSize
			numJobs, err := ReadJobCount(db)
			if err != nil {
				slog.Error("Failed to read jobs: ", "error", err)
				http.Redirect(w, r, "/schedule", http.StatusSeeOther)
				return
			}

			jobs, err := ReadJobs(db, pageSize, offset)
			if err != nil {
				slog.Error("Failed to read jobs: ", "error", err)
				http.Redirect(w, r, "/schedule", http.StatusSeeOther)
				return
			}

			totalPages := int(math.Ceil(float64(numJobs) / float64(pageSize)))
			schedule := pages.SchedulePage(jobs, page, int(totalPages), isLoggedIn)
			templ.Handler(schedule).ServeHTTP(w, r)
		})
	})

	http.ListenAndServe(":8000", router)
}
