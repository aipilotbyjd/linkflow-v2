# Credential Sharing - Frontend Guide

## Overview

Credentials now support 3 sharing scopes:
- **private** - Only the owner can see/use
- **workspace** - All workspace members can use (default)
- **specific** - Only owner + explicitly shared users

## API Response Changes

Every credential now includes permission info:

```typescript
interface Credential {
  id: string;
  name: string;
  type: string;
  // ... other fields
  
  // NEW: Sharing fields
  sharing_scope: 'private' | 'workspace' | 'specific';
  is_owner: boolean;      // Is current user the owner?
  can_edit: boolean;      // Can current user edit/delete?
  can_share: boolean;     // Can current user share?
  shares?: CredentialShare[];  // Who it's shared with (only for owner)
}

interface CredentialShare {
  id: string;
  user_id: string;
  user?: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
  };
  permission: 'use';
  shared_by: string;
  created_at: number;
}
```

## TypeScript Types

```typescript
// types/credentials.ts

export type SharingScope = 'private' | 'workspace' | 'specific';

export interface Credential {
  id: string;
  workspace_id: string;
  created_by: string;
  name: string;
  type: 'api_key' | 'oauth2' | 'basic' | 'bearer' | 'custom';
  description?: string;
  provider?: string;
  provider_account_id?: string;
  token_expires_at?: number;
  sharing_scope: SharingScope;
  is_owner: boolean;
  can_edit: boolean;
  can_share: boolean;
  shares?: CredentialShare[];
  last_used_at?: number;
  created_at: number;
  updated_at: number;
}

export interface CredentialShare {
  id: string;
  user_id: string;
  user?: UserSummary;
  permission: 'use';
  shared_by: string;
  created_at: number;
}

export interface UserSummary {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

// Request types
export interface CreateCredentialRequest {
  name: string;
  type: string;
  data: Record<string, any>;
  description?: string;
  sharing_scope?: SharingScope; // Optional, defaults to 'workspace'
}

export interface ShareCredentialRequest {
  user_ids: string[];
}

export interface UpdateSharingScopeRequest {
  sharing_scope: SharingScope;
}
```

## API Endpoints

### List Credentials (with access control)
```typescript
// GET /api/v1/workspaces/{workspaceId}/credentials
// Only returns credentials the user can access

const response = await api.get(`/workspaces/${workspaceId}/credentials`);
// Returns: { data: Credential[], meta: {...}, links: {...} }
```

### Create Credential with Sharing Scope
```typescript
// POST /api/v1/workspaces/{workspaceId}/credentials

await api.post(`/workspaces/${workspaceId}/credentials`, {
  name: "My API Key",
  type: "api_key",
  data: { api_key: "sk-xxx" },
  sharing_scope: "private"  // Optional: 'private' | 'workspace' | 'specific'
});
```

### Share Credential with Users
```typescript
// POST /api/v1/workspaces/{workspaceId}/credentials/{credentialId}/share

await api.post(`/workspaces/${workspaceId}/credentials/${credentialId}/share`, {
  user_ids: ["user-uuid-1", "user-uuid-2"]
});
// Returns: CredentialShare[]
```

### Get Shares (who has access)
```typescript
// GET /api/v1/workspaces/{workspaceId}/credentials/{credentialId}/shares

const shares = await api.get(`/workspaces/${workspaceId}/credentials/${credentialId}/shares`);
// Returns: CredentialShare[]
```

### Remove Share
```typescript
// DELETE /api/v1/workspaces/{workspaceId}/credentials/{credentialId}/shares/{userId}

await api.delete(`/workspaces/${workspaceId}/credentials/${credentialId}/shares/${userId}`);
```

### Update Sharing Scope
```typescript
// PUT /api/v1/workspaces/{workspaceId}/credentials/{credentialId}/sharing-scope

await api.put(`/workspaces/${workspaceId}/credentials/${credentialId}/sharing-scope`, {
  sharing_scope: "specific"
});
```

## React Hooks

