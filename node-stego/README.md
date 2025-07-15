# Node Stego - TypeScript Steganography Library

A TypeScript library for hiding and extracting text messages in images using LSB (Least Significant Bit) steganography techniques. Works both in Node.js (for PNG files) and browsers (with Canvas elements).

## 🎯 Features

- 🔐 Hide text messages in images using LSB steganography
- 🎯 Tile-based processing (64x64 pixel tiles) for redundancy and image crop-resistance
- 📱 Works in both Node.js and browser environments
- 🖼️ PNG file support (Node.js) and Canvas support (browser)
- 📊 Message detection visualization
- 🔧 TypeScript with full type definitions

## 🚀 Quick Start

```bash
npm install
npm run demo
```

This will create a test image, hide a message in it, and then extract the message back out!

## 📦 Installation & Build

```bash
npm install
npm run build
```

## 💡 Usage

### Node.js (PNG Files)

```typescript
import { encodePngFile, decodePngFile } from './dist/node';

// Encode a message into a PNG file
await encodePngFile('input.png', 'output.png', 'Secret message!');

// Decode a message from a PNG file (with optional visualization)
const message = await decodePngFile('output.png', 'visualization.png');
console.log('Decoded message:', message);
```

### Browser (Canvas)

```typescript
import {
  encodeCanvas,
  decodeCanvas,
  loadImageToCanvas
} from './dist/browser';

// Load image into canvas
const canvas = await loadImageToCanvas('image.png');

// Encode message into canvas
const encodedCanvas = encodeCanvas(canvas, 'Secret message!');

// Decode message from canvas
const result = decodeCanvas(encodedCanvas);
if (result) {
  console.log('Message:', result.message);
  console.log('Confidence:', result.votes);
}
```

### Core Functions (Both Node.js and Browser)

```typescript
import { encodeImageData, decodeImageData } from './dist/index';

// Work with raw image data (RGBA format)
const imageData = {
  data: new Uint8Array(width * height * 4), // RGBA pixel data
  width: 800,
  height: 600
};

const encoded = encodeImageData(imageData, 'Secret message!');
const result = decodeImageData(encoded);
```

## 🎮 Demo Scripts

### Working Demo
```bash
npm run demo
```


## 📋 API Reference

### Node.js Functions

#### `encodePngFile(inputPath, outputPath, text)`
- `inputPath`: Path to input PNG file
- `outputPath`: Path for output PNG file
- `text`: Text message to encode
- Returns: `Promise<void>`

#### `decodePngFile(inputPath, visualizationPath?)`
- `inputPath`: Path to PNG file to decode
- `visualizationPath`: Optional path for visualization output
- Returns: `Promise<string | null>`

### Browser Functions

#### `encodeCanvas(canvas, text)`
- `canvas`: HTMLCanvasElement to encode
- `text`: Text message to encode
- Returns: `HTMLCanvasElement` (modified canvas)

#### `decodeCanvas(canvas)`
- `canvas`: HTMLCanvasElement to decode
- Returns: `DecodeResult | null`

#### `loadImageToCanvas(imageUrl)`
- `imageUrl`: URL or path to image
- Returns: `Promise<HTMLCanvasElement>`

### Core Functions

#### `encodeImageData(imageData, text)`
- `imageData`: ImageData object with RGBA pixel data
- `text`: Text message to encode
- Returns: `ImageData` (encoded image data)

#### `decodeImageData(imageData)`
- `imageData`: ImageData object to decode
- Returns: `DecodeResult | null`

### Types

```typescript
interface ImageData {
  data: Uint8Array;  // RGBA pixel data
  width: number;
  height: number;
}

interface DecodeResult {
  message: string;
  votes: number;  // Number of tiles that detected this message
  detectedTiles: Array<{x: number, y: number}>;
}
```

## 🔧 How It Works

### Simple Implementation (Recommended)
1. **Magic Header**: Uses a 4-byte magic header (0xDEADBEEF) to identify encoded data
2. **Message Format**: Magic header + length (2 bytes) + message data + checksum (1 byte)
3. **LSB Embedding**: Stores bits in the least significant bit of the red channel
1. **Reed-Solomon Error Correction**: Advanced error correction coding
4. **Tile Processing**: Uses 64x64 pixel tiles for organization
2. **Multiple Tiles**: Distributes data across multiple tiles for redundancy
3. **Channel Selection**: Smart selection of color channels based on pixel characteristics
4. **Sync Pattern**: More complex synchronization pattern

## 🎨 Example: Browser Usage

```html
<!DOCTYPE html>
<html>
<head>
    <title>Steganography Demo</title>
</head>
<body>
    <input type="file" id="imageInput" accept="image/*">
    <canvas id="canvas"></canvas>
    <input type="text" id="messageInput" placeholder="Enter secret message">
    <button onclick="encode()">Encode</button>
    <button onclick="decode()">Decode</button>

    <script type="module">
        import { encodeCanvas, decodeCanvas } from './dist/browser.js';

        window.encode = function() {
            const canvas = document.getElementById('canvas');
            const message = document.getElementById('messageInput').value;
            encodeCanvas(canvas, message);
            alert('Message encoded!');
        };

        window.decode = function() {
            const canvas = document.getElementById('canvas');
            const result = decodeCanvas(canvas);
            if (result) {
                alert(`Decoded: ${result.message}`);
            } else {
                alert('No message found!');
            }
        };
    </script>
</body>
</html>
```

## 📚 Technical Details

- **Input Format**: PNG images (Node.js) or Canvas ImageData (browser)
- **Output Format**: Modified images with hidden data in LSB
- **Message Capacity**: Depends on image size; ~4096 bits per 64x64 tile
- **Supported Characters**: Full UTF-8 text including emojis
- **Error Detection**: Checksum-based integrity verification

## ⚠️ Requirements

- Images must be at least 64x64 pixels
- PNG format recommended for lossless storage
- For browser usage, images must be CORS-enabled if loaded from URLs
- Node.js 18+ recommended

## 📄 License

MIT License - feel free to use this for educational purposes and projects!