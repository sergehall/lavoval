import { redirect } from 'next/navigation';
import { signInHref } from '@/shared/lib/auth-navigation';

export default function LoginPage() {
  redirect(signInHref);
}
