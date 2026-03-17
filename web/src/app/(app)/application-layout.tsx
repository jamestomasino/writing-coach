'use client'

import { Avatar } from '@/components/avatar'
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
import {
  ArrowRightStartOnRectangleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  LockClosedIcon,
  Cog6ToothIcon,
  UserPlusIcon,
} from '@heroicons/react/16/solid'
import {
  ChartBarSquareIcon,
  HomeIcon,
  PencilSquareIcon,
  Squares2X2Icon,
  UserGroupIcon,
} from '@heroicons/react/20/solid'
import { usePathname } from 'next/navigation'

function AccountDropdownMenu({ anchor }: { anchor: 'top start' | 'bottom end' }) {
  return (
    <DropdownMenu className="min-w-64" anchor={anchor}>
      <DropdownItem href="/tree">
        <Squares2X2Icon />
        <DropdownLabel>Skill map</DropdownLabel>
      </DropdownItem>
      <DropdownItem href="/progress">
        <ChartBarSquareIcon />
        <DropdownLabel>Progress</DropdownLabel>
      </DropdownItem>
      <DropdownItem href="/new-assignment">
        <PencilSquareIcon />
        <DropdownLabel>New assignment</DropdownLabel>
      </DropdownItem>
      <DropdownItem href="/settings">
        <Cog6ToothIcon />
        <DropdownLabel>Settings</DropdownLabel>
      </DropdownItem>
      <DropdownDivider />
      <DropdownItem href="/.ory/kratos/public/self-service/logout/browser">
        <ArrowRightStartOnRectangleIcon />
        <DropdownLabel>Sign out</DropdownLabel>
      </DropdownItem>
    </DropdownMenu>
  )
}

export function ApplicationLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  return (
    <SidebarLayout
      navbar={
        <Navbar>
          <NavbarSpacer />
          <NavbarSection>
            <Dropdown>
              <DropdownButton as={NavbarItem}>
                <Avatar initials="WC" square className="bg-stone-800 text-white" />
              </DropdownButton>
              <AccountDropdownMenu anchor="bottom end" />
            </Dropdown>
          </NavbarSection>
        </Navbar>
      }
      sidebar={
        <Sidebar>
          <SidebarHeader>
            <Dropdown>
              <DropdownButton as={SidebarItem}>
                <Avatar initials="WC" className="bg-stone-800 text-white" />
                <SidebarLabel>Writing Coach</SidebarLabel>
                <ChevronDownIcon />
              </DropdownButton>
              <DropdownMenu className="min-w-80 lg:min-w-64" anchor="bottom start">
                <DropdownItem href="/progress">
                  <ChartBarSquareIcon />
                  <DropdownLabel>Progress board</DropdownLabel>
                </DropdownItem>
                <DropdownDivider />
                <DropdownItem href="/admin">
                  <LockClosedIcon />
                  <DropdownLabel>Admin</DropdownLabel>
                </DropdownItem>
              </DropdownMenu>
            </Dropdown>
          </SidebarHeader>

          <SidebarBody>
            <SidebarSection>
              <SidebarItem href="/" current={pathname === '/'}>
                <HomeIcon />
                <SidebarLabel>Current assignment</SidebarLabel>
              </SidebarItem>
              <SidebarItem href="/new-assignment" current={pathname.startsWith('/new-assignment')}>
                <PencilSquareIcon />
                <SidebarLabel>New assignment</SidebarLabel>
              </SidebarItem>
              <SidebarItem href="/progress" current={pathname.startsWith('/progress')}>
                <ChartBarSquareIcon />
                <SidebarLabel>Progress</SidebarLabel>
              </SidebarItem>
              <SidebarItem href="/tree" current={pathname.startsWith('/tree')}>
                <Squares2X2Icon />
                <SidebarLabel>Skill map</SidebarLabel>
              </SidebarItem>
              <SidebarItem href="/admin" current={pathname.startsWith('/admin')}>
                <UserGroupIcon />
                <SidebarLabel>Admin</SidebarLabel>
              </SidebarItem>
            </SidebarSection>

            <SidebarSection className="max-lg:hidden">
              <SidebarHeading>Track</SidebarHeading>
              <SidebarItem href="/new-assignment">
                <PencilSquareIcon />
                <SidebarLabel>Assignment builder</SidebarLabel>
              </SidebarItem>
              <SidebarItem href="/progress">
                <ChartBarSquareIcon />
                <SidebarLabel>Progress board</SidebarLabel>
              </SidebarItem>
              <SidebarItem href="/tree">
                <Squares2X2Icon />
                <SidebarLabel>Skill map</SidebarLabel>
              </SidebarItem>
            </SidebarSection>

            <SidebarSpacer />

            <SidebarSection>
              <SidebarItem href="/register">
                <UserPlusIcon />
                <SidebarLabel>Register</SidebarLabel>
              </SidebarItem>
            </SidebarSection>
          </SidebarBody>

          <SidebarFooter className="max-lg:hidden">
            <Dropdown>
              <DropdownButton as={SidebarItem}>
                <span className="flex min-w-0 items-center gap-3">
                  <Avatar initials="WC" className="size-10 bg-stone-800 text-white" square alt="" />
                  <span className="min-w-0">
                    <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">Workshop</span>
                    <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">
                      Scholastic coaching loop
                    </span>
                  </span>
                </span>
                <ChevronUpIcon />
              </DropdownButton>
              <AccountDropdownMenu anchor="top start" />
            </Dropdown>
          </SidebarFooter>
        </Sidebar>
      }
    >
      {children}
    </SidebarLayout>
  )
}
