export interface ImageData {
    data: Uint8Array;
    width: number;
    height: number;
}
export interface DecodeResult {
    message: string;
    votes: number;
    detectedTiles: Array<{
        x: number;
        y: number;
    }>;
}
export declare function textToData(text: string): Uint8Array;
export declare function dataToText(data: Uint8Array): string | null;
export declare function bytesToBits(data: Uint8Array): number[];
export declare function bitsToBytes(bits: number[]): Uint8Array;
export declare function encodeMessage(data: Uint8Array): Uint8Array;
export declare function decodeMessage(raw: Uint8Array): Uint8Array | null;
export declare function embedBitsInTile(tile: Uint8Array, width: number, height: number, bits: number[]): Uint8Array;
export declare function extractBitsFromTile(tile: Uint8Array, width: number, height: number, lengthBits: number): number[];
export declare function encodeImageData(imageData: ImageData, text: string): ImageData;
export declare function decodeImageData(imageData: ImageData): DecodeResult | null;
//# sourceMappingURL=simple-stego.d.ts.map