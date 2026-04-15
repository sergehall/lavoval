import { loadMFAState } from '@/features/auth/actions';
import { MFASettings } from '@/features/auth/mfa-settings';

export default async function AccountSecurityPage() {
  const mfaState = await loadMFAState();

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
      <MFASettings initialState={mfaState} />
    </div>
  );
}
