import sharp from 'sharp';
import { decodeImageData, decodeImageDataCropped, DecodeResult, encodeImageData, ImageData } from './index';

// Convert Sharp image to our ImageData format
async function sharpToImageData(image: sharp.Sharp): Promise<ImageData> {
  const { data, info } = await image.raw().ensureAlpha().toBuffer({ resolveWithObject: true });
  return {
    data: new Uint8Array(data),
    width: info.width,
    height: info.height
  };
}

// Convert our ImageData format back to Sharp
function imageDataToSharp(imageData: ImageData): sharp.Sharp {
  return sharp(Buffer.from(imageData.data), {
    raw: {
      width: imageData.width,
      height: imageData.height,
      channels: 4
    }
  });
}

// Encode text into a PNG file
export async function encodePngFile(inputPath: string, outputPath: string, text: string): Promise<void> {
  try {
    const image = sharp(inputPath);
    const imageData = await sharpToImageData(image);
    const encodedData = encodeImageData(imageData, text);

    await imageDataToSharp(encodedData)
      .png()
      .toFile(outputPath);

    console.log(`✅ Encoded image saved: ${outputPath}`);
  } catch (error) {
    throw new Error(`Failed to encode PNG file: ${error}`);
  }
}

// Decode text from a PNG file
export async function decodePngFile(inputPath: string, visualizationPath?: string): Promise<string | null> {
  try {
    const image = sharp(inputPath);
    const imageData = await sharpToImageData(image);
    const result = decodeImageData(imageData);

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
  } catch (error) {
    throw new Error(`Failed to decode PNG file: ${error}`);
  }
}

// Create visualization showing detected tiles
async function createVisualization(
  inputPath: string,
  detectedTiles: Array<{x: number, y: number}>,
  outputPath: string
): Promise<void> {
  const TILE_SIZE = 64;
  const image = sharp(inputPath);
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

// Export the core functions as well
export { decodeImageData, decodeImageDataCropped, DecodeResult, encodeImageData, ImageData };

// Decode text from a cropped PNG file using the cropped decoder
export async function decodePngFileCropped(inputPath: string, visualizationPath?: string): Promise<string | null> {
  try {
    const image = sharp(inputPath);
    const imageData = await sharpToImageData(image);
    const result = decodeImageDataCropped(imageData);

    if (!result) {
      console.log("❌ No message recovered from cropped image.");
      return null;
    }

    console.log(`✅ Decoded from cropped image: ${result.message} (votes=${result.votes})`);

    // Create visualization if requested
    if (visualizationPath && result.detectedTiles.length > 0) {
      await createVisualization(inputPath, result.detectedTiles, visualizationPath);
      console.log(`Visualization saved to ${visualizationPath}`);
    }

    return result.message;
  } catch (error) {
    throw new Error(`Failed to decode cropped PNG file: ${error}`);
  }
}

// Randomly crop an image for testing purposes
export async function randomlyCropImage(inputPath: string, outputPath: string, minCropSize: number = 128): Promise<{cropX: number, cropY: number, cropWidth: number, cropHeight: number}> {
  try {
    const image = sharp(inputPath);
    const { width, height } = await image.metadata();

    if (!width || !height) {
      throw new Error('Could not get image dimensions');
    }

    // Ensure we can crop at least minCropSize in both dimensions
    if (width < minCropSize || height < minCropSize) {
      throw new Error(`Image too small to crop. Need at least ${minCropSize}x${minCropSize}`);
    }

    // Random crop parameters
    const maxCropWidth = width - minCropSize;
    const maxCropHeight = height - minCropSize;

    const cropX = Math.floor(Math.random() * maxCropWidth);
    const cropY = Math.floor(Math.random() * maxCropHeight);

    // Random crop size between minCropSize and remaining space
    const maxWidth = width - cropX;
    const maxHeight = height - cropY;
    const cropWidth = Math.floor(Math.random() * (maxWidth - minCropSize)) + minCropSize;
    const cropHeight = Math.floor(Math.random() * (maxHeight - minCropSize)) + minCropSize;

    await image
      .extract({ left: cropX, top: cropY, width: cropWidth, height: cropHeight })
      .png()
      .toFile(outputPath);

    console.log(`✅ Randomly cropped image saved: ${outputPath}`);
    console.log(`   Original: ${width}x${height}, Cropped: ${cropWidth}x${cropHeight} at (${cropX}, ${cropY})`);

    return { cropX, cropY, cropWidth, cropHeight };
  } catch (error) {
    throw new Error(`Failed to crop image: ${error}`);
  }
}
