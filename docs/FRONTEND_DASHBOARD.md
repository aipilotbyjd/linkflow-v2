# Frontend Dashboard Implementation Guide

This document provides comprehensive specifications for implementing the Dashboard UI in the LinkFlow frontend application.

---

## Table of Contents

1. [API Endpoints](#api-endpoints)
2. [TypeScript Interfaces](#typescript-interfaces)
3. [UI Wireframes](#ui-wireframes)
4. [Component Breakdown](#component-breakdown)
5. [Chart Specifications](#chart-specifications)
6. [Implementation Checklist](#implementation-checklist)

---

## API Endpoints

### 1. Full Dashboard Data

```
GET /api/v1/workspaces/{workspace_id}/dashboard?period=7d
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| period | string | "7d" | Chart data period: "7d", "30d", "90d" |

**Response:**
```json
{
  "data": {
    "summary": {
      "total_workflows": 45,
      "active_workflows": 32,
      "inactive_workflows": 10,
      "draft_workflows": 3,
      "total_executions_today": 156,
      "total_executions_week": 1089,
      "total_executions_month": 4523,
      "success_rate": 94.5,
      "avg_duration_ms": 2340,
      "total_credentials": 12,
      "total_schedules": 8,
      "active_schedules": 6,
      "running_executions": 3,
      "queued_executions": 5
    },
    "recent_executions": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "workflow_id": "660e8400-e29b-41d4-a716-446655440001",
        "workflow_name": "Daily Report Generator",
        "status": "completed",
        "trigger_type": "schedule",
        "duration_ms": 1523,
        "started_at": 1704326400,
        "completed_at": 1704326401,
        "created_at": 1704326400
      }
    ],
    "top_workflows": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "name": "Daily Report Generator",
        "status": "active",
        "execution_count": 245,
        "success_count": 240,
        "failed_count": 5,
        "success_rate": 97.96,
        "avg_duration_ms": 1850,
        "last_executed_at": 1704326400
      }
    ],
    "recent_failures": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "workflow_id": "660e8400-e29b-41d4-a716-446655440003",
        "workflow_name": "Email Sender",
        "error_message": "SMTP connection timeout",
        "error_node_id": "node_5",
        "failed_at": 1704320000
      }
    ],
    "executions_by_day": [
      {
        "date": "2024-01-01",
        "total": 180,
        "success": 172,
        "failed": 8
      }
    ],
    "executions_by_hour": [
      {"hour": 0, "count": 5},
      {"hour": 1, "count": 3},
      {"hour": 2, "count": 2}
    ],
    "upcoming_schedules": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "workflow_id": "660e8400-e29b-41d4-a716-446655440001",
        "workflow_name": "Daily Report Generator",
        "cron_expression": "0 9 * * *",
        "timezone": "America/New_York",
        "next_run_at": 1704369600,
        "is_active": true
      }
    ],
    "executions_by_status": [
      {"status": "completed", "count": 4200},
      {"status": "failed", "count": 200},
      {"status": "running", "count": 15},
      {"status": "queued", "count": 8}
    ],
    "trigger_type_stats": [
      {"trigger_type": "schedule", "count": 2500},
      {"trigger_type": "webhook", "count": 1500},
      {"trigger_type": "manual", "count": 423}
    ]
  },
  "links": {
    "self": "/api/v1/workspaces/123/dashboard"
  }
}
```

### 2. Quick Stats (Lightweight)

```
GET /api/v1/workspaces/{workspace_id}/stats
```

**Use Case:** Sidebar, header widgets, polling for real-time updates

**Response:**
```json
{
  "data": {
    "workflows": {
      "total": 45,
      "active": 32
    },
    "executions": {
      "running": 3,
      "queued": 5,
      "today": 156
    },
    "credentials": {
      "total": 12,
      "expiring_soon": 2
    },
    "schedules": {
      "total": 8,
      "active": 6
    }
  },
  "links": {
    "self": "/api/v1/workspaces/123/stats"
  }
}
```

---

## TypeScript Interfaces

```typescript
// ===========================================
// DASHBOARD TYPES
// ===========================================

export interface DashboardSummary {
  total_workflows: number;
  active_workflows: number;
  inactive_workflows: number;
  draft_workflows: number;
  total_executions_today: number;
  total_executions_week: number;
  total_executions_month: number;
  success_rate: number;
  avg_duration_ms: number;
  total_credentials: number;
  total_schedules: number;
  active_schedules: number;
  running_executions: number;
  queued_executions: number;
}

