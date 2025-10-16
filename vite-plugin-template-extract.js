import { readFileSync, readdirSync, existsSync } from 'fs';
import { resolve, join, relative, dirname, isAbsolute, normalize } from 'path';
import { minimatch } from 'minimatch';
import MagicString from 'magic-string';

export default function templateExtract(options = {}) {
  const {
    watchFiles = true,
    minimatchOptions = {},
  } = options;

  let server;

  /**
   * Match file path against glob pattern using minimatch
   */
  function matchPattern(filePath, pattern) {
    return minimatch(filePath, pattern, {
      dot: true,
      matchBase: false,
      ...minimatchOptions,
    });
  }

  /**
   * Split path into base directory and glob pattern
   */
  function splitGlobPath(fullPath) {
    const normalized = normalize(fullPath).replace(/\\/g, '/');
    const globChars = ['*', '?', '[', '{'];
    let globStartIndex = -1;

    for (const char of globChars) {
      const index = normalized.indexOf(char);
      if (index !== -1 && (globStartIndex === -1 || index < globStartIndex)) {
        globStartIndex = index;
      }
    }

    if (globStartIndex === -1) {
      return { base: normalized, pattern: '**/*' };
    }

    const lastSepBeforeGlob = normalized.lastIndexOf('/', globStartIndex);

    if (lastSepBeforeGlob === -1) {
      return { base: process.cwd(), pattern: normalized };
    }

    const base = normalized.slice(0, lastSepBeforeGlob);
    const pattern = normalized.slice(lastSepBeforeGlob + 1);

    return { base, pattern };
  }

  /**
   * Recursively find template files matching the pattern
   */
  function findTemplateFiles(baseDir, pattern) {
    if (!existsSync(baseDir)) {
      return [];
    }

    const files = [];

    function traverse(currentDir) {
      try {
        const items = readdirSync(currentDir, { withFileTypes: true });

        for (const item of items) {
          const fullPath = join(currentDir, item.name);
          const relativePath = relative(baseDir, fullPath).replace(/\\/g, '/');

          if (item.isDirectory()) {
            traverse(fullPath);
          } else if (matchPattern(relativePath, pattern)) {
            files.push(fullPath);
          }
        }
      } catch (error) {
        // Silently skip directories we cannot read
      }
    }

    traverse(baseDir);
    return files;
  }

  /**
   * Extract content from <script data-extract> tags
   */
  function extractScripts(content) {
    const scripts = [];
    const scriptRegex = /<script\s[^>]*\bdata-extract\b[^>]*>([\s\S]*?)<\/script>/gi;
    let match;

    while ((match = scriptRegex.exec(content)) !== null) {
      scripts.push(match[1].trim());
    }

    return scripts;
  }

  /**
   * Extract content from <style data-extract> tags
   */
  function extractStyles(content) {
    const styles = [];
    const styleRegex = /<style\s[^>]*\bdata-extract\b[^>]*>([\s\S]*?)<\/style>/gi;
    let match;

    while ((match = styleRegex.exec(content)) !== null) {
      styles.push(match[1].trim());
    }

    return styles;
  }

  /**
   * Check if this is an extract request
   */
  function isExtractRequest(id) {
    return id.endsWith('.extract.js') || id.endsWith('.extract.css');
  }

  /**
   * Parse extract request
   */
  function parseExtractRequest(id) {
    if (id.endsWith('.extract.js')) {
      return {
        filename: id.replace(/\.extract\.js$/, ''),
        type: 'script'
      };
    }

    if (id.endsWith('.extract.css')) {
      return {
        filename: id.replace(/\.extract\.css$/, ''),
        type: 'style'
      };
    }

    return null;
  }

  /**
   * Handle file changes
   */
  function handleFileChange(filePath) {
    if (server) {
      const moduleGraph = server.moduleGraph;

      // Invalidate the .extract.js and .extract.css modules for this HTML file
      const extractJsPath = `${filePath}.extract.js`;
      const extractCssPath = `${filePath}.extract.css`;

      const jsModule = moduleGraph.getModuleById(extractJsPath);
      const cssModule = moduleGraph.getModuleById(extractCssPath);

      if (jsModule) {
        moduleGraph.invalidateModule(jsModule);
      }

      if (cssModule) {
        moduleGraph.invalidateModule(cssModule);
      }

      // Send HMR update
      server.ws.send({
        type: 'full-reload',
        path: '*'
      });
    }
  }

  /**
   * Generate import statements for template files
   */
  function generateImportStatements(base, pattern) {
    const templateFiles = findTemplateFiles(base, pattern);
    const imports = [];

    for (const filePath of templateFiles) {
      try {
        const content = readFileSync(filePath, 'utf-8');
        const scripts = extractScripts(content);
        const styles = extractStyles(content);

        if (scripts.length > 0) {
          imports.push(`import '${filePath}.extract.js';`);
        }

        if (styles.length > 0) {
          imports.push(`import '${filePath}.extract.css';`);
        }
      } catch (error) {
        // Silently skip files we cannot read
      }
    }

    return imports.join('\n');
  }

  /**
   * Collect all template files for watching
   */
  function getAllTemplateFiles() {
    const files = new Set();
    const projectRoot = process.cwd();

    function traverse(currentDir, depth = 0) {
      if (depth > 10) return;

      try {
        if (!existsSync(currentDir)) return;

        const items = readdirSync(currentDir, { withFileTypes: true });

        for (const item of items) {
          if (item.isDirectory() && ['node_modules', '.git', 'dist', 'build'].includes(item.name)) {
            continue;
          }

          const fullPath = join(currentDir, item.name);

          if (item.isDirectory()) {
            traverse(fullPath, depth + 1);
          } else if (item.name.endsWith('.html')) {
            files.add(fullPath);
          }
        }
      } catch (error) {
        // Silently skip directories we cannot read
      }
    }

    traverse(projectRoot);
    return [...files];
  }

  return {
    name: 'vite-plugin-template-extract',

    enforce: 'pre',

    configureServer(devServer) {
      server = devServer;

      if (watchFiles) {
        const allTemplateFiles = getAllTemplateFiles();

        for (const filePath of allTemplateFiles) {
          server.watcher.add(filePath);
        }

        server.watcher.on('change', handleFileChange);
        server.watcher.on('add', handleFileChange);
        server.watcher.on('unlink', handleFileChange);
      }
    },

    buildStart() {
      const allTemplateFiles = getAllTemplateFiles();

      for (const filePath of allTemplateFiles) {
        try {
          const content = readFileSync(filePath, 'utf-8');
          const scripts = extractScripts(content);
          const styles = extractStyles(content);

          if (scripts.length > 0 || styles.length > 0) {
            this.addWatchFile(filePath);
          }
        } catch (error) {
          // Silently skip files we cannot read
        }
      }
    },

    resolveId(id) {
      if (isExtractRequest(id)) {
        return id;
      }

      return null;
    },

    load(id) {
      if (isExtractRequest(id)) {
        const { filename, type } = parseExtractRequest(id);

        try {
          const content = readFileSync(filename, 'utf-8');

          if (type === 'script') {
            const scripts = extractScripts(content);
            if (scripts.length > 0) {
              return scripts.join('\n\n');
            }
          } else if (type === 'style') {
            const styles = extractStyles(content);
            if (styles.length > 0) {
              return styles.join('\n\n');
            }
          }

          return '';
        } catch (error) {
          this.error(`Failed to load ${filename}: ${error.message}`);
          return '';
        }
      }

      return null;
    },

    transform(code, id) {
      if (isExtractRequest(id)) {
        return null;
      }

      const virtualImportRegex = /import\s+['"]virtual:template-extract:([^'"]+)['"]/g;

      if (!virtualImportRegex.test(code)) {
        return null;
      }

      virtualImportRegex.lastIndex = 0;

      const s = new MagicString(code);
      let match;
      let hasReplacements = false;

      while ((match = virtualImportRegex.exec(code)) !== null) {
        const pattern = match[1];
        const fullMatch = match[0];
        const startIndex = match.index;

        let resolvedPath;
        if (isAbsolute(pattern)) {
          resolvedPath = pattern;
        } else {
          const importerDir = dirname(id);
          resolvedPath = resolve(importerDir, pattern);
        }

        const { base, pattern: globPattern } = splitGlobPath(resolvedPath);
        const importStatements = generateImportStatements(base, globPattern);

        if (importStatements) {
          s.overwrite(startIndex, startIndex + fullMatch.length, importStatements);
        } else {
          s.overwrite(startIndex, startIndex + fullMatch.length, '');
        }

        hasReplacements = true;
      }

      if (hasReplacements) {
        return {
          code: s.toString(),
          map: s.generateMap({ hires: true })
        };
      }

      return null;
    },
  };
}
