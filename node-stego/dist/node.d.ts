import { decodeImageData, decodeImageDataCropped, DecodeResult, encodeImageData, ImageData } from './index';
export declare function encodePngFile(inputPath: string, outputPath: string, text: string): Promise<void>;
export declare function decodePngFile(inputPath: string, visualizationPath?: string): Promise<string | null>;
export { decodeImageData, decodeImageDataCropped, DecodeResult, encodeImageData, ImageData };
export declare function decodePngFileCropped(inputPath: string, visualizationPath?: string): Promise<string | null>;
export declare function randomlyCropImage(inputPath: string, outputPath: string, minCropSize?: number): Promise<{
    cropX: number;
    cropY: number;
    cropWidth: number;
    cropHeight: number;
}>;
//# sourceMappingURL=node.d.ts.map