export interface ExecutionSummary {
  id: string;
  workflow_id: string;
  workflow_name: string;
  status: ExecutionStatus;
  trigger_type: TriggerType;
  duration_ms?: number;
  started_at?: number;
  completed_at?: number;
  created_at: number;
}

export interface WorkflowStats {
  id: string;
  name: string;
  status: WorkflowStatus;
  execution_count: number;
  success_count: number;
  failed_count: number;
  success_rate: number;
  avg_duration_ms: number;
  last_executed_at?: number;
}

export interface FailureSummary {
  id: string;
  workflow_id: string;
  workflow_name: string;
  error_message: string;
  error_node_id?: string;
  failed_at: number;
}

export interface DailyExecutions {
  date: string; // YYYY-MM-DD
  total: number;
  success: number;
  failed: number;
}

export interface HourlyExecutions {
  hour: number; // 0-23
  count: number;
}

export interface ScheduleSummary {
  id: string;
  workflow_id: string;
  workflow_name: string;
  cron_expression: string;
  timezone: string;
  next_run_at?: number;
  is_active: boolean;
}

export interface StatusCount {
  status: ExecutionStatus;
  count: number;
}

export interface TriggerTypeCount {
  trigger_type: TriggerType;
  count: number;
}

export interface DashboardData {
  summary: DashboardSummary;
  recent_executions: ExecutionSummary[];
  top_workflows: WorkflowStats[];
  recent_failures: FailureSummary[];
  executions_by_day: DailyExecutions[];
  executions_by_hour: HourlyExecutions[];
  upcoming_schedules: ScheduleSummary[];
  executions_by_status: StatusCount[];
  trigger_type_stats: TriggerTypeCount[];
}

export interface QuickStats {
  workflows: {
    total: number;
    active: number;
  };
  executions: {
    running: number;
    queued: number;
    today: number;
  };
  credentials: {
    total: number;
    expiring_soon: number;
  };
  schedules: {
    total: number;
    active: number;
  };
}

