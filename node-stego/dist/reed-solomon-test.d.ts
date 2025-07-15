interface TestResult {
    testName: string;
    passed: boolean;
    input: string;
    encoded: Uint8Array | null;
    decoded: Uint8Array | null;
    decodedText: string | null;
    error?: string;
}
declare function runTest(testName: string, input: string): TestResult;
declare function testCorruption(originalData: Uint8Array, corruptionLevel?: number): TestResult;
export declare function runReedSolomonTests(): Promise<void>;
export { runTest, testCorruption };
//# sourceMappingURL=reed-solomon-test.d.ts.map