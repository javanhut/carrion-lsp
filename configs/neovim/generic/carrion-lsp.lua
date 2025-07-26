-- Generic Carrion LSP configuration for Neovim
-- This configuration works with any Neovim setup that has nvim-lspconfig

local M = {}

-- Check if lspconfig is available
local has_lspconfig, lspconfig = pcall(require, "lspconfig")
if not has_lspconfig then
  vim.notify("nvim-lspconfig is required for Carrion LSP", vim.log.levels.ERROR)
  return M
end

local util = require("lspconfig.util")

-- Setup Carrion file detection
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  pattern = "*.crl",
  callback = function()
    vim.bo.filetype = "carrion"
    vim.bo.commentstring = "# %s"
  end,
})

-- Default configuration
local default_config = {
  cmd = { "carrion-lsp" },
  filetypes = { "carrion" },
  root_dir = function(fname)
    return util.find_git_ancestor(fname) or util.path.dirname(fname)
  end,
  settings = {
    carrion = {
      diagnostics = {
        enable = true,
      },
      completion = {
        enable = true,
        memberAccess = true,
      },
      formatting = {
        enable = true,
      },
    },
  },
  init_options = {},
}

-- Default on_attach function
local function default_on_attach(client, bufnr)
  -- Enable completion triggered by <c-x><c-o>
  vim.api.nvim_buf_set_option(bufnr, "omnifunc", "v:lua.vim.lsp.omnifunc")

  -- Buffer local mappings
  local opts = { noremap = true, silent = true, buffer = bufnr }
  
  -- Navigation mappings
  vim.keymap.set("n", "gD", vim.lsp.buf.declaration, opts)
  vim.keymap.set("n", "gd", vim.lsp.buf.definition, opts)
  vim.keymap.set("n", "K", vim.lsp.buf.hover, opts)
  vim.keymap.set("n", "gi", vim.lsp.buf.implementation, opts)
  vim.keymap.set("n", "gr", vim.lsp.buf.references, opts)
  
  -- Code action mappings
  vim.keymap.set("n", "<space>rn", vim.lsp.buf.rename, opts)
  vim.keymap.set("n", "<space>ca", vim.lsp.buf.code_action, opts)
  vim.keymap.set("n", "<space>f", function()
    vim.lsp.buf.format({ async = true })
  end, opts)
  
  -- Signature help
  vim.keymap.set("n", "<C-k>", vim.lsp.buf.signature_help, opts)
  vim.keymap.set("i", "<C-k>", vim.lsp.buf.signature_help, opts)
  
  -- Workspace folder mappings
  vim.keymap.set("n", "<space>wa", vim.lsp.buf.add_workspace_folder, opts)
  vim.keymap.set("n", "<space>wr", vim.lsp.buf.remove_workspace_folder, opts)
  vim.keymap.set("n", "<space>wl", function()
    print(vim.inspect(vim.lsp.buf.list_workspace_folders()))
  end, opts)
  
  -- Type definition
  vim.keymap.set("n", "<space>D", vim.lsp.buf.type_definition, opts)

  -- Diagnostics mappings
  vim.keymap.set("n", "<space>e", vim.diagnostic.open_float, opts)
  vim.keymap.set("n", "[d", vim.diagnostic.goto_prev, opts)
  vim.keymap.set("n", "]d", vim.diagnostic.goto_next, opts)
  vim.keymap.set("n", "<space>q", vim.diagnostic.setloclist, opts)

  -- Auto-format on save
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

  -- Configure diagnostics
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

  print("Carrion LSP attached to buffer " .. bufnr)
end

-- Register Carrion LSP configuration
local configs = require("lspconfig.configs")
if not configs.carrion_lsp then
  configs.carrion_lsp = {
    default_config = default_config,
    docs = {
      description = "Language Server for the Carrion programming language",
      default_config = {
        root_dir = "util.find_git_ancestor(fname) or util.path.dirname(fname)",
      },
    },
  }
end

-- Setup function
function M.setup(opts)
  opts = opts or {}
  
  -- Merge user options with defaults
  local config = vim.tbl_deep_extend("force", {
    on_attach = default_on_attach,
    capabilities = vim.lsp.protocol.make_client_capabilities(),
  }, default_config, opts)

  -- If nvim-cmp is available, enhance capabilities
  local has_cmp, cmp_nvim_lsp = pcall(require, "cmp_nvim_lsp")
  if has_cmp then
    config.capabilities = cmp_nvim_lsp.default_capabilities(config.capabilities)
  end

  -- Setup the LSP
  lspconfig.carrion_lsp.setup(config)
  
  vim.notify("Carrion LSP configured successfully", vim.log.levels.INFO)
end

-- Auto-setup if this file is sourced directly
M.setup()

return M