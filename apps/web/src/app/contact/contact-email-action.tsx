'use client';

import type { ReactNode } from 'react';
import { Button } from '@/shared/ui/button';

const contactEmailCodes = [
  115, 101, 114, 103, 101, 46, 104, 97, 108, 108, 46, 100, 101, 118, 64, 103, 109, 97, 105, 108, 46,
  99, 111, 109,
] as const;

type ContactEmailActionProps = {
  children: ReactNode;
  className?: string;
  variant?: 'primary' | 'secondary' | 'ghost';
};

function getContactEmail() {
  return String.fromCharCode(...contactEmailCodes);
}

export function ContactEmailAction({
  children,
  className,
  variant = 'primary',
}: ContactEmailActionProps) {
  return (
    <Button
      type="button"
      className={className}
      variant={variant}
      aria-label="Start an email to Lavoval"
      onClick={() => {
        const mailProtocol = String.fromCharCode(109, 97, 105, 108, 116, 111, 58);
        window.location.assign(`${mailProtocol}${getContactEmail()}`);
      }}
    >
      {children}
    </Button>
  );
}
