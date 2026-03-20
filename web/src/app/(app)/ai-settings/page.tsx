import { AIProviderSettingsView } from '@/components/ai-provider-settings-view'

export default function AISettingsPage({
  searchParams,
}: {
  searchParams?: { required?: string; next?: string }
}) {
  return <AIProviderSettingsView required={searchParams?.required === '1'} nextPath={searchParams?.next} />
}
