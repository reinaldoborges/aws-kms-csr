Type: grilling
Status: resolved

## Question

Should the TUI inform the user where resolved AWS credentials came from, and if so, how? Specifically: does it name the specific source (environment variables, shared config/credentials file profile, SSO, container/instance role, etc.) versus a generic "checked standard locations" message; where does this surface given the single-form layout and persistent status/error area already decided (Screen flow and Error handling tickets); and does it change the existing Credential Fallback prompting behavior at all?

## Answer

Display the specific resolved credential source by name (e.g. "Credentials: environment variables", "Credentials: shared config profile `default`", "Credentials: SSO", "Credentials: container/instance role"), not a generic "checked standard locations" message.

Surface it as a single informational line in the persistent status area already designed for errors (below the fields), using non-error styling, distinct from error messages shown in the same area. Show this line only when the default credential chain actually supplied something; show no line when the user ends up typing all credential fields by hand (there's no "source" to report beyond "typed").

This is purely a display/transparency addition. It does not change the existing Credential Fallback decision: the TUI still prompts only for whatever the default chain left missing, and prompting logic is unaffected.
