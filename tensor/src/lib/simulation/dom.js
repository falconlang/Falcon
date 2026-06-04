import { tick } from 'svelte';

export async function triggerNativePicker(input) {
  await tick();
  if (!input || input.disabled) return;
  input.focus();
  try {
    if (typeof input.showPicker === 'function') input.showPicker();
    else input.click();
  } catch {}
}

export async function hideKeyboardFor(...candidates) {
  await tick();
  const virtualKeyboard = typeof navigator !== 'undefined' ? navigator.virtualKeyboard : null;
  if (virtualKeyboard && typeof virtualKeyboard.hide === 'function') {
    try {
      virtualKeyboard.hide();
      return true;
    } catch {}
  }
  const active = typeof document !== 'undefined' ? document.activeElement : null;
  const target = candidates.find(Boolean) || active;
  if (target && typeof target.blur === 'function') {
    target.blur();
    return true;
  }
  return false;
}

export async function focusElement(element) {
  await tick();
  if (element && !element.disabled) element.focus();
}

export async function blurElement(element) {
  await tick();
  if (element) element.blur();
}
