# Setup Guide

This guide walks you through setting up the Carrion Language Server Protocol (LSP) with various editors and IDEs.

## Prerequisites

### System Requirements
- **Operating System**: Linux, macOS, or Windows
- **Go**: Version 1.19 or later (for building from source)
- **Memory**: At least 100MB available RAM
- **Disk**: 50MB for installation

### Installing Go (if needed)
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# macOS with Homebrew
brew install go

# Arch Linux
sudo pacman -S go

# Or download from https://golang.org/dl/
```

## Building the LSP Server

### Clone and Build
```bash
# Clone the repository
git clone https://github.com/javanhut/carrion-lsp.git
cd carrion-lsp

# Build the server
go build -o carrion-lsp ./cmd/carrion-lsp

# Test the build
./carrion-lsp --version  # (if version flag is implemented)
```

### Install System-wide
```bash
# Copy to system path
sudo cp carrion-lsp /usr/local/bin/
chmod +x /usr/local/bin/carrion-lsp

# Or use Go install
go install ./cmd/carrion-lsp
```

### Verify Installation
```bash
# Check if carrion-lsp is accessible
which carrion-lsp
carrion-lsp --help  # (if help flag is implemented)
```

## Editor Integration

### Visual Studio Code

#### Method 1: Generic LSP Extension
1. Install the "Generic LSP Client" extension
2. Create a `.vscode/settings.json` in your workspace:

```json
{
  "genericLanguageServer.languageServers": [
    {
      "languageId": "carrion",
      "command": "carrion-lsp",
      "args": [],
      "fileExtensions": [".carrion"],
      "workspaceRoot": true
    }
  ]
}
```

#### Method 2: Custom Extension (Advanced)
Create a custom VS Code extension:

1. Create `package.json`:
```json
{
  "name": "carrion-language-support",
  "version": "0.1.0",
  "engines": { "vscode": "^1.50.0" },
  "contributes": {
    "languages": [
      {
        "id": "carrion",
        "aliases": ["Carrion"],
        "extensions": [".carrion"],
        "configuration": "./language-configuration.json"
      }
    ],
    "grammars": [
      {
        "language": "carrion",
        "scopeName": "source.carrion", 
        "path": "./syntaxes/carrion.tmGrammar.json"
      }
    ]
  },
  "activationEvents": ["onLanguage:carrion"],
  "main": "./out/extension.js"
}
```

2. Create `src/extension.ts`:
```typescript
import * as vscode from 'vscode';
import { LanguageClient, LanguageClientOptions, ServerOptions } from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    const serverOptions: ServerOptions = {
        command: 'carrion-lsp',
        args: []
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'carrion' }]
    };

    client = new LanguageClient('carrion-lsp', 'Carrion Language Server', serverOptions, clientOptions);
    client.start();
}

export function deactivate(): Thenable<void> | undefined {
    return client?.stop();
}
```

### Neovim

#### Using nvim-lspconfig
1. Install `nvim-lspconfig`:
```vim
" Using vim-plug
Plug 'neovim/nvim-lspconfig'

" Or using packer.nvim
use 'neovim/nvim-lspconfig'
```

2. Configure in your `init.lua`:
```lua
local lspconfig = require('lspconfig')

-- Define carrion-lsp configuration
local configs = require('lspconfig.configs')
if not configs.carrion_lsp then
  configs.carrion_lsp = {
    default_config = {
      cmd = {'carrion-lsp'},
      filetypes = {'carrion'},
      root_dir = lspconfig.util.root_pattern('*.carrion', '.git'),
      settings = {}
    }
  }
end

