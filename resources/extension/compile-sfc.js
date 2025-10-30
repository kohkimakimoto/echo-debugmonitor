const fs = require('fs');
const path = require('path');
const { glob } = require('glob');
const crypto = require('crypto');

// Namespace -> scopeKey mapping to detect collisions
const scopeKeyMap = new Map();

/**
 * Generate a unique scope key for a given namespace
 */
function generateScopeKey(namespace) {
  // If scopeKey already exists for this namespace, return it
  if (scopeKeyMap.has(namespace)) {
    return scopeKeyMap.get(namespace);
  }

  // Hash the namespace using MD5
  const hash = crypto.createHash('md5')
    .update(namespace)
    .digest('hex')
    .substring(0, 8);

  let scopeKey = `scope_${hash}`;
  let counter = 0;

  // Collision check: if the same scopeKey is already assigned to a different namespace
  while ([...scopeKeyMap.entries()].some(([ns, k]) => k === scopeKey && ns !== namespace)) {
    counter++;
    // In case of collision, recalculate by appending a counter to the hash
    const modifiedNamespace = `${namespace}_${counter}`;
    const newHash = crypto.createHash('md5')
      .update(modifiedNamespace)
      .digest('hex')
      .substring(0, 8);
    scopeKey = `scope_${newHash}`;
  }

  // Save the mapping
  scopeKeyMap.set(namespace, scopeKey);

  return scopeKey;
}

/**
 * Transform $data() to Alpine.data()
 */
