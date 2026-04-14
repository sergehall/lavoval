import { redirect } from 'next/navigation';

export default async function AccountSkillDetailRedirect({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  redirect(`/skills/${id}`);
}
