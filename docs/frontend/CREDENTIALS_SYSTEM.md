# Credentials System - Frontend Implementation Guide

> Complete documentation for implementing the Credentials Management system in LinkFlow's frontend.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [API Reference](#api-reference)
4. [TypeScript Types](#typescript-types)
5. [Service Registry Configuration](#service-registry-configuration)
6. [React Hooks](#react-hooks)
7. [Component Specifications](#component-specifications)
8. [OAuth Flow Implementation](#oauth-flow-implementation)
9. [Error Handling](#error-handling)
10. [UI/UX Guidelines](#uiux-guidelines)
11. [Testing Checklist](#testing-checklist)

---

## Overview

The Credentials System allows users to securely store and manage authentication credentials for third-party services (Google, Slack, OpenAI, databases, etc.). It supports two authentication methods:

| Method | Description | Example Services |
|--------|-------------|------------------|
| **OAuth 2.0** | Redirects user to provider for authorization | Google, Slack, GitHub, Microsoft |
| **API Key / Basic** | User enters credentials manually | OpenAI, Stripe, Twilio, Databases |

### Key Principles

1. **Frontend owns the service catalog** - All service definitions (names, icons, fields) live in frontend
2. **Backend provides OAuth infrastructure** - Only OAuth flow and credential storage in backend
3. **Credentials are encrypted** - All sensitive data is AES-256-GCM encrypted at rest
4. **Workspace-scoped** - All credentials belong to a workspace

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 FRONTEND                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │ services.config  │    │   React Hooks    │    │   Components     │       │
│  │                  │    │                  │    │                  │       │
│  │ - Service list   │───▶│ - useCredentials │───▶│ - CredentialsList│       │
│  │ - Icons          │    │ - useOAuthFlow   │    │ - AddCredModal   │       │
│  │ - Form fields    │    │ - useProviders   │    │ - ServiceGrid    │       │
│  │ - Categories     │    │                  │    │ - ApiKeyForm     │       │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘       │
│                                    │                      │                  │
└────────────────────────────────────┼──────────────────────┼──────────────────┘
                                     │                      │
                                     ▼                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                  BACKEND                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                           REST API                                      │ │
│  │                                                                         │ │
│  │  GET  /oauth/providers              → List configured OAuth providers   │ │
│  │  GET  /workspaces/{ws}/oauth/authorize/{provider}  → Start OAuth       │ │
│  │  GET  /oauth/callback/{provider}    → Handle OAuth callback (redirect) │ │
│  │                                                                         │ │
│  │  GET  /workspaces/{ws}/credentials  → List credentials                 │ │
│  │  POST /workspaces/{ws}/credentials  → Create credential (API key)      │ │
│  │  GET  /workspaces/{ws}/credentials/{id}  → Get credential              │ │
│  │  PUT  /workspaces/{ws}/credentials/{id}  → Update credential           │ │
│  │  DELETE /workspaces/{ws}/credentials/{id} → Delete credential          │ │
│  │  POST /workspaces/{ws}/credentials/{id}/test → Test credential         │ │
│  │  POST /workspaces/{ws}/credentials/{id}/refresh → Refresh OAuth token  │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## API Reference

### Base URL
```
Production: https://api.linkflow.ai/api/v1
Development: http://localhost:8080/api/v1
```

### Authentication
All API requests require a Bearer token in the Authorization header:
```
Authorization: Bearer <access_token>
```

---

### 1. List OAuth Providers

Returns which OAuth providers are configured (have client_id/secret set up).

```http
GET /oauth/providers
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "google",
      "name": "Google",
      "configured": true,
      "scopes": [
        "https://www.googleapis.com/auth/spreadsheets",
        "https://www.googleapis.com/auth/drive",
        "https://www.googleapis.com/auth/calendar",
        "https://www.googleapis.com/auth/gmail.modify"
      ]
    },
    {
      "id": "slack",
      "name": "Slack",
      "configured": false,
      "scopes": ["chat:write", "channels:read", "users:read"]
    },
    {
      "id": "github",
      "name": "GitHub",
      "configured": true,
      "scopes": ["repo", "read:user", "read:org"]
    },
    {
      "id": "microsoft",
      "name": "Microsoft",
      "configured": false,
      "scopes": ["openid", "email", "profile", "User.Read"]
    },
    {
      "id": "notion",
      "name": "Notion",
      "configured": false,
      "scopes": []
    },
    {
      "id": "hubspot",
      "name": "HubSpot",
      "configured": false,
      "scopes": ["crm.objects.contacts.read", "crm.objects.contacts.write"]
    },
    {
      "id": "salesforce",
      "name": "Salesforce",
      "configured": false,
      "scopes": ["api", "refresh_token"]
    },
    {
      "id": "airtable",
      "name": "Airtable",
      "configured": false,
      "scopes": ["data.records:read", "data.records:write", "schema.bases:read"]
    }
  ],
  "links": {
    "self": "/api/v1/oauth/providers"
  }
}
```

**Usage:** Use this to determine which OAuth services to show as "available" vs "coming soon".

---

### 2. Start OAuth Flow

Initiates an OAuth authorization flow. Returns a URL to redirect the user to.

```http
GET /workspaces/{workspaceId}/oauth/authorize/{provider}
Authorization: Bearer <token>

Query Parameters:
  credential_name (optional): Name for the credential to be created
  redirect_url (optional): URL to redirect after completion
  scopes (optional): Comma-separated scopes to request
```

**Example Request:**
```
GET /workspaces/abc123/oauth/authorize/google?credential_name=My%20Google%20Sheets
```

**Response 200:**
```json
{
  "url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=xxx&redirect_uri=xxx&response_type=code&scope=xxx&state=xxx&access_type=offline&prompt=consent",
  "state": "abc123def456..."
}
```

**Frontend Action:** Redirect user to `url`:
```javascript
window.location.href = response.url;
```

---

### 3. OAuth Callback (Backend Handles)

This endpoint is called by the OAuth provider, NOT by your frontend directly. After successful OAuth:

**Success Redirect:**
```
{frontend_url}/credentials?oauth=success&credential_id=uuid-here
```

**Error Redirect:**
```
{frontend_url}/credentials?oauth=error&error=access_denied&error_description=User%20denied%20access
```

**Frontend Action:** Check URL params on `/credentials` page load:
```javascript
const params = new URLSearchParams(window.location.search);
if (params.get('oauth') === 'success') {
  const credentialId = params.get('credential_id');
  showSuccessToast(`Credential created: ${credentialId}`);
  refetchCredentials();
}
if (params.get('oauth') === 'error') {
  const error = params.get('error');
  const description = params.get('error_description');
  showErrorToast(`OAuth failed: ${description || error}`);
}
// Clean up URL
window.history.replaceState({}, '', '/credentials');
```

---

### 4. List Credentials

Returns all credentials for a workspace.

```http
GET /workspaces/{workspaceId}/credentials
Authorization: Bearer <token>

Query Parameters:
  type (optional): Filter by type (api_key, oauth2, basic, bearer, custom)
  search (optional): Search by name or description
  sort_by (optional): Sort field (name, created_at, last_used_at, type)
  order (optional): Sort order (asc, desc)
  page (optional): Page number (default: 1)
  per_page (optional): Items per page (default: 20, max: 100)
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "workspace_id": "660e8400-e29b-41d4-a716-446655440001",
      "created_by": "770e8400-e29b-41d4-a716-446655440002",
      "name": "My Google Sheets",
      "type": "oauth2",
      "description": null,
      "provider": "google",
      "provider_account_id": "user@gmail.com",
      "token_expires_at": 1704153600,
      "last_used_at": 1704067200,
      "created_at": 1703980800,
      "updated_at": 1704067200,
      "actions": [
        {
          "name": "edit",
          "method": "PUT",
          "href": "/api/v1/workspaces/xxx/credentials/550e8400-e29b-41d4-a716-446655440000",
          "label": "Edit Credential"
        },
        {
          "name": "test",
          "method": "POST",
          "href": "/api/v1/workspaces/xxx/credentials/550e8400-e29b-41d4-a716-446655440000/test",
          "label": "Test Connection"
        },
        {
          "name": "delete",
          "method": "DELETE",
          "href": "/api/v1/workspaces/xxx/credentials/550e8400-e29b-41d4-a716-446655440000",
          "label": "Delete"
        }
      ]
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "workspace_id": "660e8400-e29b-41d4-a716-446655440001",
      "created_by": "770e8400-e29b-41d4-a716-446655440002",
      "name": "OpenAI API",
      "type": "api_key",
      "description": "Production API key",
      "provider": null,
      "provider_account_id": null,
      "token_expires_at": null,
      "last_used_at": 1704060000,
      "created_at": 1703900000,
      "updated_at": 1703900000,
      "actions": [...]
    }
  ],
  "meta": {
    "total": 15,
    "page": 1,
    "per_page": 20,
    "total_pages": 1
  },
  "links": {
    "self": "/api/v1/workspaces/xxx/credentials?page=1&per_page=20",
    "first": "/api/v1/workspaces/xxx/credentials?page=1&per_page=20",
    "last": "/api/v1/workspaces/xxx/credentials?page=1&per_page=20"
  }
}
```

---

### 5. Create Credential (API Key / Basic / Custom)

Creates a new non-OAuth credential.

```http
POST /workspaces/{workspaceId}/credentials
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "My OpenAI Key",
  "type": "api_key",
  "description": "Production API key for GPT-4",
  "data": {
    "api_key": "sk-proj-xxxxxxxxxxxx"
  }
}
```

**Credential Types and Required Data:**

| Type | Required Fields in `data` |
|------|---------------------------|
| `api_key` | `api_key` |
| `basic` | `username`, `password` |
| `bearer` | `token` |
| `custom` | Any fields in `custom` object |

**Examples:**

API Key:
```json
{
  "name": "Stripe API",
  "type": "api_key",
  "data": {
    "api_key": "sk_live_xxxx"
  }
}
```

Basic Auth:
```json
{
  "name": "Twilio",
  "type": "basic",
  "data": {
    "username": "ACxxxxx",
    "password": "auth_token_here"
  }
}
```

Database:
```json
{
  "name": "Production PostgreSQL",
  "type": "basic",
  "data": {
    "host": "db.example.com",
    "port": 5432,
    "database": "myapp",
    "username": "admin",
    "password": "secret"
  }
}
```

Bearer Token:
```json
{
  "name": "Internal API",
  "type": "bearer",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

Custom:
```json
{
  "name": "Custom Service",
  "type": "custom",
  "data": {
    "custom": {
      "api_key": "xxx",
      "api_secret": "yyy",
      "region": "us-east-1"
    }
  }
}
```

**Response 201:**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440004",
    "workspace_id": "xxx",
    "created_by": "xxx",
    "name": "My OpenAI Key",
    "type": "api_key",
    "description": "Production API key for GPT-4",
    "provider": null,
    "provider_account_id": null,
    "token_expires_at": null,
    "last_used_at": null,
    "created_at": 1704067200,
    "updated_at": 1704067200
  }
}
```

**Error Responses:**

400 Bad Request:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "name", "message": "name is required" },
      { "field": "type", "message": "type must be one of: api_key, oauth2, basic, bearer, custom" }
    ]
  }
}
```

---

### 6. Get Single Credential

```http
GET /workspaces/{workspaceId}/credentials/{credentialId}
Authorization: Bearer <token>
```

**Response 200:** Same as single item in list response.

**Response 404:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "credential not found"
  }
}
```

---

### 7. Update Credential

```http
PUT /workspaces/{workspaceId}/credentials/{credentialId}
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body (all fields optional):**
```json
{
  "name": "Updated Name",
  "description": "New description",
  "data": {
    "api_key": "new-api-key"
  }
}
```

**Response 200:** Updated credential object.

---

### 8. Delete Credential

```http
DELETE /workspaces/{workspaceId}/credentials/{credentialId}
Authorization: Bearer <token>
```

**Response 204:** No content (success)

**Response 404:** Credential not found

---

### 9. Test Credential

Tests if the credential is valid/working.

```http
POST /workspaces/{workspaceId}/credentials/{credentialId}/test
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "success": true
}
```

**Response 200 (failed test):**
```json
{
  "success": false
}
```

---

### 10. Refresh OAuth Token

Manually triggers a token refresh for OAuth credentials.

```http
POST /workspaces/{workspaceId}/credentials/{credentialId}/refresh
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "message": "Token refreshed successfully",
  "credential_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response 400:**
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "no refresh token available"
  }
}
```

---

## TypeScript Types

Create a file `src/types/credentials.ts`:

```typescript
// =====================================================
// CREDENTIAL TYPES
// =====================================================

/**
 * Credential types supported by the backend
 */
export type TCredentialType = 'api_key' | 'oauth2' | 'basic' | 'bearer' | 'custom';

/**
 * Authentication types for service configuration
 */
export type TAuthType = 'oauth' | 'api_key' | 'basic' | 'bearer' | 'custom';

/**
 * Service categories for grouping in UI
 */
export type TServiceCategory =
  | 'AI'
  | 'Communication'
  | 'CRM'
  | 'Database'
  | 'Development'
  | 'E-commerce'
  | 'File Storage'
  | 'Marketing'
  | 'Payment'
  | 'Productivity'
  | 'Social Media'
  | 'Other';

// =====================================================
// API RESPONSE TYPES
// =====================================================

/**
 * Credential object returned from API
 */
export interface ICredential {
  id: string;
  workspace_id: string;
  created_by: string;
  name: string;
  type: TCredentialType;
  description: string | null;
  provider: string | null;
  provider_account_id: string | null;
  token_expires_at: number | null;  // Unix timestamp
  last_used_at: number | null;      // Unix timestamp
  created_at: number;               // Unix timestamp
  updated_at: number;               // Unix timestamp
}

/**
 * Credential with HATEOAS actions
 */
export interface ICredentialWithActions extends ICredential {
  actions: ICredentialAction[];
}

/**
 * Action available on a credential
 */
export interface ICredentialAction {
  name: 'edit' | 'test' | 'delete' | 'refresh';
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  href: string;
  label: string;
}

/**
 * OAuth provider status from backend
 */
export interface IOAuthProvider {
  id: string;
  name: string;
  configured: boolean;
  scopes: string[];
}

/**
 * OAuth authorization URL response
 */
export interface IOAuthAuthResponse {
  url: string;
  state: string;
}

// =====================================================
// REQUEST TYPES
// =====================================================

/**
 * Credential data structure (matches backend models.CredentialData)
 */
export interface ICredentialData {
  // Provider (for OAuth)
  provider?: string;

  // API Key
  api_key?: string;

  // OAuth2
  client_id?: string;
  client_secret?: string;
  access_token?: string;
  refresh_token?: string;
  token_type?: string;
  scope?: string;
  expires_at?: string;

  // Basic Auth
  username?: string;
  password?: string;

  // Bearer Token
  token?: string;

  // Database credentials
  host?: string;
  port?: number;
  database?: string;

  // Connection string
  connectionString?: string;

  // Custom fields
  custom?: Record<string, string>;

  // Generic data
  data?: Record<string, unknown>;
}

/**
 * Create credential request
 */
export interface ICreateCredentialRequest {
  name: string;
  type: TCredentialType;
  data: ICredentialData;
  description?: string;
}

/**
 * Update credential request
 */
export interface IUpdateCredentialRequest {
  name?: string;
  data?: ICredentialData;
  description?: string;
}

/**
 * Credential list filters
 */
export interface ICredentialFilters {
  type?: TCredentialType;
  search?: string;
  sort_by?: 'name' | 'created_at' | 'last_used_at' | 'type';
  order?: 'asc' | 'desc';
  page?: number;
  per_page?: number;
}

// =====================================================
// PAGINATION TYPES
// =====================================================

export interface IPaginationMeta {
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface ILinks {
  self: string;
  first?: string;
  last?: string;
  next?: string;
  prev?: string;
}

export interface IListResponse<T> {
  data: T[];
  meta: IPaginationMeta;
  links: ILinks;
}

export interface ISingleResponse<T> {
  data: T;
  links?: ILinks;
}

export interface IErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Array<{ field: string; message: string }>;
  };
}

