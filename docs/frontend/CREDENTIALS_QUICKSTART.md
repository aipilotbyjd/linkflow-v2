# Credentials System - Quick Start Guide

> Get started implementing the credentials feature in 30 minutes.

## TL;DR

| What | Where |
|------|-------|
| Full Documentation | [CREDENTIALS_SYSTEM.md](./CREDENTIALS_SYSTEM.md) |
| API Base URL | `http://localhost:8080/api/v1` |
| Service List | Frontend config (static) |
| OAuth Status | Backend API call |

## 5-Minute Overview

### Architecture

```
Frontend                          Backend
────────                          ───────
services.config.ts ──────────────→ GET /oauth/providers (which OAuth is configured)
   (static list of                    ↓
    all services)                 Returns: [{id: "google", configured: true}, ...]
        │                             ↓
        ├── OAuth services ───────→ GET /oauth/authorize/{provider}
        │       │                     ↓
        │       │                 Returns: {url: "https://google.com/...", state: "xxx"}
        │       │                     ↓
        │       └──────────────→ Redirect to OAuth provider
        │                             ↓
        │                         Provider redirects to backend callback
        │                             ↓
        │                         Backend creates credential, redirects to:
        │                         {frontend}/credentials?oauth=success&credential_id=xxx
        │
        └── API Key services ─────→ POST /credentials
                                     {name, type, data: {api_key: "xxx"}}
```

### Quick Code Examples

**1. Get services with availability:**

```tsx
const { data: providers } = useQuery({
  queryKey: ['oauth-providers'],
  queryFn: () => api.get('/oauth/providers'),
});

const configuredOAuth = new Set(providers?.data?.filter(p => p.configured).map(p => p.id));

const services = SERVICES.map(s => ({
  ...s,
  isAvailable: s.authType !== 'oauth' || configuredOAuth.has(s.oauthProvider)
}));
```

**2. Start OAuth flow:**

```tsx
const startOAuth = async (provider: string, name: string) => {
  const { url } = await api.get(
    `/workspaces/${workspaceId}/oauth/authorize/${provider}?credential_name=${name}`
  );
  window.location.href = url;
};
```

**3. Handle OAuth callback:**

```tsx
useEffect(() => {
  const params = new URLSearchParams(location.search);
  if (params.get('oauth') === 'success') {
    toast.success('Connected!');
    refetchCredentials();
    history.replaceState({}, '', '/credentials');
  }
}, []);
```

**4. Create API key credential:**

```tsx
await api.post(`/workspaces/${workspaceId}/credentials`, {
  name: "My OpenAI",
  type: "api_key",
  data: { api_key: "sk-xxx" }
});
```

## Files to Create

```
src/
├── types/
│   └── credentials.ts          # Copy from docs
├── config/
│   └── services.ts             # Copy from docs
├── hooks/
│   └── useCredentials.ts       # Copy from docs
├── pages/
│   └── CredentialsPage.tsx     # Main page
└── components/
    └── credentials/
        ├── AddCredentialModal.tsx
        ├── ServiceGrid.tsx
        ├── OAuthConnectForm.tsx
        ├── ApiKeyForm.tsx
        └── CredentialCard.tsx
```

## API Cheat Sheet

| Action | Method | Endpoint |
|--------|--------|----------|
| List providers | GET | `/oauth/providers` |
| Start OAuth | GET | `/workspaces/{ws}/oauth/authorize/{provider}` |
| List credentials | GET | `/workspaces/{ws}/credentials` |
| Create credential | POST | `/workspaces/{ws}/credentials` |
| Update credential | PUT | `/workspaces/{ws}/credentials/{id}` |
| Delete credential | DELETE | `/workspaces/{ws}/credentials/{id}` |
| Test credential | POST | `/workspaces/{ws}/credentials/{id}/test` |
| Refresh OAuth | POST | `/workspaces/{ws}/credentials/{id}/refresh` |

## Service Types Quick Reference

```typescript
// API Key (most common)
{ name: "OpenAI", type: "api_key", data: { api_key: "sk-xxx" } }

// Basic Auth (databases, Twilio)
{ name: "PostgreSQL", type: "basic", data: { username: "x", password: "y", host: "...", port: 5432 } }

// Bearer Token
{ name: "API", type: "bearer", data: { token: "xxx" } }

// Custom (AWS, special cases)
{ name: "AWS", type: "custom", data: { custom: { access_key_id: "x", secret: "y" } } }

// OAuth - created by backend after OAuth flow
// Frontend doesn't POST these, they're created via callback
```

## OAuth Callback URL Params

**Success:**
```
/credentials?oauth=success&credential_id=uuid-here
```

**Error:**
```
/credentials?oauth=error&error=access_denied&error_description=User%20denied
```

## Common Gotchas

1. **OAuth services show "Coming Soon"** - Backend needs `GOOGLE_CLIENT_ID`, `GITHUB_CLIENT_ID` etc. configured
2. **OAuth redirect fails** - Check `APP_FRONTEND_URL` is set correctly in backend
3. **Credential creation fails** - Check `type` is one of: `api_key`, `oauth2`, `basic`, `bearer`, `custom`
4. **Search not working** - Use `?search=query` param, searches name and description

## Test Data

For testing without real credentials:

```typescript
// Create fake API key
await api.post('/workspaces/xxx/credentials', {
  name: "Test OpenAI",
  type: "api_key", 
  data: { api_key: "sk-test-fake-key-12345" }
});

// Create fake database
await api.post('/workspaces/xxx/credentials', {
  name: "Test Database",
  type: "basic",
  data: {
    host: "localhost",
    port: 5432,
    database: "test",
    username: "test",
    password: "test"
  }
});
```

---

For full details, see [CREDENTIALS_SYSTEM.md](./CREDENTIALS_SYSTEM.md)
