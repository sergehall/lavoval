'use client';

import { useActionState, useEffect, useState } from 'react';
import { useFormStatus } from 'react-dom';
import type { AccountSecuritySummary } from '@lavoval/contracts';
import type { ReactNode } from 'react';
import QRCode from 'qrcode';
import { mfaSettingsAction, type MFAState } from '@/features/auth/actions';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';

export function MFASettings({
  initialState,
  accountSecurity,
}: {
  initialState: MFAState;
  accountSecurity: AccountSecuritySummary;
}) {
  const [state, action] = useActionState(mfaSettingsAction, initialState);

  return (
    <div className="security-page stack stack--lg">
      <Card>
        <div className="security-section security-section--compact stack stack--md">
          <div className="security-section__heading stack stack--xs">
            <h2>Sign-in methods</h2>
            <p className="muted">
              Review how this account can be accessed today before we add provider linking and
              session controls.
            </p>
          </div>
          <SignInMethodsPanel accountSecurity={accountSecurity} state={state} />
        </div>
      </Card>

      <Card>
        <div className="security-section stack stack--md">
          <div className="security-section__heading stack stack--xs">
            <h2>Authenticator App</h2>
            <p className="muted">
              Protect your Lavoval account with a time-based 6-digit code from apps like 1Password,
              Google Authenticator, or Authy.
            </p>
          </div>

          <StatusBlock state={state} />

          {state.error ? (
            <p className="form-message form-message--error" role="alert">
              {state.error}
            </p>
          ) : null}
          {state.success ? <SuccessPanel state={state} /> : null}

          {!state.enabled && !state.pendingEnrollment ? (
            <form action={action}>
              <input type="hidden" name="intent" value="enroll" />
              <SubmitButton idleLabel="Set up MFA" pendingLabel="Starting setup..." />
            </form>
          ) : null}

          {!state.enabled && state.pendingEnrollment ? (
            <div className="security-flow stack stack--md">
              <div className="stack stack--xs">
                <p>
                  Add this shared secret to your authenticator app, then enter the current 6-digit
                  code to finish enrollment.
                </p>
                <div className="security-inline-actions">
                  <form action={action}>
                    <input type="hidden" name="intent" value="enroll" />
                    <Button type="submit" variant="secondary" className="security-method-card__button">
                      Start setup again
                    </Button>
                  </form>
                  <form action={action}>
                    <input type="hidden" name="intent" value="cancel" />
                    <Button type="submit" variant="ghost" className="security-method-card__button">
                      Cancel setup
                    </Button>
                  </form>
                </div>
                <SecretPanel state={state} />
              </div>

              <form action={action} className="security-form stack stack--md">
                <input type="hidden" name="intent" value="verify" />
                <label className="security-form__label">
                  <span>Authenticator code</span>
                  <Input
                    name="code"
                    inputMode="numeric"
                    pattern="[0-9]{6}"
                    minLength={6}
                    maxLength={6}
                    placeholder="123456"
                    required
                  />
                </label>
                <SubmitButton idleLabel="Verify and enable MFA" pendingLabel="Verifying..." />
              </form>
            </div>
          ) : null}

          {state.enabled ? (
            <div className="stack stack--lg">
              <Card>
                <div className="security-nested-card stack stack--md">
                  <div className="stack stack--xs">
                    <h3>Recovery codes</h3>
                    <p className="muted">
                      Regenerate your emergency backup codes if you think the current set has been
                      exposed or misplaced.
                    </p>
                  </div>
                  {state.recoveryCodes?.length ? (
                    <RecoveryCodesList codes={state.recoveryCodes} />
                  ) : null}
                  <form action={action} className="security-form stack stack--md">
                    <input type="hidden" name="intent" value="regenerate" />
                    <label className="security-form__label">
                      <span>Current password</span>
                      <Input
                        type="password"
                        name="password"
                        placeholder="Enter your password"
                        required
                        minLength={8}
                      />
                    </label>
                    <label className="security-form__label">
                      <span>Authenticator code</span>
                      <Input
                        name="code"
                        inputMode="numeric"
                        pattern="[0-9]{6}"
                        minLength={6}
                        maxLength={6}
                        placeholder="123456"
                        required
                      />
                    </label>
                    <SubmitButton
                      idleLabel="Regenerate recovery codes"
                      pendingLabel="Regenerating..."
                    />
                  </form>
                </div>
              </Card>

              <Card>
                <div className="security-nested-card stack stack--md">
                  <div className="stack stack--xs">
                    <h3>Disable MFA</h3>
                    <p className="muted">
                      Turn off authenticator-based protection for this account. We will revoke other
                      active sessions after this change.
                    </p>
                  </div>
                  <form action={action} className="security-form stack stack--md">
                    <input type="hidden" name="intent" value="disable" />
                    <label className="security-form__label">
                      <span>Current password</span>
                      <Input
                        type="password"
                        name="password"
                        placeholder="Enter your password"
                        required
                        minLength={8}
                      />
                    </label>
                    <label className="security-form__label">
                      <span>Authenticator code</span>
                      <Input
                        name="code"
                        inputMode="numeric"
                        pattern="[0-9]{6}"
                        minLength={6}
                        maxLength={6}
                        placeholder="123456"
                        required
                      />
                    </label>
                    <SubmitButton
                      idleLabel="Disable MFA"
                      pendingLabel="Disabling..."
                      variant="danger"
                    />
                  </form>
                </div>
              </Card>
            </div>
          ) : null}
        </div>
      </Card>
    </div>
  );
}