### useCredentials (updated)
```typescript
// hooks/useCredentials.ts

export function useCredentials(workspaceId: string) {
  return useQuery({
    queryKey: ['credentials', workspaceId],
    queryFn: async () => {
      const res = await api.get(`/workspaces/${workspaceId}/credentials`);
      return res.data as Credential[];
    }
  });
}
```

### useShareCredential
```typescript
export function useShareCredential(workspaceId: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ credentialId, userIds }: { credentialId: string; userIds: string[] }) => {
      const res = await api.post(
        `/workspaces/${workspaceId}/credentials/${credentialId}/share`,
        { user_ids: userIds }
      );
      return res.data as CredentialShare[];
    },
    onSuccess: (_, { credentialId }) => {
      queryClient.invalidateQueries(['credentials', workspaceId]);
      queryClient.invalidateQueries(['credential-shares', credentialId]);
    }
  });
}
```

### useUnshareCredential
```typescript
export function useUnshareCredential(workspaceId: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ credentialId, userId }: { credentialId: string; userId: string }) => {
      await api.delete(
        `/workspaces/${workspaceId}/credentials/${credentialId}/shares/${userId}`
      );
    },
    onSuccess: (_, { credentialId }) => {
      queryClient.invalidateQueries(['credentials', workspaceId]);
      queryClient.invalidateQueries(['credential-shares', credentialId]);
    }
  });
}
```

### useCredentialShares
```typescript
export function useCredentialShares(workspaceId: string, credentialId: string) {
  return useQuery({
    queryKey: ['credential-shares', credentialId],
    queryFn: async () => {
      const res = await api.get(
        `/workspaces/${workspaceId}/credentials/${credentialId}/shares`
      );
      return res.data as CredentialShare[];
    },
    enabled: !!credentialId
  });
}
```

### useUpdateSharingScope
```typescript
export function useUpdateSharingScope(workspaceId: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ credentialId, scope }: { credentialId: string; scope: SharingScope }) => {
      await api.put(
        `/workspaces/${workspaceId}/credentials/${credentialId}/sharing-scope`,
        { sharing_scope: scope }
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries(['credentials', workspaceId]);
    }
  });
}
```

## UI Components

### 1. Credential Card (updated)

```tsx
// components/credentials/CredentialCard.tsx

interface CredentialCardProps {
  credential: Credential;
  onEdit: () => void;
  onDelete: () => void;
  onShare: () => void;
}

export function CredentialCard({ credential, onEdit, onDelete, onShare }: CredentialCardProps) {
  return (
    <div className="credential-card">
      <div className="header">
        <h3>{credential.name}</h3>
        <SharingScopeBadge scope={credential.sharing_scope} />
      </div>
      
      <div className="meta">
        <span className="type">{credential.type}</span>
        {!credential.is_owner && (
          <span className="shared-badge">Shared with you</span>
        )}
      </div>
      
      <div className="actions">
        {/* Only show edit/delete if user can edit */}
        {credential.can_edit && (
          <>
            <button onClick={onEdit}>Edit</button>
            <button onClick={onDelete}>Delete</button>
          </>
        )}
        
        {/* Only show share if user can share */}
        {credential.can_share && (
          <button onClick={onShare}>Share</button>
        )}
        
        {/* Everyone can test */}
        <button onClick={() => testCredential(credential.id)}>Test</button>
      </div>
    </div>
  );
}
```

### 2. Sharing Scope Badge

```tsx
// components/credentials/SharingScopeBadge.tsx

const scopeConfig = {
  private: { label: 'Private', icon: '🔒', color: 'gray' },
  workspace: { label: 'Workspace', icon: '👥', color: 'blue' },
  specific: { label: 'Shared', icon: '🔗', color: 'green' }
};

export function SharingScopeBadge({ scope }: { scope: SharingScope }) {
  const config = scopeConfig[scope];
  return (
    <span className={`badge badge-${config.color}`}>
      {config.icon} {config.label}
    </span>
  );
}
```

### 3. Sharing Scope Selector

