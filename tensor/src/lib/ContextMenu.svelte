<script>
  import { get } from 'svelte/store';
  import {
    ctxMenu, hideCtx, addCodeCell, deleteCellById, cells, requestBlocklyPreview,
  } from './stores.js';
  import {
    blocklyXmlToPng,
    copyPngBlobToClipboard,
    downloadPngBlob,
    falconCellToBlocklyPng,
    readPngBlobFromClipboard,
  } from './blockly-preview.js';

  let busyAction = null;

  function addCodeCellAfter() {
    addCodeCell();
    hideCtx();
  }

  function deleteActiveCell() {
    if ($ctxMenu.cellId) {
      deleteCellById($ctxMenu.cellId);
      hideCtx();
    }
  }

  function targetCodeCell() {
    const id = $ctxMenu.cellId;
    return get(cells).find(cell => cell.id === id && cell.type === 'code') || null;
  }

  function blockFilename(cell) {
    return `falcon-blocks-${cell?.id || 'cell'}.png`;
  }

  async function runBlockAction(action, fn) {
    if (busyAction) return;
    busyAction = action;
    hideCtx();
    try {
      await fn();
    } catch (e) {
      console.error(`[${action}]`, e);
    } finally {
      busyAction = null;
    }
  }

  async function generateCellBlocksPng(cell) {
    const source = cell?.code || '';
    if (!source.trim()) throw new Error('No Falcon code to convert');
    return falconCellToBlocklyPng(cell.id, source);
  }

  function copyBlocks() {
    const cell = targetCodeCell();
    if (!cell) return;
    runBlockAction('copy-blocks', async () => {
      const { blob } = await generateCellBlocksPng(cell);
      await copyPngBlobToClipboard(blob);
    });
  }

  function downloadBlocks() {
    const cell = targetCodeCell();
    if (!cell) return;
    runBlockAction('download-blocks', async () => {
      const { blob } = await generateCellBlocksPng(cell);
      downloadPngBlob(blob, blockFilename(cell));
    });
  }

  async function readBlocksClipboardBlob() {
    try {
      return await readPngBlobFromClipboard();
    } catch (imageError) {
      const text = await navigator.clipboard?.readText?.();
      if (text?.trim().startsWith('<')) {
        return (await blocklyXmlToPng(text)).blob;
      }
      throw imageError;
    }
  }

  function pasteBlocks() {
    const cell = targetCodeCell();
    if (!cell) return;
    runBlockAction('paste-blocks', async () => {
      const blob = await readBlocksClipboardBlob();
      requestBlocklyPreview(cell.id, { blob });
    });
  }
</script>

<div
  class="ctx-menu"
  class:show={$ctxMenu.show}
  style="left: {$ctxMenu.x}px; top: {$ctxMenu.y}px"
