const sharp = require('sharp');
const { encodePngFile, decodePngFile, decodePngFileCropped, randomlyCropImage } = require('./dist/node');

async function createTestImage(outputPath, width = 512, height = 512) {
  // Create a simple test image with random colors
  const data = Buffer.alloc(width * height * 4);

  for (let i = 0; i < data.length; i += 4) {
    // Create a simple gradient pattern
    const x = Math.floor(i / 4) % width;
    const y = Math.floor(i / 4 / width);

    data[i] = (x * 255) / width;     // Red
    data[i + 1] = (y * 255) / height; // Green
    data[i + 2] = 128;               // Blue
    data[i + 3] = 255;               // Alpha
  }

  await sharp(data, {
    raw: {
      width,
      height,
      channels: 4
    }
  })
  .png()
  .toFile(outputPath);

  console.log(`✅ Created test image: ${outputPath}`);
}

async function runDemo() {
  console.log('🎨 Node-Stego Demo\n');

  const testImage = 'demo-input.png';
  const encodedImage = 'demo-encoded.png';
  const visualImage = 'demo-visualization.png';
  const secretMessage = 'This is a secret message hidden using steganography! 🔒 It is very long with lots of words and characters and stuff. It is very long with lots of words and characters and stuff. It is very long with lots of words and characters and stuff. It is very long with lots of words and characters and stuff.'

  try {
    // Step 1: Create a test image
    console.log('1️⃣ Creating test image...');
    await createTestImage(testImage);

    // Step 2: Encode the secret message
    console.log('\n2️⃣ Encoding secret message...');
    console.log(`Message: "${secretMessage}"`);
    await encodePngFile(testImage, encodedImage, secretMessage);

    // Step 3: Decode the message
    console.log('\n3️⃣ Decoding message...');
    const decodedMessage = await decodePngFile(encodedImage, visualImage);

    // Step 4: Verify results
    console.log('\n4️⃣ Results:');
    if (decodedMessage === secretMessage) {
      console.log('🎉 SUCCESS! Steganography worked perfectly!');
      console.log(`✅ Original:  "${secretMessage}"`);
      console.log(`✅ Decoded:   "${decodedMessage}"`);
      console.log(`\n📁 Files created:`);
      console.log(`   • ${testImage} - Original test image`);
      console.log(`   • ${encodedImage} - Image with hidden message`);
      console.log(`   • ${visualImage} - Visualization showing hidden data locations`);

      // Step 5: Test cropped decoding
      console.log('\n5️⃣ Testing cropped image decoding...');
      await runCroppedDemo(encodedImage, secretMessage);

    } else {
      console.log('❌ Message mismatch!');
      console.log(`Expected: "${secretMessage}"`);
      console.log(`Got:      "${decodedMessage}"`);
    }

  } catch (error) {
    console.error('❌ Demo failed:', error.message);
  }
}

async function runCroppedDemo(encodedImagePath, expectedMessage) {
  console.log('🔄 Testing cropped image decoding capabilities...\n');

  const numTests = 3; // Fewer tests for demo
  let successfulDecodes = 0;

  for (let i = 1; i <= numTests; i++) {
    try {
      console.log(`📐 Crop test ${i}/${numTests}: Randomly cropping and decoding...`);

      // Generate unique filenames for this test
      const croppedImagePath = `demo-cropped-${i}.png`;
      const croppedVisualizationPath = `demo-cropped-visualization-${i}.png`;

      // Randomly crop the encoded image
      const cropInfo = await randomlyCropImage(encodedImagePath, croppedImagePath, 200);

      // Try to decode the cropped image using the cropped decoder
      const decodedMessage = await decodePngFileCropped(croppedImagePath, croppedVisualizationPath);

      if (decodedMessage === expectedMessage) {
        console.log(`   ✅ Crop test ${i} PASSED! Message decoded from cropped image.`);
        console.log(`   📏 Crop: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
        successfulDecodes++;
      } else if (decodedMessage === null) {
        console.log(`   ⚠️  Crop test ${i}: No message found (crop may have removed too much data)`);
        console.log(`   📏 Crop: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
      } else {
        console.log(`   ❌ Crop test ${i} FAILED! Wrong message decoded.`);
        console.log(`   Expected: "${expectedMessage}"`);
        console.log(`   Decoded:  "${decodedMessage}"`);
      }

    } catch (error) {
      console.error(`   ❌ Crop test ${i} failed with error:`, error.message);
    }
  }

  // Test with one aggressive crop
  try {
    console.log(`\n🎯 Bonus: Testing aggressive crop scenario...`);
    const aggressiveCroppedPath = 'demo-aggressive-crop.png';
    const aggressiveVisualizationPath = 'demo-aggressive-visualization.png';

    // Smaller minimum size for more challenging crop
    const cropInfo = await randomlyCropImage(encodedImagePath, aggressiveCroppedPath, 150);
    const decodedMessage = await decodePngFileCropped(aggressiveCroppedPath, aggressiveVisualizationPath);

    if (decodedMessage === expectedMessage) {
      console.log(`   ✅ Aggressive crop PASSED! Robust decoding confirmed.`);
      console.log(`   📏 Crop: ${cropInfo.cropWidth}x${cropInfo.cropHeight} at (${cropInfo.cropX}, ${cropInfo.cropY})`);
      successfulDecodes++;
    } else if (decodedMessage === null) {
      console.log(`   ℹ️  Aggressive crop: No message found (expected for extreme crops)`);
    } else {
      console.log(`   ❌ Aggressive crop: Wrong message decoded!`);
    }
  } catch (error) {
    console.error(`   ❌ Aggressive crop test failed:`, error.message);
  }

  // Summary
  console.log(`\n📊 Cropped Decoding Summary:`);
  console.log(`   ✅ Successful decodes: ${successfulDecodes}/${numTests + 1}`);
  console.log(`   📈 Success rate: ${Math.round(successfulDecodes/(numTests + 1) * 100)}%`);

  if (successfulDecodes > 0) {
    console.log(`   🎉 Cropped decoding capability confirmed!`);
    console.log(`\n📁 Additional files created:`);
    for (let i = 1; i <= numTests; i++) {
      console.log(`   • demo-cropped-${i}.png - Randomly cropped test image ${i}`);
      console.log(`   • demo-cropped-visualization-${i}.png - Visualization for crop ${i}`);
    }
    console.log(`   • demo-aggressive-crop.png - Aggressively cropped test image`);
    console.log(`   • demo-aggressive-visualization.png - Aggressive crop visualization`);
  } else {
    console.log(`   ⚠️  No successful decodes from cropped images.`);
    console.log(`   This may occur with certain crop positions or image characteristics.`);
  }
}

// Run the demo
if (require.main === module) {
  runDemo().catch(console.error);
}

module.exports = { runDemo, createTestImage, runCroppedDemo };