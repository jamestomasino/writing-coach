# AI Provider Rollout Plan

## Objective

Allow users to bring their own LLM provider credentials and use them for assignment generation, revision generation, and review generation without breaking existing accounts or exposing secrets.

Phase 1 will add app-owned provider settings with a system-provider fallback. Existing users continue working without interruption. New provider-backed flows will be introduced behind a provider resolution layer so future providers can be added without rewriting prompt and review services again.

## Rollout Policy

Phase 1 policy:
- Existing users continue to work with the system provider by default.
- Users may opt into a personal provider from a new app-owned settings page.
- The app treats AI as ready when either:
  - the user has a valid personal provider configured, or
  - system fallback is enabled.

Phase 2 policy candidates:
- Require provider setup for new users only.
- Eventually retire the system provider for all users after advance notice.

Phase 1 does not force existing users to enter credentials before continuing.

## Phase 1 Providers

Supported in the initial rollout:
- OpenAI
- Groq
- xAI

Deferred to later phases:
- Anthropic
- Gemini
- Mistral
- Cohere

Reason: OpenAI, Groq, and xAI are the closest to the current backend request flow and will let us establish the provider abstraction with less provider-specific divergence.

## User Experience

### Existing users

- Continue into the app normally.
- See a non-blocking banner or settings prompt explaining that they can connect a personal provider.
- Can keep using the system provider until a later migration phase changes policy.

### New users

Phase 1:
- Authenticate
- Complete onboarding
- Land in the workspace
- See AI provider setup as optional but recommended

Possible later policy:
- Authenticate
- Complete onboarding
- Complete AI provider setup before entering the workspace

### Settings surface

Do not overload the Kratos account settings flow.

Add a new app-owned page:
- `Settings > AI provider`

Controls:
- Provider dropdown
- API key field
- Advanced section:
  - base URL override
  - prompt model override
  - review model override
- Save button
- Validate connection button
- Remove provider button
- Status card showing effective provider mode

### Failure handling

User-facing errors must be clear and local to the provider setup flow:
- invalid credentials
- quota exhausted
- provider unavailable
- unsupported model

Secrets must never appear in logs or API responses.

## Architecture

### Data model

Add a `user_ai_provider_settings` table keyed by `user_id`.

Fields:
- `user_id`
- `provider`
- `api_key_encrypted`
- `api_key_last4`
- `base_url_override`
- `prompt_model_override`
- `review_model_override`
- `enabled`
- `validated_at`
- `last_validation_error`
- `created_at`
- `updated_at`

### Security

Store API keys encrypted at rest using a server-held master secret.

Rules:
- never store plaintext keys in the database
- never return plaintext keys to clients
- only return masked state such as the last 4 characters
- never emit secrets in logs, panic messages, or API errors

### Provider abstraction

Replace the singleton configured OpenAI client pattern with:
- provider settings resolver
- provider factory
- provider interface

Provider interface:
- `GenerateExercise`
- `GenerateRevisionExercise`
- `ReviewSubmission`
- `ValidateCredentials`

### Credential resolution

For each generation action:
1. Resolve the acting user.
2. Load AI provider settings for that user.
3. If a personal provider is enabled and valid, use it.
4. Otherwise, if system fallback is enabled, use the system provider.
5. Otherwise return an `AI provider setup required` error.

### Provenance

Persist effective provider provenance on generated artifacts via existing fields:
- `GenerationKind`
- `ReviewKind`
- `ProviderNote`

Examples:
- `system/openai`
- `user/openai`
- `user/groq`
- `user/xai`

## API Plan

New app-owned endpoints:
- `GET /api/ai/settings`
- `PUT /api/ai/settings`
- `DELETE /api/ai/settings`
- `POST /api/ai/settings/validate`

Responses must include only:
- provider selection
- enabled state
- masked key status
- validation state
- effective provider mode
- whether system fallback remains available

Responses must never include the raw API key.

## Implementation Sequence

1. Add migration and domain model for user AI provider settings.
2. Add DB methods and DB tests.
3. Add encryption helper and tests.
4. Add provider settings API and API tests.
5. Introduce provider abstraction and resolver.
6. Refactor prompt and review services to use provider resolution.
7. Add AI settings page and client-side flows.
8. Add readiness state and non-blocking phase-1 prompts.
9. Add observability and provider provenance.
10. Run end-to-end verification.

## Test Plan

### Unit tests

- encryption and decryption round-trip
- masking logic for stored keys
- provider resolution precedence
- fallback behavior
- provider validation response parsing
- unsupported provider rejection

### DB tests

- save provider settings
- update provider settings
- delete provider settings
- load not-found behavior
- validated timestamp persistence

### API tests

- get settings
- save settings
- delete settings
- validate settings success
- validate settings failure
- user scoping
- raw key never returned

### Service tests

- prompt generation uses personal provider when configured
- prompt generation falls back to system provider when allowed
- prompt generation fails cleanly when no provider is available
- review generation follows the same rules
- revision generation follows the same rules

### Integration tests

- existing user can continue on system provider
- user can switch to personal provider
- invalid personal provider fails without affecting stored historical data
- review jobs execute using the submission owner’s provider settings

### UI tests

- settings form save and validate flow
- settings remove flow
- status card copy and state transitions
- optional banner for existing users
- required setup gate for later migration phases

## Risks

- storing secrets unsafely
- using the wrong provider in background job execution
- coupling provider-specific payload logic too tightly to the current OpenAI client
- forcing a blocking migration too early

## Definition of Done for Phase 1

- Existing users continue to work without interruption.
- Users can save and validate a personal provider.
- Prompt, revision, and review flows resolve provider credentials per user.
- Provider provenance is visible on generated artifacts.
- Secrets are encrypted at rest and masked in API responses.
- The feature is covered by unit, DB, API, and service-level tests.
