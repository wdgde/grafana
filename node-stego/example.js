// Example usage of the node-stego library

// For Node.js usage
const { encodePngFile, decodePngFile } = require('./dist/node');

// For browser usage (uncomment these imports instead)
// import { encodeCanvas, decodeCanvas, loadImageToCanvas } from './dist/browser.js';

async function nodeExample() {
  console.log('🚀 Node.js Steganography Example\n');

  const inputImage = 'input.png';   // Your source image
  const outputImage = 'output.png'; // Encoded output image
  const message = 'Hello, this is a secret message hidden in the image!';

  try {
    // Encode a message into an image
    console.log('📝 Encoding message...');
    await encodePngFile(inputImage, outputImage, message);

    // Decode the message from the image
    console.log('🔍 Decoding message...');
    const decodedMessage = await decodePngFile(outputImage, 'visualization.png');

    if (decodedMessage === message) {
      console.log('✅ Success! Message correctly encoded and decoded.');
      console.log(`Original:  "${message}"`);
      console.log(`Decoded:   "${decodedMessage}"`);
    } else {
      console.log('❌ Message mismatch!');
      console.log(`Original:  "${message}"`);
      console.log(`Decoded:   "${decodedMessage}"`);
    }

  } catch (error) {
    console.error('❌ Error:', error.message);
    console.log('\n📋 To run this example:');
    console.log('1. Place a PNG image named "input.png" in the project root');
    console.log('2. The image should be at least 64x64 pixels');
    console.log('3. Run: node example.js');
  }
}

// Browser example (for reference)
function browserExample() {
  console.log(`
🌐 Browser Usage Example:

<script type="module">
import { encodeCanvas, decodeCanvas, loadImageToCanvas } from './dist/browser.js';

// Load an image into a canvas
const canvas = await loadImageToCanvas('image.png');

// Encode a message
const encodedCanvas = encodeCanvas(canvas, 'Secret message!');

// Decode the message
const result = decodeCanvas(encodedCanvas);
if (result) {
  console.log('Decoded message:', result.message);
  console.log('Confidence:', result.votes);
}
</script>
  `);
}

// Core library example (works in both Node.js and browser)
function coreLibraryExample() {
  console.log(`
🔧 Core Library Usage:

const { encodeImageData, decodeImageData } = require('./dist/index');

// Your image data in RGBA format
const imageData = {
  data: new Uint8Array(width * height * 4), // RGBA pixel data
  width: 800,
  height: 600
};

// Encode
const encoded = encodeImageData(imageData, 'Secret message!');

// Decode
const result = decodeImageData(encoded);
if (result) {
  console.log('Message:', result.message);
  console.log('Votes:', result.votes);
  console.log('Detected tiles:', result.detectedTiles);
}
  `);
}

// Run the Node.js example
if (require.main === module) {
  nodeExample().catch(console.error);
}