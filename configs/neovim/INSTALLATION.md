# Neovim Configuration for Carrion LSP

This directory contains configurations for integrating the Carrion Language Server with various Neovim distributions and setups.

## Prerequisites

1. **Install the Carrion LSP server:**
   ```bash
   curl -fsSL https://raw.githubusercontent.com/javanhut/carrion-lsp/main/install.sh | bash
   ```

2. **Verify installation:**
   ```bash
   carrion-lsp --version
   ```

3. **Ensure `carrion-lsp` is in your PATH:**
   ```bash
   which carrion-lsp
   ```

## Configuration by Distribution

### 1. Lazy.nvim

**File:** `lazy.nvim/carrion-lsp.lua`

**Installation:**
1. Copy the configuration to your lazy.nvim plugins directory:
   ```bash
   # For most setups
   cp configs/neovim/lazy.nvim/carrion-lsp.lua ~/.config/nvim/lua/plugins/
   
   # Or add to your existing plugins file
   ```

2. The configuration includes all necessary dependencies and will be loaded automatically.

**Features enabled:**
-  Auto-completion with nvim-cmp
-  Format on save
-  Go to definition
-  Find references
-  Hover documentation
-  Member access completion (obj.method)
-  Diagnostics/linting

### 2. Mason.nvim

**File:** `mason/carrion-lsp.lua`

**Installation:**
1. Copy the configuration:
   ```bash
   cp configs/neovim/mason/carrion-lsp.lua ~/.config/nvim/lua/configs/
   ```

2. Add to your Neovim configuration:
   ```lua
   require("configs.carrion-lsp")
   ```

**Note:** Mason doesn't include carrion-lsp in its registry, so you must install it manually using the install script.

### 3. NvChad

**Files:** 
- `nvchad/custom/configs/lspconfig.lua`
- `nvchad/custom/plugins.lua`
- `nvchad/custom/configs/overrides.lua`
- `nvchad/custom/configs/null-ls.lua`

**Installation:**
1. Copy the configurations to your NvChad custom directory:
   ```bash
   # Create custom directory if it doesn't exist
   mkdir -p ~/.config/nvim/lua/custom/configs/
   
   # Copy configurations
   cp configs/neovim/nvchad/custom/configs/* ~/.config/nvim/lua/custom/configs/
   cp configs/neovim/nvchad/custom/plugins.lua ~/.config/nvim/lua/custom/
   ```

2. Restart Neovim and run `:PackerSync` or `:Lazy sync` depending on your setup.

**Enhanced features:**
-  Enhanced UI with lspsaga
-  Trouble.nvim for diagnostics
-  Better file type detection
-  Integrated with NvChad's theming

### 4. Generic/Custom Setup

**File:** `generic/carrion-lsp.lua`

**Installation:**
1. Copy the configuration:
   ```bash
   cp configs/neovim/generic/carrion-lsp.lua ~/.config/nvim/lua/
   ```

2. Add to your `init.lua`:
   ```lua
   require("carrion-lsp").setup({
     -- Optional: override default settings
     settings = {
       carrion = {
         diagnostics = { enable = true },
         completion = { memberAccess = true },
       },
     },
   })
   ```

This configuration works with any Neovim setup that has `nvim-lspconfig` installed.

## Default Keybindings

All configurations include these default keybindings:

| Key | Mode | Action |
|-----|------|--------|
| `gd` | Normal | Go to definition |
| `gD` | Normal | Go to declaration |
| `gr` | Normal | Find references |
| `K` | Normal | Hover documentation |
| `<space>rn` | Normal | Rename symbol |
| `<space>ca` | Normal | Code actions |
| `<space>f` | Normal | Format document |
| `<C-k>` | Normal/Insert | Signature help |
| `[d` | Normal | Previous diagnostic |
| `]d` | Normal | Next diagnostic |
| `<space>e` | Normal | Open diagnostic float |

## File Type Detection

All configurations automatically detect `.crl` files as Carrion files and enable the LSP.

## Features

###  Enabled Features

- **Syntax highlighting** - Basic highlighting for Carrion keywords
- **Auto-completion** - Intelligent code completion
- **Member access completion** - `obj.method()` completion after class instantiation
- **Go to definition** - Navigate to symbol definitions
- **Find references** - Find all references to a symbol
- **Hover documentation** - Show documentation on hover
- **Diagnostics/Linting** - Real-time error and warning detection
- **Format on save** - Automatic code formatting when saving
- **Signature help** - Function signature information while typing
- **Rename refactoring** - Rename symbols across the workspace
- **Code actions** - Quick fixes and refactoring actions

###  Advanced Features

- **Workspace symbol search** - Find symbols across the entire project
- **Document symbols** - Outline view of current file
- **Call hierarchy** - View function call relationships
- **Type definitions** - Navigate to type definitions
- **Implementation finder** - Find implementations of interfaces/abstract methods

## Troubleshooting

### LSP Not Starting

1. **Check if carrion-lsp is installed:**
   ```bash
   which carrion-lsp
   carrion-lsp --version
   ```

2. **Check LSP logs:**
   ```vim
   :LspLog
   ```

3. **Restart LSP client:**
   ```vim
   :LspRestart
   ```

### File Type Not Detected

1. **Manually set filetype:**
   ```vim
   :set filetype=carrion
   ```

2. **Check if autocmd is working:**
   ```vim
   :autocmd BufRead,BufNewFile *.crl
   ```

### Completion Not Working

1. **Ensure nvim-cmp is installed** (for configurations that use it)
2. **Check LSP capabilities:**
   ```vim
   :LspInfo
   ```

### Format on Save Not Working

1. **Check if formatting is enabled:**
   ```vim
   :LspInfo
   ```
   Look for "documentFormattingProvider: true"

2. **Manually format:**
   ```vim
   :lua vim.lsp.buf.format()
   ```

## Customization

You can customize the LSP behavior by modifying the settings in your configuration:

```lua
require("lspconfig").carrion_lsp.setup({
  settings = {
    carrion = {
      diagnostics = {
        enable = true,
      },
      completion = {
        enable = true,
        memberAccess = true,  -- Enable obj.method completion
      },
      formatting = {
        enable = true,
      },
    },
  },
})
```

## Support

For issues specific to the Neovim configurations:
1. Check the LSP logs (`:LspLog`)
2. Verify the Carrion LSP server is working independently
3. Open an issue in the [carrion-lsp repository](https://github.com/javanhut/carrion-lsp/issues)