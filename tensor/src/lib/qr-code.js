// Minimal QR encoder for the five-digit MIT Companion code.
// Implements QR Code version 1 with medium error correction and numeric mode.
const SIZE = 21;
const DATA_CODEWORDS = 16;
const ECC_CODEWORDS = 10;
const MASK = 0;
const FORMAT_XOR = 0x5412;
const PAD_BYTES = [0xec, 0x11];

function appendBits(out, value, length) {
  if (length < 0 || length > 31 || value >>> length !== 0) {
    throw new RangeError('QR bit value out of range');
  }
  for (let i = length - 1; i >= 0; i -= 1) {
    out.push((value >>> i) & 1);
  }
}

function bitAt(byte, index) {
  return ((byte >>> index) & 1) !== 0;
}

function encodeNumericData(text) {
  if (!/^[0-9]{1,16}$/.test(text)) {
    throw new Error('Companion QR codes must be 1 to 16 digits');
  }

  const bits = [];
  appendBits(bits, 0b0001, 4);
  appendBits(bits, text.length, 10);

  for (let i = 0; i < text.length; i += 3) {
    const chunk = text.slice(i, i + 3);
    appendBits(bits, Number(chunk), chunk.length === 3 ? 10 : chunk.length === 2 ? 7 : 4);
  }

  const capacity = DATA_CODEWORDS * 8;
  appendBits(bits, 0, Math.min(4, capacity - bits.length));
  appendBits(bits, 0, (8 - (bits.length % 8)) % 8);

  let padIndex = 0;
  while (bits.length < capacity) {
    appendBits(bits, PAD_BYTES[padIndex], 8);
    padIndex = 1 - padIndex;
  }

  const codewords = [];
  for (let i = 0; i < bits.length; i += 8) {
    let value = 0;
    for (let j = 0; j < 8; j += 1) value = (value << 1) | bits[i + j];
    codewords.push(value);
  }
  return codewords;
}

function gfMultiply(x, y) {
  let z = 0;
  for (let i = 7; i >= 0; i -= 1) {
    z = (z << 1) ^ ((z >>> 7) * 0x11d);
    z ^= ((y >>> i) & 1) * x;
  }
  return z & 0xff;
}

function reedSolomonDivisor(degree) {
  const result = Array(degree - 1).fill(0).concat(1);
  let root = 1;
  for (let i = 0; i < degree; i += 1) {
    for (let j = 0; j < result.length; j += 1) {
      result[j] = gfMultiply(result[j], root);
      if (j + 1 < result.length) result[j] ^= result[j + 1];
    }
    root = gfMultiply(root, 2);
  }
  return result;
}

function reedSolomonRemainder(data, divisor) {
  const result = divisor.map(() => 0);
  for (const byte of data) {
    const factor = byte ^ result.shift();
    result.push(0);
    divisor.forEach((coef, index) => {
      result[index] ^= gfMultiply(coef, factor);
    });
  }
  return result;
}

function createMatrix() {
  return {
    modules: Array.from({ length: SIZE }, () => Array(SIZE).fill(false)),
    reserved: Array.from({ length: SIZE }, () => Array(SIZE).fill(false)),
  };
}

function setFunction(matrix, x, y, dark) {
  matrix.modules[y][x] = dark;
  matrix.reserved[y][x] = true;
}

function drawFinder(matrix, cx, cy) {
  for (let dy = -4; dy <= 4; dy += 1) {
    for (let dx = -4; dx <= 4; dx += 1) {
      const x = cx + dx;
      const y = cy + dy;
      if (x < 0 || x >= SIZE || y < 0 || y >= SIZE) continue;
      const dist = Math.max(Math.abs(dx), Math.abs(dy));
      setFunction(matrix, x, y, dist !== 2 && dist !== 4);
    }
  }
}

