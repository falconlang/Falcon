const COMMENT_WIDTH = '160';
const COMMENT_HEIGHT = '80';

function xmlChunks(xml) {
  return String(xml || '')
    .split('\0')
    .map(chunk => chunk.trim())
    .filter(Boolean);
}

function normalizeLineEndings(text) {
  return String(text ?? '').replace(/\r\n?/g, '\n');
}

function stripFalconLineCommentPrefix(line) {
  let text = String(line || '').trimStart().slice(2);
  if (text.startsWith(' ')) text = text.slice(1);
  return text;
}

function isElement(node, localName = '') {
  return node?.nodeType === 1
    && (!localName || String(node.localName || node.tagName).toLowerCase() === localName);
}

function directChild(element, localName) {
  return Array.from(element?.childNodes || [])
    .find(node => isElement(node, localName)) || null;
}

function parserError(doc) {
  return doc.getElementsByTagName('parsererror').length > 0;
}

function documentRootBlocks(doc) {
  const root = doc.documentElement;
  if (!root) return [];
  if (isElement(root, 'block')) return [root];
  return Array.from(root.childNodes).filter(node => isElement(node, 'block'));
}

function createXmlDocument() {
  return document.implementation.createDocument('http://www.w3.org/1999/xhtml', 'xml');
}

function setBlockComment(block, commentText) {
  const text = normalizeLineEndings(commentText);
  if (!text.trim()) return;

  const doc = block.ownerDocument;
  let comment = directChild(block, 'comment');
  if (!comment) {
    comment = doc.createElementNS(block.namespaceURI || null, 'comment');
    const mutation = directChild(block, 'mutation');
    if (mutation?.nextSibling) {
      block.insertBefore(comment, mutation.nextSibling);
    } else if (mutation) {
      block.appendChild(comment);
    } else {
      const firstStructuredChild = Array.from(block.childNodes)
        .find(node => isElement(node, 'field')
          || isElement(node, 'value')
          || isElement(node, 'statement')
          || isElement(node, 'next'));
      if (firstStructuredChild) block.insertBefore(comment, firstStructuredChild);
      else block.appendChild(comment);
    }
  }

  comment.setAttribute('pinned', comment.getAttribute('pinned') || 'false');
  comment.setAttribute('h', comment.getAttribute('h') || COMMENT_HEIGHT);
  comment.setAttribute('w', comment.getAttribute('w') || COMMENT_WIDTH);
  comment.textContent = text;
}

function commentedChunk(xmlString, commentText) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(xmlString, 'text/xml');
  if (parserError(doc)) return xmlString;

  const block = documentRootBlocks(doc)[0];
  if (!block) return xmlString;
  setBlockComment(block, commentText);
  return new XMLSerializer().serializeToString(doc);
}

function blockWithoutNext(block) {
  const clone = block.cloneNode(true);
  const next = directChild(clone, 'next');
  if (next) next.remove();
  return clone;
}

function nextBlock(block) {
  const next = directChild(block, 'next');
  return Array.from(next?.childNodes || []).find(node => isElement(node, 'block')) || null;
}

function blockCommentText(block) {
  return normalizeLineEndings(directChild(block, 'comment')?.textContent || '');
}

function xmlForSingleBlock(block) {
  const doc = createXmlDocument();
  doc.documentElement.appendChild(doc.importNode(blockWithoutNext(block), true));
  return new XMLSerializer().serializeToString(doc);
}

function blockType(block) {
  return block?.getAttribute?.('type') || block?.tagName || 'unknown';
}

function falconCommentLines(commentText) {
  const lines = normalizeLineEndings(commentText).split('\n');
  return lines.map(line => (line ? `// ${line}` : '//')).join('\n');
}

export function leadingFalconCommentBeforeLine(sourceCode, lineNumber) {
  const lines = normalizeLineEndings(sourceCode).split('\n');
  let index = Math.trunc(Number(lineNumber)) - 2;
  const commentLines = [];

  while (index >= 0) {
    const line = lines[index] || '';
    if (!line.trimStart().startsWith('//')) break;
    commentLines.unshift(stripFalconLineCommentPrefix(line));
    index -= 1;
  }

  return commentLines.join('\n');
}

export function injectFalconCommentsIntoBlocklyXml(xmlText, sourceCode, lineNumbers = []) {
  const chunks = xmlChunks(xmlText);
  if (!chunks.length) return String(xmlText || '');

  return chunks.map((chunk, index) => {
    const comment = leadingFalconCommentBeforeLine(sourceCode, lineNumbers[index]);
    return comment.trim() ? commentedChunk(chunk, comment) : chunk;
  }).join('\0');
}

export function blocklyXmlHasBlockComments(xmlText) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(String(xmlText || ''), 'text/xml');
  if (parserError(doc)) return false;

  return Array.from(doc.getElementsByTagName('*')).some(element =>
    isElement(element, 'block') && Boolean(blockCommentText(element).trim())
  );
}

export async function blocklyXmlToFalconCodeWithComments(xmlText, xmlToMist) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(String(xmlText || ''), 'text/xml');
  if (parserError(doc)) {
    throw new Error('Blockly XML is invalid');
  }

  const chunks = [];
  for (const topBlock of documentRootBlocks(doc)) {
    for (let block = topBlock; block; block = nextBlock(block)) {
      const code = String(await xmlToMist(xmlForSingleBlock(block)) || '').trim();
      if (!code) {
        throw new Error(`Blockly block "${blockType(block)}" could not be converted to Falcon code`);
      }
      const comment = blockCommentText(block);
      chunks.push(comment.trim()
        ? `${falconCommentLines(comment)}\n${code}`
        : code);
    }
  }

  return chunks.join('\n\n').trim();
}
