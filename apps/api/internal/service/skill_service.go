package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

var ErrSkillForbidden = errors.New("skill access forbidden")

type SkillService struct {
	skills      repository.SkillStore
	versions    repository.SkillVersionStore
	enrollments repository.EnrollmentStore
	users       repository.UserStore
}

type SkillContractValidationError struct {
	Message string
}

func (e *SkillContractValidationError) Error() string {
	return e.Message
}

type SkillMutationInput struct {
	Slug               string             `json:"slug" validate:"required,min=2"`
	Title              string             `json:"title" validate:"required,min=3"`
	Summary            string             `json:"summary" validate:"required,min=10"`
	Description        string             `json:"description" validate:"required,min=20"`
	Provider           string             `json:"provider" validate:"required,min=2"`
	Entrypoint         string             `json:"entrypoint" validate:"required,min=2"`
	Config             map[string]any     `json:"config"`
	Status             domain.SkillStatus `json:"status" validate:"required,oneof=draft published archived"`
	Visibility         domain.Visibility  `json:"visibility" validate:"required,oneof=public private"`
	InputSchema        map[string]any     `json:"inputSchema"`
	OutputSchema       map[string]any     `json:"outputSchema"`
	ErrorSchema        map[string]any     `json:"errorSchema"`
	PromptTemplate     string             `json:"promptTemplate"`
	SystemInstructions string             `json:"systemInstructions"`
	Changelog          string             `json:"changelog"`
}

func NewSkillService(
	skills repository.SkillStore,
	versions repository.SkillVersionStore,
	enrollments repository.EnrollmentStore,
	users repository.UserStore,
) *SkillService {
	return &SkillService{skills: skills, versions: versions, enrollments: enrollments, users: users}
}

func (s *SkillService) ListPublic(ctx context.Context, filter domain.SkillFilter) ([]domain.Skill, error) {
	items, err := s.skills.ListPublished(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list public skills: %w", err)
	}
	return items, nil
}

func (s *SkillService) FindByID(ctx context.Context, id string) (domain.Skill, error) {
	skill, err := s.skills.FindByID(ctx, id)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("find skill: %w", err)
	}
	if err := s.attachCurrentVersion(ctx, &skill); err != nil {
		return domain.Skill{}, err
	}
	return skill, nil
}

func (s *SkillService) ListByCreator(ctx context.Context, creatorID string) ([]domain.Skill, error) {
	items, err := s.skills.ListByCreatorID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("list creator skills: %w", err)
	}
	return items, nil
}

func (s *SkillService) FindOwnedByCreator(ctx context.Context, id string, creatorID string) (domain.Skill, error) {
	skill, err := s.FindByID(ctx, id)
	if err != nil {
		return domain.Skill{}, err
	}
	if skill.CreatedBy != creatorID {
		return domain.Skill{}, ErrSkillForbidden
	}
	return skill, nil
}

func (s *SkillService) Create(ctx context.Context, actorID string, input SkillMutationInput) (domain.Skill, error) {
	if err := s.ensureCreatorCanMutate(ctx, actorID); err != nil {
		return domain.Skill{}, err
	}
	if err := validateSkillContract(input); err != nil {
		return domain.Skill{}, err
	}

	skill := domain.Skill{
		ID:          uuid.NewString(),
		Slug:        input.Slug,
		Title:       input.Title,
		Summary:     input.Summary,
		Description: input.Description,
		Provider:    input.Provider,
		Entrypoint:  input.Entrypoint,
		Config:      input.Config,
		Status:      input.Status,
		Visibility:  input.Visibility,
		CreatedBy:   actorID,
	}

	createdSkill, err := s.skills.Create(ctx, skill)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("create skill: %w", err)
	}
	if err := s.saveCurrentVersion(ctx, createdSkill.ID, actorID, input); err != nil {
		return domain.Skill{}, err
	}
	return s.FindByID(ctx, createdSkill.ID)
}

func (s *SkillService) Update(ctx context.Context, id string, actorID string, input SkillMutationInput) (domain.Skill, error) {
	if err := validateSkillContract(input); err != nil {
		return domain.Skill{}, err
	}

	skill := domain.Skill{
		ID:          id,
		Slug:        input.Slug,
		Title:       input.Title,
		Summary:     input.Summary,
		Description: input.Description,
		Provider:    input.Provider,
		Entrypoint:  input.Entrypoint,
		Config:      input.Config,
		Status:      input.Status,
		Visibility:  input.Visibility,
	}

	updatedSkill, err := s.skills.Update(ctx, skill)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("update skill: %w", err)
	}
	if err := s.saveCurrentVersion(ctx, updatedSkill.ID, actorID, input); err != nil {
		return domain.Skill{}, err
	}
	return s.FindByID(ctx, updatedSkill.ID)
}

func (s *SkillService) UpdateOwnedByCreator(ctx context.Context, id string, creatorID string, input SkillMutationInput) (domain.Skill, error) {
	if err := s.ensureCreatorCanMutate(ctx, creatorID); err != nil {
		return domain.Skill{}, err
	}
	if err := validateSkillContract(input); err != nil {
		return domain.Skill{}, err
	}

	skill, err := s.FindOwnedByCreator(ctx, id, creatorID)
	if err != nil {
		return domain.Skill{}, err
	}

	updatedSkill, err := s.skills.Update(ctx, domain.Skill{
		ID:          skill.ID,
		Slug:        input.Slug,
		Title:       input.Title,
		Summary:     input.Summary,
		Description: input.Description,
		Provider:    input.Provider,
		Entrypoint:  input.Entrypoint,
		Config:      input.Config,
		Status:      input.Status,
		Visibility:  input.Visibility,
	})
	if err != nil {
		return domain.Skill{}, fmt.Errorf("update owned skill: %w", err)
	}
	if err := s.saveCurrentVersion(ctx, updatedSkill.ID, creatorID, input); err != nil {
		return domain.Skill{}, err
	}
	return s.FindByID(ctx, updatedSkill.ID)
}

