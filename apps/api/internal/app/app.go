package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/handler"
	"github.com/sergehall/lavoval/apps/api/internal/mailer"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
	appRuntime "github.com/sergehall/lavoval/apps/api/internal/runtime"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type Application struct {
	Config config.Config
	Server *http.Server
	Store  *pgxpool.Pool
}

func New() (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	tokenManager := auth.NewTokenManager(cfg)
	userRepo := repository.NewUserRepository(pool)
	profileRepo := repository.NewProfileRepository(pool)
	verificationRepo := repository.NewEmailVerificationRepository(pool)
	passwordResetRepo := repository.NewPasswordResetRepository(pool)
	mfaRecoveryCodeRepo := repository.NewMFARecoveryCodeRepository(pool)
	signInChallengeRepo := repository.NewSignInChallengeRepository(pool)
	oauthStateRepo := repository.NewOAuthStateRepository(pool)
	oauthIdentityRepo := repository.NewOAuthIdentityRepository(pool)
	skillRepo := repository.NewSkillRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	skillRunRepo := repository.NewSkillRunRepository(pool)
	runtimeRegistry := appRuntime.DefaultRegistry()
	verificationMailer := mailer.NewSMTPVerificationMailer(cfg)

	authService := service.NewAuthService(
		userRepo,
		profileRepo,
		verificationRepo,
		passwordResetRepo,
		mfaRecoveryCodeRepo,
		signInChallengeRepo,
		oauthStateRepo,
		oauthIdentityRepo,
		tokenManager,
		verificationMailer,
		service.NoopSessionRevoker{},
		cfg,
	)
	profileService := service.NewProfileService(profileRepo)
	accountSecurityService := service.NewAccountSecurityService(userRepo, oauthIdentityRepo)
	skillService := service.NewSkillService(skillRepo, enrollmentRepo)
	runtimeService := service.NewRuntimeService(skillRepo, skillRunRepo, runtimeRegistry)
	adminService := service.NewAdminService(userRepo, skillRepo)

	router := handler.NewRouter(cfg, tokenManager, authService, profileService, accountSecurityService, skillService, runtimeService, adminService)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Application{Config: cfg, Server: server, Store: pool}, nil
}
