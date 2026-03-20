export function hasUnsavedTrackDraft({ hasExercise, draft, baseline }) {
  return Boolean(hasExercise) && draft !== baseline
}

export function shouldConfirmTrackSwitch(hasUnsavedDraftFlag) {
  return hasUnsavedDraftFlag === true
}
