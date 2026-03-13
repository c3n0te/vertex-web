package main

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
	"vertex/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	Key       string
	DB        *sqlx.DB
	TokenAuth *jwtauth.JWTAuth
}

func NewServer(key string, db *sqlx.DB, tokenAuth *jwtauth.JWTAuth) Server {
	return Server{
		Key:       key,
		DB:        db,
		TokenAuth: tokenAuth,
	}
}

func (srv *Server) InitDB() error {
	Migrate(srv.DB)
	stns, err := ReadStns(srv.DB)
	if err != nil {
		slog.Error("Failed to read stations: ", "error", err)
		return err
	}

	sats, err := ReadSats(srv.DB)
	if err != nil {
		slog.Error("Failed to read satellites: ", "error", err)
		return err
	}

	if len(stns) == 0 {
		err = InsertStns(srv.DB)
		if err != nil {
			slog.Error("Failed to insert stations: ", "error", err)
			return err
		}
	}

	if len(sats) == 0 {
		err = InsertSats(srv.DB)
		if err != nil {
			slog.Error("Failed to insert satellites: ", "error", err)
			return err
		}
	}

	return nil
}

func (srv *Server) GetHome(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	jwtStr, err := r.Cookie("jwt")
	if err != nil {
		slog.Error("Failed to parse JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	isLoggedIn = ParseToken(jwtStr.Value, []byte(srv.Key))
	if !isLoggedIn {
		slog.Error("Invalid JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	home := pages.HomePage(isLoggedIn)
	templ.Handler(home).ServeHTTP(w, r)
}

func (srv *Server) GetAuthLogin(w http.ResponseWriter, r *http.Request) {
	login := pages.LoginPage()
	templ.Handler(login).ServeHTTP(w, r)
}

func (srv *Server) GetSignUp(w http.ResponseWriter, r *http.Request) {
	signup := pages.SignUpPage()
	templ.Handler(signup).ServeHTTP(w, r)
}

func (srv *Server) PostSignUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.Error("Failed to parse login form: ", "error", err)
	}

	user, err := ReadUser(srv.DB, r.Form)
	if err != nil || user.Email != "" {
		slog.Error("Failed to authenticate or user has account: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	err = InsertUser(srv.DB, r.Form)
	if err != nil {
		slog.Error("Failed to insert user: ", "error", err)
		http.Redirect(w, r, "/auth/sign-up", http.StatusSeeOther)
		return
	}

	_, jwtStr, err := srv.TokenAuth.Encode(map[string]interface{}{})
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
}

func (srv *Server) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.Error("Failed to parse login form: ", "error", err)
	}

	user, err := ReadUser(srv.DB, r.Form)
	if err != nil || user.Email == "" {
		slog.Error("Failed to authenticate user or user does not exist: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	_, jwtStr, err := srv.TokenAuth.Encode(map[string]interface{}{})
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
}

func (srv *Server) GetDashboard(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	jwtStr, err := r.Cookie("jwt")
	if err != nil {
		slog.Error("Failed to parse JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	isLoggedIn = ParseToken(jwtStr.Value, []byte(srv.Key))
	if !isLoggedIn {
		slog.Error("Invalid JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	stns, err := ReadStns(srv.DB)
	if err != nil {
		slog.Error("Failed to read stations: ", "error", err)
	}

	sats, err := ReadSats(srv.DB)
	if err != nil {
		slog.Error("Failed to read satellites: ", "error", err)
	}

	dashboard := pages.DashboardPage(stns, sats, isLoggedIn)
	templ.Handler(dashboard).ServeHTTP(w, r)
}

func (srv *Server) GetPlan(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	jwtStr, err := r.Cookie("jwt")
	if err != nil {
		slog.Error("Failed to parse JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	isLoggedIn = ParseToken(jwtStr.Value, []byte(srv.Key))
	if !isLoggedIn {
		slog.Error("Invalid JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	sats, err := ReadSats(srv.DB)
	if err != nil {
		slog.Error("Failed to read satellites: ", "error", err)
	}

	plan := pages.PlanPage(sats, isLoggedIn)
	templ.Handler(plan).ServeHTTP(w, r)
}

func (srv *Server) PostPlanSubmit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.Error("Failed to parse plan form: ", "error", err)
	}

	err = InsertTask(srv.DB, r.Form)
	if err != nil {
		slog.Error("Failed to insert task: ", "error", err)
	}

	http.Redirect(w, r, "/plan", http.StatusSeeOther)
}

func (srv *Server) GetSchedule(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	cookie, err := r.Cookie("jwt")
	if err != nil {
		slog.Error("Failed to parse JWT: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	isLoggedIn = ParseToken(cookie.Value, []byte(srv.Key))
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
	numJobs, err := ReadJobCount(srv.DB)
	if err != nil {
		slog.Error("Failed to read jobs: ", "error", err)
		http.Redirect(w, r, "/schedule", http.StatusSeeOther)
		return
	}

	jobs, err := ReadJobs(srv.DB, pageSize, offset)
	if err != nil {
		slog.Error("Failed to read jobs: ", "error", err)
		http.Redirect(w, r, "/schedule", http.StatusSeeOther)
		return
	}

	totalPages := int(math.Ceil(float64(numJobs) / float64(pageSize)))
	schedule := pages.SchedulePage(jobs, page, int(totalPages), isLoggedIn)
	templ.Handler(schedule).ServeHTTP(w, r)
}
