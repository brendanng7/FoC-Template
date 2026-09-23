# Configuration

Current implementation: `Load` reads required `DATABASE_URL`, `AUTH0_DOMAIN`, and `AUTH0_AUDIENCE`, plus optional `HTTP_ADDRESS` (default `:8080`). `AUTH0_DOMAIN` is the Auth0 tenant hostname without a scheme; `AUTH0_AUDIENCE` is the custom API Identifier. The entry point loads a development `.env` file when present and wires the HTTP and Auth0 settings. Repository connection setup validates and checks connectivity without exposing credentials. Bootstrap and email configuration below remain future work.

- **F1.8.2:** Read initial super-administrator credentials from the deployment environment. Exact variable names remain to be defined.
- **F1.8.5:** Support validation of required bootstrap values when initialization is needed. The service determines this from persisted initialization/account state; configuration alone cannot decide.
- **F1.8.6:** Ensure configuration errors and diagnostics never expose credential values.
- **F1.1.4–F1.1.5, F1.3.2, F1.7, F1.9.3:** Provide future email-delivery, link, session-expiration, and activation-expiration settings as needed. NUS domain eligibility remains a service rule.

Session durations, token lifetimes, provider details, and other deployment settings are not specified by these requirements and remain to be decided. Do not put credentials in source files, documentation, or image defaults.