>
  <div class="ctx-item" on:click={addCodeCellAfter}>
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 1v12M1 7h12"/></svg>
    Insert code cell below
  </div>
  <div class="ctx-sep"></div>
  <div class="ctx-item">
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1" y="3" width="12" height="9" rx="1"/></svg>
    Copy cell
  </div>
  <div class="ctx-item">
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1" y="3" width="12" height="9" rx="1"/></svg>
    Paste cell below
  </div>
  <div class="ctx-item" class:ctx-item-disabled={busyAction !== null} on:click={copyBlocks}>
    <svg viewBox="18 7 64 81" fill="currentColor" xmlns="http://www.w3.org/2000/svg"><path stroke="currentColor" stroke-width="4" stroke-linejoin="round" d="M 50 9 C 43.936593 9 39 13.936593 39 20 C 39 22.259221 39.844873 24.249457 41.017578 26 L 26 26 C 22.698375 26 20 28.698375 20 32 L 20 47 L 20 48.859375 A 1.0001 1.0001 0 0 0 21.699219 49.572266 C 23.324036 47.978972 25.540569 47 28 47 C 32.982593 47 37 51.017407 37 56 C 37 60.982593 32.982593 65 28 65 C 25.540569 65 23.324036 64.021028 21.699219 62.427734 A 1.0001 1.0001 0 0 0 20 63.140625 L 20 64 L 20 80 C 20 83.301625 22.698375 86 26 86 L 74 86 C 77.301625 86 80 83.301625 80 80 L 80 32 C 80 28.698375 77.301625 26 74 26 L 58.982422 26 C 60.154932 24.249881 61 22.259603 61 20 C 61 13.936593 56.063407 9 50 9 z M 50 11 C 54.982593 11 59 15.017407 59 20 C 59 22.459431 58.02067 24.675276 56.427734 26.298828 A 1.0001 1.0001 0 0 0 57.140625 28 L 74 28 C 76.220375 28 78 29.779625 78 32 L 78 80 C 78 82.220375 76.220375 84 74 84 L 26 84 C 23.779625 84 22 82.220375 22 80 L 22 64.982422 C 23.750312 66.155107 25.740098 67 28 67 C 34.063407 67 39 62.063407 39 56 C 39 49.936593 34.063407 45 28 45 C 25.740098 45 23.750312 45.844893 22 47.017578 L 22 32 C 22 29.779625 23.779625 28 26 28 L 42.859375 28 A 1.0001 1.0001 0 0 0 43.572266 26.300781 C 41.979101 24.676095 41 22.458333 41 20 C 41 15.017407 45.017407 11 50 11 z"/></svg>
    Copy blocks
  </div>
  <div class="ctx-item" class:ctx-item-disabled={busyAction !== null} on:click={pasteBlocks}>
    <svg viewBox="18 7 64 81" fill="currentColor" xmlns="http://www.w3.org/2000/svg"><path stroke="currentColor" stroke-width="4" stroke-linejoin="round" d="M 50 9 C 43.936593 9 39 13.936593 39 20 C 39 22.259221 39.844873 24.249457 41.017578 26 L 26 26 C 22.698375 26 20 28.698375 20 32 L 20 47 L 20 48.859375 A 1.0001 1.0001 0 0 0 21.699219 49.572266 C 23.324036 47.978972 25.540569 47 28 47 C 32.982593 47 37 51.017407 37 56 C 37 60.982593 32.982593 65 28 65 C 25.540569 65 23.324036 64.021028 21.699219 62.427734 A 1.0001 1.0001 0 0 0 20 63.140625 L 20 64 L 20 80 C 20 83.301625 22.698375 86 26 86 L 74 86 C 77.301625 86 80 83.301625 80 80 L 80 32 C 80 28.698375 77.301625 26 74 26 L 58.982422 26 C 60.154932 24.249881 61 22.259603 61 20 C 61 13.936593 56.063407 9 50 9 z M 50 11 C 54.982593 11 59 15.017407 59 20 C 59 22.459431 58.02067 24.675276 56.427734 26.298828 A 1.0001 1.0001 0 0 0 57.140625 28 L 74 28 C 76.220375 28 78 29.779625 78 32 L 78 80 C 78 82.220375 76.220375 84 74 84 L 26 84 C 23.779625 84 22 82.220375 22 80 L 22 64.982422 C 23.750312 66.155107 25.740098 67 28 67 C 34.063407 67 39 62.063407 39 56 C 39 49.936593 34.063407 45 28 45 C 25.740098 45 23.750312 45.844893 22 47.017578 L 22 32 C 22 29.779625 23.779625 28 26 28 L 42.859375 28 A 1.0001 1.0001 0 0 0 43.572266 26.300781 C 41.979101 24.676095 41 22.458333 41 20 C 41 15.017407 45.017407 11 50 11 z"/></svg>
    Paste blocks
  </div>
  <div class="ctx-sep"></div>
  <div class="ctx-item" class:ctx-item-disabled={busyAction !== null} on:click={downloadBlocks}>
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 1v8M4 6l3 3 3-3M1 11h12v2H1z"/></svg>
    Download blocks
  </div>
  <div class="ctx-sep"></div>
  <div class="ctx-item danger" on:click={deleteActiveCell}>
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4h10M5 4V2h4v2M11 4l-.8 8H3.8L3 4"/></svg>
    Delete cell
  </div>
</div>
