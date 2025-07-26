-- Health check for carrion-lsp.nvim
-- Used by :checkhealth carrion-lsp

local M = {}

function M.check()
  local health = vim.health or require("health")
  local installer = require("carrion-lsp.installer")
  
  health.report_start("carrion-lsp.nvim")
  
  -- Check Neovim version
  if vim.fn.has('nvim-0.8') == 1 then
    health.report_ok("Neovim version: " .. vim.version().major .. "." .. vim.version().minor)
  else
    health.report_error("Neovim 0.8+ required", {
      "Update Neovim to version 0.8 or higher"
    })
  end
  
  -- Check if carrion-lsp binary is available
  if installer.is_installed() then
    local version = installer.get_version()
    health.report_ok("carrion-lsp binary found: " .. (version or "unknown version"))
  else
    health.report_error("carrion-lsp binary not found", {
      "Run :CarrionLspInstall to install automatically",
      "Or install manually: curl -fsSL https://raw.githubusercontent.com/javanhut/carrion-lsp/main/install.sh | bash"
    })
  end
  
  -- Check if nvim-lspconfig is available
  local has_lspconfig, lspconfig = pcall(require, "lspconfig")
  if has_lspconfig then
    health.report_ok("nvim-lspconfig found")
    
    -- Check if carrion-lsp is configured
    local configs = require("lspconfig.configs")
    if configs.carrion_lsp then
      health.report_ok("carrion-lsp configuration loaded")
    else
      health.report_warn("carrion-lsp not yet configured (will be configured on first .crl file)")
    end
  else
    health.report_error("nvim-lspconfig not found", {
      "Install nvim-lspconfig: https://github.com/neovim/nvim-lspconfig"
    })
  end
  
  -- Check if cmp is available (optional)
  local has_cmp, _ = pcall(require, "cmp")
  if has_cmp then
    health.report_ok("nvim-cmp found (enhanced autocompletion enabled)")
    
    local has_cmp_lsp, _ = pcall(require, "cmp_nvim_lsp")
    if has_cmp_lsp then
      health.report_ok("cmp-nvim-lsp found (LSP completions enabled)")
    else
      health.report_warn("cmp-nvim-lsp not found", {
        "Install cmp-nvim-lsp for LSP completions: https://github.com/hrsh7th/cmp-nvim-lsp"
      })
    end
  else
    health.report_warn("nvim-cmp not found (basic autocompletion only)", {
      "Install nvim-cmp for enhanced autocompletion: https://github.com/hrsh7th/nvim-cmp"
    })
  end
  
  -- Check filetype detection
  local ft = vim.filetype.match({ filename = "test.crl" })
  if ft == "carrion" then
    health.report_ok("Carrion filetype detection working")
  else
    health.report_error("Carrion filetype detection not working")
  end
  
  -- Check if LSP is currently running
  local lsp = require("carrion-lsp.lsp")
  if lsp.is_running() then
    local client = lsp.get_client()
    health.report_ok("carrion-lsp server running (ID: " .. client.id .. ")")
  else
    health.report_info("carrion-lsp server not currently running (will start when .crl file is opened)")
  end
  
  -- Check PATH for carrion-lsp
  if vim.fn.executable("carrion-lsp") == 1 then
    local which_result = vim.trim(vim.fn.system("which carrion-lsp"))
    health.report_ok("carrion-lsp found in PATH: " .. which_result)
  else
    health.report_warn("carrion-lsp not in PATH", {
      "Add ~/.local/bin to your PATH if you installed using the installer",
      "Or make sure carrion-lsp is installed and accessible"
    })
  end
  
  -- Check for common optional dependencies
  local optional_deps = {
    { "mason.nvim", "williamboman/mason.nvim" },
    { "telescope.nvim", "nvim-telescope/telescope.nvim" },
    { "trouble.nvim", "folke/trouble.nvim" },
  }
  
  for _, dep in ipairs(optional_deps) do
    local has_dep, _ = pcall(require, dep[1])
    if has_dep then
      health.report_ok(dep[1] .. " found (enhanced features available)")
    else
      health.report_info(dep[1] .. " not found (optional)", {
        "Install " .. dep[2] .. " for enhanced features"
      })
    end
  end
end

return M