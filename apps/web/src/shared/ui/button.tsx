import clsx from 'clsx';
import type { ButtonHTMLAttributes, PropsWithChildren } from 'react';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';

type ButtonProps = PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement>> & {
  variant?: ButtonVariant;
  fullWidth?: boolean;
};

const variantClassName: Record<ButtonVariant, string> = {
  primary: 'button button--primary',
  secondary: 'button button--secondary',
  ghost: 'button button--ghost',
  danger: 'button button--danger',
};

export function Button({
  children,
  variant = 'primary',
  fullWidth = false,
  className,
  ...props
}: ButtonProps) {
  return (
    <button
      className={clsx(variantClassName[variant], fullWidth && 'button--full', className)}
      {...props}
    >
      {children}
    </button>
  );
}
