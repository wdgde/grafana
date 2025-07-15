#!/usr/bin/env node

const { encodePngFile } = require('./dist/node');
const fs = require('fs');
const path = require('path');

async function encodeCLI() {
  const args = process.argv.slice(2);

  if (args.length < 2) {
    console.log('❌ Error: Please provide input file and message');
    console.log('Usage:');
    console.log('  yarn run encode <input.png> <message> [output.png]');
    console.log('');
    console.log('Examples:');
    console.log('  yarn run encode photo.png "Secret message!"');
    console.log('  yarn run encode photo.png "Secret message!" encoded-photo.png');
    console.log('');
    console.log('If no output file is specified, it will be auto-generated.');
    process.exit(1);
  }

  const inputPath = args[0];
  const message = args[1];
  let outputPath = args[2];

  // Check if input file exists
  if (!fs.existsSync(inputPath)) {
    console.log(`❌ Error: Input file '${inputPath}' not found`);
    process.exit(1);
  }

  // Check if it's a PNG file
  if (!inputPath.toLowerCase().endsWith('.png')) {
    console.log('⚠️  Warning: Input file does not have .png extension');
  }

  // Auto-generate output path if not provided
  if (!outputPath) {
    const parsedPath = path.parse(inputPath);
    outputPath = path.join(parsedPath.dir, `${parsedPath.name}-encoded${parsedPath.ext}`);
  }

  // Check if output file already exists
  if (fs.existsSync(outputPath)) {
    console.log(`⚠️  Warning: Output file '${outputPath}' already exists and will be overwritten`);
  }

  // Validate message
  if (message.length === 0) {
    console.log('❌ Error: Message cannot be empty');
    process.exit(1);
  }

  if (message.length > 1000) {
    console.log('⚠️  Warning: Very long message may not encode properly in small images');
  }

  console.log(`🔒 Encoding steganographic message into PNG...\n`);
  console.log(`📁 Input:   ${inputPath}`);
  console.log(`📁 Output:  ${outputPath}`);
  console.log(`📄 Message: "${message}" (${message.length} characters)\n`);

  try {
    console.log('🎨 Processing image and embedding message...');
    await encodePngFile(inputPath, outputPath, message);

    console.log(`\n🎉 SUCCESS! Message encoded successfully!`);
    console.log(`💾 Encoded image saved to: ${outputPath}`);
    console.log(`\n💡 To decode the message later, use:`);
    console.log(`   yarn run decode ${outputPath}`);

  } catch (error) {
    console.error('❌ Error during encoding:', error.message);

    // Provide helpful error messages for common issues
    if (error.message.includes('too large')) {
      console.log('\n💡 Tips to fix this:');
      console.log('   • Use a larger image');
      console.log('   • Use a shorter message');
      console.log('   • Try an image with dimensions that are multiples of 64');
    } else if (error.message.includes('format') || error.message.includes('PNG')) {
      console.log('\n💡 Make sure the input file is a valid PNG image');
    }

    process.exit(1);
  }
}

// Run the CLI if this script is executed directly
if (require.main === module) {
  encodeCLI().catch(error => {
    console.error('❌ Unexpected error:', error.message);
    process.exit(1);
  });
}

module.exports = { encodeCLI };