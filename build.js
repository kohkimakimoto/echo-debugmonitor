const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');

async function build() {
  // Create output directory if it doesn't exist
  const outdir = path.join(__dirname, 'public/build');
  if (!fs.existsSync(outdir)) {
    fs.mkdirSync(outdir, { recursive: true });
  }

  try {
    // Build CSS and JS files with esbuild
    await esbuild.build({
      entryPoints: [
        'resources/assets/app.css',
        'resources/assets/app.js'
      ],
      bundle: true,
      outdir: 'resources/public',
      minify: true,
      conditions: ['style'],
      loader: {
        '.css': 'css',
      },
    });

    console.log('✓ Build completed successfully');

  } catch (error) {
    console.error('Build failed:', error);
    process.exit(1);
  }
}

build();
