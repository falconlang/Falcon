# TinyDB Simulator Review

## Overview

`TinyDB` is a non-visible, persistent key-value storage component in App Inventor. It uses Android's `SharedPreferences` under the hood and stores serialised JSON values keyed by string tags under a named namespace. Multiple `TinyDB` instances sharing the same `Namespace` share the same data store within an app.

In the Tensor simulator, TinyDB is backed by an in-memory `map[string]map[string]runtime.Value` keyed first by namespace, then by tag. The component has no visual representation and is handled entirely in the Go WASM runtime (`simulation_wasm.go`).

---

## Properties Analysis

### Supported

- `Namespace` — declared as a default in `simulation-capabilities.js` (`'TinyDB1'`) matching the AI constant `DEFAULT_NAMESPACE = "TinyDB1"`. The Go backend reads it correctly via `h.GetProperty(componentName, "TinyDB", "Namespace").AsStr()`.

### Missing / Unsupported

- None. `Namespace` is the only designer-settable property on `TinyDB` and it is supported.

### Wrong Defaults or Types

- **[Medium]** `Namespace` is listed in `SIMULATION_VISUAL_PROPS` (line 99 of `simulation-capabilities.js`). `Namespace` is a behaviour property, not a visual/rendering prop. While this doesn't break anything functionally in the current code, it causes `Namespace` to be included in the set of props that trigger visual re-renders in the frontend, which is wasteful and semantically wrong for a non-visible component.

---

## Events Analysis

### Supported

- No events are defined on TinyDB in the App Inventor source — it has no `@SimpleEvent` annotations. The Tensor implementation correctly defines no events for TinyDB.

### Missing / Incorrect

- None identified. TinyDB has no events in the App Inventor specification.

---

## Methods Analysis

### Supported

- `StoreValue(tag, valueToStore)` — implemented in `callTinyDBMethod`. Stores `args[1]` under `args[0].AsStr()` in the namespace-keyed store.
- `GetValue(tag, valueIfTagNotThere)` — implemented. Returns stored value if found, otherwise returns `args[1]` (the default). Falls back to `runtime.NullVal()` if fewer than 2 args are given.
- `ClearTag(tag)` — implemented. Deletes the key from the in-memory store.
- `ClearAll()` — implemented. Replaces the namespace's map with a fresh empty map.
- `GetTags()` — implemented. Collects all keys, sorts them, and returns a `runtime.ListVal` of `runtime.StrVal` items. Sort order matches AI (`java.util.Collections.sort`).

### Missing / Incorrect

- **[High]** `GetEntries()` — present in the AI Java source (`@SimpleFunction`) since at least 2022 but **not implemented** in `callTinyDBMethod`. Calling it from user code will fall through to `h.Unsupported("method", "TinyDB.GetEntries")` and return `VoidVal()`. Code that relies on `GetEntries` will silently receive a void/null instead of a dictionary.

- **[Medium]** `GetValue` argument-order bug risk: the Go implementation checks `len(args) >= 1` to attempt a lookup, then checks `len(args) >= 2` to return the default. If the user calls `GetValue` with only one argument (no default), the function returns `runtime.NullVal()`. This is consistent with the AI behaviour (which would return `valueIfTagNotThere` which is null), but the second guard block is placed **after** a successful-lookup early return, so only the no-hit path can reach it. This is logically correct but fragile to read. Not a bug as written, but worth noting.

- **[Low]** `getDataValue(key)` is an internal interface method in AI used by Chart components as an `ObservableDataSource`. It is not user-callable as a `@SimpleFunction` and does not need to be simulated. Correctly absent.

---

## Behaviour Gaps

### Persistence Across Sessions

- **[Critical]** The Tensor simulator stores TinyDB data in a Go `map` inside the `simulationHost` struct. This map exists only for the lifetime of a single simulation session. When the session is disposed (`disposeSimulationSession`) or a new session is created (`createSimulationSession`), all stored data is lost. The App Inventor specification says TinyDB data **persists between app runs** (it uses `SharedPreferences` on Android). There is no mechanism — `localStorage`, `IndexedDB`, or any other browser-side store — to persist data across simulator restarts. A user testing an app that reads previously saved data on startup will always see the default value, making it impossible to simulate the persistence lifecycle.

### Namespace Scoping

- **[Medium]** Namespace-based data isolation is correctly implemented: `tinyDBStore` reads the `Namespace` property of the component and keys the map accordingly. However, if `Namespace` is changed at runtime (via `SetProperty`), the component will read from a different store without any migration of existing data. The AI runtime also switches `SharedPreferences` files on `Namespace` change, so the absence of the old data is actually AI-compliant — but any data written to the old namespace before the change will accumulate orphaned in `h.tinyDB` for the session lifetime with no way to clean it up.

### JSON Serialisation / Deserialisation

- **[High]** On Android, `StoreValue` serialises the value to JSON via `JsonUtil.getJsonRepresentation()` and `GetValue` deserialises it back via `JsonUtil.getObjectFromJson(value, true)`. This means:
  - Lists are stored as JSON arrays and round-trip correctly.
  - Numbers stored as integers round-trip as numbers.
  - Booleans are stored/retrieved as booleans.
  
  In the Tensor simulator, values are stored as `runtime.Value` objects directly — **no serialisation occurs at all**. This is invisible in simple get/set scenarios, but causes a behavioural difference: on Android, storing a `YailList` and retrieving it returns a new `YailList` constructed from JSON; in the simulator, the exact same `runtime.Value` pointer (or shallow copy) is returned. Any mutation of the returned list on Android would not affect the stored copy (since it was deserialised from JSON), but in the simulator, depending on how `runtime.Value` implements list semantics, this may or may not produce the same isolation.

