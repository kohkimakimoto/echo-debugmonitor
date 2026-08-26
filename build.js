const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');

async function build() {
  // Create output directories if they don't exist
  const assetsOutdir = path.join(__dirname, 'resources/build/assets');

  if (!fs.existsSync(assetsOutdir)) {
    fs.mkdirSync(assetsOutdir, { recursive: true });
  }

  try {
    // Copy standalone TailwindCSS browser runtime for third-party monitors
    console.log('Copying TailwindCSS runtime...');
    fs.copyFileSync(
      'resources/assets/static/tailwindcss.js',
      'resources/build/assets/tailwindcss.js'
    );

    // Build JS file with esbuild
    console.log('Building JavaScript...');
    await esbuild.build({
      entryPoints: ['resources/assets/app.js'],
      bundle: true,
      target: 'es2020',
      outfile: 'resources/build/assets/app.js',
      minify: true,
    });

    console.log('✓ Build completed successfully');
  } catch (error) {
    console.error('Build failed:', error);
    process.exit(1);
  }
}

build();