function transformAlpineComponents(scriptContent, scopeKey) {
  const componentNames = [];

  // Pattern: const variableName = $data(
  const pattern = /(const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*\$data\s*\(/g;

  let match;
  while ((match = pattern.exec(scriptContent)) !== null) {
    const varName = match[2];
    componentNames.push(varName);
  }

  // Transform the original code
  let transformedCode = scriptContent;

  if (componentNames.length > 0) {
    // Check if Alpine import already exists
    const hasAlpineImport = /import\s+.*\bAlpine\b.*\s+from\s+['"]alpinejs['"]/.test(transformedCode) ||
      /import\s+['"]alpinejs['"]/.test(transformedCode);

    // Add Alpine.data() calls
    const alpineDataCalls = componentNames.map(varName => {
      const scopedName = `${scopeKey}_${varName}`;
      return `Alpine.data("${scopedName}", ${varName});`;
    }).join('\n');

    // Content to add at the beginning of the code
    let prefixCode = '';

    // Add Alpine import (only if it doesn't already exist)
    if (!hasAlpineImport) {
      prefixCode += `import Alpine from 'alpinejs';\n\n`;
    }

    transformedCode = prefixCode + transformedCode + '\n\n' + alpineDataCalls;
  }

  return {
    code: transformedCode,
    componentNames
  };
}

/**
 * Scope CSS class names
 */
function transformCssWithScope(cssContent, scopeKey) {
  // Add scope class to class selectors
  // .example -> .example.scope_abc123
  const classPattern = /\.([a-zA-Z_][\w-]*)/g;

  return cssContent.replace(classPattern, (match, className) => {
    return `.${className}.${scopeKey}`;
  });
}

/**
 * Process CSS extraction from HTML templates
 * Finds @import 'extract-sfc-css:pattern' and replaces it with imports to generated temp files
 */
function processCssExtraction(inputFile, outputFile) {
  const content = fs.readFileSync(inputFile, 'utf8');

  // Find extract-sfc-css: pseudo-import statements
  const templateCssRegex = /@import\s+['"]extract-sfc-css:([^'"]+)['"]\s*;?/g;

  const transformedContent = content.replace(templateCssRegex, (match, pseudoImport) => {
    const fullPattern = path.resolve(path.dirname(inputFile), pseudoImport);

    // Search for HTML files matching the glob pattern
    const htmlFiles = glob.sync(fullPattern);
    let imports = [];

    htmlFiles.forEach(htmlFile => {
      const htmlContent = fs.readFileSync(htmlFile, 'utf8');

      // Extract <style extract> tags
      const extractPattern = /<style[^>]*\bextract\b[^>]*>([\s\S]*?)<\/style>/i;
      const styleMatch = extractPattern.exec(htmlContent);

      if (styleMatch && styleMatch[1].trim()) {
        // Generate namespace from an absolute file path
        const namespace = path.resolve(htmlFile);
        const scopeKey = generateScopeKey(namespace);

        // Transform CSS with scope class
        const transformedCss = transformCssWithScope(styleMatch[1].trim(), scopeKey);

        const htmlBasename = path.basename(htmlFile, '.html');
        const tempFileName = `${htmlBasename}.temp.css`;
        const tempFilePath = path.resolve(path.dirname(htmlFile), tempFileName);

        // Create temporary file for template CSS
        fs.writeFileSync(tempFilePath, transformedCss);

        const relativePath = path.relative(path.dirname(outputFile), tempFilePath);
        imports.push(`@import './${relativePath}';`);
      }
    });

    // Return import statements
    if (imports.length > 0) {
      return imports.join('\n');
    } else {
      return '/* No template styles found */';
    }
  });

  // Write the transformed content to output file
  fs.writeFileSync(outputFile, transformedContent);
}

/**
 * Process JS extraction from HTML templates
 * Finds import 'extract-sfc-script:pattern' and replaces it with imports to generated temp files
 */
function processScriptExtraction(inputFile, outputFile) {
  const content = fs.readFileSync(inputFile, 'utf8');

  // Find extract-sfc-script: pseudo-import statements
  const templateScriptRegex = /import\s+['"]extract-sfc-script:([^'"]+)['"]\s*;?/g;

  const transformedContent = content.replace(templateScriptRegex, (match, pseudoImport) => {
    const fullPattern = path.resolve(path.dirname(inputFile), pseudoImport);

    // Search for HTML files matching the glob pattern
    const htmlFiles = glob.sync(fullPattern);
    let imports = [];

    htmlFiles.forEach(htmlFile => {
      const htmlContent = fs.readFileSync(htmlFile, 'utf8');

      // Extract <script extract> tags
      const extractPattern = /<script[^>]*\bextract\b[^>]*>([\s\S]*?)<\/script>/i;
      const scriptMatch = extractPattern.exec(htmlContent);

      if (scriptMatch && scriptMatch[1].trim()) {
        const extractedScript = scriptMatch[1].trim();

        // Generate namespace from absolute file path
        const namespace = path.resolve(htmlFile);
        const scopeKey = generateScopeKey(namespace);

        // Transform $data() to Alpine.data()
        const { code: transformedScript } = transformAlpineComponents(
          extractedScript,
          scopeKey
        );

        const htmlBasename = path.basename(htmlFile, '.html');
        const tempFileName = `${htmlBasename}.temp.js`;
        const tempFilePath = path.resolve(path.dirname(htmlFile), tempFileName);

        // Create temporary file for template script
        fs.writeFileSync(tempFilePath, transformedScript);

        const relativePath = path.relative(path.dirname(outputFile), tempFilePath);
        imports.push(`import './${relativePath}';`);
      }
    });

    // Return import statements
    if (imports.length > 0) {
      return imports.join('\n');
    } else {
      return '// No template scripts found';
    }
  });

  // Write the transformed content to output file
  fs.writeFileSync(outputFile, transformedContent);
}

/**
 * Add scope class to HTML class attributes
 */
function addScopeClassToAttributes(content, scopeKey) {
  // Handle both class="..." and class='...'
  // Exclude Alpine.js dynamic attributes (:class, x-bind:class)

  // Double-quote class attributes
  // Ensure no : or x-bind: prefix
  const doubleQuotePattern = /(?:^|[^:\w-])class="([^"]*)"/g;
  content = content.replace(doubleQuotePattern, (match, classValue) => {
    // Skip if scope class is already included
    if (classValue.includes(scopeKey)) {
      return match;
    }

    // Add scope class to the end of the class value
    const newClassValue = classValue === '' ? scopeKey : classValue + ' ' + scopeKey;

    // Replace the entire matched string (including the preceding character)
    return match.replace(`class="${classValue}"`, `class="${newClassValue}"`);
  });

  // Single-quote class attributes
  const singleQuotePattern = /(?:^|[^:\w-])class='([^']*)'/g;
  content = content.replace(singleQuotePattern, (match, classValue) => {
    // Skip if scope class is already included
    if (classValue.includes(scopeKey)) {
      return match;
    }

    // Add scope class to the end of the class value
    const newClassValue = classValue === '' ? scopeKey : classValue + ' ' + scopeKey;

    // Replace the entire matched string (including the preceding character)
    return match.replace(`class='${classValue}'`, `class='${newClassValue}'`);
  });

  return content;
}

