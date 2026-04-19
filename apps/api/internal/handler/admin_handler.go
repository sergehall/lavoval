package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type AdminHandler struct {
	validate     *validator.Validate
	adminService *service.AdminService
	skillService *service.SkillService
}

func NewAdminHandler(validate *validator.Validate, adminService *service.AdminService, skillService *service.SkillService) *AdminHandler {
	return &AdminHandler{validate: validate, adminService: adminService, skillService: skillService}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "users_load_failed", "Could not list users")
		return
	}
	httpx.JSON(w, http.StatusOK, users)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	detail, err := h.adminService.GetUser(r.Context(), chi.URLParam(r, "userID"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}
	httpx.JSON(w, http.StatusOK, detail)
}

func (h *AdminHandler) GetUserAuditLog(w http.ResponseWriter, r *http.Request) {
	entries, err := h.adminService.GetUserAuditLog(r.Context(), chi.URLParam(r, "userID"), parseLimit(r, 50))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "audit_log_load_failed", "Could not load user audit log")
		return
	}
	httpx.JSON(w, http.StatusOK, entries)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	actorID := ""
	if claims, ok := appmiddleware.ClaimsFromContext(r.Context()); ok && claims != nil {
		actorID = claims.UserID
	}
	var input service.UpdateUserInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, err := h.adminService.UpdateUser(r.Context(), actorID, chi.URLParam(r, "userID"), input)
	if err != nil {
		if errors.Is(err, service.ErrAdminReasonRequired) {
			httpx.Error(w, http.StatusBadRequest, "reason_required", "A reason is required for this status change")
			return
		}
		if errors.Is(err, service.ErrOnlyRootOwnerCanManageRoles) {
			httpx.Error(w, http.StatusForbidden, "role_change_forbidden", "Only root_owner can change elevated roles")
			return
		}
		if errors.Is(err, service.ErrCannotChangeOwnRole) {
			httpx.Error(w, http.StatusBadRequest, "self_role_change_forbidden", "You cannot change your own role")
			return
		}
		if errors.Is(err, service.ErrRootOwnerRequiresMFA) {
			httpx.Error(w, http.StatusBadRequest, "root_owner_requires_mfa", "root_owner requires MFA to be enabled")
			return
		}
		if errors.Is(err, service.ErrPrivilegedRoleRequiresActiveVerifiedAccount) {
			httpx.Error(w, http.StatusBadRequest, "privileged_role_requires_active_verified_account", "Privileged roles require an active, verified account")
			return
		}
		if errors.Is(err, service.ErrLastRootOwnerDemotion) {
			httpx.Error(w, http.StatusBadRequest, "last_root_owner_demotion_forbidden", "Cannot demote the last root_owner")
			return
		}
		if errors.Is(err, service.ErrPrivilegedUserModerationRequiresRootOwner) {
			httpx.Error(w, http.StatusForbidden, "privileged_user_moderation_forbidden", "Only root_owner can moderate admin or root_owner accounts")
			return
		}
		if errors.Is(err, service.ErrLastRootOwnerStatusLockout) {
			httpx.Error(w, http.StatusBadRequest, "last_root_owner_status_lockout_forbidden", "Cannot suspend or block the last root_owner")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "user_update_failed", "Could not update user")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	actorID := ""
	if claims, ok := appmiddleware.ClaimsFromContext(r.Context()); ok && claims != nil {
		actorID = claims.UserID
	}

	var input service.UpdateUserStatusInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := h.adminService.UpdateUserStatus(r.Context(), actorID, chi.URLParam(r, "userID"), input)
	if err != nil {
		if errors.Is(err, service.ErrAdminReasonRequired) {
			httpx.Error(w, http.StatusBadRequest, "reason_required", "A reason is required for this status change")
			return
		}
		if errors.Is(err, service.ErrPrivilegedUserModerationRequiresRootOwner) {
			httpx.Error(w, http.StatusForbidden, "privileged_user_moderation_forbidden", "Only root_owner can moderate admin or root_owner accounts")
			return
		}
		if errors.Is(err, service.ErrLastRootOwnerStatusLockout) {
			httpx.Error(w, http.StatusBadRequest, "last_root_owner_status_lockout_forbidden", "Cannot suspend or block the last root_owner")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "user_status_update_failed", "Could not update user status")
		return
	}

	httpx.JSON(w, http.StatusOK, user)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	actorID := ""
	if claims, ok := appmiddleware.ClaimsFromContext(r.Context()); ok && claims != nil {
		actorID = claims.UserID
	}

	var input service.UpdateUserRoleInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := h.adminService.UpdateUserRole(r.Context(), actorID, chi.URLParam(r, "userID"), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminReasonRequired):
			httpx.Error(w, http.StatusBadRequest, "reason_required", "A reason is required for this role change")
		case errors.Is(err, service.ErrOnlyRootOwnerCanManageRoles):
			httpx.Error(w, http.StatusForbidden, "role_change_forbidden", "Only root_owner can change elevated roles")
		case errors.Is(err, service.ErrCannotChangeOwnRole):
			httpx.Error(w, http.StatusBadRequest, "self_role_change_forbidden", "You cannot change your own role")
		case errors.Is(err, service.ErrRootOwnerRequiresMFA):
			httpx.Error(w, http.StatusBadRequest, "root_owner_requires_mfa", "root_owner requires MFA to be enabled")
		case errors.Is(err, service.ErrPrivilegedRoleRequiresActiveVerifiedAccount):
			httpx.Error(w, http.StatusBadRequest, "privileged_role_requires_active_verified_account", "Privileged roles require an active, verified account")
		case errors.Is(err, service.ErrLastRootOwnerDemotion):
			httpx.Error(w, http.StatusBadRequest, "last_root_owner_demotion_forbidden", "Cannot demote the last root_owner")
		default:
			httpx.Error(w, http.StatusInternalServerError, "user_role_update_failed", "Could not update user role")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, user)
}

