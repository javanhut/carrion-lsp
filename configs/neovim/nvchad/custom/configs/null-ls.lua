local present, null_ls = pcall(require, "null-ls")

if not present then
  return
end

local b = null_ls.builtins

local sources = {
  -- webdev stuff
  b.formatting.deno_fmt.with({ filetypes = { "javascript", "typescript", "json" } }),
  b.formatting.prettier.with({ filetypes = { "html", "markdown", "css" } }),

  -- Lua
  b.formatting.stylua,

  -- cpp
  b.formatting.clang_format,

  -- python
  b.formatting.black,
  b.diagnostics.pylint,

  -- Note: For Carrion files, formatting is handled by the carrion-lsp server
  -- Additional diagnostics can be added here if needed
}

null_ls.setup({
  debug = true,
  sources = sources,
  
  -- Format on save for supported file types
  on_attach = function(client, bufnr)
    if client.supports_method("textDocument/formatting") then
      vim.api.nvim_clear_autocmds({ group = augroup, buffer = bufnr })
      vim.api.nvim_create_autocmd("BufWritePre", {
        group = augroup,
        buffer = bufnr,
        callback = function()
          -- Don't format Carrion files here as it's handled by carrion-lsp
          if vim.bo[bufnr].filetype ~= "carrion" then
            vim.lsp.buf.format({ bufnr = bufnr })
          end
        end,
      })
    end
  end,
})

local augroup = vim.api.nvim_create_augroup("LspFormatting", {})