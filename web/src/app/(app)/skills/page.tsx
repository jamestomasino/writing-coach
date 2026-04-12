import { TGOLibraryView } from '@/components/tgo-library-view'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Skill Objectives',
}

export default function SkillsPage() {
  return <TGOLibraryView />
}