// Enums
export type ExecutionStatus = 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
export type TriggerType = 'manual' | 'schedule' | 'webhook';
export type WorkflowStatus = 'active' | 'inactive' | 'draft';
export type DashboardPeriod = '7d' | '30d' | '90d';
```

---

## UI Wireframes

### Main Dashboard Layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  LINKFLOW DASHBOARD                                      [Period: 7d ▼] [↻ Refresh] │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                    │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │  WORKFLOWS  │ │ EXECUTIONS  │ │   RUNNING   │ │   SUCCESS   │ │ AVG DURATION │ │
│  │     45      │ │    Today    │ │      3      │ │    RATE     │ │    2.3s      │ │
│  │ ────────────│ │    156      │ │   ⬤ Active  │ │   94.5%     │ │  ↓ 12%       │ │
│  │ 32 Active   │ │  Week: 1089 │ │   5 Queued  │ │  ▲ 2.1%     │ │              │ │
│  │ 10 Inactive │ │ Month: 4523 │ │             │ │             │ │              │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
│                                                                                    │
├───────────────────────────────────────────┬────────────────────────────────────────┤
│  EXECUTION TREND                          │  EXECUTIONS BY STATUS                   │
│  ────────────────────────────             │  ──────────────────────                 │
│                                           │                                          │
│      200 ┤                  ╭─╮           │         ┌──────────────┐                │
│          │        ╭─╮      │ │           │    ┌────┤  Completed   │                │
│      150 ┤   ╭───╯  ╰╮    │ │           │    │    │    94.5%     │                │
│          │  ╭╯        ╰╮  │ │           │    │    └──────────────┘                │
│      100 ┤ ╭╯          ╰──╯ │           │    │    ┌──────────────┐                │
│          │╭╯                 │           │    ├────┤   Failed     │                │
│       50 ┤│                  │           │    │    │     4.5%     │                │
│          ├─────────────────────          │    │    └──────────────┘                │
│          M  T  W  T  F  S  S            │    │    ┌──────────────┐                │
│                                          │    └────┤  Running/Q   │                │
│  [─ Total ── Success ── Failed]          │         │     1.0%     │                │
│                                          │         └──────────────┘                │
├───────────────────────────────────────────┴────────────────────────────────────────┤
│                                                                                    │
│  ┌──────────────────────────────────────┐ ┌──────────────────────────────────────┐│
│  │  TOP WORKFLOWS                       │ │  TRIGGER DISTRIBUTION                ││
│  │  ─────────────                       │ │  ────────────────────                ││
│  │                                      │ │                                      ││
│  │  1. Daily Report Generator           │ │     Schedule ████████████ 57%       ││
│  │     245 runs │ 97.9% │ 1.8s         │ │      Webhook ███████     34%       ││
│  │     ██████████████████░░             │ │       Manual ███         9%        ││
│  │                                      │ │                                      ││
│  │  2. Email Sender                     │ │                                      ││
│  │     189 runs │ 95.2% │ 2.1s         │ │                                      ││
│  │     ██████████████░░░░               │ │                                      ││
│  │                                      │ │                                      ││
│  │  3. Data Sync                        │ │                                      ││
│  │     156 runs │ 98.1% │ 4.5s         │ │                                      ││
│  │     ████████████░░░░░░               │ │                                      ││
│  │                                      │ │                                      ││
│  │  [View All Workflows →]              │ │                                      ││
│  └──────────────────────────────────────┘ └──────────────────────────────────────┘│
│                                                                                    │
├────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                    │
│  ┌────────────────────────────────────────────────────────────────────────────────┐│
│  │  RECENT EXECUTIONS                                               [View All →] ││
│  │  ─────────────────                                                            ││
│  │                                                                                ││
│  │  ┌─────────────────────────────────────────────────────────────────────────┐  ││
│  │  │ ● Daily Report Generator      completed   schedule   1.5s   2 min ago   │  ││
│  │  │ ● Email Sender                completed   webhook    2.1s   5 min ago   │  ││
│  │  │ ● Data Sync                   running     manual     --     8 min ago   │  ││
│  │  │ ○ Slack Notifier              queued      webhook    --     10 min ago  │  ││
│  │  │ ✕ Invoice Generator           failed      schedule   0.8s   15 min ago  │  ││
│  │  │ ● CRM Update                  completed   webhook    3.2s   20 min ago  │  ││
│  │  │ ● Customer Sync               completed   schedule   5.1s   25 min ago  │  ││
│  │  │ ● Report Emailer              completed   manual     1.9s   30 min ago  │  ││
│  │  │ ● Database Backup             completed   schedule   12.3s  35 min ago  │  ││
│  │  │ ● Lead Scorer                 completed   webhook    0.5s   40 min ago  │  ││
│  │  └─────────────────────────────────────────────────────────────────────────┘  ││
│  │                                                                                ││
│  └────────────────────────────────────────────────────────────────────────────────┘│
│                                                                                    │
├───────────────────────────────────────────┬────────────────────────────────────────┤
│  RECENT FAILURES                          │  UPCOMING SCHEDULES                     │
│  ───────────────                          │  ─────────────────                      │
│                                           │                                          │
│  ⚠ Invoice Generator                      │  ⏰ Daily Report Generator              │
│    SMTP connection timeout                │     0 9 * * * (America/New_York)        │
│    Node: email_sender                     │     Next: Today at 9:00 AM              │
│    15 min ago  [View →] [Retry →]        │                                          │
│                                           │  ⏰ Database Backup                      │
│  ⚠ API Sync                               │     0 0 * * * (UTC)                      │
│    Rate limit exceeded                    │     Next: Tomorrow at 12:00 AM          │
│    Node: http_request                     │                                          │
│    2 hours ago  [View →] [Retry →]       │  ⏰ Weekly Report                         │
│                                           │     0 9 * * 1 (America/Los_Angeles)     │
│  ⚠ Webhook Processor                      │     Next: Monday at 9:00 AM             │
│    Invalid JSON payload                   │                                          │
│    Node: json_parser                      │  ⏰ Monthly Cleanup                       │
│    5 hours ago  [View →] [Retry →]       │     0 0 1 * * (UTC)                       │
│                                           │     Next: Feb 1 at 12:00 AM             │
│  [View All Failures →]                    │                                          │
│                                           │  [View All Schedules →]                  │
└───────────────────────────────────────────┴────────────────────────────────────────┘
```