```tsx
// components/credentials/SharingScopeSelector.tsx

interface SharingScopeSelectorProps {
  value: SharingScope;
  onChange: (scope: SharingScope) => void;
  disabled?: boolean;
}

export function SharingScopeSelector({ value, onChange, disabled }: SharingScopeSelectorProps) {
  return (
    <div className="sharing-scope-selector">
      <label>Who can use this credential?</label>
      
      <div className="options">
        <label className={value === 'private' ? 'selected' : ''}>
          <input
            type="radio"
            name="sharing_scope"
            value="private"
            checked={value === 'private'}
            onChange={() => onChange('private')}
            disabled={disabled}
          />
          <div>
            <strong>🔒 Private</strong>
            <p>Only you can see and use this credential</p>
          </div>
        </label>
        
        <label className={value === 'workspace' ? 'selected' : ''}>
          <input
            type="radio"
            name="sharing_scope"
            value="workspace"
            checked={value === 'workspace'}
            onChange={() => onChange('workspace')}
            disabled={disabled}
          />
          <div>
            <strong>👥 Workspace</strong>
            <p>All workspace members can use this credential</p>
          </div>
        </label>
        
        <label className={value === 'specific' ? 'selected' : ''}>
          <input
            type="radio"
            name="sharing_scope"
            value="specific"
            checked={value === 'specific'}
            onChange={() => onChange('specific')}
            disabled={disabled}
          />
          <div>
            <strong>🔗 Specific People</strong>
            <p>Only people you share with can use this credential</p>
          </div>
        </label>
      </div>
    </div>
  );
}
```

### 4. Share Credential Modal

```tsx
// components/credentials/ShareCredentialModal.tsx

interface ShareCredentialModalProps {
  credential: Credential;
  workspaceId: string;
  workspaceMembers: User[];
  onClose: () => void;
}

export function ShareCredentialModal({ 
  credential, 
  workspaceId, 
  workspaceMembers,
  onClose 
}: ShareCredentialModalProps) {
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const { data: shares } = useCredentialShares(workspaceId, credential.id);
  const shareMutation = useShareCredential(workspaceId);
  const unshareMutation = useUnshareCredential(workspaceId);
  const updateScopeMutation = useUpdateSharingScope(workspaceId);
  
  // Filter out owner and already shared users
  const availableUsers = workspaceMembers.filter(
    u => u.id !== credential.created_by && 
         !shares?.some(s => s.user_id === u.id)
  );
  
  const handleShare = async () => {
    if (selectedUsers.length === 0) return;
    
    await shareMutation.mutateAsync({
      credentialId: credential.id,
      userIds: selectedUsers
    });
    
    setSelectedUsers([]);
  };
  
  const handleUnshare = async (userId: string) => {
    await unshareMutation.mutateAsync({
      credentialId: credential.id,
      userId
    });
  };
  
  const handleScopeChange = async (scope: SharingScope) => {
    await updateScopeMutation.mutateAsync({
      credentialId: credential.id,
      scope
    });
  };
  
  return (
    <Modal onClose={onClose}>
      <h2>Share "{credential.name}"</h2>
      
      {/* Sharing Scope */}
      <SharingScopeSelector
        value={credential.sharing_scope}
        onChange={handleScopeChange}
      />
      
      {/* Only show user selection for 'specific' scope */}
      {credential.sharing_scope === 'specific' && (
        <>
          {/* Current Shares */}
          <div className="current-shares">
            <h3>Shared with</h3>
            {shares?.length === 0 ? (
              <p className="empty">Not shared with anyone yet</p>
            ) : (
              <ul>
                {shares?.map(share => (
                  <li key={share.id}>
                    <div className="user-info">
                      <span>{share.user?.email}</span>
                      <span className="name">
                        {share.user?.first_name} {share.user?.last_name}
                      </span>
                    </div>
                    <button 
                      onClick={() => handleUnshare(share.user_id)}
                      className="remove-btn"
                    >
                      Remove
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
          
          {/* Add Users */}
          <div className="add-users">
            <h3>Add people</h3>
            <select
              multiple
              value={selectedUsers}
              onChange={e => setSelectedUsers(
                Array.from(e.target.selectedOptions, o => o.value)
              )}
            >
              {availableUsers.map(user => (
                <option key={user.id} value={user.id}>
                  {user.email} ({user.first_name} {user.last_name})
                </option>
              ))}
            </select>
            
            <button 
              onClick={handleShare}
              disabled={selectedUsers.length === 0 || shareMutation.isLoading}
            >
              {shareMutation.isLoading ? 'Sharing...' : 'Share'}
            </button>
          </div>
        </>
      )}
      
      <div className="modal-footer">
        <button onClick={onClose}>Done</button>
      </div>
    </Modal>
  );
}
```