function SignInMethodsPanel({
  accountSecurity,
  state,
}: {
  accountSecurity: AccountSecuritySummary;
  state: MFAState;
}) {
  const providerMap = new Map(accountSecurity.providers.map((provider) => [provider.provider, provider]));
  const recoveryCodesCount = state.recoveryCodes?.length ?? 0;

  return (
    <div className="security-methods">
      <MethodCard
        title="Password"
        description={
          accountSecurity.hasPassword
            ? 'A password is set for this account and can be used with the standard sign-in flow.'
            : 'No password is currently set for this account.'
        }
        badge={accountSecurity.hasPassword ? <Badge tone="success">Available</Badge> : <Badge tone="warning">Not set</Badge>}
        meta={
          accountSecurity.passwordUpdatedAt
            ? `Last updated ${formatSecurityDate(accountSecurity.passwordUpdatedAt)}.`
            : 'Set or change password controls will land here next.'
        }
      />
      <MethodCard
        title="Google"
        description={
          providerMap.has('google')
            ? 'Google sign-in is connected for this account.'
            : 'Google sign-in is available from the sign-in screen.'
        }
        badge={providerMap.has('google') ? <Badge tone="success">Connected</Badge> : <Badge>Not connected</Badge>}
        meta={
          providerMap.get('google')?.connectedAt
            ? `Connected ${formatSecurityDate(providerMap.get('google')!.connectedAt)}.`
            : 'Linking this provider inside Security will land next.'
        }
        actions={
          <Button type="button" variant="secondary" disabled className="security-method-card__button">
            {providerMap.has('google') ? 'Manage Google soon' : 'Link Google soon'}
          </Button>
        }
      />
      <MethodCard
        title="GitHub"
        description={
          providerMap.has('github')
            ? 'GitHub sign-in is connected for this account.'
            : 'GitHub sign-in is available from the sign-in screen.'
        }
        badge={providerMap.has('github') ? <Badge tone="success">Connected</Badge> : <Badge>Not connected</Badge>}
        meta={
          providerMap.get('github')?.connectedAt
            ? `Connected ${formatSecurityDate(providerMap.get('github')!.connectedAt)}.`
            : 'Linking this provider inside Security will land next.'
        }
        actions={
          <Button type="button" variant="secondary" disabled className="security-method-card__button">
            {providerMap.has('github') ? 'Manage GitHub soon' : 'Link GitHub soon'}
          </Button>
        }
      />
      <MethodCard
        title="Authenticator app"
        description={
          state.enabled
            ? 'Authenticator-based MFA is enabled for this account.'
            : state.pendingEnrollment
              ? 'Authenticator setup has started but still needs a 6-digit confirmation code.'
              : 'Authenticator-based MFA is not enabled yet.'
        }
        badge={
          state.enabled ? (
            <Badge tone="success">Enabled</Badge>
          ) : state.pendingEnrollment ? (
            <Badge tone="warning">Pending</Badge>
          ) : (
            <Badge>Disabled</Badge>
          )
        }
        meta={
          state.enrolledAt
            ? `Enabled ${formatSecurityDate(state.enrolledAt)}.`
            : 'Set up an authenticator app below to harden sign-in.'
        }
        extra={
          <div className="security-method-card__summary">
            <strong>Recovery codes</strong>
            <span className="muted">
              {state.enabled
                ? recoveryCodesCount > 0
                  ? `${recoveryCodesCount} backup codes are ready to use if you lose your authenticator.`
                  : 'MFA is enabled, but no recovery codes are currently loaded in this view.'
                : state.pendingEnrollment
                  ? 'Recovery codes will appear as soon as you verify the setup.'
                  : 'Recovery codes become available after MFA is enabled.'}
            </span>
          </div>
        }
      />
    </div>
  );
}

