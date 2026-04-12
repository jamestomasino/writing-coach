import { SkillsLibraryView } from '@/components/skills-library-view'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Skills',
}

export default function SkillsPage() {
  return <SkillsLibraryView />
}
