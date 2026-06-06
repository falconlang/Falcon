import { writable, derived } from 'svelte/store';
import { defaultProjectProperties } from '../project-properties.js';

export const initialDesignCode = `Screen.Screen1 { Title: "Calculator",
  Label { Text: "First number: " }
  TextBox.firstNumberTextBox { NumbersOnly: true, Hint: "Enter first number" }

  Label { Text: "Second number: " }
  TextBox.secondNumberTextBox { NumbersOnly: true, Hint: "Enter second number" }

  HorizontalArrangement {
    Button.AddButton { Text: "+" }
    Button.SubtractButton { Text: "-" }
    Button.MultiplyButton { Text: "*" }
    Button.DivideButton { Text: "/" }
  }

  Notifier.Notifier1
}`;

export const initialCells = [
  {
    id: 'c1', type: 'code', execCount: 1,
    code: `func checkTextBoxes() = {
  if (!(firstNumberTextBox.Text ? number) || !(secondNumberTextBox.Text ? number)) {
    Notifier1.ShowAlert("Please enter numeric values in the textbox!")
    yield false
  }
  true
}`,
  },
  {
    id: 'c2', type: 'code', execCount: 2,
    code: `when AddButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text + secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'c3', type: 'code', execCount: 3,
    code: `when SubtractButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text - secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'c4', type: 'code', execCount: 4,
    code: `when MultiplyButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text * secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'c5', type: 'code', execCount: 5,
    code: `when DivideButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text / secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
];

// ── Notebook content ──
export const cells = writable(initialCells);
export const designCode = writable(initialDesignCode);
export const designAssets = writable([]);
export const projectExtensionComponents = writable([]);
export const projectName = writable('falcon_tour');
export const projectProperties = writable(defaultProjectProperties());
export const activeCellId = writable('c1');
export const execCounter = writable(6);
export const lastRunAt = writable(null);

// ── UI state ──
export const ctxMenu = writable({ show: false, x: 0, y: 0, cellId: null });
export const liveTestOpen = writable(false);
export const companionCommand = writable(null);
export const liveTestState = writable({
  status: 'idle',
  code: null,
  error: null,
  messageCount: 0,
});
export const doItCellId = writable(null);
export const doItResults = writable({});
export const unifiedSelectionActive = writable(false);
export const blocklyPreviewRequest = writable(null);
export const sidebarVisible = writable(typeof window !== 'undefined' ? window.innerWidth > 1024 : true);
export const debugCollapsed = writable(true);
export const notebookMode = writable('cells'); // 'cells' | 'unified'
export const debugOpenHeight = writable(200);
export const copiedCellAvailable = writable(false);
export const sourceNavigationHighlight = writable(null);

// ── Universal search ──
export const searchOpen = writable(false);
export const unifiedSearchSource = writable('');
export const designerSearchIndex = writable([]);
export const designerTreeActive = writable(false);
export const searchNavigation = writable(null);

// ── Deleted-cell undo/redo history (per screen) ──
export const deletedCellUndoStack = writable([]);
export const deletedCellRedoStack = writable([]);
export const canUndoDeletedCell = derived(deletedCellUndoStack, s => s.length > 0);
export const canRedoDeletedCell = derived(deletedCellRedoStack, s => s.length > 0);

// ── Screen management ──
export const screenList = writable(['Screen1']);
export const activeScreen = writable('Screen1');
export const rawBlocklyXml = writable('');
export const sourceScm = writable('');
export const sourceDesignCode = writable('');
export const sourceScmUpgradeWarnings = writable([]);

// ── Debug state ──
export const debugLogs = writable([]);
export const debugModeEnabled = writable(false);
export const debugUserEnabled = writable(false);
export const runtimeErrorNotice = writable({ show: false, error: null });
export const debugExecutionState = writable({
  status: 'idle',
  sessionId: null,
  startedAt: null,
  pausedAt: null,
  hitId: null,
});
export const debugLineMap = writable([]);
export const debugActiveLocation = writable(null);
export const debugRuntimeErrors = writable({});
export const debugAnnotationActive = writable(false);
export const debugBreakpoints = writable({});
export const debugPausedFrame = writable(null);
export const debugExpressionCatalog = writable([]);
export const debugExpressionValues = writable({});
