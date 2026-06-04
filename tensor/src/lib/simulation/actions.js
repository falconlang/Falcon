export function actionValue(actionState, names) {
  for (const name of names) {
    const value = Number(actionState?.[name] ?? 0);
    if (value > 0) return value;
  }
  return 0;
}

export function runAction(actionTokens, actionState, key, names, fn) {
  const value = actionValue(actionState, names);
  if (value === (actionTokens[key] ?? 0)) return;
  actionTokens[key] = value;
  if (value > 0) fn();
}
