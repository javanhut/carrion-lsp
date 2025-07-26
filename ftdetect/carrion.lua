-- Filetype detection for Carrion files

vim.filetype.add({
  extension = {
    crl = "carrion",
  },
  filename = {
    [".carrion"] = "carrion",
  },
  pattern = {
    [".*%.carrion%..*"] = "carrion",
  },
})