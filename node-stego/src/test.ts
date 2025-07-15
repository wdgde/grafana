import { decodePngFile, decodePngFileCropped, encodePngFile, randomlyCropImage } from './node';

async function runTests() {
  console.log('🚀 Starting steganography tests...\n');

  // Test data
  const inputImage = 'test-input.png';
  const outputImage = 'test-output.png';
  const visualizationImage = 'test-visualization.png';
  const testMessage = 'Hello, this is a secret message hidden in the image!';

  try {
    // Test encoding
    console.log('📝 Encoding message into image...');
    await encodePngFile(inputImage, outputImage, testMessage);

    // Test decoding
    console.log('🔍 Decoding message from image...');
    const decodedMessage = await decodePngFile(outputImage, visualizationImage);

    if (decodedMessage === testMessage) {
      console.log('✅ Test passed! Message successfully encoded and decoded.');
      console.log(`Original: "${testMessage}"`);
      console.log(`Decoded:  "${decodedMessage}"`);
    } else {
      console.log('❌ Test failed! Messages do not match.');
      console.log(`Original: "${testMessage}"`);
      console.log(`Decoded:  "${decodedMessage}"`);
    }

    // Test cropped decoding
    await runCroppedTests(outputImage, testMessage);

    // Test with dark/light images
    await runDarkLightImageTests();

  } catch (error: any) {
    console.error('❌ Test failed with error:', error.message);
    console.log('\n📋 To run this test, you need:');
    console.log('1. A test image named "test-input.png" in the project root');
    console.log('2. The image should be at least 64x64 pixels');
    console.log('3. Run: npm run build && npm test');
  }
}

