export function forwardSvelteEvent(dispatch, event) {
  dispatch(event.type, event.detail);
}

export function emitInteraction(dispatch, properties = [], event = null) {
  dispatch('interaction', { properties, event });
}

export async function emitEvent(dispatch, eventRunner, component, event, args = []) {
  const detail = { component, event, args };
  if (eventRunner) {
    await eventRunner(detail);
  } else {
    dispatch('event', detail);
  }
}
