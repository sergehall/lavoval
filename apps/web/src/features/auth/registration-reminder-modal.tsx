'use client';

type RegistrationReminderModalProps = {
  isOpen: boolean;
  email?: string;
  onClose: () => void;
  onShowSignIn: () => void;
};

export function RegistrationReminderModal({
  isOpen,
  email,
  onClose,
  onShowSignIn,
}: RegistrationReminderModalProps) {
  if (!isOpen) {
    return null;
  }

  return (
    <div className="auth-modal auth-modal--visible" role="dialog" aria-modal="true">
      <button
        type="button"
        className="auth-modal__overlay"
        aria-label="Close registration reminder"
        onClick={onClose}
      />
      <div className="auth-modal__panel auth-modal__panel--compact auth-reminder">
        <button type="button" className="auth-modal__close" onClick={onClose} aria-label="Close">
          ×
        </button>
        <div className="stack stack--lg">
          <div className="stack stack--sm">
            <span className="badge badge--success">Registration complete</span>
            <h2>Check your inbox and confirm your email.</h2>
            <p className="muted">
              We just sent a confirmation link to the address below. Open that email, confirm your
              Lavoval account, and then come back to sign in.
            </p>
          </div>

          <div className="auth-reminder__email">
            <span className="auth-reminder__label">Confirmation address</span>
            <strong>{email || 'Use the email address you just registered.'}</strong>
          </div>

          <div className="auth-reminder__callout">
            <strong>What happens now</strong>
            <p>
              Lavoval now waits for email confirmation before the first sign-in. If the message
              does not land right away, check spam or request a fresh confirmation link from the
              sign-in screen.
            </p>
          </div>

          <div className="toolbar">
            <button type="button" className="site-nav__link site-nav__link--subtle" onClick={onClose}>
              Close
            </button>
            <button type="button" className="site-nav__link site-nav__link--cta" onClick={onShowSignIn}>
              Return to sign in
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