// =====================================================
// SERVICE CONFIGURATION TYPES
// =====================================================

/**
 * Field definition for credential forms
 */
export interface IServiceField {
  /** Field key in credential data object */
  name: string;
  /** Display label */
  label: string;
  /** Input type */
  type: 'text' | 'password' | 'url' | 'number' | 'email' | 'textarea';
  /** Is this field required? */
  required: boolean;
  /** Placeholder text */
  placeholder?: string;
  /** Help text shown below input */
  helpText?: string;
  /** Default value */
  defaultValue?: string | number;
  /** Validation pattern (regex) */
  pattern?: string;
  /** Min value for number inputs */
  min?: number;
  /** Max value for number inputs */
  max?: number;
}

/**
 * Service configuration for the service registry
 */
export interface IServiceConfig {
  /** Unique service identifier */
  id: string;
  /** Display name */
  name: string;
  /** Icon identifier (component name, URL, or icon library key) */
  icon: string;
  /** Category for grouping */
  category: TServiceCategory;
  /** Authentication type */
  authType: TAuthType;
  /** Backend OAuth provider name (if authType is 'oauth') */
  oauthProvider?: string;
  /** Credential type to use when creating */
  credentialType: TCredentialType;
  /** Form fields (for non-OAuth services) */
  fields?: IServiceField[];
  /** Service description */
  description?: string;
  /** Link to provider's API docs */
  helpUrl?: string;
  /** Default credential name template */
  defaultName?: string;
  /** Is this service available? (false for "coming soon") */
  isAvailable?: boolean;
  /** Tags for search/filtering */
  tags?: string[];
}

