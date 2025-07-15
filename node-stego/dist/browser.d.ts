import { decodeImageData, DecodeResult, encodeImageData, ImageData } from './index';
export declare function canvasToImageData(canvas: HTMLCanvasElement): ImageData;
export declare function imageDataToCanvas(imageData: ImageData, canvas?: HTMLCanvasElement): HTMLCanvasElement;
export declare function encodeCanvas(canvas: HTMLCanvasElement, text: string): HTMLCanvasElement;
export declare function decodeCanvas(canvas: HTMLCanvasElement): DecodeResult | null;
export declare function loadImageToCanvas(imageUrl: string): Promise<HTMLCanvasElement>;
export declare function createVisualization(originalCanvas: HTMLCanvasElement, detectedTiles: Array<{
    x: number;
    y: number;
}>): HTMLCanvasElement;
export declare function encodeImageFromUrl(imageUrl: string, text: string): Promise<HTMLCanvasElement>;
export declare function decodeImageFromUrl(imageUrl: string): Promise<{
    result: DecodeResult | null;
    visualization?: HTMLCanvasElement;
}>;
export { decodeImageData, DecodeResult, encodeImageData, ImageData };
//# sourceMappingURL=browser.d.ts.map