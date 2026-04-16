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
	Config     config.Config
	Server     *http.Server
	Store      *pgxpool.Pool
	dispatcher interface {
		Close(context.Context) error
	}
}

type lifecycleGroup struct {
	members []interface {
		Close(context.Context) error
	}
}

func (g lifecycleGroup) Close(ctx context.Context) error {
	for _, member := range g.members {
		if member == nil {
			continue
		}
		if err := member.Close(ctx); err != nil {
			return err
		}
	}
	return nil
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
	mailJobRepo := repository.NewMailJobRepository(pool)
	mailEventRepo := repository.NewMailEventRepository(pool)
	mailSuppressionRepo := repository.NewMailSuppressionRepository(pool)
	mailCleanupRunRepo := repository.NewMailCleanupRunRepository(pool)
	mfaRecoveryCodeRepo := repository.NewMFARecoveryCodeRepository(pool)
	signInChallengeRepo := repository.NewSignInChallengeRepository(pool)
	oauthStateRepo := repository.NewOAuthStateRepository(pool)
	oauthIdentityRepo := repository.NewOAuthIdentityRepository(pool)
	skillRepo := repository.NewSkillRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	moduleRepo := repository.NewModuleRepository(pool)
	skillRunRepo := repository.NewSkillRunRepository(pool)
	auditLogRepo := repository.NewAdminAuditLogRepository(pool)
	skillAccessRepo := repository.NewSkillAccessRepository(pool)
	runtimeRegistry := appRuntime.DefaultRegistry()
	mailMetrics := mailer.NewPrometheusHandler(mailJobRepo)
	verificationMailer := mailer.NewPostgresVerificationMailer(cfg, mailJobRepo, mailEventRepo, mailSuppressionRepo, mailMetrics)
	mailDispatcher := mailer.NewMailDispatcher(cfg, mailJobRepo, mailEventRepo, mailMetrics)
	mailRetentionWorker := mailer.NewMailRetentionWorker(cfg, mailJobRepo, mailEventRepo, mailCleanupRunRepo, mailMetrics)
	sessionRevoker := service.NewUserSessionRevoker(userRepo)

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
		sessionRevoker,
		cfg,
	)
	profileService := service.NewProfileService(profileRepo)
	accountSecurityService := service.NewAccountSecurityService(userRepo, oauthIdentityRepo)
	skillService := service.NewSkillService(skillRepo, enrollmentRepo, userRepo)
	runtimeService := service.NewRuntimeService(skillRepo, skillRunRepo, runtimeRegistry)
	adminService := service.NewAdminService(
		userRepo,
		profileRepo,
		skillRepo,
		enrollmentRepo,
		moduleRepo,
		mailJobRepo,
		mailEventRepo,
		mailSuppressionRepo,
		mailCleanupRunRepo,
		service.MailRetentionPolicy{
			JobsRetention:          cfg.MailJobsRetention,
			EventsRetention:        cfg.MailEventsRetention,
			CleanupBatchSize:       cfg.MailCleanupBatchSize,
			CleanupInterval:        cfg.MailCleanupInterval,
			CleanupDryRun:          cfg.MailCleanupDryRun,
			AlertJobsThreshold:     cfg.MailCleanupAlertJobsThreshold,
			AlertEventsThreshold:   cfg.MailCleanupAlertEventsThreshold,
			AlertFailureStreak:     cfg.MailCleanupAlertFailureStreak,
			AlertStaleAfter:        cfg.MailCleanupAlertStaleAfter,
			WebhookAlertingEnabled: cfg.MailAlertWebhookURL != "",
		},
		service.WithAuditLog(auditLogRepo),
		service.WithSkillAccess(skillAccessRepo),
		service.WithSessionRevoker(sessionRevoker),
	)

	router := handler.NewRouter(cfg, tokenManager, userRepo, authService, profileService, accountSecurityService, skillService, runtimeService, adminService, mailMetrics)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Application{
		Config: cfg,
		Server: server,
		Store:  pool,
		dispatcher: lifecycleGroup{
			members: []interface {
				Close(context.Context) error
			}{
				mailDispatcher,
				mailRetentionWorker,
			},
		},
	}, nil
}

func (a *Application) Close(ctx context.Context) error {
	if a.dispatcher != nil {
		if err := a.dispatcher.Close(ctx); err != nil {
			return fmt.Errorf("close mail dispatcher: %w", err)
		}
	}

	if a.Store != nil {
		a.Store.Close()
	}

	return nil
}
