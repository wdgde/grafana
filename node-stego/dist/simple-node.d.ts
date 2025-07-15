import { decodeImageData, DecodeResult, encodeImageData, ImageData } from './simple-stego';
export declare function encodePngFile(inputPath: string, outputPath: string, text: string): Promise<void>;
export declare function decodePngFile(inputPath: string, visualizationPath?: string): Promise<string | null>;
export { decodeImageData, DecodeResult, encodeImageData, ImageData };
//# sourceMappingURL=simple-node.d.ts.map