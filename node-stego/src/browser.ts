import { decodeImageData, DecodeResult, encodeImageData, ImageData } from './index';

// Convert Canvas to our ImageData format
export function canvasToImageData(canvas: HTMLCanvasElement): ImageData {
  const ctx = canvas.getContext('2d');
  if (!ctx) {
    throw new Error('Could not get 2D context from canvas');
  }

  const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
  return {
    data: new Uint8Array(imageData.data),
    width: canvas.width,
    height: canvas.height
  };
}

// Convert our ImageData format back to Canvas
export function imageDataToCanvas(imageData: ImageData, canvas?: HTMLCanvasElement): HTMLCanvasElement {
  const targetCanvas = canvas || document.createElement('canvas');
  targetCanvas.width = imageData.width;
  targetCanvas.height = imageData.height;

  const ctx = targetCanvas.getContext('2d');
  if (!ctx) {
    throw new Error('Could not get 2D context from canvas');
  }

  const canvasImageData = new ImageData(
    new Uint8ClampedArray(imageData.data),
    imageData.width,
    imageData.height
  );

  ctx.putImageData(canvasImageData, 0, 0);
  return targetCanvas;
}

// Encode text into a Canvas element
export function encodeCanvas(canvas: HTMLCanvasElement, text: string): HTMLCanvasElement {
  const imageData = canvasToImageData(canvas);
  const encodedData = encodeImageData(imageData, text);
  return imageDataToCanvas(encodedData, canvas);
}

// Decode text from a Canvas element
export function decodeCanvas(canvas: HTMLCanvasElement): DecodeResult | null {
  const imageData = canvasToImageData(canvas);
  return decodeImageData(imageData);
}

// Load image from URL into Canvas
export function loadImageToCanvas(imageUrl: string): Promise<HTMLCanvasElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = 'anonymous'; // Enable CORS for external images

    img.onload = () => {
      const canvas = document.createElement('canvas');
      canvas.width = img.width;
      canvas.height = img.height;

      const ctx = canvas.getContext('2d');
      if (!ctx) {
        reject(new Error('Could not get 2D context from canvas'));
        return;
      }

      ctx.drawImage(img, 0, 0);
      resolve(canvas);
    };

    img.onerror = () => {
      reject(new Error(`Failed to load image: ${imageUrl}`));
    };

    img.src = imageUrl;
  });
}

// Create visualization showing detected tiles
export function createVisualization(
  originalCanvas: HTMLCanvasElement,
  detectedTiles: Array<{x: number, y: number}>
): HTMLCanvasElement {
  const TILE_SIZE = 64;
  const visualCanvas = document.createElement('canvas');
  visualCanvas.width = originalCanvas.width;
  visualCanvas.height = originalCanvas.height;

  const ctx = visualCanvas.getContext('2d');
  if (!ctx) {
    throw new Error('Could not get 2D context from canvas');
  }

  // Draw original image
  ctx.drawImage(originalCanvas, 0, 0);

  // Draw red rectangles over detected tiles
  ctx.strokeStyle = 'red';
  ctx.fillStyle = 'rgba(255, 0, 0, 0.3)';
  ctx.lineWidth = 2;

  detectedTiles.forEach(({ x, y }) => {
    ctx.fillRect(x, y, TILE_SIZE, TILE_SIZE);
    ctx.strokeRect(x, y, TILE_SIZE, TILE_SIZE);
  });

  return visualCanvas;
}

// Complete workflow: encode image from URL and return Canvas
export async function encodeImageFromUrl(imageUrl: string, text: string): Promise<HTMLCanvasElement> {
  const canvas = await loadImageToCanvas(imageUrl);
  return encodeCanvas(canvas, text);
}

// Complete workflow: decode image from URL and return result with visualization
export async function decodeImageFromUrl(imageUrl: string): Promise<{
  result: DecodeResult | null,
  visualization?: HTMLCanvasElement
}> {
  const canvas = await loadImageToCanvas(imageUrl);
  const result = decodeCanvas(canvas);

  if (result && result.detectedTiles.length > 0) {
    const visualization = createVisualization(canvas, result.detectedTiles);
    return { result, visualization };
  }

  return { result };
}

// Export the core functions as well
export { decodeImageData, DecodeResult, encodeImageData, ImageData };
