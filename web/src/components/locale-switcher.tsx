'use client'

import { useRouter } from 'next/navigation'
import { useLocale, useTranslations } from 'next-intl'
import { LanguageIcon } from '@heroicons/react/20/solid'
import { availableLocales, localeLabels } from '@/i18n/config'
import { DropdownDivider, DropdownHeading, DropdownItem, DropdownLabel } from '@/components/dropdown'

export function LocaleSwitcherMenuItems() {
  const t = useTranslations('applicationLayout')
  const locale = useLocale()
  const router = useRouter()

  if (availableLocales.length < 2) {
    return null
  }

  async function handleSelect(nextLocale: string) {
    if (nextLocale === locale) {
      return
    }
    await fetch('/api/locale', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ locale: nextLocale }),
    })
    router.refresh()
  }

  return (
    <>
      <DropdownDivider />
      <DropdownHeading>{t('language')}</DropdownHeading>
      {availableLocales.map((item) => (
        <DropdownItem key={item} onClick={() => handleSelect(item)}>
          <LanguageIcon />
          <DropdownLabel>
            {localeLabels[item]}
            {item === locale ? ` (${t('currentLanguage')})` : ''}
          </DropdownLabel>
        </DropdownItem>
      ))}
    </>
  )
}
