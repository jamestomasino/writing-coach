import { NextResponse } from 'next/server'
import { isSupportedLocale, localeCookieName } from '@/i18n/config'

const localeCookieMaxAge = 60 * 60 * 24 * 365

export async function POST(request: Request) {
  let payload: { locale?: string } = {}
  try {
    payload = (await request.json()) as { locale?: string }
  } catch {}

  if (!isSupportedLocale(payload.locale)) {
    return NextResponse.json({ error: 'invalid locale' }, { status: 400 })
  }

  const response = NextResponse.json({ ok: true })
  response.cookies.set(localeCookieName, payload.locale, {
    maxAge: localeCookieMaxAge,
    path: '/',
    sameSite: 'lax',
  })
  return response
}
