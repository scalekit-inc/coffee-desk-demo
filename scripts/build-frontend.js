import fs from 'fs';
import path from 'path';

// Build the React app (now in web/ directory)
const reactProjectPath = './web';
const outputPath = './internal/templates';

// Ensure output directory exists
if (!fs.existsSync(outputPath)) {
  fs.mkdirSync(outputPath, { recursive: true });
}

// Copy the built React assets
const distPath = path.join(reactProjectPath, 'dist');
if (fs.existsSync(distPath)) {
  // Copy all files except index.html (we'll create templates)
  fs.cpSync(distPath, outputPath, {
    recursive: true,
    filter: (source) => !source.endsWith('index.html')
  });
  
  // Create template versions
  const indexHtml = fs.readFileSync(path.join(distPath, 'index.html'), 'utf8');
  
  // Create main template
  const mainTemplate = `{{ define "main" -}}${indexHtml}{{- end }}`;
  fs.writeFileSync(path.join(outputPath, 'index.html'), mainTemplate);
  
  console.log('Frontend assets copied and templates created successfully');
} else {
  console.error('React dist directory not found. Please build the React app first.');
  process.exit(1);
} 