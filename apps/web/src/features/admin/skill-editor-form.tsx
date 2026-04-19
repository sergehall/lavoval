'use client';

import { useState } from 'react';
import {
  defaultSkillProvider,
  formatSkillConfig,
  suggestedSkillEntrypoints,
} from '@lavoval/registry';
import type { SkillDetail } from '@lavoval/registry';
import type { Category, Tag } from '@lavoval/contracts';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Textarea } from '@/shared/ui/textarea';

type ExtendedSkill = SkillDetail & {
  categoryId?: string;
  subcategoryId?: string;
  skillType?: string;
  difficulty?: string;
  isAgentReady?: boolean;
  estimatedTimeMinutes?: number;
  languageCode?: string;
  tags?: Array<{ id: string; name: string; slug: string }>;
};

type SkillEditorFormProps = {
  skill?: ExtendedSkill;
  categories?: Category[];
  tags?: Tag[];
  action: (formData: FormData) => void | Promise<void>;
  archiveAction?: (formData: FormData) => void | Promise<void>;
  submitLabel?: string;
  archiveLabel?: string;
};

const SKILL_TYPE_OPTIONS = [
  { value: '', label: 'Not specified' },
  { value: 'guide', label: 'Guide' },
  { value: 'workflow', label: 'Workflow' },
  { value: 'prompt-pack', label: 'Prompt pack' },
  { value: 'agent-ready', label: 'Agent-ready' },
  { value: 'service', label: 'Service' },
];

const DIFFICULTY_OPTIONS = [
  { value: '', label: 'Not specified' },
  { value: 'junior', label: 'Junior' },
  { value: 'middle', label: 'Middle' },
  { value: 'senior', label: 'Senior' },
];

function FieldLabel({ label, hint }: { label: string; hint: string }) {
  return (
    <span className="field-label">
      <span>{label}</span>
      <span className="field-hint" tabIndex={0} aria-label={`${label} help`}>
        ?<span className="field-hint__tooltip">{hint}</span>
      </span>
    </span>
  );
}

function formatJsonValue(value: Record<string, unknown> | null | undefined) {
  if (!value || Object.keys(value).length === 0) {
    return '';
  }
  return JSON.stringify(value, null, 2);
}

function parsePreviewInput(raw: string) {
  const value = raw.trim();
  if (!value) {
    return {};
  }

  const parsed = JSON.parse(value) as Record<string, unknown>;
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error('Preview input must be a JSON object.');
  }
  return parsed;
}