func (s *SkillService) Archive(ctx context.Context, id string) error {
	if err := s.skills.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("archive skill: %w", err)
	}
	return nil
}

func (s *SkillService) ArchiveOwnedByCreator(ctx context.Context, id string, creatorID string) error {
	if err := s.ensureCreatorCanMutate(ctx, creatorID); err != nil {
		return err
	}
	if _, err := s.FindOwnedByCreator(ctx, id, creatorID); err != nil {
		return err
	}
	if err := s.skills.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("archive owned skill: %w", err)
	}
	return nil
}

func (s *SkillService) ensureCreatorCanMutate(ctx context.Context, creatorID string) error {
	if s.users == nil {
		return nil
	}

	user, err := s.users.FindByID(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("find creator account: %w", err)
	}

	switch user.Status {
	case domain.AccountStatusBlocked:
		return ErrAccountBlocked
	case domain.AccountStatusSuspended:
		return ErrAccountSuspended
	default:
		return nil
	}
}

func (s *SkillService) attachCurrentVersion(ctx context.Context, skill *domain.Skill) error {
	if s.versions == nil || skill == nil {
		return nil
	}

	version, err := s.versions.FindCurrentBySkillID(ctx, skill.ID)
	if err != nil {
		return fmt.Errorf("find current skill version: %w", err)
	}
	skill.CurrentVersion = version
	return nil
}

func (s *SkillService) saveCurrentVersion(ctx context.Context, skillID, actorID string, input SkillMutationInput) error {
	if s.versions == nil {
		return nil
	}

	version := domain.SkillVersion{
		SkillID:      skillID,
		ContentMD:    input.Description,
		InputSchema:  cloneJSONMap(input.InputSchema),
		OutputSchema: cloneJSONMap(input.OutputSchema),
		ErrorSchema:  cloneJSONMap(input.ErrorSchema),
		CreatedBy:    actorID,
	}

	if changelog := strings.TrimSpace(input.Changelog); changelog != "" {
		version.Changelog = &changelog
	}
	if promptTemplate := strings.TrimSpace(input.PromptTemplate); promptTemplate != "" {
		version.PromptTemplate = &promptTemplate
	}
	if systemInstructions := strings.TrimSpace(input.SystemInstructions); systemInstructions != "" {
		version.SystemInstructions = &systemInstructions
	}

	if _, err := s.versions.SaveCurrent(ctx, version); err != nil {
		return fmt.Errorf("save skill version: %w", err)
	}
	return nil
}

func validateSkillContract(input SkillMutationInput) error {
	inputSchema, err := normalizeSchemaObject("input_schema", input.InputSchema)
	if err != nil {
		return err
	}
	if _, err := normalizeSchemaObject("output_schema", input.OutputSchema); err != nil {
		return err
	}
	if _, err := normalizeSchemaObject("error_schema", input.ErrorSchema); err != nil {
		return err
	}

	templateVariables := extractTemplateVariables(input.PromptTemplate)
	if len(templateVariables) == 0 {
		return nil
	}

	allowedVariables := extractInputSchemaProperties(inputSchema)
	if len(allowedVariables) == 0 {
		return &SkillContractValidationError{
			Message: "prompt_template uses variables, but input_schema.properties is empty",
		}
	}

	for _, variable := range templateVariables {
		if !slices.Contains(allowedVariables, variable) {
			return &SkillContractValidationError{
				Message: fmt.Sprintf("prompt_template variable %q is not declared in input_schema.properties", variable),
			}
		}
	}

	return nil
}

func normalizeSchemaObject(field string, value map[string]any) (map[string]any, error) {
	if len(value) == 0 {
		return nil, nil
	}

	propertiesValue, hasProperties := value["properties"]
	if hasProperties {
		properties, ok := propertiesValue.(map[string]any)
		if !ok {
			return nil, &SkillContractValidationError{
				Message: fmt.Sprintf("%s.properties must be a JSON object", field),
			}
		}
		for propertyName, propertyValue := range properties {
			if _, ok := propertyValue.(map[string]any); !ok {
				return nil, &SkillContractValidationError{
					Message: fmt.Sprintf("%s.properties.%s must be a JSON object", field, propertyName),
				}
			}
		}
	}

	requiredValue, hasRequired := value["required"]
	if hasRequired {
		requiredFields, ok := requiredValue.([]any)
		if !ok {
			return nil, &SkillContractValidationError{
				Message: fmt.Sprintf("%s.required must be an array of strings", field),
			}
		}
		for _, requiredField := range requiredFields {
			if _, ok := requiredField.(string); !ok {
				return nil, &SkillContractValidationError{
					Message: fmt.Sprintf("%s.required must contain only strings", field),
				}
			}
		}
	}

	return value, nil
}

var promptVariablePattern = regexp.MustCompile(`{{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}`)

func extractTemplateVariables(template string) []string {
	matches := promptVariablePattern.FindAllStringSubmatch(template, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	var variables []string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		if _, ok := seen[match[1]]; ok {
			continue
		}
		seen[match[1]] = struct{}{}
		variables = append(variables, match[1])
	}
	return variables
}

func extractInputSchemaProperties(schema map[string]any) []string {
	if len(schema) == 0 {
		return nil
	}

	propertiesValue, ok := schema["properties"]
	if !ok {
		return nil
	}

	properties, ok := propertiesValue.(map[string]any)
	if !ok {
		return nil
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func cloneJSONMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
