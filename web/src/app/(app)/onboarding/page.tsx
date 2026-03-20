import { OnboardingView } from '@/components/onboarding-view'

export default async function OnboardingPage({ searchParams }: { searchParams?: Promise<{ mode?: string }> }) {
  const params = (await searchParams) ?? {}
  const mode = params.mode === 'create' ? 'create' : 'edit'
  return <OnboardingView mode={mode} />
}
