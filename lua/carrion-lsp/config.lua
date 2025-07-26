-- Configuration handling for carrion-lsp.nvim

local M = {}

-- Deep merge two tables
local function merge_tables(defaults, overrides)
  local result = vim.deepcopy(defaults)
  
  for key, value in pairs(overrides) do
    if type(value) == "table" and type(result[key]) == "table" then
      result[key] = merge_tables(result[key], value)
    else
      result[key] = value
    end
  end
  
  return result
end

-- Validate configuration
local function validate_config(config)
  local errors = {}
  
  -- Validate keymaps
  if config.keymaps then
    for action, keymap in pairs(config.keymaps) do
      if type(keymap) ~= "string" then
        table.insert(errors, string.format("keymaps.%s must be a string, got %s", action, type(keymap)))
      end
    end
  end
  
  -- Validate server command
  if config.server and config.server.cmd then
    if type(config.server.cmd) ~= "table" then
      table.insert(errors, "server.cmd must be a table")
    elseif #config.server.cmd == 0 then
      table.insert(errors, "server.cmd cannot be empty")
    end
  end
  
  if #errors > 0 then
    error("carrion-lsp configuration errors:\n" .. table.concat(errors, "\n"))
  end
end

-- Setup configuration
function M.setup(defaults, user_config)
  local config = merge_tables(defaults, user_config)
  
  -- Validate the merged configuration
  validate_config(config)
  
  -- Store config for other modules
  M.current = config
  
  return config
end

-- Get current configuration
function M.get()
  return M.current or {}
end

-- Get specific configuration value with fallback
function M.get_value(path, fallback)
  local config = M.get()
  local keys = vim.split(path, ".", { plain = true })
  local value = config
  
  for _, key in ipairs(keys) do
    if type(value) == "table" and value[key] ~= nil then
      value = value[key]
    else
      return fallback
    end
  end
  
  return value
end

return M