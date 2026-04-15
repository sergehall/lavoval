'use client';

import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { mfaSettingsAction, type MFAState } from '@/features/auth/actions';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';

export function MFASettings({ initialState }: { initialState: MFAState }) {
  const [state, action] = useActionState(mfaSettingsAction, initialState);

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="stack stack--md">
          <div className="stack stack--xs">
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
            <div className="stack stack--md">
              <div className="stack stack--xs">
                <p>
                  Add this shared secret to your authenticator app, then enter the current 6-digit
                  code to finish enrollment.
                </p>
                <SecretPanel state={state} />
              </div>

              <form action={action} className="stack stack--md">
                <input type="hidden" name="intent" value="verify" />
                <label>
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
                <div className="stack stack--md">
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
                  <form action={action} className="stack stack--md">
                    <input type="hidden" name="intent" value="regenerate" />
                    <label>
                      <span>Current password</span>
                      <Input
                        type="password"
                        name="password"
                        placeholder="Enter your password"
                        required
                        minLength={8}
                      />
                    </label>
                    <label>
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
                <div className="stack stack--md">
                  <div className="stack stack--xs">
                    <h3>Disable MFA</h3>
                    <p className="muted">
                      Turn off authenticator-based protection for this account. We will revoke other
                      active sessions after this change.
                    </p>
                  </div>
                  <form action={action} className="stack stack--md">
                    <input type="hidden" name="intent" value="disable" />
                    <label>
                      <span>Current password</span>
                      <Input
                        type="password"
                        name="password"
                        placeholder="Enter your password"
                        required
                        minLength={8}
                      />
                    </label>
                    <label>
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
  return (
    <div className="stack stack--sm">
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
    <div className="stack stack--xs">
      <strong>
        Status:{' '}
        {state.enabled
          ? 'Enabled'
          : state.pendingEnrollment
            ? 'Pending verification'
            : 'Not enabled'}
      </strong>
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
