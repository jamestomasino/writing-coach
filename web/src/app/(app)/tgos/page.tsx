import { TGOLibraryView } from '@/components/tgo-library-view'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Skill objective guides',
}

export default function TGOPage() {
  return <TGOLibraryView />
}
