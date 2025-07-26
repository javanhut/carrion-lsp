-- Main carrion-lsp module
local M = {}

local installer = require("carrion-lsp.installer")

-- Default configuration
local default_config = {
  cmd = nil, -- Will be set dynamically
  filetypes = { "carrion" },
  root_dir = function(fname)
    return vim.fs.dirname(vim.fs.find({ "go.mod", ".git" }, { upward = true })[1])
  end,
  settings = {},
  capabilities = nil,
  on_attach = nil,
}

-- Current configuration
M.config = nil

-- Setup function
function M.setup(opts)
  opts = opts or {}
  M.config = vim.tbl_deep_extend("force", default_config, opts)
  
  -- Set the command to the installed binary
  if not M.config.cmd then
    M.config.cmd = { installer.get_binary_path() }
  end
  
  -- Auto-install if binary doesn't exist
  if not installer.is_installed() then
    vim.schedule(function()
      vim.notify("carrion-lsp binary not found. Run :CarrionLspInstall to install it.", vim.log.levels.WARN)
    end)
    return
  end
  
  -- Configure LSP
  local lspconfig = require("lspconfig")
  local configs = require("lspconfig.configs")
  
  -- Register carrion-lsp if not already registered
  if not configs.carrion_lsp then
    configs.carrion_lsp = {
      default_config = M.config,
    }
  end
  
  -- Setup the LSP server
  lspconfig.carrion_lsp.setup(M.config)
end

-- Check installation and configuration
function M.check()
  print("Carrion LSP Status:")
  print("------------------")
  
  -- Check if binary exists
  local binary_path = installer.get_binary_path()
  local is_installed = installer.is_installed()
  
  print("Binary path: " .. binary_path)
  print("Installed: " .. (is_installed and "Yes" or "No"))
  
  if is_installed then
    -- Try to get version
    local handle = vim.loop.spawn(binary_path, {
      args = {"--version"},
      stdio = {nil, "pipe", "pipe"}
    }, function(code, signal)
      if code == 0 then
        print("Binary is working")
      else
        print("Binary failed to run (exit code: " .. code .. ")")
      end
    end)
    
    if handle then
      vim.loop.close(handle)
    end
  end
  
  -- Check configuration
  if M.config then
    print("Configuration: Loaded")
    print("Filetypes: " .. table.concat(M.config.filetypes, ", "))
  else
    print("Configuration: Not loaded (call setup() first)")
  end
  
  -- Check if LSP is active for current buffer
  local clients = vim.lsp.get_active_clients({ name = "carrion_lsp" })
  print("Active clients: " .. #clients)
end

return M