function renderPromptPreview(template: string, previewInput: string) {
  const trimmedTemplate = template.trim();
  if (!trimmedTemplate) {
    return {
      text: 'Add a prompt template above to see a rendered preview.',
      error: '',
    };
  }

  try {
    const values = parsePreviewInput(previewInput);
    const rendered = trimmedTemplate.replace(
      /{{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}/g,
      (_, variable) => {
        const value = values[variable];
        if (value == null) {
          return '';
        }
        return typeof value === 'string' ? value : JSON.stringify(value, null, 2);
      },
    );

    return {
      text: rendered,
      error: '',
    };
  } catch (error) {
    return {
      text: '',
      error: error instanceof Error ? error.message : 'Could not render preview.',
    };
  }
}

export function SkillEditorForm({
  skill,
  categories = [],
  tags = [],
  action,
  archiveAction,
  submitLabel,
  archiveLabel = 'Archive skill',
}: SkillEditorFormProps) {
  const skillTagIds = skill?.tags?.map((t) => t.id) ?? [];
  const currentVersion = skill?.currentVersion;

  const [inputSchemaText, setInputSchemaText] = useState(
    formatJsonValue(currentVersion?.inputSchema as Record<string, unknown> | undefined),
  );
  const [outputSchemaText, setOutputSchemaText] = useState(
    formatJsonValue(currentVersion?.outputSchema as Record<string, unknown> | undefined),
  );
  const [errorSchemaText, setErrorSchemaText] = useState(
    formatJsonValue(currentVersion?.errorSchema as Record<string, unknown> | undefined),
  );
  const [promptTemplate, setPromptTemplate] = useState(currentVersion?.promptTemplate ?? '');
  const [previewInput, setPreviewInput] = useState(
    formatJsonValue(currentVersion?.inputSchema?.example as Record<string, unknown> | undefined),
  );

  const preview = renderPromptPreview(promptTemplate, previewInput);

  return (
    <form action={action} className="stack stack--md form-grid">
      <label>
        <FieldLabel
          label="Slug"
          hint="Defines the URL key for this skill. Expected format: lowercase English words separated with hyphens, for example `prompt-engineering-basics`. Avoid spaces, uppercase characters, and special symbols."
        />
        <Input
          name="slug"
          defaultValue={skill?.slug}
          placeholder="prompt-engineering-basics"
          required
        />
      </label>
      <label>
        <FieldLabel
          label="Title"
          hint="Defines the marketplace-facing name shown in discovery, detail, and account views. Keep it short, clear, and compelling as an offer title."
        />
        <Input
          name="title"
          defaultValue={skill?.title}
          placeholder="Prompt Engineering Basics"
          required
        />
      </label>
      <label>
        <FieldLabel
          label="Provider"
          hint="Defines which execution provider this skill uses. For the first runtime milestone, keep this as `internal` so built-in executors can handle the request."
        />
        <Input
          name="provider"
          defaultValue={skill?.provider ?? defaultSkillProvider}
          placeholder={defaultSkillProvider}
          required
        />
      </label>
      <label>
        <FieldLabel
          label="Entrypoint"
          hint="Defines the executor key used by runtime lookup. Good starter values are `echo`, `text-summary-mock`, and `keyword-extract-mock`."
        />
        <Input
          name="entrypoint"
          defaultValue={skill?.entrypoint ?? suggestedSkillEntrypoints[0]}
          placeholder={suggestedSkillEntrypoints[0]}
          required
        />
        <span className="muted">
          Suggested starter entrypoints: {suggestedSkillEntrypoints.join(', ')}.
        </span>
      </label>
      <label className="form-grid__full">
        <FieldLabel
          label="Summary"
          hint="Defines the short preview used in cards and lists. Recommended format: 1-2 sentences covering what the offer teaches, who it helps, and why it matters."
        />
        <Textarea
          name="summary"
          defaultValue={skill?.summary}
          placeholder="A short 1-2 sentence description explaining what this offer helps someone learn and why it is valuable."
          required
          rows={3}
        />
      </label>
      <label className="form-grid__full">
        <FieldLabel
          label="Description"
          hint="Defines the main offer overview in markdown. Use it for context, audience, outcomes, structure, exchange expectations, references, and any content that should appear on the full page."
        />
        <Textarea
          name="description"
          defaultValue={skill?.description}
          placeholder={`# What you will learn

This offer helps people learn...

## Why it matters now

- AI-native teams who still need human judgment
- Operators who want reusable practical know-how

## Who it is for

- People trying to level up quickly
- Teams looking for real-world experience

## Expected outcome

After working through this offer, someone will be able to...`}
          required
          rows={10}
        />
        <span className="muted">
          Use markdown here. This becomes the full marketplace page for the offer: context,
          audience, outcomes, structure, and any key notes.
        </span>
      </label>
      <label className="form-grid__full">
        <FieldLabel
          label="Runtime config"
          hint="Defines executor-specific configuration as JSON. Keep this valid JSON so runtime can persist and pass it into the selected entrypoint."
        />
        <Textarea
          name="config"
          defaultValue={formatSkillConfig(skill?.config)}
          placeholder={`{
  "maxTokens": 500
}`}
          rows={6}
        />
        <span className="muted">
          Use valid JSON here. For the first mock executors, an empty object is completely fine.
        </span>
      </label>

      <div className="form-grid__full stack stack--md" style={{ paddingTop: '8px' }}>
        <div className="section-heading">
          <div className="stack stack--sm">
            <h2 style={{ margin: 0 }}>AI contract</h2>
            <p className="muted" style={{ margin: 0 }}>
              Make this skill machine-readable for agents: schemas, prompt template, system
              instructions, and changelog are versioned together.
            </p>
          </div>
          {currentVersion ? (
            <span className="badge badge--neutral">v{currentVersion.versionNo}</span>
          ) : null}
        </div>

        <label className="form-grid__full">
          <FieldLabel
            label="Input schema"
            hint="JSON Schema-like object describing what the agent must send. Keep `properties` and `required` aligned because prompt variables are validated against this object."
          />
          <Textarea
            name="inputSchema"
            value={inputSchemaText}
            onChange={(event) => setInputSchemaText(event.target.value)}
            placeholder={`{
  "type": "object",
  "required": ["topic", "level"],
  "properties": {
    "topic": { "type": "string", "description": "Subject to explain" },
    "level": { "type": "string", "enum": ["beginner", "intermediate", "expert"] }
  }
}`}
            rows={12}
          />
        </label>

        <label className="form-grid__full">
          <FieldLabel
            label="Output schema"
            hint="Structured response contract the skill should produce. This becomes the machine-readable output promise for downstream tooling."
          />
          <Textarea
            name="outputSchema"
            value={outputSchemaText}
            onChange={(event) => setOutputSchemaText(event.target.value)}
            placeholder={`{
  "type": "object",
  "properties": {
    "explanation": { "type": "string" },
    "keyPoints": { "type": "array", "items": { "type": "string" } }
  }
}`}
            rows={10}
          />
        </label>

        <label className="form-grid__full">
          <FieldLabel
            label="Error schema"
            hint="Describe machine-readable error codes so an agent can react predictably to validation, policy, or runtime failures."
          />
          <Textarea
            name="errorSchema"
            value={errorSchemaText}
            onChange={(event) => setErrorSchemaText(event.target.value)}
            placeholder={`{
  "type": "object",
  "properties": {
    "code": { "type": "string" },
    "message": { "type": "string" }
  }
}`}
            rows={8}
          />
        </label>

        <label className="form-grid__full">
          <FieldLabel
            label="System instructions"
            hint="Persistent behavior rules for the skill runtime. Use this for guardrails, tone, boundaries, and non-negotiable operating instructions."
          />
          <Textarea
            name="systemInstructions"
            defaultValue={currentVersion?.systemInstructions ?? ''}
            placeholder="You are a precise teaching assistant. Stay within the provided schema and do not invent unsupported facts."
            rows={6}
          />
        </label>

        <label className="form-grid__full">
          <FieldLabel
            label="Prompt template"
            hint="Templated prompt body. Use `{{variable_name}}` placeholders that match fields from `inputSchema.properties`."
          />
          <Textarea
            name="promptTemplate"
            value={promptTemplate}
            onChange={(event) => setPromptTemplate(event.target.value)}
            placeholder={`Explain {{topic}} for a {{level}} audience.

Return JSON that matches the output schema exactly.`}
            rows={8}
          />
          <span className="muted">
            Current validation supports <code>{'{{variable}}'}</code> placeholders and checks them
            against the input schema field names.
          </span>
        </label>

        <label className="form-grid__full">
          <FieldLabel
            label="Preview input"
            hint="Sample JSON input used only for previewing the rendered prompt before save."
          />
          <Textarea
            value={previewInput}
            onChange={(event) => setPreviewInput(event.target.value)}
            placeholder={`{
  "topic": "vector databases",
  "level": "beginner"
}`}
            rows={8}
          />
        </label>

        <div
          className="form-grid__full stack stack--sm"
          style={{ border: '1px solid var(--border)', borderRadius: '20px', padding: '16px' }}
        >
          <div className="section-heading">
            <strong>Prompt preview</strong>
            <span className="muted" style={{ fontSize: '0.85rem' }}>
              Rendered locally from the current template and sample input
            </span>
          </div>
          {preview.error ? (
            <p className="muted" style={{ color: 'var(--danger, #b42318)' }}>
              {preview.error}
            </p>
          ) : (
            <pre
              style={{
                margin: 0,
                whiteSpace: 'pre-wrap',
                fontFamily: 'var(--font-mono, monospace)',
              }}
            >
              {preview.text}
            </pre>
          )}
        </div>

        <label className="form-grid__full">
          <FieldLabel
            label="Version changelog"
            hint="Short note describing what changed in this version: schema updates, prompt edits, safety changes, or behavior shifts."
          />
          <Textarea
            name="changelog"
            defaultValue=""
            placeholder="Added structured output schema and aligned prompt variables with input_schema."
            rows={4}
          />
        </label>
      </div>

      {categories.length > 0 && (
        <label>
          <FieldLabel
            label="Category"
            hint="Groups this skill in the marketplace browse tree so buyers can filter by domain."
          />
          <select name="categoryId" defaultValue={skill?.categoryId ?? ''} className="input">
            <option value="">No category</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
      )}

      <label>
        <FieldLabel
          label="Skill type"
          hint="Describes the format of the offer. Guide = readable walkthrough, Workflow = multi-step process, Prompt pack = reusable prompts, Agent-ready = can be invoked by an AI agent, Service = ongoing delivery."
        />
        <select name="skillType" defaultValue={skill?.skillType ?? ''} className="input">
          {SKILL_TYPE_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      </label>

      <label>
        <FieldLabel
          label="Difficulty"
          hint="Sets the expected experience level for someone benefiting from this offer. Helps buyers self-select the right match."
        />
        <select name="difficulty" defaultValue={skill?.difficulty ?? ''} className="input">
          {DIFFICULTY_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      </label>

      <label>
        <FieldLabel
          label="Estimated time (minutes)"
          hint="Approximate time in minutes needed to work through or apply this skill. Shown on the detail page to help buyers plan."
        />
        <Input
          type="number"
          name="estimatedTimeMinutes"
          defaultValue={skill?.estimatedTimeMinutes ?? ''}
          placeholder="30"
          min={1}
          max={9999}
        />
      </label>

      <label>
        <FieldLabel
          label="Language"
          hint="ISO 639-1 language code for the primary language of this skill. Use `en` for English, `ru` for Russian, etc."
        />
        <Input
          name="languageCode"
          defaultValue={skill?.languageCode ?? 'en'}
          placeholder="en"
          maxLength={5}
        />
      </label>

      <label>
        <FieldLabel
          label="Agent-ready"
          hint="Mark this skill as usable by AI agents. Only check this if the skill has a well-defined entrypoint and structured input/output."
        />
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', paddingTop: '4px' }}>
          <input
            type="checkbox"
            name="isAgentReady"
            value="true"
            defaultChecked={skill?.isAgentReady ?? false}
          />
          <span className="muted" style={{ fontSize: '0.9rem' }}>
            This skill can be invoked by an AI agent
          </span>
        </div>
      </label>

      {tags.length > 0 && (
        <label className="form-grid__full">
          <FieldLabel
            label="Tags"
            hint="Select tags that describe the topics and tools covered. Buyers filter by tag in the catalog. Hold Ctrl (or Cmd) to select multiple."
          />
          <select
            name="tagIds"
            multiple
            className="input"
            style={{ height: `${Math.min(tags.length, 8) * 2 + 2}rem` }}
            defaultValue={skillTagIds}
          >
            {tags.map((tag) => (
              <option key={tag.id} value={tag.id}>
                {tag.name}
              </option>
            ))}
          </select>
          <span className="muted">Hold Ctrl / Cmd to select multiple tags.</span>
        </label>
      )}

      <label>
        <FieldLabel
          label="Status"
          hint="Controls lifecycle state. `Draft` keeps the offer in progress, `Published` makes it discoverable, and `Archived` removes it from active circulation."
        />
        <select name="status" defaultValue={skill?.status ?? 'draft'} className="input">
          <option value="draft">Draft</option>
          <option value="published">Published</option>
          <option value="archived">Archived</option>
        </select>
      </label>
      <label>
        <FieldLabel
          label="Visibility"
          hint="Controls audience scope. `Private` keeps the offer inside account and governance workflows. `Public` allows it to appear in the marketplace once ready."
        />
        <select name="visibility" defaultValue={skill?.visibility ?? 'private'} className="input">
          <option value="private">Private</option>
          <option value="public">Public</option>
        </select>
      </label>
      <div className="form-grid__full toolbar">
        <Button type="submit">{submitLabel ?? (skill ? 'Save offer' : 'Create offer')}</Button>
        {skill && archiveAction ? (
          <button className="button button--danger" formAction={archiveAction}>
            {archiveLabel}
          </button>
        ) : null}
      </div>
    </form>
  );
}