### Hourly Activity Chart (24-Hour View)

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│  HOURLY ACTIVITY (Last 24 Hours)                                                   │
│  ───────────────────────────────                                                   │
│                                                                                    │
│   50 ┤                                    ██                                       │
│      │                                    ██ ██                                    │
│   40 ┤                                 ██ ██ ██ ██                                 │
│      │                              ██ ██ ██ ██ ██                                 │
│   30 ┤                           ██ ██ ██ ██ ██ ██ ██                              │
│      │                        ██ ██ ██ ██ ██ ██ ██ ██ ██                           │
│   20 ┤                     ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██                        │
│      │               ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██                     │
│   10 ┤         ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██              │
│      │   ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ ██ │
│    0 ├─────────────────────────────────────────────────────────────────────────────│
│      0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23       │
│                                     Hour of Day                                    │
│                                                                                    │
│  Peak: 45 executions at 2:00 PM │ Low: 2 executions at 3:00 AM                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Sidebar Quick Stats Widget

```
┌─────────────────────┐
│  QUICK STATS        │
│  ───────────        │
│                     │
│  📊 Workflows       │
│     45 total        │
│     32 active       │
│                     │
│  ⚡ Executions      │
│     156 today       │
│     3 running       │
│     5 queued        │
│                     │
│  🔑 Credentials     │
│     12 total        │
│     ⚠ 2 expiring    │
│                     │
│  ⏰ Schedules       │
│     8 total         │
│     6 active        │
│                     │
└─────────────────────┘
```

### Mobile Dashboard Layout

```
┌──────────────────────────────┐
│  DASHBOARD        [☰] [↻]   │
├──────────────────────────────┤
│                              │
│  ┌────────────┐┌────────────┐│
│  │ WORKFLOWS  ││ EXECUTIONS ││
│  │    45      ││   Today    ││
│  │ 32 Active  ││    156     ││
│  └────────────┘└────────────┘│
│                              │
│  ┌────────────┐┌────────────┐│
│  │  SUCCESS   ││  RUNNING   ││
│  │   94.5%    ││     3      ││
│  │  ▲ 2.1%    ││  5 queued  ││
│  └────────────┘└────────────┘│
│                              │
│  EXECUTION TREND             │
│  ──────────────              │
│  ┌──────────────────────────┐│
│  │    ╭─╮     ╭──╮          ││
│  │ ╭──╯ ╰╮   ╭╯  ╰╮         ││
│  │─╯      ╰──╯    ╰─        ││
│  │ M T W T F S S            ││
│  └──────────────────────────┘│
│                              │
│  RECENT EXECUTIONS           │
│  ─────────────────           │
│  ┌──────────────────────────┐│
│  │● Report Gen   ✓   1.5s   ││
│  │● Email Send   ✓   2.1s   ││
│  │● Data Sync    ⟳   --     ││
│  │○ Notifier     ◌   --     ││
│  │✕ Invoice      ✗   0.8s   ││
│  └──────────────────────────┘│
│  [View All →]                │
│                              │
│  RECENT FAILURES             │
│  ───────────────             │
│  ┌──────────────────────────┐│
│  │⚠ Invoice Generator       ││
│  │  SMTP timeout            ││
│  │  15 min ago              ││
│  │  [Retry] [View]          ││
│  └──────────────────────────┘│
│                              │
└──────────────────────────────┘
```

---

## Component Breakdown

### Required Components

| Component | Description | Props |
|-----------|-------------|-------|
| `DashboardPage` | Main dashboard container | workspaceId |
| `SummaryCards` | Top row stat cards | summary: DashboardSummary |
| `ExecutionTrendChart` | Line/area chart for daily executions | data: DailyExecutions[], period |
| `StatusDistributionChart` | Pie/donut chart for status | data: StatusCount[] |
| `TopWorkflowsList` | Ranked workflow performance | workflows: WorkflowStats[] |
| `TriggerTypeChart` | Horizontal bar chart | data: TriggerTypeCount[] |
| `RecentExecutionsTable` | Scrollable execution list | executions: ExecutionSummary[] |
| `RecentFailuresList` | Failed executions with actions | failures: FailureSummary[] |
| `UpcomingSchedulesList` | Next scheduled runs | schedules: ScheduleSummary[] |
| `HourlyActivityChart` | 24-hour bar chart | data: HourlyExecutions[] |
| `QuickStatsWidget` | Compact sidebar widget | stats: QuickStats |
| `PeriodSelector` | Dropdown for period selection | value, onChange |
| `RefreshButton` | Manual refresh with loading state | onRefresh, isLoading |