/**
 * Service with availability status (after merging with backend data)
 */
export interface IServiceWithStatus extends IServiceConfig {
  /** Is OAuth configured on backend? (only for OAuth services) */
  isOAuthConfigured?: boolean;
  /** Is the service available for use? */
  isAvailable: boolean;
}
```

---

## Service Registry Configuration

Create a file `src/config/services.ts`:

```typescript
import { IServiceConfig, TServiceCategory } from '@/types/credentials';

/**
 * Complete service registry
 * 
 * This is the single source of truth for all supported services.
 * Add new services here - no backend changes needed for API key services.
 */
export const SERVICES: IServiceConfig[] = [
  // =========================================================
  // AI SERVICES
  // =========================================================
  {
    id: 'openai',
    name: 'OpenAI',
    icon: 'openai',
    category: 'AI',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'GPT-4, GPT-3.5, DALL-E, Whisper APIs',
    defaultName: 'OpenAI API',
    fields: [
      {
        name: 'api_key',
        label: 'API Key',
        type: 'password',
        required: true,
        placeholder: 'sk-proj-...',
        helpText: 'Get your API key from platform.openai.com/api-keys',
      },
    ],
    helpUrl: 'https://platform.openai.com/api-keys',
    tags: ['ai', 'gpt', 'chatgpt', 'llm'],
  },
  {
    id: 'anthropic',
    name: 'Anthropic',
    icon: 'anthropic',
    category: 'AI',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'Claude AI models',
    defaultName: 'Anthropic API',
    fields: [
      {
        name: 'api_key',
        label: 'API Key',
        type: 'password',
        required: true,
        placeholder: 'sk-ant-...',
        helpText: 'Get your API key from console.anthropic.com',
      },
    ],
    helpUrl: 'https://console.anthropic.com/',
    tags: ['ai', 'claude', 'llm'],
  },
  {
    id: 'replicate',
    name: 'Replicate',
    icon: 'replicate',
    category: 'AI',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'Run ML models in the cloud',
    fields: [
      {
        name: 'api_key',
        label: 'API Token',
        type: 'password',
        required: true,
        placeholder: 'r8_...',
        helpText: 'Get from replicate.com/account/api-tokens',
      },
    ],
    helpUrl: 'https://replicate.com/account/api-tokens',
    tags: ['ai', 'ml', 'models'],
  },

  // =========================================================
  // PRODUCTIVITY (OAuth)
  // =========================================================
  {
    id: 'google_sheets',
    name: 'Google Sheets',
    icon: 'google-sheets',
    category: 'Productivity',
    authType: 'oauth',
    oauthProvider: 'google',
    credentialType: 'oauth2',
    description: 'Read and write Google Sheets data',
    defaultName: 'Google Sheets Connection',
    helpUrl: 'https://developers.google.com/sheets',
    tags: ['google', 'spreadsheet', 'sheets'],
  },
  {
    id: 'google_drive',
    name: 'Google Drive',
    icon: 'google-drive',
    category: 'File Storage',
    authType: 'oauth',
    oauthProvider: 'google',
    credentialType: 'oauth2',
    description: 'Upload, download, and manage files',
    defaultName: 'Google Drive Connection',
    tags: ['google', 'storage', 'files'],
  },
  {
    id: 'google_calendar',
    name: 'Google Calendar',
    icon: 'google-calendar',
    category: 'Productivity',
    authType: 'oauth',
    oauthProvider: 'google',
    credentialType: 'oauth2',
    description: 'Create and manage calendar events',
    defaultName: 'Google Calendar Connection',
    tags: ['google', 'calendar', 'events'],
  },
  {
    id: 'gmail',
    name: 'Gmail',
    icon: 'gmail',
    category: 'Communication',
    authType: 'oauth',
    oauthProvider: 'google',
    credentialType: 'oauth2',
    description: 'Send and read emails',
    defaultName: 'Gmail Connection',
    tags: ['google', 'email', 'mail'],
  },
  {
    id: 'notion',
    name: 'Notion',
    icon: 'notion',
    category: 'Productivity',
    authType: 'oauth',
    oauthProvider: 'notion',
    credentialType: 'oauth2',
    description: 'Access Notion pages and databases',
    defaultName: 'Notion Connection',
    helpUrl: 'https://developers.notion.com/',
    tags: ['notion', 'wiki', 'database'],
  },
  {
    id: 'airtable',
    name: 'Airtable',
    icon: 'airtable',
    category: 'Productivity',
    authType: 'oauth',
    oauthProvider: 'airtable',
    credentialType: 'oauth2',
    description: 'Access Airtable bases and records',
    defaultName: 'Airtable Connection',
    helpUrl: 'https://airtable.com/developers/web/api/oauth-reference',
    tags: ['airtable', 'database', 'spreadsheet'],
  },

  // =========================================================
  // COMMUNICATION
  // =========================================================
  {
    id: 'slack',
    name: 'Slack',
    icon: 'slack',
    category: 'Communication',
    authType: 'oauth',
    oauthProvider: 'slack',
    credentialType: 'oauth2',
    description: 'Send messages and interact with Slack',
    defaultName: 'Slack Connection',
    helpUrl: 'https://api.slack.com/',
    tags: ['slack', 'chat', 'messaging'],
  },
  {
    id: 'discord',
    name: 'Discord',
    icon: 'discord',
    category: 'Communication',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'Discord bot integration',
    defaultName: 'Discord Bot',
    fields: [
      {
        name: 'api_key',
        label: 'Bot Token',
        type: 'password',
        required: true,
        placeholder: 'Enter your bot token',
        helpText: 'Get from discord.com/developers/applications',
      },
    ],
    helpUrl: 'https://discord.com/developers/docs',
    tags: ['discord', 'chat', 'bot'],
  },
  {
    id: 'twilio',
    name: 'Twilio',
    icon: 'twilio',
    category: 'Communication',
    authType: 'basic',
    credentialType: 'basic',
    description: 'SMS, voice calls, and WhatsApp',
    defaultName: 'Twilio Account',
    fields: [
      {
        name: 'username',
        label: 'Account SID',
        type: 'text',
        required: true,
        placeholder: 'ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
        helpText: 'Find in your Twilio Console',
      },
      {
        name: 'password',
        label: 'Auth Token',
        type: 'password',
        required: true,
        placeholder: 'Your auth token',
      },
    ],
    helpUrl: 'https://www.twilio.com/docs',
    tags: ['sms', 'voice', 'whatsapp', 'messaging'],
  },
  {
    id: 'sendgrid',
    name: 'SendGrid',
    icon: 'sendgrid',
    category: 'Communication',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'Email delivery service',
    defaultName: 'SendGrid API',
    fields: [
      {
        name: 'api_key',
        label: 'API Key',
        type: 'password',
        required: true,
        placeholder: 'SG.xxxxxx',
        helpText: 'Create an API key at app.sendgrid.com/settings/api_keys',
      },
    ],
    helpUrl: 'https://docs.sendgrid.com/',
    tags: ['email', 'mail', 'transactional'],
  },

  // =========================================================
  // DEVELOPMENT
  // =========================================================
  {
    id: 'github_oauth',
    name: 'GitHub (OAuth)',
    icon: 'github',
    category: 'Development',
    authType: 'oauth',
    oauthProvider: 'github',
    credentialType: 'oauth2',
    description: 'Full GitHub access via OAuth',
    defaultName: 'GitHub Connection',
    helpUrl: 'https://docs.github.com/en/rest',
    tags: ['github', 'git', 'code', 'repository'],
  },
  {
    id: 'github',
    name: 'GitHub (Token)',
    icon: 'github',
    category: 'Development',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'GitHub Personal Access Token',
    defaultName: 'GitHub Token',
    fields: [
      {
        name: 'api_key',
        label: 'Personal Access Token',
        type: 'password',
        required: true,
        placeholder: 'ghp_xxxxxxxxxxxx',
        helpText: 'Create at github.com/settings/tokens',
      },
    ],
    helpUrl: 'https://github.com/settings/tokens',
    tags: ['github', 'git', 'code', 'repository'],
  },
  {
    id: 'gitlab',
    name: 'GitLab',
    icon: 'gitlab',
    category: 'Development',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'GitLab Personal Access Token',
    defaultName: 'GitLab Token',
    fields: [
      {
        name: 'api_key',
        label: 'Personal Access Token',
        type: 'password',
        required: true,
        placeholder: 'glpat-xxxxxxxxxxxx',
        helpText: 'Create at gitlab.com/-/profile/personal_access_tokens',
      },
    ],
    helpUrl: 'https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html',
    tags: ['gitlab', 'git', 'code', 'repository'],
  },
  {
    id: 'jira',
    name: 'Jira',
    icon: 'jira',
    category: 'Development',
    authType: 'basic',
    credentialType: 'basic',
    description: 'Jira project management',
    defaultName: 'Jira Account',
    fields: [
      {
        name: 'host',
        label: 'Jira URL',
        type: 'url',
        required: true,
        placeholder: 'https://yourcompany.atlassian.net',
      },
      {
        name: 'username',
        label: 'Email',
        type: 'email',
        required: true,
        placeholder: 'your-email@company.com',
      },
      {
        name: 'password',
        label: 'API Token',
        type: 'password',
        required: true,
        helpText: 'Create at id.atlassian.com/manage-profile/security/api-tokens',
      },
    ],
    helpUrl: 'https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/',
    tags: ['jira', 'atlassian', 'project', 'issues'],
  },

  // =========================================================
  // PAYMENT
  // =========================================================
  {
    id: 'stripe',
    name: 'Stripe',
    icon: 'stripe',
    category: 'Payment',
    authType: 'api_key',
    credentialType: 'api_key',
    description: 'Payment processing',
    defaultName: 'Stripe API',
    fields: [
      {
        name: 'api_key',
        label: 'Secret Key',
        type: 'password',
        required: true,
        placeholder: 'sk_live_... or sk_test_...',
        helpText: 'Get from dashboard.stripe.com/apikeys',
      },
    ],
    helpUrl: 'https://stripe.com/docs/api',
    tags: ['stripe', 'payment', 'billing'],
  },
  {
    id: 'paypal',
    name: 'PayPal',
    icon: 'paypal',
    category: 'Payment',
    authType: 'basic',
    credentialType: 'basic',
    description: 'PayPal REST API',
    defaultName: 'PayPal API',
    fields: [
      {
        name: 'username',
        label: 'Client ID',
        type: 'text',
        required: true,
        placeholder: 'Your Client ID',
      },
      {
        name: 'password',
        label: 'Client Secret',
        type: 'password',
        required: true,
        placeholder: 'Your Client Secret',
      },
    ],
    helpUrl: 'https://developer.paypal.com/docs/api/overview/',
    tags: ['paypal', 'payment'],
  },

  // =========================================================
  // CRM
  // =========================================================
  {
    id: 'hubspot',
    name: 'HubSpot',
    icon: 'hubspot',
    category: 'CRM',
    authType: 'oauth',
    oauthProvider: 'hubspot',
    credentialType: 'oauth2',
    description: 'CRM and marketing automation',
    defaultName: 'HubSpot Connection',
    helpUrl: 'https://developers.hubspot.com/',
    tags: ['hubspot', 'crm', 'marketing'],
  },
  {
    id: 'salesforce',
    name: 'Salesforce',
    icon: 'salesforce',
    category: 'CRM',
    authType: 'oauth',
    oauthProvider: 'salesforce',
    credentialType: 'oauth2',
    description: 'Salesforce CRM',
    defaultName: 'Salesforce Connection',
    helpUrl: 'https://developer.salesforce.com/',
    tags: ['salesforce', 'crm'],
  },

  // =========================================================
  // DATABASES
  // =========================================================
  {
    id: 'postgresql',
    name: 'PostgreSQL',
    icon: 'postgresql',
    category: 'Database',
    authType: 'basic',
    credentialType: 'basic',
    description: 'PostgreSQL database',
    defaultName: 'PostgreSQL Database',
    fields: [
      {
        name: 'host',
        label: 'Host',
        type: 'text',
        required: true,
        placeholder: 'localhost or db.example.com',
      },
      {
        name: 'port',
        label: 'Port',
        type: 'number',
        required: true,
        placeholder: '5432',
        defaultValue: 5432,
      },
      {
        name: 'database',
        label: 'Database Name',
        type: 'text',
        required: true,
        placeholder: 'myapp',
      },
      {
        name: 'username',
        label: 'Username',
        type: 'text',
        required: true,
        placeholder: 'postgres',
      },
      {
        name: 'password',
        label: 'Password',
        type: 'password',
        required: true,
      },
    ],
    tags: ['postgresql', 'postgres', 'database', 'sql'],
  },
  {
    id: 'mysql',
    name: 'MySQL',
    icon: 'mysql',
    category: 'Database',
    authType: 'basic',
    credentialType: 'basic',
    description: 'MySQL database',
    defaultName: 'MySQL Database',
    fields: [
      {
        name: 'host',
        label: 'Host',
        type: 'text',
        required: true,
        placeholder: 'localhost',
      },
      {
        name: 'port',
        label: 'Port',
        type: 'number',
        required: true,
        placeholder: '3306',
        defaultValue: 3306,
      },
      {
        name: 'database',
        label: 'Database Name',
        type: 'text',
        required: true,
      },
      {
        name: 'username',
        label: 'Username',
        type: 'text',
        required: true,
      },
      {
        name: 'password',
        label: 'Password',
        type: 'password',
        required: true,
      },
    ],
    tags: ['mysql', 'database', 'sql'],
  },
  {
    id: 'mongodb',
    name: 'MongoDB',
    icon: 'mongodb',
    category: 'Database',
    authType: 'custom',
    credentialType: 'custom',
    description: 'MongoDB database',
    defaultName: 'MongoDB Connection',
    fields: [
      {
        name: 'connectionString',
        label: 'Connection String',
        type: 'password',
        required: true,
        placeholder: 'mongodb+srv://user:pass@cluster.mongodb.net/dbname',
        helpText: 'Get from MongoDB Atlas or your MongoDB instance',
      },
    ],
    tags: ['mongodb', 'database', 'nosql'],
  },
  {
    id: 'redis',
    name: 'Redis',
    icon: 'redis',
    category: 'Database',
    authType: 'custom',
    credentialType: 'custom',
    description: 'Redis cache/database',
    defaultName: 'Redis Connection',
    fields: [
      {
        name: 'host',
        label: 'Host',
        type: 'text',
        required: true,
        placeholder: 'localhost',
      },
      {
        name: 'port',
        label: 'Port',
        type: 'number',
        required: true,
        placeholder: '6379',
        defaultValue: 6379,
      },
      {
        name: 'password',
        label: 'Password',
        type: 'password',
        required: false,
        placeholder: 'Optional',
      },
    ],
    tags: ['redis', 'cache', 'database'],
  },

  // =========================================================
  // CLOUD / AWS
  // =========================================================
  {
    id: 'aws',
    name: 'AWS',
    icon: 'aws',
    category: 'Other',
    authType: 'custom',
    credentialType: 'custom',
    description: 'Amazon Web Services',
    defaultName: 'AWS Credentials',
    fields: [
      {
        name: 'custom.access_key_id',
        label: 'Access Key ID',
        type: 'text',
        required: true,
        placeholder: 'AKIAIOSFODNN7EXAMPLE',
      },
      {
        name: 'custom.secret_access_key',
        label: 'Secret Access Key',
        type: 'password',
        required: true,
        placeholder: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
      },
      {
        name: 'custom.region',
        label: 'Default Region',
        type: 'text',
        required: false,
        placeholder: 'us-east-1',
        defaultValue: 'us-east-1',
      },
    ],
    helpUrl: 'https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html',
    tags: ['aws', 'amazon', 'cloud', 's3'],
  },

  // =========================================================
  // HTTP / GENERIC
  // =========================================================
  {
    id: 'http_header',
    name: 'HTTP Header Auth',
    icon: 'http',
    category: 'Other',
    authType: 'custom',
    credentialType: 'custom',
    description: 'Custom HTTP header authentication',
    defaultName: 'HTTP Auth',
    fields: [
      {
        name: 'custom.header_name',
        label: 'Header Name',
        type: 'text',
        required: true,
        placeholder: 'X-API-Key',
        defaultValue: 'Authorization',
      },
      {
        name: 'custom.header_value',
        label: 'Header Value',
        type: 'password',
        required: true,
        placeholder: 'Bearer your-token',
      },
    ],
    tags: ['http', 'api', 'custom'],
  },
  {
    id: 'webhook',
    name: 'Webhook',
    icon: 'webhook',
    category: 'Other',
    authType: 'custom',
    credentialType: 'custom',
    description: 'Webhook with optional authentication',
    defaultName: 'Webhook Config',
    fields: [
      {
        name: 'custom.url',
        label: 'Webhook URL',
        type: 'url',
        required: true,
        placeholder: 'https://example.com/webhook',
      },
      {
        name: 'custom.secret',
        label: 'Secret (optional)',
        type: 'password',
        required: false,
        placeholder: 'For signature verification',
      },
    ],
    tags: ['webhook', 'http'],
  },
];

