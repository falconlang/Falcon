import { writable, get } from 'svelte/store';

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

const initialCells = [
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

export const cells = writable(initialCells);
export const designCode = writable(initialDesignCode);
export const activeCellId = writable('c1');
export const execCounter = writable(6);
export const ctxMenu = writable({ show: false, x: 0, y: 0, cellId: null });
export const liveTestOpen = writable(false);
export const liveTestState = writable({
  status: 'idle',
  code: null,
  error: null,
  messageCount: 0,
});
export const doItCellId = writable(null);
export const sidebarVisible = writable(true);
export const debugCollapsed = writable(false);
export const debugOpenHeight = writable(200);

export function setActive(id) {
  activeCellId.set(id);
}

export function addCodeCell() {
  const id = 'c' + Date.now();
  const activeId = get(activeCellId);
  const currentCells = get(cells);
  const idx = activeId ? currentCells.findIndex(c => c.id === activeId) + 1 : currentCells.length;
  cells.update(cs => {
    const next = [...cs];
    next.splice(idx, 0, { id, type: 'code', code: '', execCount: null });
    return next;
  });
  setActive(id);
  return id;
}

export function addMarkdownCell() {
  const id = 'c' + Date.now();
  const activeId = get(activeCellId);
  const currentCells = get(cells);
  const idx = activeId ? currentCells.findIndex(c => c.id === activeId) + 1 : currentCells.length;
  cells.update(cs => {
    const next = [...cs];
    next.splice(idx, 0, { id, type: 'markdown', content: '<div class="md-p">New text cell</div>' });
    return next;
  });
  setActive(id);
  return id;
}

export function moveCellById(id, dir) {
  cells.update(cs => {
    const idx = cs.findIndex(c => c.id === id);
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= cs.length) return cs;
    const next = [...cs];
    [next[idx], next[newIdx]] = [next[newIdx], next[idx]];
    return next;
  });
  setActive(id);
}

export function deleteCellById(id) {
  const currentCells = get(cells);
  if (currentCells.length <= 1) return;
  const idx = currentCells.findIndex(c => c.id === id);
  cells.update(cs => cs.filter(c => c.id !== id));
  const nextCells = get(cells);
  const nextActive = nextCells[Math.min(idx, nextCells.length - 1)]?.id || null;
  if (nextActive) setActive(nextActive);
}

export function updateCellExecCount(id) {
  const count = get(execCounter);
  cells.update(cs => cs.map(c => c.id === id ? { ...c, execCount: count } : c));
  execCounter.update(n => n + 1);
}

export function updateCellCode(id, code) {
  cells.update(cs => cs.map(c => c.id === id ? { ...c, code } : c));
}

export function updateDesignCode(code) {
  designCode.set(code);
}

export function getFalconSource() {
  return get(cells)
    .filter(cell => cell.type === 'code')
    .map(cell => cell.code || '')
    .join('\n\n');
}

export function getDesignSource() {
  return get(designCode);
}

export function showCtx(e, id) {
  e.preventDefault();
  e.stopPropagation();
  setActive(id);
  ctxMenu.set({
    show: true,
    x: Math.min(e.clientX, window.innerWidth - 180),
    y: Math.min(e.clientY, window.innerHeight - 260),
    cellId: id,
  });
}

export function hideCtx() {
  ctxMenu.update(m => ({ ...m, show: false, cellId: null }));
}
