'use client';

export function GuestEntryBar({ onOpenSignIn }: { onOpenSignIn: () => void }) {
  return (
    <div className="site-auth-actions">
      <button type="button" className="guest-sign-in" onClick={onOpenSignIn}>
        Sign in
      </button>
    </div>
  );
}