// =========================================================
// HELPER FUNCTIONS
// =========================================================

/**
 * Get a service by its ID
 */
export const getServiceById = (id: string): IServiceConfig | undefined => {
  return SERVICES.find(s => s.id === id);
};

/**
 * Get services grouped by category
 */
export const getServicesByCategory = (): Record<TServiceCategory, IServiceConfig[]> => {
  const result: Partial<Record<TServiceCategory, IServiceConfig[]>> = {};
  
  SERVICES.forEach(service => {
    if (!result[service.category]) {
      result[service.category] = [];
    }
    result[service.category]!.push(service);
  });
  
  return result as Record<TServiceCategory, IServiceConfig[]>;
};

/**
 * Get all OAuth services
 */
export const getOAuthServices = (): IServiceConfig[] => {
  return SERVICES.filter(s => s.authType === 'oauth');
};

/**
 * Get all API key services
 */
export const getApiKeyServices = (): IServiceConfig[] => {
  return SERVICES.filter(s => s.authType === 'api_key');
};

/**
 * Search services by name, description, or tags
 */
export const searchServices = (query: string): IServiceConfig[] => {
  const lowerQuery = query.toLowerCase();
  return SERVICES.filter(s => 
    s.name.toLowerCase().includes(lowerQuery) ||
    s.description?.toLowerCase().includes(lowerQuery) ||
    s.tags?.some(tag => tag.toLowerCase().includes(lowerQuery))
  );
};

