"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runReedSolomonTests = runReedSolomonTests;
exports.runTest = runTest;
exports.testCorruption = testCorruption;
const index_1 = require("./index");
function runTest(testName, input) {
    const result = {
        testName,
        passed: false,
        input,
        encoded: null,
        decoded: null,
        decodedText: null
    };
    try {
        // Step 1: Convert text to data
        const inputData = (0, index_1.textToData)(input);
        console.log(`📝 Input: "${input}" (${inputData.length} bytes)`);
        // Step 2: Encode with Reed-Solomon
        const encoded = (0, index_1.encodePayloadWithRS)(inputData);
        result.encoded = encoded;
        console.log(`🔐 Encoded length: ${encoded.length} bytes`);
        // Step 3: Decode with Reed-Solomon
        const decoded = (0, index_1.decodePayloadWithRS)(encoded);
        result.decoded = decoded;
        if (!decoded) {
            result.error = 'Decoding returned null';
            return result;
        }
        console.log(`🔓 Decoded length: ${decoded.length} bytes`);
        // Step 4: Convert back to text
        const decodedText = (0, index_1.dataToText)(decoded);
        result.decodedText = decodedText;
        if (!decodedText) {
            result.error = 'Failed to convert decoded data to text';
            return result;
        }
        // Step 5: Compare
        if (decodedText === input) {
            result.passed = true;
            console.log(`✅ SUCCESS: "${input}" → "${decodedText}"`);
        }
        else {
            result.error = `Text mismatch: expected "${input}", got "${decodedText}"`;
            console.log(`❌ FAILED: Expected "${input}", got "${decodedText}"`);
        }
    }
    catch (error) {
        result.error = `Exception: ${error}`;
        console.log(`💥 ERROR: ${error}`);
    }
    return result;
}
function testCorruption(originalData, corruptionLevel = 1) {
    const result = {
        testName: `Corruption Test (${corruptionLevel} bytes)`,
        passed: false,
        input: `corrupted-${corruptionLevel}`,
        encoded: originalData,
        decoded: null,
        decodedText: null
    };
    try {
        // Create a copy and corrupt it
        const corrupted = new Uint8Array(originalData);
        // Corrupt random bytes
        for (let i = 0; i < corruptionLevel; i++) {
            const randomIndex = Math.floor(Math.random() * corrupted.length);
            corrupted[randomIndex] = corrupted[randomIndex] ^ 0xFF; // Flip all bits
        }
        console.log(`🔥 Corrupted ${corruptionLevel} byte(s) at random positions`);
        // Try to decode corrupted data
        const decoded = (0, index_1.decodePayloadWithRS)(corrupted);
        result.decoded = decoded;
        if (decoded) {
            const decodedText = (0, index_1.dataToText)(decoded);
            result.decodedText = decodedText;
            result.passed = true;
            console.log(`✅ Reed-Solomon successfully recovered from corruption: "${decodedText}"`);
        }
        else {
            console.log(`❌ Reed-Solomon could not recover from ${corruptionLevel} byte(s) of corruption`);
        }
    }
    catch (error) {
        result.error = `Exception: ${error}`;
        console.log(`💥 ERROR: ${error}`);
    }
    return result;
}
async function runReedSolomonTests() {
    console.log('🧪 Reed-Solomon Encode/Decode Tests\n');
    const testCases = [
        'Hello World!',
        'A',
        'Short message',
        'This is a longer message to test Reed-Solomon encoding and decoding capabilities.',
        'Special chars: !@#$%^&*()[]{}',
        'Unicode: 🌍🔐📝✅❌💡🎯',
        'Numbers: 1234567890',
        'Mixed: Hello 🌍! Test #123 with special chars @2024',
        '', // Empty string edge case
        'x'.repeat(50), // Medium length
        'Lorem ipsum dolor sit amet, consectetur adipiscing elit.' // Longer text
    ];
    const results = [];
    // Run basic encode/decode tests
    console.log('='.repeat(60));
    console.log('📋 BASIC ENCODE/DECODE TESTS');
    console.log('='.repeat(60));
    for (let i = 0; i < testCases.length; i++) {
        const testCase = testCases[i];
        console.log(`\n🧪 Test ${i + 1}/${testCases.length}: ${testCase.length === 0 ? '(empty string)' : testCase.substring(0, 50)}${testCase.length > 50 ? '...' : ''}`);
        console.log('-'.repeat(40));
        const result = runTest(`Test ${i + 1}`, testCase);
        results.push(result);
    }
    // Test corruption recovery (if we have successful encoding)
    console.log('\n' + '='.repeat(60));
    console.log('🔥 CORRUPTION RECOVERY TESTS');
    console.log('='.repeat(60));
    const successfulResult = results.find(r => r.passed && r.encoded);
    if (successfulResult && successfulResult.encoded) {
        console.log(`\nUsing successful encoding from: "${successfulResult.input}"`);
        // Test different levels of corruption
        for (let corruptionLevel = 1; corruptionLevel <= 5; corruptionLevel++) {
            console.log(`\n🔥 Testing ${corruptionLevel} byte(s) of corruption:`);
            console.log('-'.repeat(40));
            const corruptionResult = testCorruption(successfulResult.encoded, corruptionLevel);
            results.push(corruptionResult);
        }
    }
    // Summary
    console.log('\n' + '='.repeat(60));
    console.log('📊 TEST SUMMARY');
    console.log('='.repeat(60));
    const passed = results.filter(r => r.passed).length;
    const total = results.length;
    console.log(`\n📈 Results: ${passed}/${total} tests passed`);
    if (passed === total) {
        console.log('🎉 ALL TESTS PASSED! Reed-Solomon implementation is working correctly.');
    }
    else {
        console.log(`⚠️  ${total - passed} test(s) failed. Here are the failures:`);
        results.filter(r => !r.passed).forEach(result => {
            console.log(`  ❌ ${result.testName}: ${result.error || 'Unknown error'}`);
        });
    }
    // Detailed breakdown
    console.log('\n📋 Detailed Results:');
    results.forEach(result => {
        const status = result.passed ? '✅' : '❌';
        const input = result.input.length > 30 ? result.input.substring(0, 30) + '...' : result.input;
        console.log(`  ${status} ${result.testName}: "${input}"`);
    });
    console.log('\n🏁 Reed-Solomon testing complete!');
}
//# sourceMappingURL=reed-solomon-test.js.map