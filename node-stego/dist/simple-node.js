"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.encodeImageData = exports.decodeImageData = void 0;
exports.encodePngFile = encodePngFile;
exports.decodePngFile = decodePngFile;
const sharp_1 = __importDefault(require("sharp"));
const simple_stego_1 = require("./simple-stego");
Object.defineProperty(exports, "decodeImageData", { enumerable: true, get: function () { return simple_stego_1.decodeImageData; } });
Object.defineProperty(exports, "encodeImageData", { enumerable: true, get: function () { return simple_stego_1.encodeImageData; } });
// Convert Sharp image to our ImageData format
async function sharpToImageData(image) {
    const { data, info } = await image.raw().ensureAlpha().toBuffer({ resolveWithObject: true });
    return {
        data: new Uint8Array(data),
        width: info.width,
        height: info.height
    };
}
// Convert our ImageData format back to Sharp
function imageDataToSharp(imageData) {
    return (0, sharp_1.default)(Buffer.from(imageData.data), {
        raw: {
            width: imageData.width,
            height: imageData.height,
            channels: 4
        }
    });
}
// Encode text into a PNG file
async function encodePngFile(inputPath, outputPath, text) {
    try {
        const image = (0, sharp_1.default)(inputPath);
        const imageData = await sharpToImageData(image);
        const encodedData = (0, simple_stego_1.encodeImageData)(imageData, text);
        await imageDataToSharp(encodedData)
            .png()
            .toFile(outputPath);
        console.log(`✅ Encoded image saved: ${outputPath}`);
    }
    catch (error) {
        throw new Error(`Failed to encode PNG file: ${error}`);
    }
}
// Decode text from a PNG file
async function decodePngFile(inputPath, visualizationPath) {
    try {
        const image = (0, sharp_1.default)(inputPath);
        const imageData = await sharpToImageData(image);
        const result = (0, simple_stego_1.decodeImageData)(imageData);
        if (!result) {
            console.log("❌ No message recovered.");
            return null;
        }
        console.log(`✅ Decoded: ${result.message} (votes=${result.votes})`);
        // Create visualization if requested
        if (visualizationPath && result.detectedTiles.length > 0) {
            await createVisualization(inputPath, result.detectedTiles, visualizationPath);
            console.log(`Visualization saved to ${visualizationPath}`);
        }
        return result.message;
    }
    catch (error) {
        throw new Error(`Failed to decode PNG file: ${error}`);
    }
}
// Create visualization showing detected tiles
async function createVisualization(inputPath, detectedTiles, outputPath) {
    const TILE_SIZE = 64;
    const image = (0, sharp_1.default)(inputPath);
    const { width, height } = await image.metadata();
    if (!width || !height) {
        throw new Error('Could not get image dimensions');
    }
    // Create overlay with red rectangles for detected tiles
    const overlayElements = detectedTiles.map(({ x, y }) => ({
        input: Buffer.from(`
      <svg width="${TILE_SIZE}" height="${TILE_SIZE}">
        <rect x="0" y="0" width="${TILE_SIZE}" height="${TILE_SIZE}"
              fill="red" fill-opacity="0.3" stroke="red" stroke-width="2"/>
      </svg>
    `),
        top: y,
        left: x
    }));
    await image
        .composite(overlayElements)
        .png()
        .toFile(outputPath);
}
//# sourceMappingURL=simple-node.js.map