/**
 * Process HTML files
 * Remove <style extract> and <script extract> tags, and transform class and x-data attributes
 */
function processHtmlFile(inputFile, outputFile) {
  // Generate namespace from an absolute file path
  const namespace = path.resolve(inputFile);
  const scopeKey = generateScopeKey(namespace);

  // Get template content
  let content = fs.readFileSync(inputFile, 'utf8');

  // Check if <style extract> tag exists
  const extractCssPattern = /<style[^>]*\bextract\b[^>]*>[\s\S]*?<\/style>/i;
  const hasExtractStyle = extractCssPattern.test(content);

  // Extract component definitions from <script extract> tags
  const extractScriptPattern = /<script[^>]*\bextract\b[^>]*>([\s\S]*?)<\/script>/gi;
  const componentNames = [];

  let scriptMatch;
  while ((scriptMatch = extractScriptPattern.exec(content)) !== null) {
    const scriptContent = scriptMatch[1];

    // Extract component names defined in the script tag
    const componentDefPattern = /(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*\$data\s*\(/g;
    let compMatch;

    while ((compMatch = componentDefPattern.exec(scriptContent)) !== null) {
      componentNames.push(compMatch[1]);
    }
  }

  // Transform class attributes if <style extract> exists
  if (hasExtractStyle) {
    content = addScopeClassToAttributes(content, scopeKey);
  }

  // Transform x-data attributes if component names are defined
  componentNames.forEach(componentName => {
    const scopedName = `${scopeKey}_${componentName}`;

    // Replace x-data="componentName" or x-data="componentName(...)" with
    // x-data="scopeKey_componentName" or x-data="scopeKey_componentName(...)"
    const xDataPattern = new RegExp(`\\bx-data="${componentName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}(\\([^)]*\\))?`, 'g');
    content = content.replace(xDataPattern, `x-data="${scopedName}$1`);
  });

  // Remove <style extract> tags
  content = content.replace(extractCssPattern, '');

  // Remove <script extract> tags
  content = content.replace(extractScriptPattern, '');

  // Write the transformed content to an output file
  fs.writeFileSync(outputFile, content);
}

/**
 * Clean up temporary files
 */
function cleanTempFiles(patterns) {
  if (!Array.isArray(patterns)) {
    patterns = [patterns];
  }

  let deletedCount = 0;

  patterns.forEach(pattern => {
    const matchingFiles = glob.sync(pattern);

    matchingFiles.forEach(filePath => {
      try {
        if (fs.existsSync(filePath)) {
          fs.unlinkSync(filePath);
          deletedCount++;
        }
      } catch (error) {
        console.warn(`Failed to delete temporary file ${filePath}:`, error.message);
      }
    });
  });

  if (deletedCount > 0) {
    console.log(`Cleaned up ${deletedCount} temporary file(s)`);
  }
}

module.exports = {
  processCssExtraction,
  processScriptExtraction,
  processHtmlFile,
  cleanTempFiles,
};