### Component Details

#### 1. SummaryCards

```tsx
interface SummaryCardsProps {
  summary: DashboardSummary;
  previousSummary?: DashboardSummary; // For comparison/trends
  isLoading?: boolean;
}

// Cards to display:
// 1. Total Workflows (with active/inactive breakdown)
// 2. Executions (today/week/month toggle)
// 3. Running Now (with queued count)
// 4. Success Rate (with trend indicator)
// 5. Avg Duration (with trend indicator)
```

#### 2. ExecutionTrendChart

```tsx
interface ExecutionTrendChartProps {
  data: DailyExecutions[];
  period: DashboardPeriod;
  height?: number;
}

// Chart type: Multi-line or stacked area
// Lines: Total (primary), Success (green), Failed (red)
// X-axis: Dates
// Y-axis: Count
// Tooltip: Date, Total, Success, Failed, Success Rate
```

#### 3. StatusDistributionChart

```tsx
interface StatusDistributionChartProps {
  data: StatusCount[];
  showLegend?: boolean;
}

// Chart type: Donut/Pie
// Colors:
//   - completed: #10B981 (green)
//   - failed: #EF4444 (red)
//   - running: #3B82F6 (blue)
//   - queued: #F59E0B (amber)
//   - cancelled: #6B7280 (gray)
//   - paused: #8B5CF6 (purple)
```

#### 4. TopWorkflowsList

```tsx
interface TopWorkflowsListProps {
  workflows: WorkflowStats[];
  maxItems?: number;
  onWorkflowClick?: (id: string) => void;
}

// Display:
// - Rank number
// - Workflow name (link)
// - Execution count
// - Success rate (with color indicator)
// - Avg duration
// - Progress bar showing relative execution count
```

#### 5. RecentExecutionsTable

```tsx
interface RecentExecutionsTableProps {
  executions: ExecutionSummary[];
  onExecutionClick?: (id: string) => void;
  onWorkflowClick?: (id: string) => void;
}

// Columns:
// - Status icon
// - Workflow name (link)
// - Status badge
// - Trigger type icon
// - Duration
// - Time ago
```

#### 6. RecentFailuresList

```tsx
interface RecentFailuresListProps {
  failures: FailureSummary[];
  onRetry?: (id: string) => void;
  onView?: (id: string) => void;
}

// Display:
// - Workflow name
// - Error message (truncated)
// - Error node (if available)
// - Time ago
// - Action buttons: View, Retry
```

#### 7. UpcomingSchedulesList

```tsx
interface UpcomingSchedulesListProps {
  schedules: ScheduleSummary[];
  onScheduleClick?: (id: string) => void;
  userTimezone?: string;
}

// Display:
// - Workflow name
// - Cron expression (human readable)
// - Timezone
// - Next run time (relative and absolute)
// - Active status indicator
```

---

## Chart Specifications

### Color Palette

```typescript
export const DASHBOARD_COLORS = {
  // Status colors
  completed: '#10B981',  // Emerald 500
  failed: '#EF4444',     // Red 500
  running: '#3B82F6',    // Blue 500
  queued: '#F59E0B',     // Amber 500
  cancelled: '#6B7280',  // Gray 500
  paused: '#8B5CF6',     // Purple 500
  
  // Trend colors
  success: '#10B981',
  error: '#EF4444',
  total: '#6366F1',      // Indigo 500
  
  // Trigger type colors
  schedule: '#8B5CF6',   // Purple 500
  webhook: '#F59E0B',    // Amber 500
  manual: '#6366F1',     // Indigo 500
  
  // Chart backgrounds
  chartBg: '#F9FAFB',
  gridLines: '#E5E7EB',
};
```

### Chart Libraries Recommendation

- **Recharts** - React-native, good for simple charts
- **Chart.js + react-chartjs-2** - More customization
- **Tremor** - Pre-built dashboard components
- **ApexCharts** - Interactive charts

### Responsive Breakpoints

```typescript
export const DASHBOARD_BREAKPOINTS = {
  mobile: 640,   // sm
  tablet: 768,   // md
  desktop: 1024, // lg
  wide: 1280,    // xl
};

// Grid layout:
// Mobile: 1 column
// Tablet: 2 columns
// Desktop: 3-4 columns
```

---

