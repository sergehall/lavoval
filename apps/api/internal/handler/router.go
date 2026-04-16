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

func NewRouter(cfg config.Config, tokens auth.TokenManager, authService *service.AuthService, profileService *service.ProfileService, accountSecurityService *service.AccountSecurityService, skillService *service.SkillService, runtimeService *service.RuntimeService, adminService *service.AdminService, metricsHandler http.Handler) http.Handler {
	validate := validator.New(validator.WithRequiredStructEnabled())
	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Logging)
	r.Use(appmiddleware.Recovery)
	r.Use(chimiddleware.Heartbeat("/livez"))

	authHandler := NewAuthHandler(validate, authService)
	meHandler := NewMeHandler(validate, profileService, accountSecurityService)
	skillHandler := NewSkillHandler(validate, skillService)
	mySkillsHandler := NewMySkillsHandler(validate, skillService)
	runtimeHandler := NewRuntimeHandler(validate, runtimeService)
	adminHandler := NewAdminHandler(validate, adminService, skillService)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		mailProvider := cfg.MailProvider
		if mailProvider == "" {
			mailProvider = "smtp"
		}

		mailConfigured := cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != ""
		if mailProvider == "gmail_api" {
			mailConfigured = cfg.GmailAPIAccessToken != "" ||
				(cfg.GmailAPIRefreshToken != "" && cfg.GmailAPIClientID != "" && cfg.GmailAPIClientSecret != "")
		}
		if mailProvider == "noop" {
			mailConfigured = true
		}

		httpx.JSON(w, http.StatusOK, map[string]any{
			"status":            "ok",
			"service":           cfg.AppName,
			"mail_provider":     mailProvider,
			"mail_configured":   mailConfigured,
			"smtp_configured":   cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "",
			"gmail_api_enabled": cfg.GmailAPIAccessToken != "",
		})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	if metricsHandler != nil {
		r.Handle("/metrics", metricsHandler)
	}

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/register", authHandler.Register)
			authRouter.Post("/login", authHandler.Login)
			authRouter.Get("/oauth/google/start", authHandler.GoogleOAuthStart)
			authRouter.Get("/oauth/github/start", authHandler.GitHubOAuthStart)
			authRouter.Post("/oauth/google/complete", authHandler.CompleteGoogleOAuth)
			authRouter.Post("/oauth/github/complete", authHandler.CompleteGitHubOAuth)
			authRouter.Post("/mfa/complete-sign-in", authHandler.CompleteMFASignIn)
			authRouter.Post("/verify-email", authHandler.VerifyEmail)
			authRouter.Post("/resend-verification", authHandler.ResendVerification)
			authRouter.Post("/forgot-password", authHandler.ForgotPassword)
			authRouter.Post("/reset-password", authHandler.ResetPassword)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/logout", authHandler.Logout)
			authRouter.With(appmiddleware.Authenticate(tokens)).Get("/mfa/status", authHandler.MFAStatus)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/mfa/enroll", authHandler.EnrollMFA)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/mfa/cancel-enrollment", authHandler.CancelMFAEnrollment)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/mfa/verify-enrollment", authHandler.VerifyMFAEnrollment)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/mfa/disable", authHandler.DisableMFA)
			authRouter.With(appmiddleware.Authenticate(tokens)).Post("/mfa/recovery-codes/regenerate", authHandler.RegenerateMFARecoveryCodes)
		})

		api.Get("/skills", skillHandler.ListPublic)
		api.Get("/skills/{skillID}", skillHandler.FindByID)

		api.Group(func(private chi.Router) {
			private.Use(appmiddleware.Authenticate(tokens))
			private.Get("/me", meHandler.Profile)
			private.Get("/me/security", meHandler.Security)
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
			admin.Get("/stats", adminHandler.GetAdminStats)
			admin.Get("/users", adminHandler.ListUsers)
			admin.Get("/users/{userID}", adminHandler.GetUser)
			admin.Patch("/users/{userID}", adminHandler.UpdateUser)
			admin.Get("/users/{userID}/audit", adminHandler.GetUserAuditLog)
			admin.Get("/enrollments", adminHandler.ListEnrollments)
			admin.Post("/enrollments", adminHandler.AssignSkill)
			admin.Patch("/enrollments/{enrollmentID}", adminHandler.UpdateEnrollment)
			admin.Get("/mail/ops", adminHandler.MailOperations)
			admin.Get("/mail/retention", adminHandler.MailRetentionSnapshot)
			admin.Get("/mail/cleanup-runs", adminHandler.ListMailCleanupRuns)
			admin.Post("/mail/cleanup", adminHandler.CleanupMailRetention)
			admin.Get("/mail/dead-letters", adminHandler.ListDeadLetters)
			admin.Post("/mail/dead-letters/{jobID}/requeue", adminHandler.RequeueDeadLetter)
			admin.Post("/mail/jobs/{jobID}/replay", adminHandler.ReplayMailJob)
			admin.Get("/mail/events", adminHandler.ListMailEvents)
			admin.Get("/mail/jobs/{jobID}/events", adminHandler.ListMailEventsByJob)
			admin.Get("/mail/suppressions", adminHandler.ListMailSuppressions)
			admin.Post("/mail/suppressions", adminHandler.CreateMailSuppression)
			admin.Delete("/mail/suppressions/{suppressionID}", adminHandler.DeleteMailSuppression)
			admin.Get("/skills", adminHandler.ListSkills)
			admin.Get("/skills/{skillID}", adminHandler.GetSkill)
			admin.Patch("/skills/{skillID}/governance", adminHandler.GovernSkill)
			admin.Patch("/skills/{skillID}/pricing", adminHandler.UpdateSkillPricing)
			admin.Get("/skills/{skillID}/audit", adminHandler.GetSkillAuditLog)
			admin.Get("/skills/{skillID}/modules", adminHandler.ListModules)
			admin.Get("/runs", runtimeHandler.ListAll)
			admin.Get("/runs/{runID}", runtimeHandler.GetAny)
			admin.Post("/skills", adminHandler.CreateSkill)
			admin.Patch("/skills/{skillID}", adminHandler.UpdateSkill)
			admin.Delete("/skills/{skillID}", adminHandler.DeleteSkill)
			admin.Post("/skills/{skillID}/modules", adminHandler.CreateModule)
			admin.Patch("/modules/{moduleID}", adminHandler.UpdateModule)
			admin.Delete("/modules/{moduleID}", adminHandler.DeleteModule)
		})
	})

	return r
}