-- Setup the language server
lspconfig.carrion_lsp.setup{
  on_attach = function(client, bufnr)
    -- Enable completion triggered by <c-x><c-o>
    vim.api.nvim_buf_set_option(bufnr, 'omnifunc', 'v:lua.vim.lsp.omnifunc')

    -- Mappings
    local bufopts = { noremap=true, silent=true, buffer=bufnr }
    vim.keymap.set('n', 'gD', vim.lsp.buf.declaration, bufopts)
    vim.keymap.set('n', 'gd', vim.lsp.buf.definition, bufopts)
    vim.keymap.set('n', 'K', vim.lsp.buf.hover, bufopts)
    vim.keymap.set('n', 'gi', vim.lsp.buf.implementation, bufopts)
    vim.keymap.set('n', '<C-k>', vim.lsp.buf.signature_help, bufopts)
    vim.keymap.set('n', '<space>rn', vim.lsp.buf.rename, bufopts)
    vim.keymap.set('n', '<space>ca', vim.lsp.buf.code_action, bufopts)
    vim.keymap.set('n', 'gr', vim.lsp.buf.references, bufopts)
    vim.keymap.set('n', '<space>f', vim.lsp.buf.formatting, bufopts)
  end
}
```

3. Set up filetype detection in `ftdetect/carrion.vim`:
```vim
au BufRead,BufNewFile *.carrion set filetype=carrion
```

#### Using coc.nvim
1. Install coc.nvim and add to `coc-settings.json`:
```json
{
  "languageserver": {
    "carrion": {
      "command": "carrion-lsp",
      "filetypes": ["carrion"],
      "rootPatterns": ["*.carrion"]
    }
  }
}
```

### Emacs

#### Using lsp-mode
1. Install `lsp-mode`:
```elisp
;; Using use-package
(use-package lsp-mode
  :ensure t
  :commands lsp)
```

2. Configure Carrion LSP:
```elisp
;; Add to your init.el
(add-to-list 'lsp-language-id-configuration '(carrion-mode . "carrion"))

(lsp-register-client
 (make-lsp-client
  :new-connection (lsp-stdio-connection "carrion-lsp")
  :major-modes '(carrion-mode)
  :server-id 'carrion-lsp))

;; Define carrion-mode (basic)
(define-derived-mode carrion-mode fundamental-mode "Carrion"
  "Major mode for Carrion programming language."
  (setq comment-start "#")
  (setq comment-end ""))

;; Associate .carrion files with carrion-mode
(add-to-list 'auto-mode-alist '("\\.carrion\\'" . carrion-mode))

;; Hook to start LSP
(add-hook 'carrion-mode-hook #'lsp)
```

#### Using eglot
```elisp
;; Using use-package
(use-package eglot
  :ensure t)

;; Configure Carrion LSP
(add-to-list 'eglot-server-programs '(carrion-mode . ("carrion-lsp")))

;; Hook to start eglot
(add-hook 'carrion-mode-hook #'eglot-ensure)
```

### Vim (Classic)

#### Using vim-lsp
1. Install vim-lsp:
```vim
" Using vim-plug
Plug 'prabirshrestha/vim-lsp'
Plug 'prabirshrestha/asyncomplete.vim'
Plug 'prabirshrestha/asyncomplete-lsp.vim'
```

2. Configure:
```vim
" Register carrion-lsp
if executable('carrion-lsp')
    au User lsp_setup call lsp#register_server({
        \ 'name': 'carrion-lsp',
        \ 'cmd': {server_info->['carrion-lsp']},
        \ 'whitelist': ['carrion'],
        \ })
endif

" Filetype detection
au BufRead,BufNewFile *.carrion set filetype=carrion

" Key mappings
function! s:on_lsp_buffer_enabled() abort
    setlocal omnifunc=lsp#complete
    nmap <buffer> gd <plug>(lsp-definition)
    nmap <buffer> gs <plug>(lsp-document-symbol-search)
    nmap <buffer> gS <plug>(lsp-workspace-symbol-search)
    nmap <buffer> gr <plug>(lsp-references)
    nmap <buffer> gi <plug>(lsp-implementation)
    nmap <buffer> <leader>rn <plug>(lsp-rename)
    nmap <buffer> [g <plug>(lsp-previous-diagnostic)
    nmap <buffer> ]g <plug>(lsp-next-diagnostic)
    nmap <buffer> K <plug>(lsp-hover)
