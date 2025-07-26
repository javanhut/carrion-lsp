-- Auto-loaded plugin setup for carrion-lsp.nvim
-- This file is automatically sourced when Neovim starts

-- Prevent loading if already loaded
if vim.g.loaded_carrion_lsp then
  return
end
vim.g.loaded_carrion_lsp = 1

-- Only load if Neovim version is supported
if vim.fn.has('nvim-0.8') == 0 then
  vim.api.nvim_err_writeln('carrion-lsp.nvim requires Neovim 0.8 or higher')
  return
end

-- Create user commands immediately (before setup is called)
vim.api.nvim_create_user_command("CarrionLspInstall", function()
  require("carrion-lsp.installer").install()
end, { desc = "Install carrion-lsp binary" })

vim.api.nvim_create_user_command("CarrionLspUpdate", function()
  require("carrion-lsp.installer").update()
end, { desc = "Update carrion-lsp binary" })

vim.api.nvim_create_user_command("CarrionLspUninstall", function()
  require("carrion-lsp.installer").uninstall()
end, { desc = "Uninstall carrion-lsp binary" })

vim.api.nvim_create_user_command("CarrionLspInfo", function()
  require("carrion-lsp").check()
end, { desc = "Show carrion-lsp information" })

vim.api.nvim_create_user_command("CarrionLspRestart", function()
  require("carrion-lsp.lsp").restart()
end, { desc = "Restart carrion-lsp server" })

-- Set up filetype detection for Carrion files
-- This ensures .crl files are recognized even before the LSP is set up
vim.api.nvim_create_autocmd({"BufRead", "BufNewFile"}, {
  pattern = "*.crl",
  callback = function()
    vim.bo.filetype = "carrion"
  end,
  desc = "Set filetype for Carrion files"
})

-- Auto-setup if user hasn't called setup manually
-- This provides a zero-config experience for basic usage
vim.api.nvim_create_autocmd("FileType", {
  pattern = "carrion",
  once = true,
  callback = function()
    -- Check if user has already called setup
    local carrion_lsp = require("carrion-lsp")
    if not carrion_lsp.config then
      -- Auto-setup with defaults if not already configured
      carrion_lsp.setup()
    end
  end,
  desc = "Auto-setup carrion-lsp on first Carrion file"
})