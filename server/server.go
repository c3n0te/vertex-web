package main

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
	"vertex/ui/modules"
	"vertex/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	Key       string
	DB        *sqlx.DB
	TokenAuth *jwtauth.JWTAuth
	Router    chi.Router
}

func NewServer(key string, db *sqlx.DB) Server {
	var tokenAuth *jwtauth.JWTAuth
	tokenAuth = jwtauth.New("HS256", []byte(key), nil)

	return Server{
		Key:       key,
		DB:        db,
		TokenAuth: tokenAuth,
		Router:    NewRouter(),
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

func (srv *Server) ServeAssets() {
	filesrv := http.FileServer(http.Dir("assets"))
	srv.Router.Handle("/images/*", filesrv)
	srv.Router.Handle("/css/output.css", filesrv)
	srv.Router.Handle("/assets/*", http.StripPrefix("/assets/", filesrv))
}

func (srv *Server) MountHandlers() {
	srv.Router.Get("/", srv.GetHome)
	srv.Router.Get("/auth/login", srv.GetAuthLogin)
	srv.Router.Get("/auth/sign-up", srv.GetAuthSignUp)
	srv.Router.Post("/auth/sign-up", srv.PostAuthSignUp)
	srv.Router.Post("/auth/login", srv.PostAuthLogin)
	srv.Router.Get("/auth/logout", srv.GetAuthLogout)

	srv.Router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(srv.TokenAuth))
		r.Use(jwtauth.Authenticator(srv.TokenAuth))
		r.Get("/auth/refresh", srv.GetAuthRefresh)
		r.Get("/dashboard", srv.GetDashboard)
		r.Get("/dashboard/stations", srv.GetDashboardStations)
		r.Get("/dashboard/satellites", srv.GetDashboardSatellites)
		r.Get("/plan", srv.GetPlan)
		r.Post("/plan/submit", srv.PostPlanSubmit)
		r.Get("/pending", srv.GetPending)
		r.Get("/pending/refersh", srv.GetPendingRefresh)
		r.Get("/schedule", srv.GetSchedule)
		r.Get("/schedule/refresh", srv.GetScheduleRefresh)
	})
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

func (srv *Server) GetAuthSignUp(w http.ResponseWriter, r *http.Request) {
	signup := pages.SignUpPage()
	templ.Handler(signup).ServeHTTP(w, r)
}

func (srv *Server) GetAuthRefresh(w http.ResponseWriter, r *http.Request) {
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
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * time.Minute),
		Path:     "/",
	})
}

func (srv *Server) GetAuthLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (srv *Server) PostAuthSignUp(w http.ResponseWriter, r *http.Request) {
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

	user, err = ReadUser(srv.DB, r.Form)
	if err != nil || user.Email != "" {
		slog.Error("Failed to authenticate or user has account: ", "error", err)
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	claims := map[string]interface{}{
		"user_id": user.UserID,
		"email":   user.Email,
		"role":    "operator",
	}

	_, jwtStr, err := srv.TokenAuth.Encode(claims)
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
		Expires:  time.Now().Add(30 * time.Minute),
		Path:     "/",
		MaxAge:   0,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (srv *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
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

	claims := map[string]interface{}{
		"user_id": user.UserID,
		"email":   user.Email,
		"role":    "operator",
	}

	_, jwtStr, err := srv.TokenAuth.Encode(claims)
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
		Expires:  time.Now().Add(30 * time.Minute),
		Path:     "/",
		MaxAge:   0,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (srv *Server) GetDashboardStations(w http.ResponseWriter, r *http.Request) {
	stns, err := ReadStns(srv.DB)
	if err != nil {
		slog.Error("Failed to read stations: ", "error", err)
	}

	dashboard := modules.StationsDash(stns)
	templ.Handler(dashboard).ServeHTTP(w, r)
}

func (srv *Server) GetDashboardSatellites(w http.ResponseWriter, r *http.Request) {
	sats, err := ReadSats(srv.DB)
	if err != nil {
		slog.Error("Failed to read satellites: ", "error", err)
	}

	dashboard := modules.SatellitesDash(sats)
	templ.Handler(dashboard).ServeHTTP(w, r)
}

func (srv *Server) GetDashboard(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		slog.Error("Failed to extract claims from request context", "error", err)
		return
	}

	if token != nil {
		isLoggedIn = true
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
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		slog.Error("Failed to extract claims from request context", "error", err)
		return
	}

	if token != nil {
		isLoggedIn = true
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

func (srv *Server) GetPending(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		slog.Error("Failed to extract claims from request context", "error", err)
		return
	}

	if token != nil {
		isLoggedIn = true
	}

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		slog.Error("Failed to convert page to str: ", "error", err)
		http.Redirect(w, r, "/pending", http.StatusSeeOther)
		return
	}

	if page < 1 {
		page = 1
	}

	pageSize := 10
	offset := (page - 1) * pageSize
	numTasks, err := ReadTaskCount(srv.DB)
	if err != nil {
		slog.Error("Failed to read jobs: ", "error", err)
		http.Redirect(w, r, "/pending", http.StatusSeeOther)
		return
	}

	tasks, err := ReadPendingTasks(srv.DB, pageSize, offset)
	if err != nil {
		slog.Error("Failed to read jobs: ", "error", err)
		http.Redirect(w, r, "/pending", http.StatusSeeOther)
		return
	}

	totalPages := int(math.Ceil(float64(numTasks) / float64(pageSize)))
	pending := pages.PendingPage(tasks, page, int(totalPages), isLoggedIn)
	templ.Handler(pending).ServeHTTP(w, r)
}

func (srv *Server) GetPendingRefresh(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		slog.Error("Failed to convert page to str: ", "error", err)
		http.Redirect(w, r, "/pending", http.StatusSeeOther)
		return
	}

	if page < 1 {
		page = 1
	}

	pageSize := 10
	offset := (page - 1) * pageSize
	numTasks, err := ReadTaskCount(srv.DB)
	if err != nil {
		slog.Error("Failed to read jobs: ", "error", err)
		http.Redirect(w, r, "/pending", http.StatusSeeOther)
		return
	}

	tasks, err := ReadPendingTasks(srv.DB, pageSize, offset)
	if err != nil {
		slog.Error("Failed to read jobs: ", "error", err)
		http.Redirect(w, r, "/pending", http.StatusSeeOther)
		return
	}

	totalPages := int(math.Ceil(float64(numTasks) / float64(pageSize)))
	pending := modules.PaginatedPendingTable(tasks, page, int(totalPages))
	templ.Handler(pending).ServeHTTP(w, r)
}

func (srv *Server) GetSchedule(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := false
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		slog.Error("Failed to extract claims from request context", "error", err)
		return
	}

	if token != nil {
		isLoggedIn = true
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

func (srv *Server) GetScheduleRefresh(w http.ResponseWriter, r *http.Request) {
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
	schedule := modules.PaginatedScheduleTable(jobs, page, int(totalPages))
	templ.Handler(schedule).ServeHTTP(w, r)
}