endfunction

augroup lsp_install
    au!
    autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```

### Sublime Text

#### Using LSP Package
1. Install Package Control and the LSP package
2. Create `Carrion.sublime-settings`:
```json
{
  "extensions": ["carrion"],
  "scope": "source.carrion"
}
```

3. Add to LSP settings:
```json
{
  "clients": {
    "carrion-lsp": {
      "enabled": true,
      "command": ["carrion-lsp"],
      "selector": "source.carrion"
    }
  }
}
```

### Kate/KWrite

1. Create `~/.local/share/katepart5/syntax/carrion.xml`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE language SYSTEM "language.dtd">
<language name="Carrion" version="1.0" kateversion="5.0" section="Sources" extensions="*.carrion" mimetype="text/x-carrion">
  <highlighting>
    <list name="keywords">
      <item>spell</item>
      <item>grim</item>
      <item>ignore</item>
      <item>if</item>
      <item>else</item>
      <item>while</item>
      <item>for</item>
      <item>return</item>
      <item>import</item>
      <item>as</item>
    </list>
    <contexts>
      <context attribute="Normal Text" lineEndContext="#stay" name="Normal">
        <keyword attribute="Keyword" context="#stay" String="keywords"/>
      </context>
    </contexts>
    <itemDatas>
      <itemData name="Normal Text" defStyleNum="dsNormal"/>
      <itemData name="Keyword" defStyleNum="dsKeyword"/>
    </itemDatas>
  </highlighting>
</language>
```

## Troubleshooting

### Common Issues

#### "carrion-lsp command not found"
```bash
# Check if carrion-lsp is in PATH
echo $PATH
which carrion-lsp

# If not found, ensure it's installed correctly
ls -la /usr/local/bin/carrion-lsp

# Or use full path in editor config
/full/path/to/carrion-lsp
```

#### LSP Server Not Starting
1. Check if the server runs manually:
```bash
# Test the server directly
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | carrion-lsp
```

2. Check editor LSP logs:
   - VS Code: `View > Output > Language Client`
   - Neovim: `:LspInfo` and `:LspLog`
   - Emacs: `*lsp-log*` buffer

#### No Completions/Features Working
1. Ensure the file has the correct extension (`.carrion`)
2. Check that the LSP client recognizes the filetype
3. Verify the server capabilities in initialization response

#### Performance Issues
1. Check memory usage: `ps aux | grep carrion-lsp`
2. Monitor CPU usage during editing
3. Consider closing unused documents
4. Check for large files causing analysis slowdown

### Debug Mode
Enable debug logging (if implemented):
```bash
carrion-lsp --debug --log-file /tmp/carrion-lsp.log
```

### Getting Help
1. Check the server logs for error messages
2. Verify your editor's LSP client configuration
3. Test with a minimal Carrion file
4. Open an issue with:
   - Editor/IDE version
   - LSP client configuration
   - Sample Carrion code
   - Error messages or logs

## File Associations

To ensure your editor recognizes `.carrion` files:

### System-wide (Linux)
Create `~/.local/share/mime/packages/carrion.xml`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="text/x-carrion">
    <comment>Carrion source code</comment>
    <glob pattern="*.carrion"/>
  </mime-type>
</mime-info>
```

Then run:
```bash
update-mime-database ~/.local/share/mime
```

### macOS
Add to `~/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist` or use a specific editor's file association settings.

### Windows
Associate `.carrion` files with your editor through the "Open With" dialog or registry modifications.

## Next Steps

After setup:
1. Create a test `.carrion` file
2. Try the basic features (completion, hover, go-to-definition)
3. Explore the available LSP features in your editor
4. Report any issues or feature requests
5. Consider contributing to the project

For advanced usage and API details, see [API.md](API.md).