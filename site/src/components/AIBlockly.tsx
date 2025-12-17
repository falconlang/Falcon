/* eslint-disable */
import { useEffect } from "react";

// These DONT work, chekc them
export function BlocklyPanel_checkIsAdmin() {
  return true;
}

export function BlocklyPanel_getSnapEnabled() {
  return true;
}

/**
 * Blockly.CONTROL_CATEGORY_HUE = "#F3AA44";  // [177, 143, 53]
 * Blockly.LOGIC_CATEGORY_HUE = "#5BA55B";  // [119, 171, 65]
 * Blockly.MATH_CATEGORY_HUE = "#67a5fa";  // [63, 113, 181]
 * Blockly.TEXT_CATEGORY_HUE = "#dc5e76";  // [179, 45, 94]
 * Blockly.LIST_CATEGORY_HUE = "#49a6d5";  // [73, 166, 212]
 * Blockly.COLOR_CATEGORY_HUE = "#7D7D7D";  // [125, 125, 125]
 * Blockly.VARIABLE_CATEGORY_HUE = "#fa7753";  // [208, 95, 45]
 * Blockly.PROCEDURE_CATEGORY_HUE = "#745da8";  // [124, 83, 133]
 * Blockly.DICTIONARY_CATEGORY_HUE = "#5f6cae"; // [45, 23, 153]
 */

// eslint-disable-next-line
declare const Blockly: any;
// eslint-disable-next-line
declare const AI: any;


const options = {
  toolbox: {
    kind: 'flyoutToolbox',
    contents: [],
  },
  trashcan: false,
  // TODO: add rest of all the options
  multiselectIcon: {
    hideIcon: true,
    enabledIcon: 'static/images/select.svg',
    disabledIcon: 'static/images/unselect.svg',
  },
  multiselectCopyPaste: {
    crossTab: true,
    menu: true,
  },
  readOnly: true, // readOnly if in iframe
  useDoubleClick: true,
  bumpNeighbours: true,
  renderer: "geras2_renderer",

  'zoom': { 'controls': false, 'wheel': false, 'scaleSpeed': 1.1, 'maxScale': 3, 'minScale': 0.1 },
};


export function AIBlockly() {
  useEffect(() => {
    // Inject Blockly workspace with your existing toolbox object
    const workspace = Blockly.inject('blocklyDiv', options);
    AI.Blockly.multiselect = (window as any).Multiselect;
    const multiselectPlugin = new AI.Blockly.multiselect(workspace);
    multiselectPlugin.init(options);
    const lexicalVariablesPlugin = (window as any).LexicalVariablesPlugin;
    lexicalVariablesPlugin.init(workspace);
    const searchPlugin = new (window as any).WorkspaceSearch(workspace);
    searchPlugin.init();
    const scrollOptions = new (window as any).ScrollOptions(workspace);
    // Make autoscrolling be based purely on the mouse position ands slow it down a bit.
    scrollOptions.init({
      edgeScrollOptions: {
        oversizeBlockMargin: 0,
        oversizeBlockThreshold: 0,
        slowBlockSpeed: .15
      }
    });

    // Custom workspace properties
    workspace.formName = "CatScreen";
    workspace.screenList_ = [];
    workspace.assetList_ = [];
    workspace.componentDb_ = new Blockly.ComponentDatabase();
    workspace.procedureDb_ = new Blockly.ProcedureDatabase(workspace);
    workspace.variableDb_ = new Blockly.VariableDatabase();
    workspace.blocksNeedingRendering = [];
    //workspace.addWarningHandler();

    // @ts-ignore
    workspace.addChangeListener(function (e) {
      if (e.type == Blockly.Events.BLOCK_MOVE && e.newParentId !== e.oldParentId) {
        const block = workspace.getBlockById(e.blockId);
        if (!block) {
          // This seems to be the case when the block has been deleted since it is first moved from
          // its parent then removed from the workspace, but both events will be run back to back
          // after the deletion has already happened.
          return;
        }
        // @ts-ignore
        block.getDescendants().forEach(function (block) {
          if (block.type === 'lexical_variable_get' || block.type === 'lexical_variable_set') {
            // If the block is a lexical variable, then we need to rebuild the options for the field
            // given the change in scope.
            const field = block.getField('VAR');
            field.getOptions(false);  // rebuild option cache
            field.setValue(field.getValue());
            block.queueRender();
          }
        });
      }
    });

    workspace.flyout_ = workspace.getFlyout();
    //workspace.addWarningIndicator();
    //workspace.addBackpack();
    Blockly.browserEvents.bind(workspace.svgGroup_, 'focus', workspace, workspace.markFocused);
    // Hide scrollbars by default (otherwise ghost rectangles intercept mouse events)
    // Render blocks created prior to the workspace being rendered.
    workspace.injecting = false;
    workspace.injected = true;
    workspace.notYetRendered = true;

    // Dont add a toolbox if in iframe
    // if (!inIframe) {
    //   initToolbox();
    // }

    // Resize handling
    const resizeObserver = new ResizeObserver(() => {
      Blockly.svgResize(workspace);
    });
    const blocklyDiv = document.getElementById('blocklyDiv');
    if (blocklyDiv) {
      resizeObserver.observe(blocklyDiv);
    }

    return () => {
      workspace.dispose();
      resizeObserver.disconnect();
    };
  }, []);

  console.log("Done");
  return <div id="blocklyDiv" style={{ width: '100%', height: '100%', flex: 1 }} />

}