### `ClearAll` and Observer Notification

- **[Low]** In the AI source, `ClearAll` calls `notifyDataObservers(null, null)` after clearing, and `onDelete()` does the same. These are internal observer callbacks for Chart components and have no user-facing event equivalent. The simulator correctly has no observer machinery. Not a gap that affects normal user code.

### Data Shared Across TinyDB Instances with Same Namespace

- **[Medium]** The AI spec says "there is only one data store per app. Even if you have multiple TinyDB components, they will use the same data store" if they share the same namespace. The Tensor simulator correctly models this: `tinyDBStore` always indexes by namespace string, so two components with `Namespace = "TinyDB1"` will read/write the same in-memory map. This aspect is correctly implemented.

---

## Bugs Found

### Bug 1 — `GetEntries` silently returns void (High)

**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, `callTinyDBMethod`, line 335 (`default` branch).

Any block that calls `TinyDB1.GetEntries` will receive `runtime.VoidVal()` and the call will be logged to the `unsupported` list. If the caller assigns the result to a variable and iterates over it (treating it as a dictionary), a runtime panic or silent empty-iteration will occur instead of the expected dictionary of all entries.

### Bug 2 — `GetValue` returns `NullVal` not the default when tag is missing and only one arg is passed (Low)

**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, lines 307–314.

```go
case "GetValue":
    if len(args) >= 1 {
        if value, ok := store[args[0].AsStr()]; ok {
            return value
        }
    }
    if len(args) >= 2 {
        return args[1]
    }
    return runtime.NullVal()
```

If the user calls `GetValue` with a single argument and the tag is not found, the code returns `runtime.NullVal()`. The AI source signature is `GetValue(tag, valueIfTagNotThere)` — the second parameter is mandatory in block-based AI. However, block-based users sometimes wire it with an empty string or empty list. The Falcon compiler/parser likely enforces arity here, so this path may never be reached in practice — but the fallback is semantically wrong if it is reached (should return empty string `""` per original AI default, not null).

### Bug 3 — `Namespace` in `SIMULATION_VISUAL_PROPS` (Medium)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 99.

`Namespace` appears in `SIMULATION_VISUAL_PROPS`. This set is used by `deriveStateFromDesignerProps` to decide which props to pass through to the simulation state. Because `Namespace` is a non-visible component property that does not drive any rendering, including it in `SIMULATION_VISUAL_PROPS` is a semantic error. It works accidentally because the Go host reads `Namespace` from the state map directly, but any future code that iterates `SIMULATION_VISUAL_PROPS` only for rendered components will incorrectly include `Namespace`.

### Bug 4 — No persistence between simulator restarts (Critical)

**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, `newSimulationHost`, line 67.

`tinyDB` is initialised as a fresh empty map on every `createSimulationSession` call. There is no mechanism to seed or restore it from browser-side persistent storage. Apps that rely on data surviving a restart will always behave as if TinyDB is empty at startup.

---

## Android/App Inventor Standards Compliance

| Requirement | Status |
|---|---|
| Single shared store per namespace per app | Correctly simulated |
| `Namespace` default `"TinyDB1"` | Correct |
| Tags sorted in `GetTags` | Correct (Go `sort.Strings`) |
| JSON round-trip for stored values | Not simulated — raw `runtime.Value` stored |
| Persistence across sessions/restarts | Not simulated — in-memory only |
| `GetEntries` returning a dictionary | Missing |
| `onDelete` clearing the store | Not applicable in simulator context |
| Data observer notifications to Chart components | Not applicable in simulator context |

---

## Summary

TinyDB is the simplest non-visible component in App Inventor and the Tensor implementation covers its core use-case (store, retrieve, clear individual tags, clear all, list tags) correctly within a single simulation session. The namespace scoping is also correctly modelled.

**Top 3 action items:**

1. **Implement `GetEntries`** (High) — This is a `@SimpleFunction` present in the AI source that returns a dictionary of all stored key-value pairs. It is entirely missing from `callTinyDBMethod`. Add a `case "GetEntries":` block that iterates the store, sorts keys, and builds a runtime dictionary/map value.

2. **Persist TinyDB data across simulator sessions** (Critical) — Use `localStorage` or `IndexedDB` to serialise the `tinyDB` map to the browser on every `StoreValue`/`ClearTag`/`ClearAll` call, and deserialise it back in `createSimulationSession`. Without this, apps that test the persistence lifecycle (e.g., save a high score, restart, read it back) will always fail.

3. **Remove `Namespace` from `SIMULATION_VISUAL_PROPS`** (Medium) — `Namespace` is a behaviour property of a non-visible component. It should be handled as a raw state prop (it already is, via the `TinyDB` defaults entry) without being listed alongside visual rendering props. Remove it from the `SIMULATION_VISUAL_PROPS` set and ensure `deriveStateFromDesignerProps` still passes it through via a separate non-visual prop pathway or simply by including it in the defaults map (which already happens).
