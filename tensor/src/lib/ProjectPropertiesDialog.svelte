<script>
  import { onMount } from 'svelte';
  import { projectProperties, setProjectProperty, designAssets } from './stores.js';
  import { projectPropertiesByCategory } from './project-properties.js';

  export let open = false;

  const PROPERTY_LABELS = {
    AppName: 'App Name',
    Icon: 'Icon',
    DefaultFileScope: 'Default File Scope',
    ShowListsAsJson: 'Show Lists as JSON',
    Sizing: 'Sizing',
    TutorialURL: 'Tutorial URL',
    PrimaryColor: 'Primary Color',
    PrimaryColorDark: 'Primary Color (Dark)',
    AccentColor: 'Accent Color',
    Theme: 'Theme',
    VersionCode: 'Version Code',
    VersionName: 'Version Name',
    NSBluetoothAlwaysUsageDescription: 'Bluetooth Always',
    NSBluetoothPeripheralUsageDescription: 'Bluetooth Peripheral',
    NSContactsUsageDescription: 'Contacts Access',
    NSMicrophoneUsageDescription: 'Microphone',
    NSCameraUsageDescription: 'Camera',
    NSSpeechRecognitionUsageDescription: 'Speech Recognition',
    NSLocationWhenInUseUsageDescription: 'Location (When In Use)',
  };

  const THEME_OPTIONS = [
    { value: 'Classic', label: 'Classic' },
    { value: 'AppTheme.Light.DarkActionBar', label: 'Device Default' },
    { value: 'AppTheme.Light', label: 'Black Title Text' },
    { value: 'AppTheme', label: 'Dark' },
  ];

  const categories = projectPropertiesByCategory({ includeHidden: false });
  let activeCategory = categories[0]?.id ?? '';
  let localText = {};

  // ── Custom dropdown state ──
  let openDropdown = null; // prop.name of the currently open dropdown, or null
  let dropdownX = 0;
  let dropdownY = 0;
  let dropdownMinWidth = 120;
  const triggerEls = {}; // { [propName]: HTMLElement }

  // Svelte action to register a trigger element by prop name
  function setTrigger(node, name) {
    triggerEls[name] = node;
    return { destroy() { delete triggerEls[name]; } };
  }

  function openDropdownFor(e, name) {
    e.stopPropagation();
    if (openDropdown === name) { openDropdown = null; return; }
    const rect = triggerEls[name]?.getBoundingClientRect();
    if (rect) {
      dropdownX = rect.left;
      dropdownY = rect.bottom + 4;
      dropdownMinWidth = rect.width;
    }
    openDropdown = name;
  }

  function closeDropdown() { openDropdown = null; }

  function selectOption(name, value) {
    setProjectProperty(name, value);
    openDropdown = null;
  }

  function getSelectLabel(prop) {
    const val = $projectProperties[prop.name] ?? prop.defaultValue;
    if (prop.editorType === 'theme') {
      return THEME_OPTIONS.find(o => o.value === val)?.label ?? val;
    }
    return val;
  }

  function getSelectOptions(prop) {
    if (prop.editorType === 'theme') return THEME_OPTIONS;
    return (prop.options ?? []).map(v => ({ value: v, label: v }));
  }

  // The definition of whichever dropdown is currently open
  $: allProps = categories.flatMap(c => c.properties);
  $: openPropDef = openDropdown ? allProps.find(p => p.name === openDropdown) ?? null : null;
  $: openPropOptions = openPropDef ? getSelectOptions(openPropDef) : [];
  $: openPropCurrentValue = openPropDef
    ? ($projectProperties[openPropDef.name] ?? openPropDef.defaultValue)
    : null;

  $: assetNames = $designAssets.map(a => typeof a === 'string' ? a : a.name);
  $: activeProps = categories.find(c => c.id === activeCategory)?.properties ?? [];

  function propLabel(name) { return PROPERTY_LABELS[name] ?? name; }

  function close() {
    openDropdown = null;
    open = false;
  }

  function handleOverlayClick() {
    if (openDropdown) { openDropdown = null; return; }
    close();
  }

  function handleDialogKey(e) {
    if (open && e.key === 'Escape') {
      e.stopPropagation();
      if (openDropdown) { openDropdown = null; return; }
      close();
    }
  }

  function handleDropdownMenuKey(e) {
    if (e.key === 'Escape') { e.stopPropagation(); closeDropdown(); }
  }

  onMount(() => {
    function onDocClick() { if (openDropdown) openDropdown = null; }
    document.addEventListener('click', onDocClick);
    return () => document.removeEventListener('click', onDocClick);
  });

  // ── Local text state for in-flight editing ──
  function textValue(name) {
    return name in localText ? localText[name] : String($projectProperties[name] ?? '');
  }

  function onTextInput(name, e) {
    localText = { ...localText, [name]: e.target.value };
  }

  function onTextBlur(name) {
    if (name in localText) {
      setProjectProperty(name, localText[name]);
      const next = { ...localText };
      delete next[name];
      localText = next;
    }
  }

  function onTextEnter(e) {
    if (e.key === 'Enter') e.currentTarget.blur();
  }

  // ── Boolean toggle ──
  function boolVal(name) {
    return String($projectProperties[name] ?? '').toLowerCase() === 'true';
  }

  function onToggle(name, e) {
    setProjectProperty(name, e.target.checked ? 'True' : 'False');
  }

  // ── Color ──
  function aiaToHex(v) {
    const m = String(v || '').toUpperCase().match(/^&H[0-9A-F]{2}([0-9A-F]{6})$/);
    if (m) return '#' + m[1];
    const h = String(v || '').replace(/^#/, '');
    if (/^[0-9A-Fa-f]{6}$/.test(h)) return '#' + h.toUpperCase();
    return '#000000';
  }

  function onColorPicker(name, e) {
    const aia = '&HFF' + e.target.value.replace('#', '').toUpperCase();
    setProjectProperty(name, aia);
    const next = { ...localText };
    delete next[name];
    localText = next;
  }

  // ── Non-negative integer validation ──
  function isValidNonNegInt(v) {
    return /^\d+$/.test(String(v ?? '').trim());
  }
</script>

<svelte:window on:keydown={handleDialogKey} />

{#if open}
<div class="pp-overlay">
  <button type="button" class="pp-backdrop" aria-label="Close project properties" on:click={handleOverlayClick}></button>
  <div
    class="pp-card"
    role="dialog"
    aria-modal="true"
    aria-labelledby="pp-dialog-title"
  >

    <!-- Header -->
    <div class="pp-header">
      <div class="pp-header-icon">
        <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.4">
          <circle cx="4.5" cy="4.5" r="1.3"/>
          <path d="M1 4.5h1.8M5.8 4.5H13" stroke-linecap="round"/>
          <circle cx="9.5" cy="9.5" r="1.3"/>
          <path d="M1 9.5h6.8M10.8 9.5H13" stroke-linecap="round"/>
        </svg>
      </div>
      <span class="pp-title" id="pp-dialog-title">Project Properties</span>
      <button class="sd-close" on:click={close} title="Close">
        <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
          <path d="M2 2l6 6M8 2l-6 6"/>
        </svg>
      </button>
    </div>

    <!-- Category tabs -->
    <div class="pp-tabs" role="tablist">
      {#each categories as cat}
        <button
          class="pp-tab"
          class:active={activeCategory === cat.id}
          role="tab"
          aria-selected={activeCategory === cat.id}
          on:click={() => { closeDropdown(); activeCategory = cat.id; }}
        >{cat.label}</button>
      {/each}
    </div>

    <!-- Property rows -->
    <div class="pp-body">
      {#each activeProps as prop (prop.name)}
        <div class="pp-row">
          <div class="pp-row-meta">
            <span class="pp-prop-name">{propLabel(prop.name)}</span>
            <span class="pp-prop-desc">{prop.description}</span>
          </div>

          <div class="pp-row-control">

            {#if prop.editorType === 'string'}
              <input
                class="pp-input"
                type="text"
                value={textValue(prop.name)}
                on:input={e => onTextInput(prop.name, e)}
                on:blur={() => onTextBlur(prop.name)}
                on:keydown={onTextEnter}
                placeholder={prop.defaultValue || ''}
                spellcheck="false"
              />

            {:else if prop.editorType === 'non_negative_integer'}
              <input
                class="pp-input pp-input--mono"
                class:pp-input--invalid={!isValidNonNegInt(textValue(prop.name))}
                type="text"
                inputmode="numeric"
                value={textValue(prop.name)}
                on:input={e => onTextInput(prop.name, e)}
                on:blur={() => onTextBlur(prop.name)}
                on:keydown={onTextEnter}
                placeholder={prop.defaultValue || '0'}
              />
              {#if !isValidNonNegInt(textValue(prop.name))}
                <span class="pp-field-error">Must be a whole number ≥ 0</span>
              {/if}

            {:else if prop.editorType === 'boolean'}
              <label class="pp-toggle">
                <input
                  type="checkbox"
                  checked={boolVal(prop.name)}
                  on:change={e => onToggle(prop.name, e)}
                />
                <span class="pp-toggle-track">
                  <span class="pp-toggle-thumb"></span>
                </span>
              </label>

            {:else if prop.editorType === 'file_scope' || prop.editorType === 'sizing' || prop.editorType === 'theme'}
              <button
                class="pp-select-btn"
                class:open={openDropdown === prop.name}
                use:setTrigger={prop.name}
                on:click={e => openDropdownFor(e, prop.name)}
                aria-haspopup="listbox"
                aria-expanded={openDropdown === prop.name}
              >
                <span class="pp-select-label">{getSelectLabel(prop)}</span>
                <svg class="pp-select-chevron" class:rotated={openDropdown === prop.name} viewBox="0 0 10 6" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M1 1l4 4 4-4"/>
                </svg>
              </button>

            {:else if prop.editorType === 'color'}
              <div class="pp-color-field">
                <label
                  class="pp-color-swatch"
                  style="background:{aiaToHex($projectProperties[prop.name] ?? prop.defaultValue)}"
                  title="Pick color"
                >
                  <input
                    type="color"
                    class="pp-color-native"
                    value={aiaToHex($projectProperties[prop.name] ?? prop.defaultValue)}
                    on:input={e => onColorPicker(prop.name, e)}
                  />
                </label>
                <input
                  class="pp-input pp-input--mono pp-color-text"
                  type="text"
                  value={textValue(prop.name)}
                  on:input={e => onTextInput(prop.name, e)}
                  on:blur={() => onTextBlur(prop.name)}
                  on:keydown={onTextEnter}
                  placeholder="&HFF000000"
                  spellcheck="false"
                />
              </div>

            {:else if prop.editorType === 'asset'}
              <input
                class="pp-input"
                type="text"
                list="pp-assets-{prop.name}"
                value={textValue(prop.name)}
                on:input={e => onTextInput(prop.name, e)}
                on:blur={() => onTextBlur(prop.name)}
                on:keydown={onTextEnter}
                placeholder="filename.png"
                spellcheck="false"
              />
              <datalist id="pp-assets-{prop.name}">
                {#each assetNames as aname}
                  <option value={aname}>{aname}</option>
                {/each}
              </datalist>

            {/if}

          </div>
        </div>
      {/each}
    </div>

  </div>
</div>

<!-- Custom dropdown — position:fixed, renders above the dialog overlay -->
{#if openDropdown && openPropDef}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="ctx-menu pp-dropdown show"
    style="left:{dropdownX}px; top:{dropdownY}px; min-width:{dropdownMinWidth}px"
    role="listbox"
    tabindex="-1"
    on:click|stopPropagation
    on:keydown={handleDropdownMenuKey}
  >
    {#each openPropOptions as opt}
      <button
        type="button"
        class="ctx-item"
        class:pp-option-active={opt.value === openPropCurrentValue}
        role="option"
        aria-selected={opt.value === openPropCurrentValue}
        on:click={() => selectOption(openPropDef.name, opt.value)}
      >
        {opt.label}
      </button>
    {/each}
  </div>
{/if}

{/if}
