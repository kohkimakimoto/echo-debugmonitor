const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');
const { execSync } = require('child_process');
const { processCssExtraction, cleanTempFiles } = require('./resources/extension/compile-sfc');

async function build() {
  // Create output directory if it doesn't exist
  const outdir = path.join(__dirname, 'resources/build/assets');
  if (!fs.existsSync(outdir)) {
    fs.mkdirSync(outdir, { recursive: true });
  }

  try {
    // Process SFC CSS extraction
    console.log('Extracting SFC CSS...');
    processCssExtraction(
      'resources/assets/app.css',
      'resources/assets/app.temp.css',
      'resources/views',
    );

    // Process CSS with TailwindCSS CLI
    console.log('Building CSS with TailwindCSS...');
    execSync(
      'npx tailwindcss -i resources/assets/app.temp.css -o resources/build/assets/app.css --minify',
      { stdio: 'inherit' }
    );

    // Build JS file with esbuild
    console.log('Building JavaScript...');
    await esbuild.build({
      entryPoints: ['resources/assets/app.js'],
      bundle: true,
      target: 'es2020',
      outdir: 'resources/build/assets',
      minify: true,
    });

    console.log('✓ Build completed successfully');
  } catch (error) {
    console.error('Build failed:', error);
    process.exit(1);
  } finally {
    // Clean up temporary files
    console.log('Cleaning up temporary files...');
    cleanTempFiles([
      'resources/**/*.temp.css',
      'resources/**/*.temp.js'
    ]);
  }
}

build();
