# carrion-lsp.nvim

A Neovim plugin that provides Language Server Protocol (LSP) support for the [Carrion Programming Language](https://github.com/javanhut/TheCarrionLanguage).

** Now with zero-config installation for Neovim!**

##  Features

-  **Zero-config setup** - Works out of the box with sensible defaults
-  **Auto-installation** - Automatically downloads and installs the carrion-lsp binary
-  **Full LSP support** with rich language features:
  - Real-time diagnostics and error detection
  - Intelligent code completion (IntelliSense)
  - Go-to-definition and find references
  - Hover documentation
  - Code formatting
  - Signature help
  - Rename refactoring
  - Code actions
-  **Syntax highlighting** for Carrion files (`.crl`)
-  **Health checks** with `:checkhealth carrion-lsp`
-  **Lazy loading** - Only activates for Carrion files

### Carrion Language Support
-  **Crow-themed Keywords** - `spell` (function), `grim` (class), `skip`/`stop` (continue/break)
-  **Error Handling** - `attempt`/`ensnare`/`resolve` syntax support
-  **Module System** - Support for Carrion's import resolution hierarchy
-  **Class Inheritance** - Full support for class hierarchies and inheritance
-  **Built-in Functions** - Complete support for Carrion's standard library

##  Requirements

- Neovim 0.8+
- [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig)

### Optional (recommended)

- [nvim-cmp](https://github.com/hrsh7th/nvim-cmp) + [cmp-nvim-lsp](https://github.com/hrsh7th/cmp-nvim-lsp) for enhanced autocompletion
- [mason.nvim](https://github.com/williamboman/mason.nvim) for LSP management UI

##  Installation

### With [lazy.nvim](https://github.com/folke/lazy.nvim) (recommended)

```lua
{
  "javanhut/carrion-lsp",
  ft = "carrion", -- Lazy load only for Carrion files
  dependencies = {
    "neovim/nvim-lspconfig",
    -- Optional: enhanced autocompletion
    "hrsh7th/nvim-cmp",
    "hrsh7th/cmp-nvim-lsp",
  },
  config = true, -- Use default configuration
}
```

### With [packer.nvim](https://github.com/wbthomason/packer.nvim)

```lua
use {
  "javanhut/carrion-lsp",
  ft = "carrion",
  requires = {
    "neovim/nvim-lspconfig",
    -- Optional dependencies
    "hrsh7th/nvim-cmp",
    "hrsh7th/cmp-nvim-lsp",
  },
  config = function()
    require("carrion-lsp").setup()
  end
}
```

### With [vim-plug](https://github.com/junegunn/vim-plug)

```vim
Plug 'neovim/nvim-lspconfig'
Plug 'javanhut/carrion-lsp'

" In your init.vim or init.lua:
lua require("carrion-lsp").setup()
```

## ⚙ Configuration

### Default Configuration

The plugin works with zero configuration, but you can customize it:

```lua
require("carrion-lsp").setup({
  -- Auto-install carrion-lsp binary if not found
  auto_install = true,
  
  -- LSP server settings
  server = {
    cmd = { "carrion-lsp" },
    settings = {
      carrion = {
        diagnostics = { enable = true },
        completion = { enable = true, memberAccess = true },
        formatting = { enable = true },
      },
    },
  },
  
  -- Key mappings (set to empty string to disable)
  keymaps = {
    goto_definition = "gd",
    goto_declaration = "gD", 
    hover = "K",
    implementation = "gi",
    signature_help = "<C-k>",
    rename = "<leader>rn",
    code_action = "<leader>ca",
    references = "gr",
    format = "<leader>f",
    next_diagnostic = "]d",
    prev_diagnostic = "[d",
    diagnostic_float = "<leader>e",
  },
  
  -- Auto-format on save
  format_on_save = true,
  
  -- Show diagnostics in floating window on cursor hold
  diagnostic_float_on_hold = true,
})
```

### Advanced Configuration

```lua
require("carrion-lsp").setup({
  auto_install = false, -- Disable auto-installation
  
  server = {
    cmd = { "/custom/path/to/carrion-lsp" },
    settings = {
      carrion = {
        -- Custom LSP settings
        diagnostics = { 
          enable = true,
          severity = "error", -- Only show errors
        },
        completion = { 
          enable = true,
          memberAccess = true,
          snippets = false, -- Disable snippets
        },
      },
    },
    init_options = {
      -- Custom initialization options
      carrionPath = "/path/to/carrion/interpreter",
    },
  },
  
  -- Disable specific keymaps
  keymaps = {
    goto_definition = "gd",
    hover = "K",
    -- Disable other keymaps by setting to empty string
    code_action = "",
    rename = "",
  },
  
  format_on_save = false, -- Disable auto-format
  diagnostic_float_on_hold = false, -- Disable diagnostic float
})
```

##  Usage

Once installed, the plugin automatically:

1. **Detects Carrion files** (`.crl` extension)
2. **Downloads carrion-lsp binary** (if `auto_install = true`)
3. **Starts the LSP server** when you open a Carrion file
4. **Provides language features** like completion, diagnostics, etc.

### Commands

| Command | Description |
|---------|-------------|
| `:CarrionLspInstall` | Install carrion-lsp binary |
| `:CarrionLspUpdate` | Update carrion-lsp binary |
| `:CarrionLspUninstall` | Uninstall carrion-lsp binary |
| `:CarrionLspInfo` | Show plugin information |
| `:CarrionLspRestart` | Restart LSP server |

### Default Keymaps

| Key | Action |
|-----|--------|
| `gd` | Go to definition |
| `gD` | Go to declaration |
| `gr` | Find references |
| `K` | Hover documentation |
| `gi` | Go to implementation |
| `<C-k>` | Signature help |
| `<leader>rn` | Rename symbol |
| `<leader>ca` | Code actions |
| `<leader>f` | Format document |
| `<leader>e` | Show diagnostic |
| `]d` | Next diagnostic |
| `[d` | Previous diagnostic |

##  Health Check

Run `:checkhealth carrion-lsp` to verify your setup:

```vim
:checkhealth carrion-lsp
```

This will check:
- Neovim version compatibility
- carrion-lsp binary installation
- Required dependencies
- LSP server status
- Filetype detection

##  Example Usage

### Basic Carrion Program

Create a file `example.crl`:

```carrion
# Define a class
grim Calculator:
    spell init():
        self.value = 0
    
    spell add(number):
        self.value = self.value + number
        return self.value
    
    spell multiply(factor):
        self.value = self.value * factor
        return self.value

# Main function
spell main():
    calc = Calculator()
    result = calc.add(10)
    print(f"After adding 10: {result}")
    
    calc.  # <- LSP shows: add, multiply, init methods
    
    final = calc.multiply(2)
    print(f"After multiplying by 2: {final}")

# Entry point
main:
    main()
```

### LSP Features in Action

1. **Auto-completion**: Type `calc.` and see available methods
2. **Go to definition**: Press `gd` on `Calculator` to go to class definition
3. **Find references**: Press `gr` on the `add` method to find all uses
4. **Hover documentation**: Press `K` over functions to see signatures
5. **Error detection**: Misspelled variable names are highlighted
6. **Format on save**: Code is automatically formatted when saved

##  Manual Installation

If you prefer to install the carrion-lsp binary manually:

```bash
# Install carrion-lsp binary
curl -fsSL https://raw.githubusercontent.com/javanhut/carrion-lsp/main/install.sh | bash

# Or build from source
git clone https://github.com/javanhut/carrion-lsp.git
cd carrion-lsp
make install-user
```

Then disable auto-installation:

```lua
require("carrion-lsp").setup({
  auto_install = false,
})
```

##  Troubleshooting

### LSP not starting

1. Check if carrion-lsp is installed:
   ```vim
   :CarrionLspInfo
   ```

2. Check LSP logs:
   ```vim
   :LspLog
   ```

3. Restart LSP:
   ```vim
   :CarrionLspRestart
   ```

### File not detected as Carrion

1. Check filetype:
   ```vim
   :set ft?
   ```

2. Manually set filetype:
   ```vim
   :set ft=carrion
   ```

### Autocompletion not working

1. Install recommended completion plugins:
   ```lua
   -- In your plugin manager
   "hrsh7th/nvim-cmp"
   "hrsh7th/cmp-nvim-lsp"
   ```

2. Check LSP capabilities:
   ```vim
   :LspInfo
   ```

##  Advanced Setup

### Integration with Mason.nvim

```lua
{
  "javanhut/carrion-lsp",
  ft = "carrion",
  dependencies = {
    "neovim/nvim-lspconfig",
    "williamboman/mason.nvim", -- Optional: for UI management
  },
  config = function()
    require("carrion-lsp").setup({
      auto_install = true, -- Still works with Mason
    })
  end
}
```

### Integration with Trouble.nvim

```lua
{
  "javanhut/carrion-lsp",
  ft = "carrion",
  dependencies = {
    "neovim/nvim-lspconfig",
    "folke/trouble.nvim", -- Enhanced diagnostics UI
  },
  config = function()
    require("carrion-lsp").setup()
    
    -- Optional: Configure Trouble for Carrion diagnostics
    vim.keymap.set("n", "<leader>xx", "<cmd>TroubleToggle<cr>")
    vim.keymap.set("n", "<leader>xd", "<cmd>TroubleToggle document_diagnostics<cr>")
  end
}
```

### Full-featured Setup

```lua
{
  "javanhut/carrion-lsp",
  ft = "carrion",
  dependencies = {
    "neovim/nvim-lspconfig",
    "hrsh7th/nvim-cmp",
    "hrsh7th/cmp-nvim-lsp",
    "williamboman/mason.nvim",
    "folke/trouble.nvim",
  },
  config = function()
    require("carrion-lsp").setup({
      auto_install = true,
      format_on_save = true,
      diagnostic_float_on_hold = true,
      keymaps = {
        goto_definition = "gd",
        hover = "K",
        references = "gr",
        rename = "<leader>rn",
        code_action = "<leader>ca",
        format = "<leader>f",
      },
    })
  end
}
```

##  Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

##  License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

##  Related Projects

- [TheCarrionLanguage](https://github.com/javanhut/TheCarrionLanguage) - The Carrion programming language
- [carrion-lsp](https://github.com/javanhut/carrion-lsp) - Language Server Protocol implementation for Carrion