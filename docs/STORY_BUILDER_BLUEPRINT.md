# Advanced Automation Story Builder Blueprint

This blueprint describes the architecture, UI layout, and logic for the "Story Builder" interface—a document-centric abstraction layer over the LinkFlow graph engine.

## 1. Core Philosophy: "Logic as a Document"

Instead of a node-based graph editor, we present the workflow as a **living document**.
*   **Vertical Flow:** Execution always moves down.
*   **Containers, Not Wires:** Loops and Logic are "boxes" that contain other steps.
*   **Natural Language:** Technical steps are summarized in human-readable text.

---

## 2. The UI Layout (The 4-Zone System)

### 📐 Zone A: The Story Stream (Center - 60% Width)
The main "Document" view. A scrollable vertical feed of the automation.
*   **Visual Style:** Resembles a Notion doc or messaging thread.
*   **The "Thread":** A vertical line connects cards, pulsing during execution.
*   **Interaction:** Click to edit, hover between cards to insert.

### 🪄 Zone B: The Magic Command (Bottom Floating - 100% Width)
A floating "Spotlight" input bar.
*   **Function:** AI text-to-workflow generation.
*   **Context Aware:** Applies changes to selected cards or inserts new logic based on natural language prompts.

### ⚙️ Zone C: The Inspector Panel (Right - 25% Width)
A slide-over configuration panel for the selected card.
*   **Header:** Rename step, add description.
*   **Config Form:** Inputs for node parameters (URL, Body, etc.).
*   **Test Tab:** Dedicated runner for individual steps.
*   **AI Assistant:** Debugging helper for failed tests.

### 📦 Zone D: The Data Drawer (Left - 15% Width - Collapsible)
A drawer showing available data variables from previous steps.
*   **Function:** Drag-and-drop data pills into Zone C inputs.
*   **Visuals:** Hierarchical list of variables (e.g., `Webhook > Body > Email`).

---

## 3. The "Story Card" Component System

### 1. The Trigger Card (The Headline)
*   **Role:** The entry point.
*   **UI:** Large icon, bold text.
*   **Features:** Webhook URL copy, "Listen for Event" button.

### 2. The Action Card (The Standard Step)
*   **Role:** Performs a task (HTTP Request, Database Op).
*   **UI:** Clean card with colored status border.
*   **Content:** Title and dynamic summary (e.g., "To: {{email}}").

### 3. The Logic Container (The Branch)
*   **Role:** Handles `IF/ELSE` logic without wires.
*   **UI:** A "Split Container" with two parallel columns.
    *   **Left (True):** Green tint.
    *   **Right (False):** Red/Grey tint.
*   **UX:** Action cards are dropped *inside* the columns. Columns visually merge at the bottom.

### 4. The Loop Container (The Iterator)
*   **Role:** Handles `For Each` logic.
*   **UI:** A "Sandwich" container wrapping inner steps.
    *   **Header:** "For every item in [List]..."
    *   **Body:** Indented area for repeated steps.
    *   **Footer:** "End Loop".

---

## 4. Advanced Features

### 🧠 Smart Data Mapping
*   **Feature:** Type `{` in any input to open the Data Picker.
*   **Visual:** Variables appear as "Chips" or "Tags" in text fields.

### ⚡ Live "Step-by-Step" Debugging
*   **Feature:** "Play" button on individual cards.
*   **Logic:** Uses cached **Pinned Data** from previous steps to execute the selected step in isolation.

### 🛡️ "Rescue" Blocks (Error Handling)
*   **Feature:** Detour logic for failed steps.
*   **Visual:** A branch attached to the right side of an Action Card.
*   **Logic:** "If failed, run these steps, then resume/stop."

### ⏪ Time Travel
*   **Feature:** Version history slider.
*   **Visual:** UI animates to reflect the workflow state at the selected timestamp.

---

## 5. Frontend Data Structure

The frontend maintains a simplified "Story" state that maps to the backend "Graph".

```typescript
interface StoryState {
  meta: {
    id: string;
    name: string;
    status: 'draft' | 'active';
  };
  
  // Linear representation
  blocks: StoryBlock[];
  
  // Execution cache
  executionData: Record<string, any>;
}

type StoryBlock = {
  id: string;
  type: 'trigger' | 'action' | 'logic_container' | 'loop_container';
  
  // Visuals
  title: string;
  icon: string;
  summary: string;
  
  // Configuration (matches Backend Node)
  nodeType: string;
  parameters: Record<string, any>;
  
  // Nesting for Containers
  branches?: {
    true: StoryBlock[];
    false: StoryBlock[];
  };
  loopBody?: StoryBlock[];
  onError?: StoryBlock[];
};
```

## 6. Adapter Logic

### `Graph -> Story`
Traverses the backend graph to identify patterns (branches, loops) and nests them into the `StoryBlock` container structure for the UI.

### `Story -> Graph`
Flattens the nested `StoryBlock` structure back into a standard list of `Nodes` and `Connections` for the backend API.

---

## 7. Interaction Flow Example

1.  **User:** Types "Check if order > $100 then email me" in Magic Bar.
2.  **AI:** Generates JSON for Trigger + Logic Container + Action.
3.  **UI:** Renders Trigger, then a Split Container (True/False).
4.  **User:** Clicks the Logic Container. Inspector opens.
5.  **User:** Drags `total_price` variable from Data Drawer into the Condition field.
6.  **User:** Clicks "Run Test" on the container. True path lights up green.