/**
 * Get unique OAuth providers from services
 */
export const getUniqueOAuthProviders = (): string[] => {
  const providers = new Set<string>();
  SERVICES.forEach(s => {
    if (s.oauthProvider) {
      providers.add(s.oauthProvider);
    }
  });
  return Array.from(providers);
};

/**
 * Get all categories with counts
 */
export const getCategoriesWithCounts = (): Array<{ category: TServiceCategory; count: number }> => {
  const counts: Record<string, number> = {};
  
  SERVICES.forEach(s => {
    counts[s.category] = (counts[s.category] || 0) + 1;
  });
  
  return Object.entries(counts)
    .map(([category, count]) => ({ category: category as TServiceCategory, count }))
    .sort((a, b) => b.count - a.count);
};
```

---

## React Hooks

Create a file `src/hooks/useCredentials.ts`:

```typescript
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import {
  ICredential,
  ICredentialWithActions,
  ICreateCredentialRequest,
  IUpdateCredentialRequest,
  ICredentialFilters,
  IListResponse,
  ISingleResponse,
  IOAuthProvider,
  IOAuthAuthResponse,
  IServiceWithStatus,
} from '@/types/credentials';
import { SERVICES, getUniqueOAuthProviders } from '@/config/services';

// =====================================================
// QUERY KEYS
// =====================================================

export const credentialKeys = {
  all: ['credentials'] as const,
  lists: () => [...credentialKeys.all, 'list'] as const,
  list: (workspaceId: string, filters?: ICredentialFilters) =>
    [...credentialKeys.lists(), workspaceId, filters] as const,
  details: () => [...credentialKeys.all, 'detail'] as const,
  detail: (workspaceId: string, id: string) =>
    [...credentialKeys.details(), workspaceId, id] as const,
  oauthProviders: ['oauth-providers'] as const,
};

