"use strict";
// Simple steganography implementation without Reed-Solomon error correction
// This is for demonstration purposes and educational use
Object.defineProperty(exports, "__esModule", { value: true });
exports.textToData = textToData;
exports.dataToText = dataToText;
exports.bytesToBits = bytesToBits;
exports.bitsToBytes = bitsToBytes;
exports.encodeMessage = encodeMessage;
exports.decodeMessage = decodeMessage;
exports.embedBitsInTile = embedBitsInTile;
exports.extractBitsFromTile = extractBitsFromTile;
exports.encodeImageData = encodeImageData;
exports.decodeImageData = decodeImageData;
// === Parameters ===
const MAGIC_HEADER = new Uint8Array([0xDE, 0xAD, 0xBE, 0xEF]);
const TILE_SIZE = 64;
// === Utility Functions ===
function textToData(text) {
    return new TextEncoder().encode(text);
}
function dataToText(data) {
    try {
        return new TextDecoder().decode(data);
    }
    catch (error) {
        return null;
    }
}
function bytesToBits(data) {
    const bits = [];
    for (const byte of data) {
        for (let i = 7; i >= 0; i--) {
            bits.push((byte >> i) & 1);
        }
    }
    return bits;
}
function bitsToBytes(bits) {
    const bytes = new Uint8Array(Math.ceil(bits.length / 8));
    for (let i = 0; i < bits.length; i += 8) {
        let byte = 0;
        for (let j = 0; j < 8 && i + j < bits.length; j++) {
            byte |= (bits[i + j] << (7 - j));
        }
        bytes[Math.floor(i / 8)] = byte;
    }
    return bytes;
}
function setBit(value, position, bit) {
    if (bit === 1) {
        return value | (1 << position);
    }
    else {
        return value & ~(1 << position);
    }
}
function getBit(value, position) {
    return (value >> position) & 1;
}
// === Simple Message Encoding (without Reed-Solomon) ===
function encodeMessage(data) {
    // Simple format: MAGIC_HEADER + LENGTH (2 bytes) + DATA + CHECKSUM (1 byte)
    const length = new Uint8Array(2);
    const dataView = new DataView(length.buffer);
    dataView.setUint16(0, data.length, false); // big endian
    // Simple checksum
    let checksum = 0;
    for (const byte of data) {
        checksum = (checksum + byte) & 0xFF;
    }
    const result = new Uint8Array(MAGIC_HEADER.length + 2 + data.length + 1);
    let offset = 0;
    result.set(MAGIC_HEADER, offset);
    offset += MAGIC_HEADER.length;
    result.set(length, offset);
    offset += 2;
    result.set(data, offset);
    offset += data.length;
    result[offset] = checksum;
    return result;
}
function decodeMessage(raw) {
    // Check magic header
    for (let i = 0; i < MAGIC_HEADER.length; i++) {
        if (raw[i] !== MAGIC_HEADER[i]) {
            return null;
        }
    }
    const payload = raw.slice(MAGIC_HEADER.length);
    if (payload.length < 3) { // At least length + 1 data byte + checksum
        return null;
    }
    const dataView = new DataView(payload.buffer, payload.byteOffset, payload.byteLength);
    const length = dataView.getUint16(0, false); // big endian
    if (payload.length < 2 + length + 1) {
        return null; // Not enough data
    }
    const data = payload.slice(2, 2 + length);
    const expectedChecksum = payload[2 + length];
    // Verify checksum
    let actualChecksum = 0;
    for (const byte of data) {
        actualChecksum = (actualChecksum + byte) & 0xFF;
    }
    if (actualChecksum !== expectedChecksum) {
        return null; // Checksum mismatch
    }
    return data;
}
// === Tile Processing Functions ===
function embedBitsInTile(tile, width, height, bits) {
    const result = new Uint8Array(tile);
    const totalPixels = width * height;
    if (bits.length > totalPixels) {
        throw new Error('Data too large for tile');
    }
    let bitIdx = 0;
    for (let y = 0; y < height && bitIdx < bits.length; y++) {
        for (let x = 0; x < width && bitIdx < bits.length; x++) {
            const pixelIdx = (y * width + x) * 4; // RGBA
            const bit = bits[bitIdx];
            // Use LSB of the red channel
            result[pixelIdx] = setBit(result[pixelIdx], 0, bit);
            bitIdx++;
        }
    }
    return result;
}
function extractBitsFromTile(tile, width, height, lengthBits) {
    const bits = [];
    let bitIdx = 0;
    for (let y = 0; y < height && bitIdx < lengthBits; y++) {
        for (let x = 0; x < width && bitIdx < lengthBits; x++) {
            const pixelIdx = (y * width + x) * 4; // RGBA
            const bit = getBit(tile[pixelIdx], 0); // LSB of red channel
            bits.push(bit);
            bitIdx++;
        }
    }
    return bits;
}
// === Main Encode/Decode Functions ===
function encodeImageData(imageData, text) {
    const { data, width, height } = imageData;
    const result = new Uint8Array(data);
    const payload = encodeMessage(textToData(text));
    const dataBits = bytesToBits(payload);
    // Use only the first tile for simplicity
    if (width >= TILE_SIZE && height >= TILE_SIZE) {
        const tile = new Uint8Array(TILE_SIZE * TILE_SIZE * 4);
        // Extract first tile
        for (let ty = 0; ty < TILE_SIZE; ty++) {
            for (let tx = 0; tx < TILE_SIZE; tx++) {
                const srcIdx = (ty * width + tx) * 4;
                const dstIdx = (ty * TILE_SIZE + tx) * 4;
                tile[dstIdx] = result[srcIdx];
                tile[dstIdx + 1] = result[srcIdx + 1];
                tile[dstIdx + 2] = result[srcIdx + 2];
                tile[dstIdx + 3] = result[srcIdx + 3];
            }
        }
        // Embed bits in tile
        const encodedTile = embedBitsInTile(tile, TILE_SIZE, TILE_SIZE, dataBits);
        // Copy back to result
        for (let ty = 0; ty < TILE_SIZE; ty++) {
            for (let tx = 0; tx < TILE_SIZE; tx++) {
                const srcIdx = (ty * TILE_SIZE + tx) * 4;
                const dstIdx = (ty * width + tx) * 4;
                result[dstIdx] = encodedTile[srcIdx];
                result[dstIdx + 1] = encodedTile[srcIdx + 1];
                result[dstIdx + 2] = encodedTile[srcIdx + 2];
                result[dstIdx + 3] = encodedTile[srcIdx + 3];
            }
        }
    }
    return { data: result, width, height };
}
function decodeImageData(imageData) {
    const { data, width, height } = imageData;
    if (width < TILE_SIZE || height < TILE_SIZE) {
        return null;
    }
    // Maximum expected payload size
    const maxBits = (MAGIC_HEADER.length + 2 + 100 + 1) * 8; // Header + length + max 100 bytes + checksum
    // Extract first tile
    const tile = new Uint8Array(TILE_SIZE * TILE_SIZE * 4);
    for (let ty = 0; ty < TILE_SIZE; ty++) {
        for (let tx = 0; tx < TILE_SIZE; tx++) {
            const srcIdx = (ty * width + tx) * 4;
            const dstIdx = (ty * TILE_SIZE + tx) * 4;
            tile[dstIdx] = data[srcIdx];
            tile[dstIdx + 1] = data[srcIdx + 1];
            tile[dstIdx + 2] = data[srcIdx + 2];
            tile[dstIdx + 3] = data[srcIdx + 3];
        }
    }
    const bits = extractBitsFromTile(tile, TILE_SIZE, TILE_SIZE, maxBits);
    const rawData = bitsToBytes(bits);
    const decodedBytes = decodeMessage(rawData);
    if (decodedBytes) {
        const text = dataToText(decodedBytes);
        if (text) {
            return {
                message: text,
                votes: 1,
                detectedTiles: [{ x: 0, y: 0 }]
            };
        }
    }
    return null;
}
//# sourceMappingURL=simple-stego.js.map