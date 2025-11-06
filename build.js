const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');
const { execSync } = require('child_process');
const { processCssExtraction, processScriptExtraction, processHtmlFile, cleanTempFiles } = require('./resources/extensions/sfc-build');
const { glob } = require('glob');
const { minify } = require('html-minifier-next');

async function build() {
  // Create output directories if they don't exist
  const assetsOutdir = path.join(__dirname, 'resources/build/assets');
  const viewsOutdir = path.join(__dirname, 'resources/build/views');

  if (!fs.existsSync(assetsOutdir)) {
    fs.mkdirSync(assetsOutdir, { recursive: true });
  }
  if (!fs.existsSync(viewsOutdir)) {
    fs.mkdirSync(viewsOutdir, { recursive: true });
  }

  try {
    // Process SFC CSS extraction
    console.log('Extracting SFC CSS...');
    processCssExtraction(
      'resources/assets/app.css',
      'resources/assets/app.temp.css'
    );

    // Process CSS with TailwindCSS CLI
    console.log('Building CSS with TailwindCSS...');
    execSync(
      'npx @tailwindcss/cli -i resources/assets/app.temp.css -o resources/build/assets/app.css --minify',
      { stdio: 'inherit' }
    );

    // Process SFC JS extraction
    console.log('Extracting SFC JavaScript...');
    processScriptExtraction(
      'resources/assets/app.js',
      'resources/assets/app.temp.js'
    );

    // Build JS file with esbuild
    console.log('Building JavaScript...');
    await esbuild.build({
      entryPoints: ['resources/assets/app.temp.js'],
      bundle: true,
      target: 'es2020',
      outfile: 'resources/build/assets/app.js',
      minify: true,
    });

    // Process templates
    console.log('Processing HTML templates...');
    const htmlFiles = glob.sync('resources/views/**/*.html');
    for (const htmlFile of htmlFiles) {
      const relativePath = path.relative('resources/views', htmlFile);
      const outputPath = path.join('resources/build/views', relativePath);
      const outputDir = path.dirname(outputPath);

      // Create subdirectories if needed
      if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true });
      }

      processHtmlFile(htmlFile, outputPath);

      // Minify the processed HTML file
      const htmlContent = fs.readFileSync(outputPath, 'utf8');
      const minifiedHtml = await minify(htmlContent, {
        caseSensitive: true,
        collapseWhitespace: true,
        ignoreCustomFragments: [/{{[\s\S]*?}}/, /{%[\s\S]*?%}/, /{#[\s\S]*?#}/],
        keepClosingSlash: true,
        minifyCSS: true,
        removeComments: true,
        minifyJS: true,
      });
      fs.writeFileSync(outputPath, minifiedHtml);
    }

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