func (h *AdminHandler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetAdminStats(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "stats_load_failed", "Could not load admin stats")
		return
	}
	httpx.JSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListEnrollments(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "enrollments_load_failed", "Could not list enrollments")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) AssignSkill(w http.ResponseWriter, r *http.Request) {
	var input service.AssignSkillInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	enrollment, err := h.adminService.AssignSkill(r.Context(), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "assign_skill_failed", "Could not assign skill")
		return
	}
	httpx.JSON(w, http.StatusCreated, enrollment)
}

func (h *AdminHandler) UpdateEnrollment(w http.ResponseWriter, r *http.Request) {
	var input service.UpdateEnrollmentInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	enrollment, err := h.adminService.UpdateEnrollment(r.Context(), chi.URLParam(r, "enrollmentID"), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "enrollment_update_failed", "Could not update enrollment")
		return
	}
	httpx.JSON(w, http.StatusOK, enrollment)
}

// ── Module handlers ───────────────────────────────────────────────────────

func (h *AdminHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListModules(r.Context(), chi.URLParam(r, "skillID"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "modules_load_failed", "Could not list modules")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) CreateModule(w http.ResponseWriter, r *http.Request) {
	var input service.ModuleMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	module, err := h.adminService.CreateModule(r.Context(), chi.URLParam(r, "skillID"), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "module_create_failed", "Could not create module")
		return
	}
	httpx.JSON(w, http.StatusCreated, module)
}

func (h *AdminHandler) UpdateModule(w http.ResponseWriter, r *http.Request) {
	var input service.ModuleMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	module, err := h.adminService.UpdateModule(r.Context(), chi.URLParam(r, "moduleID"), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "module_update_failed", "Could not update module")
		return
	}
	httpx.JSON(w, http.StatusOK, module)
}

func (h *AdminHandler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	if err := h.adminService.DeleteModule(r.Context(), chi.URLParam(r, "moduleID")); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "module_delete_failed", "Could not delete module")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.adminService.ListSkills(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skills_load_failed", "Could not list skills")
		return
	}
	httpx.JSON(w, http.StatusOK, skills)
}

