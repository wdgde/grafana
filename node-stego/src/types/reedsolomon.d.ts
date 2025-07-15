declare module 'reedsolomon' {
  export class ReedSolomonEncoder {
    constructor(field: GenericGF);
    encode(toEncode: Int32Array, ecBytes: number): void;
  }

  export class ReedSolomonDecoder {
    constructor(field: GenericGF);
    decode(received: Int32Array, twoS: number): void;
  }

  export interface GenericGF {
    // Add any methods you use from GenericGF
  }

  export const GenericGF: {
    AZTEC_DATA_8(): GenericGF;
    // Add other static methods as needed
  };
}