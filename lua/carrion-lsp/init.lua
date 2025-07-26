-- Main carrion-lsp module
local M = {}

local installer = require("carrion-lsp.installer")

-- Setup function
function M.setup(opts)
  opts = opts or {}
  
  -- Get binary path
  local binary_path = installer.get_binary_path()
  
  -- Auto-install if binary doesn't exist
  if not installer.is_installed() then
    vim.schedule(function()
      vim.notify("carrion-lsp binary not found. Run :CarrionLspInstall to install it.", vim.log.levels.WARN)
    end)
    return
  end
  
  -- Simple root directory function
  local function get_root_dir(fname)
    local found = vim.fs.find({ "go.mod", ".git", "Bifrost.toml" }, { upward = true, path = fname })
    if found and #found > 0 then
      return vim.fs.dirname(found[1])
    end
    return vim.fs.dirname(fname)
  end
  
  -- Configure LSP
  local lspconfig = require("lspconfig")
  local configs = require("lspconfig.configs")
  
  -- Register carrion-lsp
  configs.carrion_lsp = {
    default_config = {
      cmd = { binary_path },
      filetypes = { "carrion" },
      root_dir = get_root_dir,
      single_file_support = true,
      settings = {},
    }
  }
  
  -- Setup the LSP server
  lspconfig.carrion_lsp.setup(opts)
  
  -- Force attach to any open Carrion files
  vim.api.nvim_create_autocmd({ "FileType", "BufEnter", "BufWinEnter" }, {
    pattern = "*",
    callback = function(args)
      if vim.bo[args.buf].filetype == "carrion" then
        vim.schedule(function()
          local clients = vim.lsp.get_clients({ name = "carrion_lsp", bufnr = args.buf })
          if #clients == 0 then
            -- Force start the LSP
            lspconfig.carrion_lsp.launch()
          end
        end)
      end
    end,
  })
  
  -- Also attach to current buffer if it's already a Carrion file
  if vim.bo.filetype == "carrion" then
    vim.schedule(function()
      lspconfig.carrion_lsp.launch()
    end)
  end
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
    print("Configuration: Loaded")
    print("Filetypes: carrion")
  else
    print("Configuration: Not loaded (binary not found)")
  end
  
  -- Check if LSP is active for current buffer
  local clients = vim.lsp.get_clients({ name = "carrion_lsp" })
  print("Active clients: " .. #clients)
end

return M