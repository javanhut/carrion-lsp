local M = {}

M.treesitter = {
  ensure_installed = {
    "vim",
    "lua",
    "html",
    "css",
    "javascript",
    "typescript",
    "tsx",
    "c",
    "markdown",
    "markdown_inline",
    "python",
    "go",
    "rust",
    "json",
    "yaml",
    "toml",
    "bash",
  },
  indent = {
    enable = true,
  },
  highlight = {
    enable = true,
    additional_vim_regex_highlighting = false,
  },
}

M.mason = {
  ensure_installed = {
    -- lua stuff
    "lua-language-server",
    "stylua",

    -- web dev stuff
    "css-lsp",
    "html-lsp",
    "typescript-language-server",
    "deno",
    "prettier",

    -- c/cpp stuff
    "clangd",
    "clang-format",

    -- python
    "pyright",
    "black",

    -- Note: carrion-lsp must be installed manually using the install script
    -- Run: curl -fsSL https://raw.githubusercontent.com/javanhut/carrion-lsp/main/install.sh | bash
  },
}

-- git support in nvimtree
M.nvimtree = {
  git = {
    enable = true,
  },

  renderer = {
    highlight_git = true,
    icons = {
      show = {
        git = true,
      },
    },
  },

  view = {
    adaptive_size = true,
  },

  -- Add Carrion files to the tree
  filters = {
    custom = { "^.git$" },
  },
}

M.telescope = {
  defaults = {
    file_ignore_patterns = {
      "node_modules",
      ".git/",
      "target/",
      "build/",
      "*.o",
      "*.so",
      "*.dylib",
      "*.dll",
    },
  },
  extensions_list = { "themes", "terms" },
  extensions = {
    fzf = {
      fuzzy = true,
      override_generic_sorter = true,
      override_file_sorter = true,
      case_mode = "smart_case",
    },
  },
}

M.cmp = {
  sources = {
    { name = "nvim_lsp", priority = 1000 },
    { name = "luasnip", priority = 750 },
    { name = "buffer", priority = 500 },
    { name = "nvim_lua", priority = 400 },
    { name = "path", priority = 250 },
  },
  mapping = {
    ["<C-p>"] = require("cmp").mapping.select_prev_item(),
    ["<C-n>"] = require("cmp").mapping.select_next_item(),
    ["<C-d>"] = require("cmp").mapping.scroll_docs(-4),
    ["<C-f>"] = require("cmp").mapping.scroll_docs(4),
    ["<C-Space>"] = require("cmp").mapping.complete(),
    ["<C-e>"] = require("cmp").mapping.close(),
    ["<CR>"] = require("cmp").mapping.confirm({
      behavior = require("cmp").ConfirmBehavior.Insert,
      select = true,
    }),
    ["<Tab>"] = require("cmp").mapping(function(fallback)
      if require("cmp").visible() then
        require("cmp").select_next_item()
      elseif require("luasnip").expand_or_jumpable() then
        require("luasnip").expand_or_jump()
      else
        fallback()
      end
    end, { "i", "s" }),
    ["<S-Tab>"] = require("cmp").mapping(function(fallback)
      if require("cmp").visible() then
        require("cmp").select_prev_item()
      elseif require("luasnip").jumpable(-1) then
        require("luasnip").jump(-1)
      else
        fallback()
      end
    end, { "i", "s" }),
  },
  formatting = {
    format = function(entry, vim_item)
      local icons = require("nvchad.icons.lspkind")
      vim_item.kind = string.format("%s %s", icons[vim_item.kind], vim_item.kind)
      vim_item.menu = ({
        nvim_lsp = "[LSP]",
        luasnip = "[Snippet]",
        buffer = "[Buffer]",
        path = "[Path]",
        nvim_lua = "[Lua]",
      })[entry.source.name]
      return vim_item
    end,
  },
}

return M