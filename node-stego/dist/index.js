"use strict";
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
exports.decodeImageDataCropped = decodeImageDataCropped;
// === Parameters ===
const TILE_SIZE = 64; // in bits - seems ot be sweet spot of tile big enough for a useful string but not too big that search is horribly slow
const LIGHT = 240;
const DARK = 15;
const MAGIC_HEADER = new Uint8Array([0xDE, 0xAD, 0xBE, 0xEF]);
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
// === Message Encoding & Decoding ===
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
    for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
            if (bitIdx >= bits.length) {
                return result;
            }
            const pixelIdx = (y * width + x) * 4; // RGBA
            const r = result[pixelIdx];
            const g = result[pixelIdx + 1];
            const b = result[pixelIdx + 2];
            // Find brightest channel
            let brightestChannel = 0;
            if (g > result[pixelIdx + brightestChannel])
                brightestChannel = 1;
            if (b > result[pixelIdx + brightestChannel])
                brightestChannel = 2;
            // Safety check for light pixels
            if (r > LIGHT && g > LIGHT && b > LIGHT && result[pixelIdx + brightestChannel] > 250) {
                const safeChannels = [];
                if (r < 250)
                    safeChannels.push(0);
                if (g < 250)
                    safeChannels.push(1);
                if (b < 250)
                    safeChannels.push(2);
                if (safeChannels.length > 0) {
                    brightestChannel = safeChannels[0];
                }
            }
            // Determine bit position
            let bitPos;
            if ((r < DARK && g < DARK && b < DARK) || (r > LIGHT && g > LIGHT && b > LIGHT)) {
                bitPos = 1;
            }
            else {
                bitPos = 0;
            }
            const bit = bits[bitIdx];
            result[pixelIdx + brightestChannel] = setBit(result[pixelIdx + brightestChannel], bitPos, bit);
            bitIdx++;
        }
    }
    return result;
}
function extractBitsFromTile(tile, width, height, lengthBits) {
    const bits = [];
    let bitIdx = 0;
    for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
            if (bitIdx >= lengthBits) {
                return bits;
            }
            const pixelIdx = (y * width + x) * 4; // RGBA
            const r = tile[pixelIdx];
            const g = tile[pixelIdx + 1];
            const b = tile[pixelIdx + 2];
            // Use same channel selection logic as embedding
            let brightestChannel;
            if (r > LIGHT && g > LIGHT && b > LIGHT) {
                const nonMaxed = [];
                if (r < 253)
                    nonMaxed.push({ idx: 0, val: r }); //253 because can add 2 (bitshift in position 1) and still avoid overflow
                if (g < 253)
                    nonMaxed.push({ idx: 1, val: g });
                if (b < 253)
                    nonMaxed.push({ idx: 2, val: b });
                if (nonMaxed.length > 0) {
                    nonMaxed.sort((a, b) => a.val - b.val);
                    brightestChannel = nonMaxed[0].idx;
                }
                else {
                    brightestChannel = r <= g && r <= b ? 0 : (g <= b ? 1 : 2);
                }
            }
            else {
                brightestChannel = 0;
                if (g > tile[pixelIdx + brightestChannel])
                    brightestChannel = 1;
                if (b > tile[pixelIdx + brightestChannel])
                    brightestChannel = 2;
            }
            // Safety check
            if (r > LIGHT && g > LIGHT && b > LIGHT && tile[pixelIdx + brightestChannel] > 250) {
                const safeChannels = [];
                if (r < 250)
                    safeChannels.push(0);
                if (g < 250)
                    safeChannels.push(1);
                if (b < 250)
                    safeChannels.push(2);
                if (safeChannels.length > 0) {
                    brightestChannel = safeChannels[0];
                }
            }
            // Determine bit position
            let bitPos;
            if ((r < DARK && g < DARK && b < DARK) || (r > LIGHT && g > LIGHT && b > LIGHT)) {
                bitPos = 1;
            }
            else {
                bitPos = 0;
            }
            const bit = getBit(tile[pixelIdx + brightestChannel], bitPos);
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
    for (let y = 0; y <= height - TILE_SIZE; y += TILE_SIZE) {
        for (let x = 0; x <= width - TILE_SIZE; x += TILE_SIZE) {
            // Extract tile
            const tile = new Uint8Array(TILE_SIZE * TILE_SIZE * 4);
            for (let ty = 0; ty < TILE_SIZE; ty++) {
                for (let tx = 0; tx < TILE_SIZE; tx++) {
                    const srcIdx = ((y + ty) * width + (x + tx)) * 4;
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
                    const dstIdx = ((y + ty) * width + (x + tx)) * 4;
                    result[dstIdx] = encodedTile[srcIdx];
                    result[dstIdx + 1] = encodedTile[srcIdx + 1];
                    result[dstIdx + 2] = encodedTile[srcIdx + 2];
                    result[dstIdx + 3] = encodedTile[srcIdx + 3];
                }
            }
        }
    }
    return { data: result, width, height };
}
function decodeImageData(imageData) {
    const { data, width, height } = imageData;
    // const maxBits = (MAGIC_HEADER.length + 2 + 100 + 1) * 8; // Header + length + max 100 bytes + checksum
    const maxBits = TILE_SIZE * TILE_SIZE;
    const messages = {};
    const detectedTiles = [];
    for (let y = 0; y <= height - TILE_SIZE; y += TILE_SIZE) {
        for (let x = 0; x <= width - TILE_SIZE; x += TILE_SIZE) {
            // Extract tile
            const tile = new Uint8Array(TILE_SIZE * TILE_SIZE * 4);
            for (let ty = 0; ty < TILE_SIZE; ty++) {
                for (let tx = 0; tx < TILE_SIZE; tx++) {
                    const srcIdx = ((y + ty) * width + (x + tx)) * 4;
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
                detectedTiles.push({ x, y });
                const text = dataToText(decodedBytes);
                // console.debug("DECODED TEXT", text);
                if (text) {
                    messages[text] = (messages[text] || 0) + 1;
                }
            }
        }
    }
    if (Object.keys(messages).length === 0) {
        return null;
    }
    // console.debug("MESSAGES", messages);
    // Return most common message
    const [message, votes] = Object.entries(messages).reduce((a, b) => a[1] > b[1] ? a : b);
    return { message, votes, detectedTiles };
}
function decodeImageDataCropped(imageData) {
    const { data, width, height } = imageData;
    // const maxBits = (MAGIC_HEADER.length + 2 + 100 + 1) * 8; // Header + length + max 100 bytes + checksum
    const maxBits = TILE_SIZE * TILE_SIZE;
    // Helper function to try decoding at a specific position
    const tryDecodeAtPosition = (startX, startY) => {
        if (startX < 0 || startY < 0 || startX + TILE_SIZE > width || startY + TILE_SIZE > height) {
            return null;
        }
        // Extract tile
        const tile = new Uint8Array(TILE_SIZE * TILE_SIZE * 4);
        for (let ty = 0; ty < TILE_SIZE; ty++) {
            for (let tx = 0; tx < TILE_SIZE; tx++) {
                const srcIdx = ((startY + ty) * width + (startX + tx)) * 4;
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
            return text;
        }
        return null;
    };
    // Helper function to validate alignment by finding 5 more tiles with the same message
    const validateAlignment = (offsetX, offsetY, expectedMessage) => {
        const detectedTiles = [];
        let votes = 0;
        // Start from the offset and scan in grid pattern
        for (let y = offsetY; y <= height - TILE_SIZE; y += TILE_SIZE) {
            for (let x = offsetX; x <= width - TILE_SIZE; x += TILE_SIZE) {
                const decoded = tryDecodeAtPosition(x, y);
                if (decoded === expectedMessage) {
                    votes++;
                    detectedTiles.push({ x, y });
                }
            }
        }
        // Need at least 4 total hits (including the original one that triggered this validation)
        if (votes >= 4) {
            return { message: expectedMessage, votes, detectedTiles };
        }
        return null;
    };
    // Try scanning around top-left (0, 0) first
    const scanRange = TILE_SIZE - 1;
    for (let dy = -scanRange; dy <= scanRange; dy++) {
        for (let dx = -scanRange; dx <= scanRange; dx++) {
            const testX = 0 + dx;
            const testY = 0 + dy;
            const decoded = tryDecodeAtPosition(testX, testY);
            if (decoded) {
                // Found a hit, determine alignment and validate
                // Calculate what the tile grid offset would be for this position
                const alignmentOffsetX = ((testX % TILE_SIZE) + TILE_SIZE) % TILE_SIZE;
                const alignmentOffsetY = ((testY % TILE_SIZE) + TILE_SIZE) % TILE_SIZE;
                const result = validateAlignment(alignmentOffsetX, alignmentOffsetY, decoded);
                if (result) {
                    return result;
                }
            }
        }
    }
    // If no success at top-left, try lots of random positions
    const numRandomTries = 20;
    for (let attempt = 0; attempt < numRandomTries; attempt++) {
        // Choose random center point, avoiding edges
        const centerX = Math.floor(Math.random() * (width - TILE_SIZE * 2)) + TILE_SIZE;
        const centerY = Math.floor(Math.random() * (height - TILE_SIZE * 2)) + TILE_SIZE;
        // Scan around this center point
        for (let dy = -scanRange; dy <= scanRange; dy++) {
            for (let dx = -scanRange; dx <= scanRange; dx++) {
                const testX = centerX + dx;
                const testY = centerY + dy;
                const decoded = tryDecodeAtPosition(testX, testY);
                if (decoded) {
                    // Found a hit, determine alignment and validate
                    const alignmentOffsetX = ((testX % TILE_SIZE) + TILE_SIZE) % TILE_SIZE;
                    const alignmentOffsetY = ((testY % TILE_SIZE) + TILE_SIZE) % TILE_SIZE;
                    const result = validateAlignment(alignmentOffsetX, alignmentOffsetY, decoded);
                    if (result) {
                        return result;
                    }
                }
            }
        }
    }
    return null;
}
//# sourceMappingURL=index.js.map