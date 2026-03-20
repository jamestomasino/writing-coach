export function nextSetupPath(session) {
  if (!session?.authenticated) {
    return null
  }
  switch (session.setup_step) {
    case 'needs_ai_setup':
      return `/ai-settings?required=1&next=${encodeURIComponent(postAISetupPath(session))}`
    case 'needs_first_track':
      return '/onboarding'
    case 'needs_first_assignment':
      return '/new-assignment'
    default:
      return null
  }
}

export function requiredSetupPath(session, pathname) {
  const nextPath = nextSetupPath(session)
  if (!nextPath) {
    return null
  }
  if (session.setup_step === 'needs_ai_setup' && pathname.startsWith('/ai-settings')) {
    return null
  }
  if (session.setup_step === 'needs_first_track' && pathname.startsWith('/onboarding')) {
    return null
  }
  if (session.setup_step === 'needs_first_assignment' && pathname.startsWith('/new-assignment')) {
    return null
  }
  return nextPath
}

export function postAISetupPath(session) {
  if (!session?.authenticated) {
    return '/'
  }
  if (!session.onboarding_complete) {
    return '/onboarding'
  }
  if (session.setup_step === 'needs_first_assignment') {
    return '/new-assignment'
  }
  return '/'
}