// =====================================================
// CREDENTIALS HOOKS
// =====================================================

/**
 * Fetch list of credentials for a workspace
 */
export const useCredentials = (workspaceId: string, filters?: ICredentialFilters) => {
  return useQuery({
    queryKey: credentialKeys.list(workspaceId, filters),
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.type) params.set('type', filters.type);
      if (filters?.search) params.set('search', filters.search);
      if (filters?.sort_by) params.set('sort_by', filters.sort_by);
      if (filters?.order) params.set('order', filters.order);
      if (filters?.page) params.set('page', String(filters.page));
      if (filters?.per_page) params.set('per_page', String(filters.per_page));

      const response = await api.get<IListResponse<ICredentialWithActions>>(
        `/workspaces/${workspaceId}/credentials?${params.toString()}`
      );
      return response.data;
    },
    enabled: !!workspaceId,
  });
};

/**
 * Fetch single credential
 */
export const useCredential = (workspaceId: string, credentialId: string) => {
  return useQuery({
    queryKey: credentialKeys.detail(workspaceId, credentialId),
    queryFn: async () => {
      const response = await api.get<ISingleResponse<ICredentialWithActions>>(
        `/workspaces/${workspaceId}/credentials/${credentialId}`
      );
      return response.data;
    },
    enabled: !!workspaceId && !!credentialId,
  });
};

/**
 * Create a new credential
 */
export const useCreateCredential = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: ICreateCredentialRequest) => {
      const response = await api.post<ISingleResponse<ICredential>>(
        `/workspaces/${workspaceId}/credentials`,
        data
      );
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: credentialKeys.lists() });
    },
  });
};

/**
 * Update a credential
 */
export const useUpdateCredential = (workspaceId: string, credentialId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: IUpdateCredentialRequest) => {
      const response = await api.put<ISingleResponse<ICredential>>(
        `/workspaces/${workspaceId}/credentials/${credentialId}`,
        data
      );
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: credentialKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: credentialKeys.detail(workspaceId, credentialId),
      });
    },
  });
};

/**
 * Delete a credential
 */
