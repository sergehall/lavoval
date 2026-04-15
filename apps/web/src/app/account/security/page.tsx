import { loadMFAState } from '@/features/auth/actions';
import { MFASettings } from '@/features/auth/mfa-settings';
import { fetchAccountSecurity, withValidSession } from '@/shared/api/server-client';

export default async function AccountSecurityPage() {
  const [mfaState, accountSecurity] = await Promise.all([
    loadMFAState(),
    withValidSession((session) => fetchAccountSecurity(session.accessToken).then((response) => response.data)),
  ]);

  return (
    <div className="stack stack--lg">
      <section className="section-heading">
        <div className="stack stack--sm">
          <h1>Security</h1>
          <p className="muted">
            Add an authenticator app to your account before we enforce MFA during sign-in.
          </p>
        </div>
      </section>
      <MFASettings initialState={mfaState} accountSecurity={accountSecurity} />
    </div>
  );
}
