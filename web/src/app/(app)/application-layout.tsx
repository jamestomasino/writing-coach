'use client'

import Image from 'next/image'
import { Avatar } from '@/components/avatar'
import { Badge } from '@/components/badge'
import {
  Dropdown,
  DropdownButton,
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '@/components/dropdown'
import { Navbar, NavbarItem, NavbarSection, NavbarSpacer } from '@/components/navbar'
import {
  Sidebar,
  SidebarBody,
  SidebarFooter,
  SidebarHeader,
  SidebarHeading,
  SidebarItem,
  SidebarLabel,
  SidebarSection,
  SidebarSpacer,
} from '@/components/sidebar'
import { SidebarLayout } from '@/components/sidebar-layout'
import { getSession } from '@/lib/api'
import {
  ArrowPathIcon,
  ArrowRightStartOnRectangleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  Cog6ToothIcon,
  ArrowLeftEndOnRectangleIcon,
  UserPlusIcon,
} from '@heroicons/react/16/solid'
import {
  ChartBarSquareIcon,
  ClockIcon,
  HomeIcon,
  InformationCircleIcon,
  PencilSquareIcon,
  Squares2X2Icon,
  UserGroupIcon,
} from '@heroicons/react/20/solid'
import { useEffect, useState } from 'react'
import { usePathname } from 'next/navigation'

function initialsForName(value: string) {
  const parts = value
    .trim()
    .split(/\s+/)
    .filter(Boolean)
  if (parts.length === 0) {
    return 'WC'
  }
  return parts
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

function AccountDropdownMenu({
  anchor,
  isAdmin,
  authenticated,
}: {
  anchor: 'top start' | 'bottom end'
  isAdmin: boolean
  authenticated: boolean | null
}) {
  return (
    <DropdownMenu className="min-w-64" anchor={anchor}>
      {authenticated === true ? (
        <>
          <DropdownItem href="/settings">
            <Cog6ToothIcon />
            <DropdownLabel>Account settings</DropdownLabel>
          </DropdownItem>
          <DropdownItem href="/ai-settings">
            <Cog6ToothIcon />
            <DropdownLabel>AI provider</DropdownLabel>
          </DropdownItem>
          {isAdmin ? (
            <DropdownItem href="/admin">
              <UserGroupIcon />
              <DropdownLabel>Admin</DropdownLabel>
            </DropdownItem>
          ) : null}
          <DropdownDivider />
          <DropdownItem href="/logout">
            <ArrowRightStartOnRectangleIcon />
            <DropdownLabel>Sign out</DropdownLabel>
          </DropdownItem>
        </>
      ) : (
        <>
          <DropdownItem href="/login">
            <ArrowLeftEndOnRectangleIcon />
            <DropdownLabel>Sign in</DropdownLabel>
          </DropdownItem>
          <DropdownItem href="/register">
            <UserPlusIcon />
            <DropdownLabel>Register</DropdownLabel>
          </DropdownItem>
        </>
      )}
    </DropdownMenu>
  )
}

function providerStatusMeta({
  ready,
  hasPersonalKey,
  systemFallback,
}: {
  ready: boolean
  hasPersonalKey: boolean
  systemFallback: boolean
}) {
  if (!ready) {
    return {
      label: 'AI setup required',
      detail: 'Add a provider key to generate assignments and reviews.',
      badgeColor: 'rose' as const,
    }
  }
  if (hasPersonalKey) {
    return {
      label: 'Personal AI provider',
      detail: 'Your own provider key is active for generation.',
      badgeColor: 'emerald' as const,
    }
  }
  if (systemFallback) {
    return {
      label: 'Shared AI provider',
      detail: 'You are using the shared system provider.',
      badgeColor: 'amber' as const,
    }
  }
  return {
    label: 'AI status unavailable',
    detail: 'Provider status could not be determined.',
    badgeColor: 'zinc' as const,
  }
}

export function ApplicationLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const [authenticated, setAuthenticated] = useState<boolean | null>(null)
  const [isAdmin, setIsAdmin] = useState(false)
  const [accountName, setAccountName] = useState('Workshop')
  const [accountDetail, setAccountDetail] = useState('Scholastic coaching loop')
  const [aiProviderReady, setAIProviderReady] = useState(true)
  const [aiHasPersonalKey, setAIHasPersonalKey] = useState(false)
  const [aiSystemFallback, setAISystemFallback] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const session = await getSession()
        if (!cancelled) {
          setAuthenticated(session.authenticated)
          setIsAdmin(session.is_admin)
          setAIProviderReady(session.ai_provider_ready)
          setAIHasPersonalKey(session.ai_has_personal_key)
          setAISystemFallback(session.ai_system_fallback)
          if (session.authenticated) {
            setAccountName(session.identity?.name?.trim() || session.identity?.email?.trim() || 'Account')
            setAccountDetail(session.identity?.email?.trim() || 'Writing Coach account')
          } else {
            setAccountName('Guest')
            setAccountDetail('Sign in for settings and account actions')
          }
        }
      } catch {
        if (!cancelled) {
          setAuthenticated(false)
          setIsAdmin(false)
          setAIProviderReady(true)
          setAIHasPersonalKey(false)
          setAISystemFallback(false)
          setAccountName('Guest')
          setAccountDetail('Sign in for settings and account actions')
        }
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const accountInitials = initialsForName(accountName)
  const aiStatus = providerStatusMeta({
    ready: aiProviderReady,
    hasPersonalKey: aiHasPersonalKey,
    systemFallback: aiSystemFallback,
  })

  return (
    <SidebarLayout
      navbar={
        <Navbar>
          <NavbarSpacer />
          <NavbarSection>
            <Dropdown>
              <DropdownButton as={NavbarItem}>
                <Avatar initials={accountInitials} square className="bg-stone-800 text-white" />
              </DropdownButton>
              <AccountDropdownMenu anchor="bottom end" authenticated={authenticated} isAdmin={isAdmin} />
            </Dropdown>
          </NavbarSection>
        </Navbar>
      }
      sidebar={
        <Sidebar>
          <SidebarHeader>
            <div className="space-y-3 px-2 py-1">
              <div className="flex items-center gap-3">
                <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-zinc-950/70 ring-1 ring-white/10 dark:bg-white/5">
                  <Image src="/logo-writing-coach.svg" alt="Writing Coach logo" width={24} height={24} priority />
                </div>
                <div className="min-w-0">
                  <div className="text-lg/6 font-semibold text-zinc-950 dark:text-white">Writing Coach</div>
                  <div className="mt-1 text-xs/5 text-zinc-500 dark:text-zinc-400">Structured practice</div>
                </div>
              </div>
              {authenticated === true ? (
                <a
                  href="/ai-settings"
                  className="block rounded-2xl border border-stone-200 bg-stone-50 px-3 py-3 transition hover:border-stone-300 hover:bg-white dark:border-white/10 dark:bg-white/5 dark:hover:border-white/20 dark:hover:bg-white/8"
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="text-sm font-medium text-zinc-950 dark:text-white">{aiStatus.label}</div>
                    <Badge color={aiStatus.badgeColor}>{aiHasPersonalKey ? 'Personal' : aiProviderReady ? 'Shared' : 'Needed'}</Badge>
                  </div>
                  <div className="mt-1 text-xs/5 text-zinc-500 dark:text-zinc-400">{aiStatus.detail}</div>
                </a>
              ) : null}
            </div>
          </SidebarHeader>

          <SidebarBody>
            <SidebarSection>
              <SidebarItem href="/about" current={pathname.startsWith('/about')}>
                <InformationCircleIcon />
                <SidebarLabel>About</SidebarLabel>
              </SidebarItem>
            </SidebarSection>

            {authenticated === true ? (
              <>
                <SidebarSection>
                  <SidebarHeading>Assignments</SidebarHeading>
                  <SidebarItem href="/" current={pathname === '/'}>
                    <HomeIcon />
                    <SidebarLabel>Current assignment</SidebarLabel>
                  </SidebarItem>
                  <SidebarItem href="/new-assignment" current={pathname.startsWith('/new-assignment')}>
                    <PencilSquareIcon />
                    <SidebarLabel>New assignment</SidebarLabel>
                  </SidebarItem>
                  <SidebarItem href="/assignments" current={pathname === '/assignments' || pathname.startsWith('/assignments/')}>
                    <ClockIcon />
                    <SidebarLabel>Past assignments</SidebarLabel>
                  </SidebarItem>
                </SidebarSection>

                <SidebarSection>
                  <SidebarHeading>Track</SidebarHeading>
                  <SidebarItem href="/progress" current={pathname.startsWith('/progress')}>
                    <ChartBarSquareIcon />
                    <SidebarLabel>Track progress</SidebarLabel>
                  </SidebarItem>
                  <SidebarItem href="/onboarding" current={pathname.startsWith('/onboarding')}>
                    <ArrowPathIcon />
                    <SidebarLabel>Change track</SidebarLabel>
                  </SidebarItem>
                <SidebarItem href="/tree" current={pathname.startsWith('/tree')}>
                  <Squares2X2Icon />
                  <SidebarLabel>Skill map</SidebarLabel>
                </SidebarItem>
              </SidebarSection>

              <SidebarSection>
                <SidebarHeading>Settings</SidebarHeading>
                <SidebarItem href="/ai-settings" current={pathname.startsWith('/ai-settings')}>
                  <Cog6ToothIcon />
                  <SidebarLabel>AI provider</SidebarLabel>
                </SidebarItem>
              </SidebarSection>
            </>
          ) : null}

            <SidebarSpacer />

            <SidebarSection>
              {authenticated === false ? (
                <>
                  <SidebarItem href="/login">
                    <ArrowLeftEndOnRectangleIcon />
                    <SidebarLabel>Sign in</SidebarLabel>
                  </SidebarItem>
                  <SidebarItem href="/register">
                    <UserPlusIcon />
                    <SidebarLabel>Register</SidebarLabel>
                  </SidebarItem>
                </>
              ) : null}
            </SidebarSection>
          </SidebarBody>

          <SidebarFooter className="max-lg:hidden">
            {authenticated === true ? (
              <Dropdown>
                <DropdownButton as={SidebarItem}>
                  <span className="flex min-w-0 items-center gap-3">
                    <Avatar initials={accountInitials} className="size-10 bg-stone-800 text-white" square alt={accountName} />
                    <span className="min-w-0">
                      <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">{accountName}</span>
                      <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">{accountDetail}</span>
                    </span>
                  </span>
                  <ChevronUpIcon />
                </DropdownButton>
                <AccountDropdownMenu anchor="top start" authenticated={authenticated} isAdmin={isAdmin} />
              </Dropdown>
            ) : (
              <SidebarItem href="/login" current={pathname.startsWith('/login')}>
                <span className="flex min-w-0 items-center gap-3">
                  <Avatar initials={accountInitials} className="size-10 bg-stone-800 text-white" square alt={accountName} />
                  <span className="min-w-0">
                    <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">{accountName}</span>
                    <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">{accountDetail}</span>
                  </span>
                </span>
              </SidebarItem>
            )}
          </SidebarFooter>
        </Sidebar>
      }
    >
      {children}
    </SidebarLayout>
  )
}
