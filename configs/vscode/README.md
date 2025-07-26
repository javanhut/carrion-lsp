# Carrion Language Support for Visual Studio Code

This extension provides comprehensive language support for the Carrion programming language in Visual Studio Code.

## Features

-  **Syntax Highlighting** - Full syntax highlighting for Carrion code
-  **IntelliSense** - Smart code completion and suggestions  
-  **Go to Definition** - Navigate to symbol definitions
-  **Find References** - Find all references to symbols
-  **Hover Documentation** - View documentation on hover
-  **Error Detection** - Real-time error and warning diagnostics
-  **Code Formatting** - Format code automatically
-  **Member Access Completion** - `obj.method()` completion after class instantiation
-  **Format on Save** - Automatic formatting when saving files
-  **Symbol Navigation** - Outline view and workspace symbol search

## Prerequisites

Before installing this extension, you need to install the Carrion Language Server:

### Install Carrion LSP

**Option 1: Automatic Installation (Recommended)**
```bash
curl -fsSL https://raw.githubusercontent.com/javanhut/carrion-lsp/main/install.sh | bash
```

**Option 2: Manual Installation**
1. Download the appropriate binary from [GitHub Releases](https://github.com/javanhut/carrion-lsp/releases)
2. Extract and place `carrion-lsp` in your PATH
3. Verify installation: `carrion-lsp --version`

## Installation

### Method 1: From Source (Development)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/javanhut/carrion-lsp.git
   cd carrion-lsp/configs/vscode
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Compile the extension:**
   ```bash
   npm run compile
   ```

4. **Package the extension:**
   ```bash
   npm run package
   ```

5. **Install the packaged extension:**
   ```bash
   code --install-extension carrion-language-support-1.0.0.vsix
   ```

### Method 2: Manual Setup

If you prefer to set up the language server manually without the extension:

1. **Install the Carrion LSP server** (see prerequisites above)

2. **Configure VS Code settings** (`settings.json`):
   ```json
   {
     "files.associations": {
       "*.crl": "carrion"
     },
     "carrion.enable": true,
     "carrion.serverPath": "carrion-lsp",
     "carrion.formatting.formatOnSave": true,
     "carrion.diagnostics.enable": true,
     "carrion.completion.memberAccess": true
   }
   ```

## Configuration

The extension can be configured through VS Code settings. Open your settings (Ctrl/Cmd + ,) and search for "Carrion" or edit your `settings.json`:

```json
{
  // Enable/disable the Carrion language server
  "carrion.enable": true,
  
  // Path to the carrion-lsp executable
  "carrion.serverPath": "carrion-lsp",
  
  // Trace communication with the language server (for debugging)
  "carrion.trace.server": "off", // "off" | "messages" | "verbose"
  
  // Enable diagnostic features (error detection)
  "carrion.diagnostics.enable": true,
  
  // Enable auto-completion
  "carrion.completion.enable": true,
  
  // Enable member access completion (obj.method)
  "carrion.completion.memberAccess": true,
  
  // Enable code formatting
  "carrion.formatting.enable": true,
  
  // Format code automatically on save
  "carrion.formatting.formatOnSave": true,
  
  // Format code automatically while typing
  "carrion.formatting.formatOnType": false
}
```

## Commands

The extension provides several commands accessible through the Command Palette (Ctrl/Cmd + Shift + P):

- **Carrion: Restart Language Server** - Restart the LSP server
- **Carrion: Show Output Channel** - Show LSP communication logs

## File Association

The extension automatically associates `.crl` files with the Carrion language. You can also manually set the language mode by:

1. Opening a file
2. Clicking the language indicator in the status bar
3. Selecting "Carrion" from the list

## Features in Detail

### Smart Code Completion

The extension provides intelligent code completion including:

- **Variable and function names** - Based on scope analysis
- **Built-in functions** - `print()`, `len()`, `str()`, etc.
- **Keywords** - `spell`, `grim`, `if`, `for`, etc.
- **Class member access** - After typing `obj.`, see available methods

Example:
```carrion
grim Example:
    spell example_spell():
        print("example")

main:
    ex = Example()
    ex.  # <-- Completion shows: example_spell()
```

### Error Detection and Diagnostics

Real-time error detection shows:
- Undefined variables
- Syntax errors
- Type mismatches
- Unused imports

### Code Formatting

Automatic code formatting ensures consistent style:
- Proper indentation
- Consistent spacing
- Line length management

### Navigation Features

- **Go to Definition** - Ctrl/Cmd + Click or F12
- **Find References** - Shift + F12
- **Symbol Search** - Ctrl/Cmd + T
- **Outline View** - Shows document structure

## Carrion Language Syntax

The extension supports all Carrion language features:

### Keywords
- `spell` - Function definition
- `grim` - Class definition  
- `if`, `elif`, `else` - Conditionals
- `for`, `while` - Loops
- `attempt`, `ensnare`, `resolve` - Error handling
- `skip`, `stop` - Loop control
- `import`, `as`, `from` - Module imports

### Example Code
```carrion
# Import modules
import file as f

# Define a class
grim Calculator:
    spell init():
        self.value = 0
    
    spell add(number):
        self.value = self.value + number
        return self.value

# Main function
spell main():
    calc = Calculator()
    result = calc.add(5)
    print(f"Result: {result}")

# Entry point
main:
    main()
```

## Troubleshooting

### Language Server Not Starting

1. **Check if carrion-lsp is installed:**
   ```bash
   which carrion-lsp
   carrion-lsp --version
   ```

2. **Verify PATH configuration:**
   - Restart VS Code after installing carrion-lsp
   - Check that the installation directory is in your PATH

3. **Check extension logs:**
   - Open Command Palette (Ctrl/Cmd + Shift + P)
   - Run "Carrion: Show Output Channel"
   - Look for error messages

### File Not Recognized as Carrion

1. **Check file extension:** Ensure files have `.crl` extension
2. **Manually set language:** Click language indicator in status bar → Select "Carrion"
3. **Add file association:** Add to settings.json:
   ```json
   {
     "files.associations": {
       "*.crl": "carrion"
     }
   }
   ```

### Completion Not Working

1. **Check completion settings:** Ensure `carrion.completion.enable` is `true`
2. **Restart language server:** Command Palette → "Carrion: Restart Language Server"
3. **Check for syntax errors:** Fix any red underlined errors first

### Formatting Issues

1. **Check formatting settings:** Ensure `carrion.formatting.enable` is `true`
2. **Manual format:** Right-click → "Format Document" or Shift + Alt + F
3. **Disable conflicting formatters:** Disable other formatters for `.crl` files

### Performance Issues

1. **Disable trace logging:** Set `carrion.trace.server` to `"off"`
2. **Close unused Carrion files:** LSP analyzes all open files
3. **Restart VS Code:** Sometimes helps with memory issues

## Development

To contribute to this extension:

1. **Clone the repository:**
   ```bash
   git clone https://github.com/javanhut/carrion-lsp.git
   cd carrion-lsp/configs/vscode
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Open in VS Code:**
   ```bash
   code .
   ```

4. **Run extension:** Press F5 to launch Extension Development Host

5. **Make changes:** Edit TypeScript files in `src/`

6. **Test changes:** Extension automatically recompiles on save

### Build Commands

```bash
# Compile TypeScript
npm run compile

# Watch for changes
npm run watch

# Run linter
npm run lint

# Run tests
npm run test

# Package for distribution
npm run package

# Publish to marketplace
npm run publish
```

## Support

For issues and feature requests:

1. **Check existing issues:** [GitHub Issues](https://github.com/javanhut/carrion-lsp/issues)
2. **Create new issue:** Include VS Code version, extension version, and error logs
3. **Discussion:** [GitHub Discussions](https://github.com/javanhut/carrion-lsp/discussions)

## License

This extension is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

**Enjoy coding in Carrion! 🦅**