### 5. Create Credential Form (updated)

```tsx
// In your create credential form, add sharing scope selection

export function CreateCredentialForm({ workspaceId, onSuccess }: Props) {
  const [formData, setFormData] = useState({
    name: '',
    type: 'api_key',
    data: {},
    description: '',
    sharing_scope: 'workspace' as SharingScope  // Default
  });
  
  const createMutation = useCreateCredential(workspaceId);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createMutation.mutateAsync(formData);
    onSuccess();
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* ... name, type, data fields ... */}
      
      {/* Sharing Scope */}
      <SharingScopeSelector
        value={formData.sharing_scope}
        onChange={(scope) => setFormData(prev => ({ ...prev, sharing_scope: scope }))}
      />
      
      <button type="submit">Create Credential</button>
    </form>
  );
}
```

## Visual Guide

```
┌─────────────────────────────────────────────────────────────────┐
│  Credentials                                        [+ Add New] │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ OpenAI API Key                           🔒 Private     │   │
│  │ api_key • Created by you                                │   │
│  │                                    [Edit] [Delete] [Share]   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Google Sheets                            👥 Workspace   │   │
│  │ oauth2 • Created by alice@team.com                      │   │
│  │                                              [Test]     │   │  <-- No edit/delete (not owner)
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Slack Bot Token                          🔗 Shared      │   │
│  │ api_key • Shared with you by bob@team.com               │   │
│  │                                              [Test]     │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Permission Matrix

| Action | Owner | Shared User | Workspace Member (if scope=workspace) |
|--------|-------|-------------|--------------------------------------|
| View in list | ✅ | ✅ | ✅ |
| Use in workflow | ✅ | ✅ | ✅ |
| Test connection | ✅ | ✅ | ✅ |
| Edit | ✅ | ❌ | ❌ |
| Delete | ✅ | ❌ | ❌ |
| Share | ✅ | ❌ | ❌ |
| Change scope | ✅ | ❌ | ❌ |
| View shares | ✅ | ❌ | ❌ |

## Error Handling

```typescript
try {
  await api.put(`/workspaces/${wsId}/credentials/${credId}`, data);
} catch (error) {
  if (error.response?.status === 403) {
    toast.error("Only the owner can edit this credential");
  } else if (error.response?.status === 404) {
    toast.error("Credential not found");
  } else {
    toast.error("Failed to update credential");
  }
}
```

## Migration Notes

If you have existing frontend code:

1. **List responses now include permission flags** - Use `can_edit`, `can_share` to conditionally show buttons
2. **Create accepts `sharing_scope`** - Optional field, defaults to `workspace`
3. **Actions are filtered by permissions** - The `actions` array in responses only includes permitted actions
4. **Shared credentials appear in list** - Users will see credentials shared with them

## Quick Reference

```typescript
// Check if user can edit
if (credential.can_edit) {
  showEditButton();
}

// Check if user can share
if (credential.can_share) {
  showShareButton();
}

// Check if user is owner
if (credential.is_owner) {
  showOwnerBadge();
}

// Get sharing scope display
const scopeLabel = {
  private: 'Private',
  workspace: 'Workspace', 
  specific: 'Shared'
}[credential.sharing_scope];
```
