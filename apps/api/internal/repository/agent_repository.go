package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type AgentRepository struct {
	pool *pgxpool.Pool
}

func NewAgentRepository(pool *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{pool: pool}
}

const agentCols = `
	id, slug, name, provider, model_name, description,
	supports_text, supports_code, supports_tools, supports_web, supports_files,
	supports_multimodal, supports_json_output, max_context_tokens,
	pricing_json, status, created_at, updated_at`

func (r *AgentRepository) List(ctx context.Context) ([]domain.Agent, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+agentCols+` FROM agents WHERE status = 'active' ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var out []domain.Agent
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AgentRepository) FindByID(ctx context.Context, id string) (domain.Agent, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+agentCols+` FROM agents WHERE id = $1`, id)
	return scanAgent(row)
}

func (r *AgentRepository) FindBySlug(ctx context.Context, slug string) (domain.Agent, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+agentCols+` FROM agents WHERE slug = $1`, slug)
	return scanAgent(row)
}

func (r *AgentRepository) ListRecommendedForSkill(ctx context.Context, skillID string) ([]domain.SkillAgentCompatibility, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			sac.skill_id, sac.agent_id,
			sac.compatibility_score, sac.success_rate, sac.avg_rating, sac.runs_count,
			sac.tested_by_system, sac.tested_by_users, sac.notes,
			a.id, a.slug, a.name, a.provider, a.model_name, a.description,
			a.supports_text, a.supports_code, a.supports_tools, a.supports_web, a.supports_files,
			a.supports_multimodal, a.supports_json_output, a.max_context_tokens,
			a.pricing_json, a.status, a.created_at, a.updated_at
		FROM skill_agent_compatibility sac
		INNER JOIN agents a ON a.id = sac.agent_id
		WHERE sac.skill_id = $1 AND a.status = 'active'
		ORDER BY sac.compatibility_score DESC, sac.avg_rating DESC`, skillID)
	if err != nil {
		return nil, fmt.Errorf("list recommended agents: %w", err)
	}
	defer rows.Close()

	var out []domain.SkillAgentCompatibility
	for rows.Next() {
		var c domain.SkillAgentCompatibility
		var pricingRaw []byte
		if err := rows.Scan(
			&c.SkillID, &c.AgentID,
			&c.CompatibilityScore, &c.SuccessRate, &c.AvgRating, &c.RunsCount,
			&c.TestedBySystem, &c.TestedByUsers, &c.Notes,
			&c.Agent.ID, &c.Agent.Slug, &c.Agent.Name, &c.Agent.Provider, &c.Agent.ModelName,
			&c.Agent.Description,
			&c.Agent.SupportsText, &c.Agent.SupportsCode, &c.Agent.SupportsTools,
			&c.Agent.SupportsWeb, &c.Agent.SupportsFiles, &c.Agent.SupportsMultimodal,
			&c.Agent.SupportsJSONOutput, &c.Agent.MaxContextTokens,
			&pricingRaw, &c.Agent.Status, &c.Agent.CreatedAt, &c.Agent.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan compatibility: %w", err)
		}
		if len(pricingRaw) > 0 {
			_ = json.Unmarshal(pricingRaw, &c.Agent.PricingJSON)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanAgent(row scanner) (domain.Agent, error) {
	var a domain.Agent
	var pricingRaw []byte
	if err := row.Scan(
		&a.ID, &a.Slug, &a.Name, &a.Provider, &a.ModelName, &a.Description,
		&a.SupportsText, &a.SupportsCode, &a.SupportsTools, &a.SupportsWeb, &a.SupportsFiles,
		&a.SupportsMultimodal, &a.SupportsJSONOutput, &a.MaxContextTokens,
		&pricingRaw, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return domain.Agent{}, fmt.Errorf("scan agent: %w", err)
	}
	if len(pricingRaw) > 0 {
		_ = json.Unmarshal(pricingRaw, &a.PricingJSON)
	}
	return a, nil
}
