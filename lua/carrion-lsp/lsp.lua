-- LSP management utilities
local M = {}

-- Restart the carrion-lsp server
function M.restart()
  local clients = vim.lsp.get_active_clients({ name = "carrion_lsp" })
  
  if #clients == 0 then
    vim.notify("No active carrion-lsp clients found", vim.log.levels.WARN)
    return
  end
  
  for _, client in pairs(clients) do
    local bufs = vim.lsp.get_buffers_by_client_id(client.id)
    client.stop()
    
    vim.defer_fn(function()
      for _, buf in pairs(bufs) do
        if vim.api.nvim_buf_is_valid(buf) and vim.bo[buf].filetype == "carrion" then
          vim.cmd("edit")
        end
      end
    end, 500)
  end
  
  vim.notify("Restarted carrion-lsp server", vim.log.levels.INFO)
end

-- Get active carrion-lsp clients
function M.get_clients()
  return vim.lsp.get_active_clients({ name = "carrion_lsp" })
end

-- Check if carrion-lsp is attached to current buffer
function M.is_attached()
  local clients = M.get_clients()
  local current_buf = vim.api.nvim_get_current_buf()
  
  for _, client in pairs(clients) do
    local bufs = vim.lsp.get_buffers_by_client_id(client.id)
    for _, buf in pairs(bufs) do
      if buf == current_buf then
        return true
      end
    end
  end
  
  return false
end

return M