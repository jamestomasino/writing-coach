import type { Metadata } from 'next'
import { redirect } from 'next/navigation'

export const metadata: Metadata = {
  title: 'Skill objective guides',
}

export default function TGOPage() {
  redirect('/skills')
}