function drawFunctionPatterns(matrix) {
  for (let i = 0; i < SIZE; i += 1) {
    setFunction(matrix, 6, i, i % 2 === 0);
    setFunction(matrix, i, 6, i % 2 === 0);
  }
  drawFinder(matrix, 3, 3);
  drawFinder(matrix, SIZE - 4, 3);
  drawFinder(matrix, 3, SIZE - 4);
  drawFormatBits(matrix, MASK);
}

function drawFormatBits(matrix, mask) {
  const data = mask;
  let rem = data;
  for (let i = 0; i < 10; i += 1) rem = (rem << 1) ^ ((rem >>> 9) * 0x537);
  const bits = ((data << 10) | rem) ^ FORMAT_XOR;

  for (let i = 0; i <= 5; i += 1) setFunction(matrix, 8, i, bitAt(bits, i));
  setFunction(matrix, 8, 7, bitAt(bits, 6));
  setFunction(matrix, 8, 8, bitAt(bits, 7));
  setFunction(matrix, 7, 8, bitAt(bits, 8));
  for (let i = 9; i < 15; i += 1) setFunction(matrix, 14 - i, 8, bitAt(bits, i));
  for (let i = 0; i < 8; i += 1) setFunction(matrix, SIZE - 1 - i, 8, bitAt(bits, i));
  for (let i = 8; i < 15; i += 1) setFunction(matrix, 8, SIZE - 15 + i, bitAt(bits, i));
  setFunction(matrix, 8, SIZE - 8, true);
}

function drawCodewords(matrix, codewords) {
  let bitIndex = 0;
  const bitCount = codewords.length * 8;

  for (let right = SIZE - 1; right >= 1; right -= 2) {
    if (right === 6) right = 5;
    for (let vert = 0; vert < SIZE; vert += 1) {
      for (let j = 0; j < 2; j += 1) {
        const x = right - j;
        const upward = ((right + 1) & 2) === 0;
        const y = upward ? SIZE - 1 - vert : vert;
        if (matrix.reserved[y][x] || bitIndex >= bitCount) continue;
        matrix.modules[y][x] = bitAt(codewords[bitIndex >>> 3], 7 - (bitIndex & 7));
        bitIndex += 1;
      }
    }
  }
}

function maskBit(x, y) {
  return (x + y) % 2 === 0;
}

function applyMask(matrix) {
  for (let y = 0; y < SIZE; y += 1) {
    for (let x = 0; x < SIZE; x += 1) {
      if (!matrix.reserved[y][x] && maskBit(x, y)) {
        matrix.modules[y][x] = !matrix.modules[y][x];
      }
    }
  }
}

function matrixToPath(modules) {
  const ops = [];
  modules.forEach((row, y) => {
    let start = null;
    row.forEach((dark, x) => {
      if (dark && start === null) start = x;
      if ((!dark || x === row.length - 1) && start !== null) {
        const end = dark && x === row.length - 1 ? x + 1 : x;
        ops.push(`M${start} ${y}h${end - start}v1H${start}z`);
        start = null;
      }
    });
  });
  return ops.join('');
}

function svgEscape(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

export function createQrSvg(value, {
  size = 148,
  fgColor = '#1A1916',
  bgColor = '#ffffff',
} = {}) {
  const text = String(value ?? '');
  const data = encodeNumericData(text);
  const ecc = reedSolomonRemainder(data, reedSolomonDivisor(ECC_CODEWORDS));
  const matrix = createMatrix();

  drawFunctionPatterns(matrix);
  drawCodewords(matrix, data.concat(ecc));
  applyMask(matrix);
  drawFormatBits(matrix, MASK);

  return `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${SIZE} ${SIZE}" role="img" aria-label="Companion code ${svgEscape(text)}"><path fill="${svgEscape(bgColor)}" d="M0 0h${SIZE}v${SIZE}H0z" shape-rendering="crispEdges"/><path fill="${svgEscape(fgColor)}" d="${matrixToPath(matrix.modules)}" shape-rendering="crispEdges"/></svg>`;
}