func (h *AdminHandler) GetSkill(w http.ResponseWriter, r *http.Request) {
	skill, err := h.skillService.FindByID(r.Context(), chi.URLParam(r, "skillID"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "skill_not_found", "Skill not found")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *AdminHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	var input service.SkillMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	skill, err := h.skillService.Create(r.Context(), claims.UserID, input)
	if err != nil {
		var contractErr *service.SkillContractValidationError
		if errors.As(err, &contractErr) {
			httpx.Error(w, http.StatusBadRequest, "invalid_skill_contract", contractErr.Error())
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_create_failed", "Could not create skill")
		return
	}
	httpx.JSON(w, http.StatusCreated, skill)
}

func (h *AdminHandler) UpdateSkill(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	var input service.SkillMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	skill, err := h.skillService.Update(r.Context(), chi.URLParam(r, "skillID"), claims.UserID, input)
	if err != nil {
		var contractErr *service.SkillContractValidationError
		if errors.As(err, &contractErr) {
			httpx.Error(w, http.StatusBadRequest, "invalid_skill_contract", contractErr.Error())
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_update_failed", "Could not update skill")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *AdminHandler) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	if err := h.skillService.Archive(r.Context(), chi.URLParam(r, "skillID")); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skill_delete_failed", "Could not archive skill")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) GovernSkill(w http.ResponseWriter, r *http.Request) {
	actorID := ""
	if claims, ok := appmiddleware.ClaimsFromContext(r.Context()); ok && claims != nil {
		actorID = claims.UserID
	}
	var input service.SkillGovernanceInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	skill, err := h.adminService.GovernSkill(r.Context(), actorID, chi.URLParam(r, "skillID"), input)
	if err != nil {
		if errors.Is(err, service.ErrAdminReasonRequired) {
			httpx.Error(w, http.StatusBadRequest, "reason_required", "A reason is required for this governance change")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_govern_failed", "Could not update skill governance")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *AdminHandler) UpdateSkillPricing(w http.ResponseWriter, r *http.Request) {
	actorID := ""
	if claims, ok := appmiddleware.ClaimsFromContext(r.Context()); ok && claims != nil {
		actorID = claims.UserID
	}
	var input service.SkillPricingInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	skill, err := h.adminService.UpdateSkillPricing(r.Context(), actorID, chi.URLParam(r, "skillID"), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSkillPricing) {
			httpx.Error(w, http.StatusBadRequest, "invalid_skill_pricing", "Pricing settings are inconsistent with the selected access type")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_pricing_failed", "Could not update skill pricing")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *AdminHandler) GetSkillAuditLog(w http.ResponseWriter, r *http.Request) {
	entries, err := h.adminService.GetSkillAuditLog(r.Context(), chi.URLParam(r, "skillID"), parseLimit(r, 50))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "audit_log_load_failed", "Could not load skill audit log")
		return
	}
	httpx.JSON(w, http.StatusOK, entries)
}

func (h *AdminHandler) MailOperations(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.adminService.MailOperations(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_ops_load_failed", "Could not load mail operations snapshot")
		return
	}
	httpx.JSON(w, http.StatusOK, snapshot)
}

func (h *AdminHandler) ListDeadLetters(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListDeadLetters(r.Context(), domain.MailJobFilter{
		Query:       r.URL.Query().Get("query"),
		MessageType: r.URL.Query().Get("messageType"),
		Provider:    r.URL.Query().Get("provider"),
		ErrorCode:   r.URL.Query().Get("errorCode"),
		Limit:       parseLimit(r, 100),
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "dead_letters_load_failed", "Could not load dead-letter jobs")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) ListMailEvents(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListMailEvents(r.Context(), domain.MailEventFilter{
		Query:       r.URL.Query().Get("query"),
		JobID:       r.URL.Query().Get("jobId"),
		EventType:   r.URL.Query().Get("eventType"),
		MessageType: r.URL.Query().Get("messageType"),
		Provider:    r.URL.Query().Get("provider"),
		ErrorCode:   r.URL.Query().Get("errorCode"),
		Limit:       parseLimit(r, 100),
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_events_load_failed", "Could not load mail events")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) ListMailEventsByJob(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListMailEventsByJob(r.Context(), chi.URLParam(r, "jobID"), domain.MailEventFilter{
		Query:       r.URL.Query().Get("query"),
		EventType:   r.URL.Query().Get("eventType"),
		MessageType: r.URL.Query().Get("messageType"),
		Provider:    r.URL.Query().Get("provider"),
		ErrorCode:   r.URL.Query().Get("errorCode"),
		Limit:       parseLimit(r, 100),
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_job_events_load_failed", "Could not load mail job events")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) RequeueDeadLetter(w http.ResponseWriter, r *http.Request) {
	job, err := h.adminService.RequeueDeadLetter(r.Context(), chi.URLParam(r, "jobID"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "dead_letter_requeue_failed", "Could not requeue dead-letter job")
		return
	}
	httpx.JSON(w, http.StatusOK, job)
}

func (h *AdminHandler) ReplayMailJob(w http.ResponseWriter, r *http.Request) {
	job, err := h.adminService.ReplayMailJob(r.Context(), chi.URLParam(r, "jobID"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_job_replay_failed", "Could not replay mail job")
		return
	}
	httpx.JSON(w, http.StatusOK, job)
}

func (h *AdminHandler) ListMailSuppressions(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListMailSuppressions(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_suppressions_load_failed", "Could not load mail suppressions")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) MailRetentionSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.adminService.MailRetentionSnapshot(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_retention_load_failed", "Could not load mail retention snapshot")
		return
	}
	httpx.JSON(w, http.StatusOK, snapshot)
}

func (h *AdminHandler) CleanupMailRetention(w http.ResponseWriter, r *http.Request) {
	result, err := h.adminService.CleanupMailRetention(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_cleanup_failed", "Could not clean up retained mail data")
		return
	}
	httpx.JSON(w, http.StatusOK, result)
}

func (h *AdminHandler) ListMailCleanupRuns(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminService.ListMailCleanupRuns(r.Context(), parseLimit(r, 20))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_cleanup_runs_load_failed", "Could not load mail cleanup runs")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *AdminHandler) CreateMailSuppression(w http.ResponseWriter, r *http.Request) {
	var input service.CreateMailSuppressionInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	item, err := h.adminService.CreateMailSuppression(r.Context(), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_suppression_create_failed", "Could not create mail suppression")
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

func (h *AdminHandler) DeleteMailSuppression(w http.ResponseWriter, r *http.Request) {
	if err := h.adminService.DeleteMailSuppression(r.Context(), chi.URLParam(r, "suppressionID")); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mail_suppression_delete_failed", "Could not delete mail suppression")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func parseLimit(r *http.Request, fallback int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return fallback
	}
	if limit > 500 {
		return 500
	}
	return limit
}
