-- Installer module for carrion-lsp
local M = {}

local uv = vim.loop

-- Get platform-specific binary name and URLs
local function get_platform_info()
  local os_name = uv.os_uname().sysname:lower()
  local arch = uv.os_uname().machine:lower()
  
  local platform = ""
  if os_name:match("linux") then
    platform = "linux"
  elseif os_name:match("darwin") then
    platform = "darwin"
  elseif os_name:match("windows") then
    platform = "windows"
  else
    error("Unsupported operating system: " .. os_name)
  end
  
  local binary_name = "carrion-lsp"
  if platform == "windows" then
    binary_name = binary_name .. ".exe"
  end
  
  return platform, arch, binary_name
end

-- Get installation directory
local function get_install_dir()
  local data_path = vim.fn.stdpath("data")
  return data_path .. "/carrion-lsp"
end

-- Get binary path
local function get_binary_path()
  local install_dir = get_install_dir()
  local _, _, binary_name = get_platform_info()
  return install_dir .. "/" .. binary_name
end

-- Check if binary exists and is executable
local function is_installed()
  local binary_path = get_binary_path()
  local stat = uv.fs_stat(binary_path)
  if not stat then
    return false
  end
  
  -- Check if it's executable
  local handle = uv.spawn(binary_path, {
    args = {"--version"},
    stdio = {nil, nil, nil}
  }, function() end)
  
  if handle then
    uv.close(handle)
    return true
  end
  
  return false
end

-- Build from source
local function build_from_source()
  local install_dir = get_install_dir()
  local _, _, binary_name = get_platform_info()
  local binary_path = install_dir .. "/" .. binary_name
  
  -- Create install directory
  vim.fn.mkdir(install_dir, "p")
  
  -- Build the binary from plugin directory
  local plugin_dir = vim.fn.fnamemodify(debug.getinfo(1, "S").source:sub(2), ":h:h:h")
  local cmd = "cd " .. plugin_dir .. " && go build -o " .. binary_path .. " cmd/carrion-lsp/main.go"
  
  print("Building carrion-lsp from source...")
  local result = vim.fn.system(cmd)
  
  if vim.v.shell_error ~= 0 then
    error("Failed to build carrion-lsp: " .. result)
  end
  
  print("Successfully built carrion-lsp at: " .. binary_path)
  return binary_path
end

-- Create symlink in ~/.local/bin for PATH access
local function create_symlink()
  local binary_path = get_binary_path()
  local local_bin = os.getenv("HOME") .. "/.local/bin"
  local symlink_path = local_bin .. "/carrion-lsp"
  
  -- Create ~/.local/bin if it doesn't exist
  vim.fn.mkdir(local_bin, "p")
  
  -- Remove existing symlink if it exists
  if uv.fs_stat(symlink_path) then
    uv.fs_unlink(symlink_path)
  end
  
  -- Create symlink
  local success, err = uv.fs_symlink(binary_path, symlink_path)
  if success then
    print("Created symlink: " .. symlink_path)
  else
    print("Warning: Failed to create symlink in ~/.local/bin: " .. (err or "unknown error"))
    print("You may need to manually add " .. binary_path .. " to your PATH")
  end
end

-- Install the LSP server
function M.install()
  if is_installed() then
    print("carrion-lsp is already installed at: " .. get_binary_path())
    return
  end
  
  build_from_source()
  create_symlink()
end

-- Update the LSP server
function M.update()
  build_from_source()
  create_symlink()
end

-- Uninstall the LSP server
function M.uninstall()
  local binary_path = get_binary_path()
  local install_dir = get_install_dir()
  local symlink_path = os.getenv("HOME") .. "/.local/bin/carrion-lsp"
  
  if uv.fs_stat(binary_path) then
    uv.fs_unlink(binary_path)
    print("Removed carrion-lsp binary")
  end
  
  -- Remove symlink
  if uv.fs_stat(symlink_path) then
    uv.fs_unlink(symlink_path)
    print("Removed symlink: " .. symlink_path)
  end
  
  -- Remove install directory if empty
  local handle = uv.fs_scandir(install_dir)
  if handle then
    local name = uv.fs_scandir_next(handle)
    if not name then
      uv.fs_rmdir(install_dir)
      print("Removed empty install directory")
    end
  end
end

-- Get binary path for external use
function M.get_binary_path()
  return get_binary_path()
end

-- Check if installed
function M.is_installed()
  return is_installed()
end

-- Get version of installed binary
function M.get_version()
  if not is_installed() then
    return nil
  end
  
  local binary_path = get_binary_path()
  local result = vim.fn.system(binary_path .. " --version 2>/dev/null")
  
  if vim.v.shell_error == 0 then
    return vim.trim(result)
  end
  
  return "unknown"
end

return M