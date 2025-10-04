/**
 * @jest-environment jsdom
 */

// Set up the basic HTML structure needed for app.js
document.body.innerHTML = `
  <div id="main-container"></div>
  <div id="search-dialog-overlay" style="display: none;">
    <div id="search-dialog-dialog">
      <input id="search-dialog-input" />
      <ul id="search-dialog-list"></ul>
    </div>
  </div>
  <div id="help-text"></div>
  <select id="theme-select">
    <option value="light.css">Light</option>
  </select>
  <link rel="stylesheet" href="light.css" />
`;

// Mock global variables that would be injected by Go's html/template
global.typeData = {};
global.startTypes = [];

// Now, require the script to be tested
const app = require('./app.js');

describe('splitTypeName', () => {
  it('should correctly split a fully qualified type name', () => {
    const { pkg, type } = splitTypeName('k8s.io/api/core/v1.Pod');
    expect(pkg).toBe('k8s.io/api/core/v1');
    expect(type).toBe('Pod');
  });

  it('should handle type names without a package', () => {
    const { pkg, type } = splitTypeName('string');
    expect(pkg).toBe('');
    expect(type).toBe('string');
  });

  it('should handle empty strings', () => {
    const { pkg, type } = splitTypeName('');
    expect(pkg).toBe('');
    expect(type).toBe('');
  });
});