function MethodCard({
  title,
  description,
  badge,
  meta,
  extra,
  actions,
}: {
  title: string;
  description: string;
  badge: ReactNode;
  meta: string;
  extra?: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <article className="security-method-card">
      <div className="security-method-card__header">
        <h3>{title}</h3>
        {badge}
      </div>
      <p>{description}</p>
      <p className="muted">{meta}</p>
      {extra}
      {actions ? <div className="security-method-card__actions">{actions}</div> : null}
    </article>
  );
}

function formatSecurityDate(value: string) {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(new Date(value));
}

function SuccessPanel({ state }: { state: MFAState }) {
  return (
    <div className="stack stack--sm form-message form-message--success" role="status">
      {state.successTitle ? <strong>{state.successTitle}</strong> : null}
      <p>{state.success}</p>
      {state.recoveryCodes?.length ? <RecoveryCodesList codes={state.recoveryCodes} /> : null}
    </div>
  );
}

function SecretPanel({ state }: { state: MFAState }) {
  const [qrCodeDataURL, setQRCodeDataURL] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function renderQRCode() {
      if (!state.provisionUrl) {
        setQRCodeDataURL(null);
        return;
      }

      try {
        const dataURL = await QRCode.toDataURL(state.provisionUrl, {
          errorCorrectionLevel: 'M',
          margin: 1,
          scale: 8,
          width: 220,
          color: {
            dark: '#2f3742',
            light: '#fffaf6',
          },
        });

        if (!cancelled) {
          setQRCodeDataURL(dataURL);
        }
      } catch {
        if (!cancelled) {
          setQRCodeDataURL(null);
        }
      }
    }

    renderQRCode();

    return () => {
      cancelled = true;
    };
  }, [state.provisionUrl]);

  return (
    <div className="stack stack--sm">
      {qrCodeDataURL ? (
        <div className="mfa-qr-panel">
          <div className="stack stack--xs">
            <strong>Scan QR code</strong>
            <p className="muted">
              Open your authenticator app and scan this code instead of pasting the setup link
              manually.
            </p>
          </div>
          <img
            src={qrCodeDataURL}
            alt="QR code for authenticator app setup"
            className="mfa-qr-panel__image"
          />
        </div>
      ) : null}
      {state.secret ? (
        <div className="stack stack--xs">
          <strong>Shared secret</strong>
          <code>{state.secret}</code>
        </div>
      ) : null}
      {state.provisionUrl ? (
        <div className="stack stack--xs">
          <strong>Provisioning link</strong>
          <code>{state.provisionUrl}</code>
        </div>
      ) : null}
    </div>
  );
}

function RecoveryCodesList({ codes }: { codes: string[] }) {
  return (
    <div className="stack stack--xs">
      <strong>Recovery codes</strong>
      <p className="muted">
        Each code works once. Store them in a password manager or another secure place.
      </p>
      <div className="form-grid">
        {codes.map((code) => (
          <code key={code}>{code}</code>
        ))}
      </div>
    </div>
  );
}

function StatusBlock({ state }: { state: MFAState }) {
  return (
    <div className="security-status">
      <strong>Status</strong>
      <Badge
        tone={
          state.enabled ? 'success' : state.pendingEnrollment ? 'warning' : 'neutral'
        }
      >
        {state.enabled ? 'Enabled' : state.pendingEnrollment ? 'Pending verification' : 'Not enabled'}
      </Badge>
      {state.enrolledAt ? (
        <p className="muted">Enabled on {new Date(state.enrolledAt).toLocaleString()}.</p>
      ) : null}
    </div>
  );
}

function SubmitButton({
  idleLabel,
  pendingLabel,
  variant,
}: {
  idleLabel: string;
  pendingLabel: string;
  variant?: 'primary' | 'danger';
}) {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" variant={variant ?? 'primary'} disabled={pending} aria-disabled={pending}>
      {pending ? pendingLabel : idleLabel}
    </Button>
  );
}