async function runCroppedTests(encodedImagePath: string, expectedMessage: string) {
  console.log('\n🔄 Starting cropped image decoding tests...\n');

  const numTests = 5;
  let successfulDecodes = 0;

  for (let i = 1; i <= numTests; i++) {
    try {
      console.log(`📐 Test ${i}/${numTests}: Randomly cropping and decoding...`);

      // Generate unique filenames for this test
      const croppedImagePath = `test-cropped-${i}.png`;
      const croppedVisualizationPath = `test-cropped-visualization-${i}.png`;

      // Randomly crop the encoded image
      const cropInfo = await randomlyCropImage(encodedImagePath, croppedImagePath, 200);

      // Try to decode the cropped image using the cropped decoder
      const decodedMessage = await decodePngFileCropped(croppedImagePath, croppedVisualizationPath);

      if (decodedMessage === expectedMessage) {
        console.log(`   ✅ Crop test ${i} passed! Message decoded successfully from cropped image.`);
        console.log(`   📏 Crop details: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
        successfulDecodes++;
      } else if (decodedMessage === null) {
        console.log(`   ⚠️  Crop test ${i}: No message found in cropped image (this may be expected for some crops)`);
        console.log(`   📏 Crop details: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
      } else {
        console.log(`   ❌ Crop test ${i} failed! Wrong message decoded.`);
        console.log(`   Expected: "${expectedMessage}"`);
        console.log(`   Decoded:  "${decodedMessage}"`);
        console.log(`   📏 Crop details: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
      }

    } catch (error: any) {
      console.error(`   ❌ Crop test ${i} failed with error:`, error.message);
    }
  }

  // Test with a very aggressive crop that's likely to lose alignment
  try {
    console.log(`\n🎯 Bonus test: Testing with aggressive random crop...`);
    const aggressiveCroppedPath = 'test-aggressive-crop.png';
    const aggressiveVisualizationPath = 'test-aggressive-visualization.png';

    // Crop with smaller minimum size to make alignment loss more likely
    const cropInfo = await randomlyCropImage(encodedImagePath, aggressiveCroppedPath, 150);
    const decodedMessage = await decodePngFileCropped(aggressiveCroppedPath, aggressiveVisualizationPath);

    if (decodedMessage === expectedMessage) {
      console.log(`   ✅ Aggressive crop test passed! Message decoded despite challenging crop.`);
      console.log(`   📏 Crop details: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
      successfulDecodes++;
    } else if (decodedMessage === null) {
      console.log(`   ℹ️  Aggressive crop test: No message found (expected for very aggressive crops)`);
    } else {
      console.log(`   ❌ Aggressive crop test: Wrong message decoded!`);
      console.log(`   Expected: "${expectedMessage}"`);
      console.log(`   Decoded:  "${decodedMessage}"`);
    }
  } catch (error: any) {
    console.error(`   ❌ Aggressive crop test failed with error:`, error.message);
  }

  // Summary
  console.log(`\n📊 Cropped decoding test summary:`);
  console.log(`   Successful decodes: ${successfulDecodes}/${numTests + 1} tests`);

  if (successfulDecodes >= Math.ceil(numTests * 0.6)) {
    console.log(`   ✅ Overall cropped decoding test PASSED! (${Math.round(successfulDecodes/(numTests + 1) * 100)}% success rate)`);
  } else {
    console.log(`   ⚠️  Overall cropped decoding test had mixed results (${Math.round(successfulDecodes/(numTests + 1) * 100)}% success rate)`);
    console.log(`   This may be expected depending on crop positions and image content.`);
  }
}

async function runDarkLightImageTests() {
  console.log('\n🌓 Starting dark/light image tests...\n');

  const testMessage = 'Testing steganography with different image brightness levels!';
  const testImages = [
    { name: 'dark-test.png', type: 'dark' },
    { name: 'light-test.png', type: 'light' }
  ];

  let totalTests = 0;
  let passedTests = 0;

  for (const testImage of testImages) {
    try {
      console.log(`🖼️  Testing with ${testImage.type} image: ${testImage.name}`);

      // Test encoding and decoding
      const encodedPath = `${testImage.type}-encoded.png`;
      const visualizationPath = `${testImage.type}-visualization.png`;

      console.log(`   📝 Encoding message into ${testImage.type} image...`);
      await encodePngFile(testImage.name, encodedPath, testMessage);

      console.log(`   🔍 Decoding message from ${testImage.type} image...`);
      const decodedMessage = await decodePngFile(encodedPath, visualizationPath);

      if (decodedMessage === testMessage) {
        console.log(`   ✅ ${testImage.type} image test passed! Message successfully encoded and decoded.`);
        passedTests++;
      } else {
        console.log(`   ❌ ${testImage.type} image test failed! Messages do not match.`);
        console.log(`   Expected: "${testMessage}"`);
        console.log(`   Decoded:  "${decodedMessage}"`);
      }
      totalTests++;

      // Test cropped decoding with the encoded image
      await runCroppedTestsForImage(encodedPath, testMessage, testImage.type);

    } catch (error: any) {
      console.error(`   ❌ ${testImage.type} image test failed with error:`, error.message);
      if (testImage.name === 'light-test.png') {
        console.log(`   ℹ️  Note: ${testImage.name} may not exist. Create it to test light images.`);
      }
      totalTests++;
    }
  }

  console.log(`\n📊 Dark/Light image test summary:`);
  console.log(`   Passed: ${passedTests}/${totalTests} image types`);

  if (passedTests === totalTests) {
    console.log(`   ✅ All dark/light image tests PASSED!`);
  } else {
    console.log(`   ⚠️  Some dark/light image tests failed or couldn't run.`);
  }
}

async function runCroppedTestsForImage(encodedImagePath: string, expectedMessage: string, imageType: string) {
  console.log(`   🔄 Testing cropped decoding for ${imageType} image...`);

  const numTests = 3; // Reduced number for dark/light tests
  let successfulDecodes = 0;

  for (let i = 1; i <= numTests; i++) {
    try {
      const croppedImagePath = `${imageType}-cropped-${i}.png`;
      const croppedVisualizationPath = `${imageType}-cropped-visualization-${i}.png`;

      // Randomly crop the encoded image
      const cropInfo = await randomlyCropImage(encodedImagePath, croppedImagePath, 200);

      // Try to decode the cropped image using the cropped decoder
      const decodedMessage = await decodePngFileCropped(croppedImagePath, croppedVisualizationPath);

      if (decodedMessage === expectedMessage) {
        console.log(`     ✅ ${imageType} crop test ${i} passed! Message decoded from cropped image.`);
        successfulDecodes++;
      } else if (decodedMessage === null) {
        console.log(`     ⚠️  ${imageType} crop test ${i}: No message found in cropped image`);
      } else {
        console.log(`     ❌ ${imageType} crop test ${i} failed! Wrong message decoded.`);
      }

    } catch (error: any) {
      console.error(`     ❌ ${imageType} crop test ${i} failed with error:`, error.message);
    }
  }

  console.log(`   📊 ${imageType} cropped decoding: ${successfulDecodes}/${numTests} successful`);
}

export { runTests };

// Run tests if this module is executed directly
if (require.main === module) {
  runTests().catch(console.error);
}