## Implementation Checklist

### Phase 1: Core Dashboard
- [ ] Create DashboardPage component
- [ ] Implement API hooks (useDashboard, useQuickStats)
- [ ] Build SummaryCards component
- [ ] Add PeriodSelector
- [ ] Add RefreshButton with loading state

### Phase 2: Charts
- [ ] ExecutionTrendChart (daily line/area)
- [ ] StatusDistributionChart (pie/donut)
- [ ] HourlyActivityChart (bar)
- [ ] TriggerTypeChart (horizontal bar)

### Phase 3: Lists & Tables
- [ ] RecentExecutionsTable
- [ ] TopWorkflowsList
- [ ] RecentFailuresList
- [ ] UpcomingSchedulesList

### Phase 4: Interactivity
- [ ] Click handlers for navigation
- [ ] Retry failed execution
- [ ] Polling for real-time updates (optional)
- [ ] Period change handling

### Phase 5: Polish
- [ ] Loading skeletons
- [ ] Empty states
- [ ] Error states
- [ ] Mobile responsive layout
- [ ] Dark mode support

### Phase 6: Sidebar Widget
- [ ] QuickStatsWidget component
- [ ] Integrate into sidebar/header
- [ ] Auto-refresh every 30 seconds

---

## API Hook Examples

```typescript
// useDashboard.ts
import { useQuery } from '@tanstack/react-query';

export function useDashboard(workspaceId: string, period: DashboardPeriod = '7d') {
  return useQuery({
    queryKey: ['dashboard', workspaceId, period],
    queryFn: () => fetchDashboard(workspaceId, period),
    refetchInterval: 60000, // Refresh every minute
    staleTime: 30000,
  });
}

export function useQuickStats(workspaceId: string) {
  return useQuery({
    queryKey: ['quickStats', workspaceId],
    queryFn: () => fetchQuickStats(workspaceId),
    refetchInterval: 30000, // Refresh every 30 seconds
    staleTime: 15000,
  });
}

// API functions
async function fetchDashboard(workspaceId: string, period: DashboardPeriod): Promise<DashboardData> {
  const response = await api.get(`/workspaces/${workspaceId}/dashboard`, {
    params: { period }
  });
  return response.data.data;
}

async function fetchQuickStats(workspaceId: string): Promise<QuickStats> {
  const response = await api.get(`/workspaces/${workspaceId}/stats`);
  return response.data.data;
}
```

---

## Utility Functions

```typescript
// Format duration
export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  if (ms < 3600000) return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
  return `${Math.floor(ms / 3600000)}h ${Math.floor((ms % 3600000) / 60000)}m`;
}

// Format relative time
export function formatTimeAgo(timestamp: number): string {
  const seconds = Math.floor((Date.now() / 1000) - timestamp);
  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} hours ago`;
  return `${Math.floor(seconds / 86400)} days ago`;
}

// Format cron to human readable
export function formatCron(cron: string): string {
  // Use cronstrue library or similar
  // "0 9 * * *" -> "At 9:00 AM"
  // "0 0 * * 0" -> "At 12:00 AM, only on Sunday"
}

// Get status color
export function getStatusColor(status: ExecutionStatus): string {
  return DASHBOARD_COLORS[status] || DASHBOARD_COLORS.cancelled;
}

// Get trigger icon
export function getTriggerIcon(trigger: TriggerType): React.ReactNode {
  switch (trigger) {
    case 'schedule': return <ClockIcon />;
    case 'webhook': return <WebhookIcon />;
    case 'manual': return <PlayIcon />;
  }
}
```

---

## Notes for Frontend Team

1. **Polling Strategy**: Use React Query's `refetchInterval` for automatic updates. Full dashboard can poll every 60s, quick stats every 30s.

2. **Loading States**: Always show skeleton loaders during initial load. Don't block the entire page.

3. **Error Handling**: Show toast notifications for API errors. Keep stale data visible while retrying.

4. **Performance**: The dashboard endpoint aggregates many queries. Consider lazy loading charts below the fold.

5. **Timezone**: All timestamps are Unix seconds. Convert to user's timezone for display.

6. **Deep Linking**: Make workflow names, execution IDs clickable to navigate to detail pages.

7. **Empty States**: Design nice empty states for new workspaces with no data yet.

8. **Accessibility**: Ensure charts have proper ARIA labels and keyboard navigation.
