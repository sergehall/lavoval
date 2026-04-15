package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

func NewRouter(cfg config.Config, tokens auth.TokenManager, authService *service.AuthService, profileService *service.ProfileService, skillService *service.SkillService, runtimeService *service.RuntimeService, adminService *service.AdminService) http.Handler {
	validate := validator.New(validator.WithRequiredStructEnabled())
	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Logging)
	r.Use(appmiddleware.Recovery)
	r.Use(chimiddleware.Heartbeat("/livez"))

	authHandler := NewAuthHandler(validate, authService)
	meHandler := NewMeHandler(validate, profileService)
	skillHandler := NewSkillHandler(validate, skillService)
	mySkillsHandler := NewMySkillsHandler(validate, skillService)
	runtimeHandler := NewRuntimeHandler(validate, runtimeService)
	adminHandler := NewAdminHandler(validate, adminService, skillService)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "service": cfg.AppName})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/register", authHandler.Register)
			authRouter.Post("/login", authHandler.Login)
			authRouter.Post("/verify-email", authHandler.VerifyEmail)
			authRouter.Post("/resend-verification", authHandler.ResendVerification)
			authRouter.Post("/forgot-password", authHandler.ForgotPassword)
			authRouter.Post("/reset-password", authHandler.ResetPassword)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/logout", authHandler.Logout)
		})

		api.Get("/skills", skillHandler.ListPublic)
		api.Get("/skills/{skillID}", skillHandler.FindByID)

		api.Group(func(private chi.Router) {
			private.Use(appmiddleware.Authenticate(tokens))
			private.Get("/me", meHandler.Profile)
			private.Patch("/me/profile", meHandler.UpdateProfile)
			private.Get("/me/skills", mySkillsHandler.List)
			private.Post("/me/skills", mySkillsHandler.Create)
			private.Get("/me/skills/{skillID}", mySkillsHandler.Get)
			private.Patch("/me/skills/{skillID}", mySkillsHandler.Update)
			private.Delete("/me/skills/{skillID}", mySkillsHandler.Delete)
			private.Get("/runtime/runs", runtimeHandler.List)
			private.Get("/runtime/runs/{runID}", runtimeHandler.Get)
			private.Post("/runtime/run", runtimeHandler.Run)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(appmiddleware.Authenticate(tokens))
			admin.Use(appmiddleware.RequireRole(domain.RoleAdmin))
			admin.Get("/users", adminHandler.ListUsers)
			admin.Get("/skills", adminHandler.ListSkills)
			admin.Get("/skills/{skillID}", adminHandler.GetSkill)
			admin.Get("/runs", runtimeHandler.ListAll)
			admin.Get("/runs/{runID}", runtimeHandler.GetAny)
			admin.Post("/skills", adminHandler.CreateSkill)
			admin.Patch("/skills/{skillID}", adminHandler.UpdateSkill)
			admin.Delete("/skills/{skillID}", adminHandler.DeleteSkill)
		})
	})

	return r
}