export const useDeleteCredential = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (credentialId: string) => {
      await api.delete(`/workspaces/${workspaceId}/credentials/${credentialId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: credentialKeys.lists() });
    },
  });
};

/**
 * Test a credential
 */
export const useTestCredential = (workspaceId: string) => {
  return useMutation({
    mutationFn: async (credentialId: string) => {
      const response = await api.post<{ success: boolean }>(
        `/workspaces/${workspaceId}/credentials/${credentialId}/test`
      );
      return response.data;
    },
  });
};

/**
 * Refresh OAuth token
 */
export const useRefreshCredentialToken = (workspaceId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (credentialId: string) => {
      const response = await api.post<{ message: string; credential_id: string }>(
        `/workspaces/${workspaceId}/credentials/${credentialId}/refresh`
      );
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: credentialKeys.lists() });
    },
  });
};

// =====================================================
// OAUTH HOOKS
// =====================================================

/**
 * Fetch list of OAuth providers and their status
 */
export const useOAuthProviders = () => {
  return useQuery({
    queryKey: credentialKeys.oauthProviders,
    queryFn: async () => {
      const response = await api.get<{ data: IOAuthProvider[] }>('/oauth/providers');
      return response.data.data;
    },
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  });
};

/**
 * Start OAuth flow
 */
export const useStartOAuth = (workspaceId: string) => {
  return useMutation({
    mutationFn: async ({
      provider,
      credentialName,
      redirectUrl,
    }: {
      provider: string;
      credentialName?: string;
      redirectUrl?: string;
    }) => {
      const params = new URLSearchParams();
      if (credentialName) params.set('credential_name', credentialName);
      if (redirectUrl) params.set('redirect_url', redirectUrl);

      const response = await api.get<IOAuthAuthResponse>(
        `/workspaces/${workspaceId}/oauth/authorize/${provider}?${params.toString()}`
      );
      return response.data;
    },
  });
};

// =====================================================
// SERVICE AVAILABILITY HOOK
// =====================================================

/**
 * Get services with availability status
 * Merges frontend service config with backend OAuth provider status
 */
export const useServicesWithStatus = () => {
  const { data: oauthProviders, isLoading } = useOAuthProviders();

  // Build map of configured providers
  const configuredProviders = new Set(
    oauthProviders?.filter(p => p.configured).map(p => p.id) ?? []
  );

  // Enhance services with availability
  const services: IServiceWithStatus[] = SERVICES.map(service => {
    const isOAuthService = service.authType === 'oauth';
    const isOAuthConfigured = service.oauthProvider
      ? configuredProviders.has(service.oauthProvider)
      : undefined;

    return {
      ...service,
      isOAuthConfigured,
      // Service is available if:
      // - It's not OAuth (API key services are always available)
      // - OR it's OAuth and the provider is configured
      isAvailable: !isOAuthService || (isOAuthConfigured ?? false),
    };
  });

  return {
    services,
    oauthProviders: oauthProviders ?? [],
    isLoading,
    configuredProviders: Array.from(configuredProviders),
  };
};

// =====================================================
// URL PARAMS HOOK (for OAuth callback handling)
// =====================================================

/**
 * Handle OAuth callback URL parameters
 */
export const useOAuthCallbackParams = () => {
  const searchParams = new URLSearchParams(window.location.search);

  return {
    isOAuthCallback: searchParams.has('oauth'),
    status: searchParams.get('oauth') as 'success' | 'error' | null,
    credentialId: searchParams.get('credential_id'),
    error: searchParams.get('error'),
    errorDescription: searchParams.get('error_description'),
    clearParams: () => {
      window.history.replaceState({}, '', window.location.pathname);
    },
  };
};
```

---

## Component Specifications

### 1. CredentialsListPage

Main page component that displays all credentials and handles OAuth callbacks.

```tsx
// src/pages/CredentialsListPage.tsx

/**
 * REQUIREMENTS:
 * - Display list of credentials in a table or grid
 * - Search/filter functionality
 * - "Add Credential" button opens modal
 * - Handle OAuth callback URL params on mount
 * - Show success/error toasts for OAuth
 * - Actions: Edit, Test, Delete, Refresh (for OAuth)
 */

interface CredentialsListPageProps {
  workspaceId: string;
}

// Features to implement:
// 1. On mount, check for ?oauth=success or ?oauth=error in URL
// 2. Show appropriate toast message
// 3. Clear URL params after handling
// 4. Refetch credentials list after OAuth success
// 5. Display credentials in sortable table
// 6. Search bar to filter by name
// 7. Filter dropdown by type (OAuth, API Key, etc.)
// 8. Each row shows: Icon, Name, Type, Provider, Last Used, Actions
// 9. Actions dropdown: Edit, Test Connection, Refresh Token (OAuth only), Delete
// 10. Empty state when no credentials
```

### 2. AddCredentialModal

Multi-step modal for adding new credentials.

```tsx
// src/components/credentials/AddCredentialModal.tsx

/**
 * STEPS:
 * 1. Select Service - Grid of available services
 * 2. Configure - Either OAuth connect or API key form
 * 
 * REQUIREMENTS:
 * - Step 1: Show service grid grouped by category
 * - Search bar to filter services
 * - Show "Available" badge for configured OAuth providers
 * - Show "Coming Soon" for unconfigured OAuth providers
 * - API key services always available
 * - Step 2 (OAuth): Show service info + name input + "Connect" button
 * - Step 2 (API Key): Show dynamic form based on service.fields
 * - Back button to return to service selection
 * - Close button
 */

interface AddCredentialModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId: string;
}

// State:
// - step: 'select' | 'configure'
// - selectedService: IServiceConfig | null
// - credentialName: string
// - formData: Record<string, string>

// On service select:
// - Set selectedService
// - Set credentialName to service.defaultName or `My ${service.name}`
// - Move to 'configure' step

// On OAuth connect:
// - Call useStartOAuth mutation
// - Redirect to returned URL

// On API key submit:
// - Call useCreateCredential mutation
// - Close modal on success
// - Show error toast on failure
```

### 3. ServiceGrid

Grid component for displaying available services.

```tsx
// src/components/credentials/ServiceGrid.tsx

/**
 * REQUIREMENTS:
 * - Display services in a responsive grid
 * - Group by category with collapsible sections
 * - Each service card shows: Icon, Name, Description
 * - Visual indicator for availability (configured/not configured)
 * - Search functionality
 * - Click to select
 */

interface ServiceGridProps {
  services: IServiceWithStatus[];
  onSelect: (service: IServiceWithStatus) => void;
  searchQuery?: string;
}

// Service Card states:
// - Available: Normal styling, clickable
// - Coming Soon: Grayed out, "Coming Soon" badge, not clickable
// - Selected: Highlighted border

// Group headers: "AI", "Communication", "Database", etc.
// Show count in each group header
```

### 4. OAuthConnectForm

Form for initiating OAuth connection.

```tsx
// src/components/credentials/OAuthConnectForm.tsx

/**
 * REQUIREMENTS:
 * - Show service icon and name
 * - Name input field (pre-filled with default)
 * - Description of what will be connected
 * - "Connect with {Provider}" button styled with provider colors
 * - Back button
 * - Loading state while redirecting
 */

interface OAuthConnectFormProps {
  service: IServiceConfig;
  credentialName: string;
  onNameChange: (name: string) => void;
  onConnect: () => void;
  onBack: () => void;
  isLoading?: boolean;
}

// Button styling by provider:
// - Google: White bg, Google colors
// - GitHub: Black bg
// - Slack: Purple bg
// - Microsoft: Blue bg
```

### 5. ApiKeyForm

Dynamic form for API key and other non-OAuth credentials.

```tsx
// src/components/credentials/ApiKeyForm.tsx

/**
 * REQUIREMENTS:
 * - Dynamic form fields based on service.fields
 * - Name input (always present)
 * - Description input (optional)
 * - Show help text and links
 * - Validation
 * - Submit button
 * - Back button
 * - Loading state during submission
 */

interface ApiKeyFormProps {
  service: IServiceConfig;
  credentialName: string;
  onNameChange: (name: string) => void;
  onSubmit: (formData: ICredentialData) => void;
  onBack: () => void;
  isLoading?: boolean;
  error?: string;
}

// Field types to handle:
// - text: Regular text input
// - password: Password input with show/hide toggle
// - url: URL input with validation
// - number: Number input with min/max
// - email: Email input with validation
// - textarea: Multi-line text

// Build formData object from fields:
// - Simple fields: { api_key: "value" }
// - Nested fields (custom.xxx): { custom: { xxx: "value" } }
```

### 6. CredentialCard / CredentialRow

Display component for a single credential.

```tsx
// src/components/credentials/CredentialCard.tsx

/**
 * REQUIREMENTS:
 * - Service icon
 * - Credential name
 * - Type badge (OAuth, API Key, Basic, etc.)
 * - Provider name (for OAuth)
 * - Last used timestamp (relative: "2 hours ago")
 * - Token expiry warning (for OAuth with expiring tokens)
 * - Actions menu: Edit, Test, Refresh, Delete
 * - Click to view details (optional)
 */

interface CredentialCardProps {
  credential: ICredentialWithActions;
  onEdit: () => void;
  onTest: () => void;
  onRefresh: () => void;
  onDelete: () => void;
}

// Token expiry states:
// - Healthy: No indicator (> 24h until expiry)
// - Warning: Yellow badge "Expires soon" (< 24h)
// - Expired: Red badge "Expired"
// - N/A: No indicator (non-OAuth or no expiry info)
```

### 7. EditCredentialModal

Modal for editing existing credentials.

```tsx
// src/components/credentials/EditCredentialModal.tsx

/**
 * REQUIREMENTS:
 * - Edit name and description
 * - For API key credentials: Allow updating the secret data
 * - For OAuth credentials: Only name/description (can't edit tokens)
 * - Show "Refresh Token" button for OAuth
 * - Confirm dialog for data changes
 */

interface EditCredentialModalProps {
  isOpen: boolean;
  onClose: () => void;
  credential: ICredentialWithActions;
  workspaceId: string;
}
```

### 8. DeleteCredentialDialog

Confirmation dialog for deleting credentials.

```tsx
// src/components/credentials/DeleteCredentialDialog.tsx

/**
 * REQUIREMENTS:
 * - Show credential name
 * - Warning about workflows using this credential
 * - Type credential name to confirm (optional security)
 * - Delete button (red, destructive)
 * - Cancel button
 */

interface DeleteCredentialDialogProps {
  isOpen: boolean;
  onClose: () => void;
  credential: ICredential;
  onConfirm: () => void;
  isLoading?: boolean;
}
```

---

## OAuth Flow Implementation

### Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            USER JOURNEY                                      │
└─────────────────────────────────────────────────────────────────────────────┘

1. User clicks "Add Credential"
   │
   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        AddCredentialModal (Step 1)                           │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Search: [_______________]                                           │    │
│  │                                                                      │    │
│  │  ═══ AI ═══════════════════════════════════════════════════════════ │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐                                │    │
│  │  │ OpenAI  │ │Anthropic│ │Replicate│                                │    │
│  │  │  (key)  │ │  (key)  │ │  (key)  │                                │    │
│  │  └─────────┘ └─────────┘ └─────────┘                                │    │
│  │                                                                      │    │
│  │  ═══ Productivity ═════════════════════════════════════════════════ │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐                    │    │
│  │  │ Google  │ │  Gmail  │ │  Slack  │ │ Notion  │                    │    │
│  │  │ Sheets  │ │   ✓     │ │   ✗     │ │   ✗     │                    │    │
│  │  │   ✓     │ │         │ │ Coming  │ │ Coming  │                    │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘                    │    │
│  │                                                                      │    │
│  │  ✓ = OAuth configured    ✗ = Not configured (coming soon)           │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
   │
   │ User clicks "Google Sheets"
   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        AddCredentialModal (Step 2 - OAuth)                   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  [←] Back                                                            │    │
│  │                                                                      │    │
│  │                    [Google Sheets Icon]                              │    │
│  │                     Google Sheets                                    │    │
│  │                                                                      │    │
│  │  Connect your Google account to read and write                       │    │
│  │  spreadsheet data in your workflows.                                 │    │
│  │                                                                      │    │
│  │  Name:                                                               │    │
│  │  ┌─────────────────────────────────────────────────────────────┐    │    │
│  │  │ My Google Sheets                                             │    │    │
│  │  └─────────────────────────────────────────────────────────────┘    │    │
│  │                                                                      │    │
│  │  ┌─────────────────────────────────────────────────────────────┐    │    │
│  │  │     [G] Connect with Google                                  │    │    │
│  │  └─────────────────────────────────────────────────────────────┘    │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
   │
   │ User clicks "Connect with Google"
   │
   │ Frontend calls: GET /workspaces/{ws}/oauth/authorize/google
   │                      ?credential_name=My%20Google%20Sheets
   │
   │ Backend returns: { url: "https://accounts.google.com/...", state: "xxx" }
   │
   │ Frontend redirects: window.location.href = url
   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Google OAuth Screen                                  │
│                                                                              │
│                    Choose an account to continue                             │
│                         to LinkFlow                                          │
│                                                                              │
│                    ┌─────────────────────────────┐                          │
│                    │ 👤 user@gmail.com           │                          │
│                    └─────────────────────────────┘                          │
│                                                                              │
│                    LinkFlow wants to:                                        │
│                    ✓ View and manage spreadsheets                           │
│                    ✓ View and manage Drive files                            │
│                                                                              │
│                         [Allow]  [Deny]                                      │
└─────────────────────────────────────────────────────────────────────────────┘
   │
   │ User clicks "Allow"
   │
   │ Google redirects to: {backend}/api/v1/oauth/callback/google
   │                      ?code=xxx&state=xxx
   │
   │ Backend:
   │   1. Validates state
   │   2. Exchanges code for tokens
   │   3. Encrypts and stores credential
   │   4. Redirects to frontend
   │
   │ Backend redirects: {frontend}/credentials
   │                    ?oauth=success
   │                    &credential_id=uuid-here
   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CredentialsListPage                                  │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  ✓ Successfully connected Google Sheets!               [dismiss]    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  [+ Add Credential]                              Search: [___________]       │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ Name              │ Type   │ Provider │ Last Used │ Actions         │    │
│  ├───────────────────┼────────┼──────────┼───────────┼─────────────────┤    │
│  │ My Google Sheets  │ OAuth  │ Google   │ Just now  │ [···]           │    │
│  │ OpenAI API        │ API Key│ -        │ 2h ago    │ [···]           │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

### OAuth Callback Handler Code

```tsx
// In CredentialsListPage.tsx or a dedicated callback component

import { useEffect } from 'react';
import { useOAuthCallbackParams, useCredentials } from '@/hooks/useCredentials';
import { toast } from '@/components/ui/toast';

export const CredentialsListPage = ({ workspaceId }) => {
  const { refetch } = useCredentials(workspaceId);
  const { 
    isOAuthCallback,
    status,
    credentialId,
    error,
    errorDescription,
    clearParams 
  } = useOAuthCallbackParams();

  useEffect(() => {
    if (!isOAuthCallback) return;

    if (status === 'success') {
      toast.success('Credential created successfully!', {
        description: `Credential ID: ${credentialId}`,
      });
      refetch(); // Refresh the credentials list
    }

    if (status === 'error') {
      toast.error('Failed to connect', {
        description: errorDescription || error || 'Unknown error occurred',
      });
    }

    // Clean up URL params
    clearParams();
  }, [isOAuthCallback, status, credentialId, error, errorDescription]);

  // ... rest of component
};
```

---

## Error Handling

### API Error Codes

| Code | HTTP Status | Description | User Message |
|------|-------------|-------------|--------------|
| `VALIDATION_ERROR` | 400 | Invalid request data | "Please check your input" |
| `UNAUTHORIZED` | 401 | Not authenticated | "Please log in again" |
| `FORBIDDEN` | 403 | No permission | "You don't have permission" |
| `NOT_FOUND` | 404 | Resource not found | "Credential not found" |
| `CONFLICT` | 409 | Duplicate name | "A credential with this name already exists" |
| `INTERNAL_ERROR` | 500 | Server error | "Something went wrong. Please try again." |

### OAuth Error Codes

| Error | Description | User Message |
|-------|-------------|--------------|
| `access_denied` | User denied authorization | "Authorization was denied" |
| `invalid_state` | CSRF protection failed | "Security verification failed. Please try again." |
| `callback_failed` | Token exchange failed | "Failed to complete connection. Please try again." |
| `provider_not_configured` | OAuth not set up | "This service is not available yet" |

### Error Handling Example

```tsx
import { isAxiosError } from 'axios';
import { IErrorResponse } from '@/types/credentials';

const handleApiError = (error: unknown): string => {
  if (isAxiosError<IErrorResponse>(error)) {
    const apiError = error.response?.data?.error;
    
    if (apiError?.details) {
      // Validation errors - show first error
      return apiError.details[0].message;
    }
    
    return apiError?.message || 'An error occurred';
  }
  
  return 'An unexpected error occurred';
};

// Usage in mutation
const createCredential = useCreateCredential(workspaceId);

const handleSubmit = async (data: ICreateCredentialRequest) => {
  try {
    await createCredential.mutateAsync(data);
    toast.success('Credential created!');
    onClose();
  } catch (error) {
    toast.error(handleApiError(error));
  }
};
```

---

## UI/UX Guidelines

### Visual Design

1. **Service Icons**
   - Use official brand icons where possible
   - Consistent size: 32x32 or 40x40 in grids
   - Use neutral placeholder for unknown services

2. **Color Coding**
   - OAuth badges: Blue/Purple
   - API Key badges: Green
   - Basic Auth badges: Orange
   - Error states: Red
   - Success states: Green
   - Warning states: Yellow/Orange

3. **Loading States**
   - Skeleton loaders for lists
   - Spinner on buttons during actions
   - Disabled state during loading

4. **Empty States**
   - Friendly illustration
   - Clear CTA: "Add your first credential"
   - Brief explanation of what credentials are for

### Accessibility

1. All interactive elements must be keyboard accessible
2. Focus traps in modals
3. Aria labels for icon-only buttons
4. Color not the only indicator (use icons/text too)
5. Form field labels and error messages linked

### Responsive Design

1. Mobile: Single column grid for services
2. Tablet: 2-column grid
3. Desktop: 3-4 column grid
4. Table becomes cards on mobile

---

## Testing Checklist

### Unit Tests

- [ ] Service registry helper functions
- [ ] Form validation logic
- [ ] OAuth URL parameter parsing
- [ ] Credential data formatting

### Integration Tests

- [ ] Create API key credential flow
- [ ] Create OAuth credential (mock OAuth)
- [ ] Edit credential
- [ ] Delete credential
- [ ] Test credential connection
- [ ] Refresh OAuth token
- [ ] List with filters
- [ ] Search functionality
- [ ] Pagination

### E2E Tests

- [ ] Full OAuth flow (with test provider)
- [ ] Add API key credential end-to-end
- [ ] Edit and delete credential
- [ ] Error handling scenarios

### Manual Testing Checklist

- [ ] OAuth flow with Google (if configured)
- [ ] OAuth flow with GitHub (if configured)
- [ ] API key creation (OpenAI, Stripe)
- [ ] Database credential creation
- [ ] Edit credential name
- [ ] Test connection button
- [ ] Delete with confirmation
- [ ] Search/filter works
- [ ] Responsive design (mobile, tablet, desktop)
- [ ] Keyboard navigation
- [ ] Error toast messages

---

## Environment Variables (Frontend)

```env
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

---

## Questions for Frontend Team

1. What icon library are you using? (e.g., Lucide, Heroicons, custom SVGs)
2. What UI component library? (e.g., shadcn/ui, Radix, Chakra)
3. What form library? (e.g., React Hook Form, Formik)
4. What toast/notification library?
5. Any existing modal/dialog patterns to follow?
6. Current authentication/token storage mechanism?

---

## Summary

This documentation provides everything needed to implement the credentials system:

1. **API Reference** - Complete endpoint documentation
2. **TypeScript Types** - Full type definitions
3. **Service Registry** - Static configuration for all services
4. **React Hooks** - Data fetching and mutations
5. **Component Specs** - What each component should do
6. **OAuth Flow** - Step-by-step implementation guide
7. **Error Handling** - Error codes and user messages
8. **UI/UX Guidelines** - Design recommendations

The frontend owns the service catalog (names, icons, fields), while the backend handles OAuth infrastructure and credential storage with encryption.
