# HTTP access enforcement

Auth0 JWT authentication is implemented with `go-jwt-middleware/v3`. The service accepts only RS256 access tokens whose issuer matches `AUTH0_DOMAIN` and whose audience matches `AUTH0_AUDIENCE`. `/health` and `/public` are exact public-path exclusions; all other paths require a bearer token.

`Auth0.RequirePermissions` enforces every named permission in the token's `permissions` claim and returns `403` when any are absent. Configure RBAC and **Add Permissions in the Access Token** for this API in Auth0, then wrap future routes inside the global authentication middleware:

```go
router.Handle(
	"GET /profile",
	authentication.RequirePermissions("read:profile")(profileHandler),
)
```

Use `middleware.Subject(r.Context())` to obtain the trusted Auth0 `sub` claim. A future identity-linking workflow must map that external subject to the service's account ID; handlers must never accept a caller-supplied account ID as authenticated identity.

- **F1.3.1–F1.3.2:** Auth0 access-token validation attaches authenticated claims to requests and rejects invalid or expired tokens. Login/logout UI and application session policy remain future work.
- **F1.2, F1.4:** Require authentication for self-deletion and profile access. Services enforce ownership and permitted operations.
- **F1.6, F1.6.3–F1.6.6:** Restrict administrator-management routes to super administrators and apply least privilege and need-to-know access. Services also enforce business and data-level restrictions.
- **F1.9.2:** Reject additional super-administrator creation by users and administrators.
- **F1.10, F1.10.5, F1.10.7:** Restrict super-administrator management to authenticated super administrators and deny protected access to deactivated accounts, including when an old session is presented.
- **F1.9.5, F1.10.6:** Ensure denied creation/deactivation attempts can be recorded by the future audit workflow rather than disappearing at the middleware boundary.

Self-deactivation, last-active-super-administrator protection, and re-authentication confirmation (F1.10.1–F1.10.3) remain service rules. Other services enforce their own errand, supplier, and dispute permissions using future identity/API/event contracts.
