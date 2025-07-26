-- Carrion LSP configuration for mason.nvim
-- This configuration assumes you're using mason.nvim for LSP management

-- First, ensure carrion-lsp is installed via mason or system package manager
-- Mason doesn't include carrion-lsp by default, so you'll need to install it manually:
-- Run: curl -fsSL https://raw.githubusercontent.com/javanhut/carrion-lsp/main/install.sh | bash

local lspconfig = require("lspconfig")
local mason = require("mason")
local mason_lspconfig = require("mason-lspconfig")

-- Setup mason
mason.setup({
  ui = {
    icons = {
      package_installed = "✓",
      package_pending = "➜",
      package_uninstalled = "✗",
    },
  },
})

-- Configure mason-lspconfig
mason_lspconfig.setup({
  -- Note: carrion-lsp is not available in mason registry
  -- Install it manually using the install script
  ensure_installed = {
    -- Add other LSPs you want mason to manage
    "lua_ls",
    "pyright",
  },
})

-- Custom Carrion LSP configuration
local configs = require("lspconfig.configs")

-- Check if carrion_lsp is already defined
if not configs.carrion_lsp then
  configs.carrion_lsp = {
    default_config = {
      cmd = { "carrion-lsp" },
      filetypes = { "carrion" },
      root_dir = function(fname)
        return lspconfig.util.find_git_ancestor(fname) or lspconfig.util.path.dirname(fname)
      end,
      settings = {},
      init_options = {},
    },
    docs = {
      description = "Language Server for the Carrion programming language",
      default_config = {
        root_dir = "util.find_git_ancestor(fname) or util.path.dirname(fname)",
      },
    },
  }
end

-- Function to setup Carrion LSP with common configuration
local function setup_carrion_lsp(opts)
  opts = opts or {}
  
  local default_opts = {
    on_attach = function(client, bufnr)
      -- Enable completion triggered by <c-x><c-o>
      vim.api.nvim_buf_set_option(bufnr, "omnifunc", "v:lua.vim.lsp.omnifunc")

      -- Buffer local mappings
      local bufopts = { noremap = true, silent = true, buffer = bufnr }
      
      -- Navigation
      vim.keymap.set("n", "gD", vim.lsp.buf.declaration, bufopts)
      vim.keymap.set("n", "gd", vim.lsp.buf.definition, bufopts)
      vim.keymap.set("n", "K", vim.lsp.buf.hover, bufopts)
      vim.keymap.set("n", "gi", vim.lsp.buf.implementation, bufopts)
      vim.keymap.set("n", "gr", vim.lsp.buf.references, bufopts)
      
      -- Code actions
      vim.keymap.set("n", "<space>rn", vim.lsp.buf.rename, bufopts)
      vim.keymap.set("n", "<space>ca", vim.lsp.buf.code_action, bufopts)
      vim.keymap.set("n", "<space>f", function()
        vim.lsp.buf.format({ async = true })
      end, bufopts)
      
      -- Signature help
      vim.keymap.set("n", "<C-k>", vim.lsp.buf.signature_help, bufopts)
      vim.keymap.set("i", "<C-k>", vim.lsp.buf.signature_help, bufopts)
      
      -- Workspace folders
      vim.keymap.set("n", "<space>wa", vim.lsp.buf.add_workspace_folder, bufopts)
      vim.keymap.set("n", "<space>wr", vim.lsp.buf.remove_workspace_folder, bufopts)
      vim.keymap.set("n", "<space>wl", function()
        print(vim.inspect(vim.lsp.buf.list_workspace_folders()))
      end, bufopts)

      -- Auto-format on save for Carrion files
      if client.server_capabilities.documentFormattingProvider then
        local augroup = vim.api.nvim_create_augroup("CarrionLspFormatting", { clear = true })
        vim.api.nvim_create_autocmd("BufWritePre", {
          group = augroup,
          buffer = bufnr,
          callback = function()
            vim.lsp.buf.format({ 
              async = false,
              filter = function(c)
                return c.name == "carrion_lsp"
              end,
            })
          end,
        })
      end

      -- Enable auto-linting (diagnostics)
      vim.diagnostic.config({
        virtual_text = true,
        signs = true,
        underline = true,
        update_in_insert = false,
        severity_sort = true,
      })
    end,
    
    capabilities = vim.lsp.protocol.make_client_capabilities(),
    
    settings = {
      carrion = {
        -- Add any Carrion-specific settings here
        diagnostics = {
          enable = true,
        },
        completion = {
          enable = true,
        },
      },
    },
  }

  -- Merge user options with defaults
  for k, v in pairs(opts) do
    default_opts[k] = v
  end

  -- If nvim-cmp is available, use its capabilities
  local has_cmp, cmp_nvim_lsp = pcall(require, "cmp_nvim_lsp")
  if has_cmp then
    default_opts.capabilities = cmp_nvim_lsp.default_capabilities(default_opts.capabilities)
  end

  lspconfig.carrion_lsp.setup(default_opts)
end

-- Auto-detect Carrion files
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  pattern = "*.crl",
  callback = function()
    vim.bo.filetype = "carrion"
  end,
})

-- Setup Carrion LSP
setup_carrion_lsp()

-- Export the setup function for manual configuration
return {
  setup = setup_carrion_lsp,
}