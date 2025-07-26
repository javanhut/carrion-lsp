-- Carrion LSP configuration for NvChad
-- Add this to your custom/configs/lspconfig.lua

local on_attach = require("plugins.configs.lspconfig").on_attach
local capabilities = require("plugins.configs.lspconfig").capabilities

local lspconfig = require("lspconfig")
local util = require("lspconfig.util")

-- Configure Carrion LSP
local configs = require("lspconfig.configs")

-- Register Carrion LSP if not already defined
if not configs.carrion_lsp then
  configs.carrion_lsp = {
    default_config = {
      cmd = { "carrion-lsp" },
      filetypes = { "carrion" },
      root_dir = function(fname)
        return util.find_git_ancestor(fname) or util.path.dirname(fname)
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

-- Enhanced on_attach function for Carrion LSP
local function carrion_on_attach(client, bufnr)
  -- Call the default NvChad on_attach
  on_attach(client, bufnr)
  
  -- Additional Carrion-specific mappings
  local bufopts = { noremap = true, silent = true, buffer = bufnr }
  
  -- Carrion-specific keybindings
  vim.keymap.set("n", "<leader>cf", function()
    vim.lsp.buf.format({ 
      async = true,
      filter = function(c)
        return c.name == "carrion_lsp"
      end,
    })
  end, bufopts)
  
  -- Auto-format on save
  if client.server_capabilities.documentFormattingProvider then
    local augroup = vim.api.nvim_create_augroup("CarrionLspFormatting", { clear = true })
    vim.api.nvim_create_autocmd("BufWritePre", {
      group = augroup,
      buffer = bufnr,
      pattern = "*.crl",
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

  -- Enhanced diagnostics for Carrion
  vim.diagnostic.config({
    virtual_text = {
      prefix = "●",
      source = "always",
    },
    float = {
      source = "always",
      header = "",
      border = "rounded",
    },
    signs = true,
    underline = true,
    update_in_insert = false,
    severity_sort = true,
  })
end

-- Setup Carrion LSP
lspconfig.carrion_lsp.setup({
  on_attach = carrion_on_attach,
  capabilities = capabilities,
  settings = {
    carrion = {
      diagnostics = {
        enable = true,
      },
      completion = {
        enable = true,
        memberAccess = true, -- Enable obj.method completion
      },
      formatting = {
        enable = true,
      },
    },
  },
})

-- Auto-detect Carrion files
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  pattern = "*.crl",
  callback = function()
    vim.bo.filetype = "carrion"
    vim.bo.commentstring = "# %s"
  end,
})

-- Additional LSP servers (keep your existing ones)
-- lua
lspconfig.lua_ls.setup({
  on_attach = on_attach,
  capabilities = capabilities,
  settings = {
    Lua = {
      diagnostics = {
        globals = { "vim" },
      },
      workspace = {
        library = {
          [vim.fn.expand("$VIMRUNTIME/lua")] = true,
          [vim.fn.expand("$VIMRUNTIME/lua/vim/lsp")] = true,
          [vim.fn.stdpath("data") .. "/lazy/ui/nvchad_types"] = true,
          [vim.fn.stdpath("data") .. "/lazy/lazy.nvim/lua/lazy"] = true,
        },
        maxPreload = 100000,
        preloadFileSize = 10000,
      },
    },
  },
})

-- Add other language servers as needed