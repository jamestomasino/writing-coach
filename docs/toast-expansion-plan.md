# Toast Expansion Candidates

This note tracks UI interactions that could move into the app-level toast system if we decide to broaden the shift beyond the first low-risk rollout.

## Already moved

- AI provider settings action confirmations and action-level failures
- Admin user provisioning confirmations and failures
- Logout failure notification

## Good later candidates

These are action-level messages where a toast would likely improve flow without removing important context.

- Current assignment draft submit success in `web/src/components/current-assignment-view.tsx`
  - The app currently uses inline messaging to say the draft is saved and review is running in the background.
  - A toast could confirm save immediately while the longer-lived background-review state stays inline.

- New assignment generation and accept failures in `web/src/components/new-assignment-view.tsx`
  - Short-lived retryable failures are good toast candidates.
  - Assignment preview and required next steps should remain inline.

- Revision creation failures in:
  - `web/src/components/current-assignment-view.tsx`
  - `web/src/components/review-view.tsx`
  - `web/src/components/compare-view.tsx`
  - These are action-triggered and often followed by navigation, so a toast can work well.

- Account reset confirmation/failure in `web/src/components/reset-data-card.tsx`
  - Especially useful if the card remains on the page and we want a visible completion confirmation.

- Onboarding save confirmation/failure in `web/src/components/onboarding-view.tsx`
  - The confirmation path may still prefer redirect-first behavior, but save errors are reasonable toast candidates.

- Admin filter reload warnings in `web/src/components/admin-view.tsx`
  - Partial refresh failures can be surfaced as toasts while preserving the last successful view.

## Probably keep inline

These messages carry workflow context or longer-lived instruction and should likely remain inline even if we use toasts elsewhere.

- AI provider setup-required guidance in `web/src/components/ai-provider-settings-view.tsx`
- Onboarding instruction callouts in `web/src/components/onboarding-view.tsx`
- Background review/job progress in `web/src/components/current-assignment-view.tsx`
- Page-level unavailable/error states in `web/src/components/status-state.tsx`
- Review content and assignment guidance that changes what the user should do next

## Possible future improvements

- Add toast preferences at the user level:
  - enable non-critical toasts
  - auto-dismiss duration
  - sticky errors
  - reduced-motion preference for transitions

- Add a queued dedupe policy:
  - collapse repeated identical toasts
  - prevent action spam from stacking duplicates

- Add action toasts with follow-up links:
  - `Open review`
  - `Open AI settings`
  - `Go to current assignment`

## Decision rule

Use a toast when the message is:

- short-lived
- action-triggered
- recoverable
- not required as persistent page context

Keep inline messaging when it changes user workflow or explains page state.
