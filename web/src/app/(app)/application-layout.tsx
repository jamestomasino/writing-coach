'use client'

import { Avatar } from '@/components/avatar'
import {
  Dropdown,
  DropdownButton,
  DropdownDescription,
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
import { getSession, listTracks, setActiveTrack } from '@/lib/api'
import { formatLocalDateTime } from '@/lib/datetime'
import { shouldConfirmTrackSwitch } from '@/lib/track-switch-guard'
import type { UserTrack } from '@/lib/types'
import {
  ArrowLeftEndOnRectangleIcon,
  ArrowPathIcon,
  ArrowRightStartOnRectangleIcon,
  CheckIcon,
  ChevronUpIcon,
  Cog6ToothIcon,
  PlusIcon,
  UserPlusIcon,
} from '@heroicons/react/16/solid'
import {
  ChartBarSquareIcon,
  ClockIcon,
  FolderIcon,
  HomeIcon,
  InformationCircleIcon,
  PencilSquareIcon,
  Squares2X2Icon,
  UserGroupIcon,
} from '@heroicons/react/20/solid'
import Image from 'next/image'
import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'

function initialsForName(value: string) {
  const parts = value.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) {
    return 'WC'
  }
  return parts
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

function trackStatusLine(track: UserTrack | null) {
  if (!track) {
    return 'No active track selected'
  }
  if (track.assignment_count === 0) {
    return 'No assignments yet'
  }
  const latestAssignmentTime = formatLocalDateTime(track.latest_assignment_time)
  const activity = latestAssignmentTime ? `Latest ${latestAssignmentTime}` : ''
  if (track.current_assignment) {
    return activity ? `${track.current_assignment} · ${activity}` : track.current_assignment
  }
  return activity || `${track.assignment_count} assignment chains`
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

export function ApplicationLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const [authenticated, setAuthenticated] = useState<boolean | null>(null)
  const [isAdmin, setIsAdmin] = useState(false)
  const [accountName, setAccountName] = useState('Workshop')
  const [accountDetail, setAccountDetail] = useState('Writing practice')
  const [tracks, setTracks] = useState<UserTrack[]>([])
  const [switchingTrack, setSwitchingTrack] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const session = await getSession()
        if (cancelled) {
          return
        }

        setAuthenticated(session.authenticated)
        setIsAdmin(session.is_admin)
        if (session.authenticated) {
          const nextTracks = await listTracks().catch(() => [])
          if (!cancelled) {
            setTracks(nextTracks)
          }
        } else {
          setTracks([])
        }

        const profileName = session.identity?.name?.trim() || session.identity?.email?.trim()

        if (profileName) {
          setAccountName(profileName)
          setAccountDetail(
            session.is_admin ? 'Administrator account' : session.identity?.email?.trim() || 'Writing practice workspace'
          )
          return
        }

        const fallbackDetail = session.authenticated ? 'Writing practice workspace' : 'Writing practice'
        setAccountName(session.authenticated ? 'Writing Coach' : 'Workshop')
        setAccountDetail(session.is_admin ? 'Administrator account' : fallbackDetail)
      } catch {
        if (cancelled) {
          return
        }
        setAuthenticated(false)
        setIsAdmin(false)
        setTracks([])
        setAccountName('Workshop')
        setAccountDetail('Writing practice')
      }
    }

    void load()

    return () => {
      cancelled = true
    }
  }, [pathname])

  const accountInitials = initialsForName(accountName)
  const activeTrack = tracks.find((track) => track.is_active) ?? null

  async function handleTrackSelect(treeSlug: string) {
    if (!treeSlug || treeSlug === activeTrack?.tree_slug) {
      return
    }
    if (shouldConfirmTrackSwitch(window.__writingCoachHasUnsavedDraft)) {
      const confirmed = window.confirm(
        'You have an unsaved draft in the current track. Switch tracks and discard those unsaved edits?'
      )
      if (!confirmed) {
        return
      }
    }
    try {
      setSwitchingTrack(treeSlug)
      const nextTracks = await setActiveTrack(treeSlug)
      setTracks(nextTracks)
      const currentURL = `${pathname}${window.location.search}`
      window.location.assign(currentURL)
    } finally {
      setSwitchingTrack(null)
    }
  }

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
                <Dropdown>
                  <DropdownButton
                    data-testid="active-track-button"
                    className="flex w-full items-center justify-between rounded-2xl border border-stone-200 bg-stone-50 px-3 py-3 text-left text-sm text-zinc-950 dark:border-white/10 dark:bg-white/5 dark:text-white"
                  >
                    <span className="min-w-0">
                      <span className="block text-[11px] font-semibold tracking-[0.18em] text-zinc-500 uppercase dark:text-zinc-400">
                        Track
                      </span>
                      <span className="mt-1 block truncate text-sm font-semibold">
                        {activeTrack?.title ?? 'Select a track'}
                      </span>
                      <span className="mt-1 block truncate text-xs text-zinc-500 dark:text-zinc-400">
                        {trackStatusLine(activeTrack)}
                      </span>
                    </span>
                    <ChevronUpIcon className="size-4 rotate-180 text-zinc-500 dark:text-zinc-400" />
                  </DropdownButton>
                  <DropdownMenu className="min-w-80" anchor="bottom start">
                    {tracks.map((track) => (
                      <DropdownItem
                        key={track.enrollment_id}
                        data-testid={`track-option-${track.tree_slug}`}
                        data-track-slug={track.tree_slug}
                        onClick={() => handleTrackSelect(track.tree_slug)}
                        disabled={switchingTrack !== null}
                      >
                        <FolderIcon />
                        <DropdownLabel>{track.title}</DropdownLabel>
                        <DropdownDescription>{trackStatusLine(track)}</DropdownDescription>
                        {track.is_active ? <CheckIcon /> : null}
                      </DropdownItem>
                    ))}
                    <DropdownDivider />
                    <DropdownItem href="/onboarding?mode=create" data-testid="new-track-link">
                      <PlusIcon />
                      <DropdownLabel>New track</DropdownLabel>
                    </DropdownItem>
                  </DropdownMenu>
                </Dropdown>
              ) : null}
            </div>
          </SidebarHeader>

          <SidebarBody>
            {authenticated === true ? (
              <>
                <SidebarSection>
                  <SidebarItem href="/progress" current={pathname.startsWith('/progress')}>
                    <ChartBarSquareIcon />
                    <SidebarLabel>Track progress</SidebarLabel>
                  </SidebarItem>
                  <SidebarItem href="/tree" current={pathname.startsWith('/tree')}>
                    <Squares2X2Icon />
                    <SidebarLabel>Skill map</SidebarLabel>
                  </SidebarItem>
                  <SidebarItem href="/onboarding?mode=edit" current={pathname.startsWith('/onboarding')}>
                    <ArrowPathIcon />
                    <SidebarLabel>Edit track</SidebarLabel>
                  </SidebarItem>
                </SidebarSection>

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
                  <SidebarItem
                    href="/assignments"
                    current={pathname === '/assignments' || pathname.startsWith('/assignments/')}
                  >
                    <ClockIcon />
                    <SidebarLabel>Past assignments</SidebarLabel>
                  </SidebarItem>
                </SidebarSection>
              </>
            ) : null}

            <SidebarSpacer />

            <SidebarSection>
              <SidebarItem href="/about" current={pathname.startsWith('/about')}>
                <InformationCircleIcon />
                <SidebarLabel>About</SidebarLabel>
              </SidebarItem>
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
                    <Avatar
                      initials={accountInitials}
                      className="size-10 bg-stone-800 text-white"
                      square
                      alt={accountName}
                    />
                    <span className="min-w-0">
                      <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">
                        {accountName}
                      </span>
                      <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">
                        {accountDetail}
                      </span>
                    </span>
                  </span>
                  <ChevronUpIcon />
                </DropdownButton>
                <AccountDropdownMenu anchor="top start" authenticated={authenticated} isAdmin={isAdmin} />
              </Dropdown>
            ) : (
              <SidebarItem href="/login" current={pathname.startsWith('/login')}>
                <span className="flex min-w-0 items-center gap-3">
                  <Avatar
                    initials={accountInitials}
                    className="size-10 bg-stone-800 text-white"
                    square
                    alt={accountName}
                  />
                  <span className="min-w-0">
                    <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">
                      {accountName}
                    </span>
                    <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">
                      {accountDetail}
                    </span>
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
