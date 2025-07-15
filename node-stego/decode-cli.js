#!/usr/bin/env node

const { decodePngFile, decodePngFileCropped } = require('./dist/node');
const fs = require('fs');
const path = require('path');

async function decodeCLI() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    console.log('❌ Error: Please provide a PNG file path');
    console.log('Usage: yarn run decode <filename.png>');
    console.log('Example: yarn run decode my-image.png');
    process.exit(1);
  }

  const filePath = args[0];

  // Check if file exists
  if (!fs.existsSync(filePath)) {
    console.log(`❌ Error: File '${filePath}' not found`);
    process.exit(1);
  }

  // Check if it's a PNG file
  if (!filePath.toLowerCase().endsWith('.png')) {
    console.log('⚠️  Warning: File does not have .png extension');
  }

  console.log(`🔍 Decoding steganographic message from: ${filePath}\n`);

  try {
    // First try regular decoding
    console.log('1️⃣ Attempting standard decoding...');
    const regularResult = await decodePngFile(filePath);

    if (regularResult) {
      console.log(`\n🎉 SUCCESS! Message decoded using standard method:`);
      console.log(`📄 Message: "${regularResult}"`);
      return;
    }

    console.log('   ❌ No message found with standard decoding\n');

    // If regular decoding fails, try cropped decoding
    console.log('2️⃣ Attempting cropped image decoding...');
    const croppedResult = await decodePngFileCropped(filePath);

    if (croppedResult) {
      console.log(`\n🎉 SUCCESS! Message decoded using cropped method:`);
      console.log(`📄 Message: "${croppedResult}"`);
      return;
    }

    console.log('   ❌ No message found with cropped decoding\n');
    console.log('🚫 No steganographic message could be decoded from this image.');
    console.log('💡 This could mean:');
    console.log('   • The image contains no hidden message');
    console.log('   • The message was encoded with a different tool');
    console.log('   • The image has been heavily compressed or modified');

  } catch (error) {
    console.error('❌ Error during decoding:', error.message);
    process.exit(1);
  }
}

// Run the CLI if this script is executed directly
if (require.main === module) {
  decodeCLI().catch(error => {
    console.error('❌ Unexpected error:', error.message);
    process.exit(1);
  });
}

module.exports = { decodeCLI };