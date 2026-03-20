import assert from 'node:assert/strict'
import test from 'node:test'

import { hasUnsavedTrackDraft, shouldConfirmTrackSwitch } from './track-switch-guard.js'

test('hasUnsavedTrackDraft is false when no exercise is active', () => {
  assert.equal(hasUnsavedTrackDraft({ hasExercise: false, draft: 'Draft in progress', baseline: '' }), false)
})

test('hasUnsavedTrackDraft is false when draft matches baseline', () => {
  assert.equal(hasUnsavedTrackDraft({ hasExercise: true, draft: 'Saved draft', baseline: 'Saved draft' }), false)
})

test('hasUnsavedTrackDraft is true when draft differs from baseline', () => {
  assert.equal(hasUnsavedTrackDraft({ hasExercise: true, draft: 'Changed draft', baseline: 'Saved draft' }), true)
})

test('shouldConfirmTrackSwitch only confirms on explicit unsaved flag', () => {
  assert.equal(shouldConfirmTrackSwitch(true), true)
  assert.equal(shouldConfirmTrackSwitch(false), false)
  assert.equal(shouldConfirmTrackSwitch(undefined), false)
})
