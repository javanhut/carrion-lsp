-- NvChad custom plugins configuration
-- Add this to your custom/plugins.lua

local overrides = require("custom.configs.overrides")

local plugins = {
  -- Override plugin configs
  {
    "neovim/nvim-lspconfig",
    dependencies = {
      -- format & linting
      {
        "jose-elias-alvarez/null-ls.nvim",
        config = function()
          require("custom.configs.null-ls")
        end,
      },
    },
    config = function()
      require("plugins.configs.lspconfig")
      require("custom.configs.lspconfig")
    end,
  },

  -- Install a plugin to help with syntax highlighting for Carrion
  {
    "nvim-treesitter/nvim-treesitter",
    opts = overrides.treesitter,
  },

  -- Override telescope plugin
  {
    "nvim-telescope/telescope.nvim",
    opts = overrides.telescope,
  },

  -- Override nvim-tree plugin
  {
    "nvim-tree/nvim-tree.lua",
    opts = overrides.nvimtree,
  },

  -- Enhanced completion for Carrion LSP
  {
    "hrsh7th/nvim-cmp",
    opts = overrides.cmp,
  },

  -- Additional plugins that work well with Carrion LSP
  {
    "folke/trouble.nvim",
    lazy = false,
    dependencies = { "nvim-tree/nvim-web-devicons" },
    opts = {
      -- Configuration for trouble.nvim
    },
  },

  -- Better LSP UI
  {
    "glepnir/lspsaga.nvim",
    lazy = false,
    config = function()
      require("lspsaga").setup({
        ui = {
          winblend = 10,
          border = "rounded",
          colors = {
            normal_bg = "#002b36",
          },
        },
        lightbulb = {
          enable = false,
        },
      })
    end,
    dependencies = {
      {"nvim-tree/nvim-web-devicons"},
      {"nvim-treesitter/nvim-treesitter"},
    },
  },

  -- File type detection
  {
    "nathom/filetype.nvim",
    lazy = false,
    config = function()
      require("filetype").setup({
        overrides = {
          extensions = {
            crl = "carrion",
          },
        },
      })
    end,
  },
}

return plugins