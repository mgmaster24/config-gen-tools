return {
  "stevearc/conform.nvim",
  event = { "BufWritePre" },
  cmd = { "ConformInfo" },
  opts = {
    formatters_by_ft = {
      ["bash"] = { "shfmt" },
      ["c"] = { "clang-format" },
      ["cpp"] = { "clang-format" },
      ["cs"] = { "csharpier" },
      ["go"] = { "goimports" },
      ["javascript"] = { "prettierd" },
      ["javascriptreact"] = { "prettierd" },
      ["lua"] = { "stylua" },
      ["python"] = { "ruff" },
      ["rust"] = { "rustfmt" },
      ["sh"] = { "shfmt" },
      ["typescript"] = { "prettierd" },
      ["typescriptreact"] = { "prettierd" },
      ["yaml"] = { "yamlfmt" },
    },
    format_on_save = {
      timeout_ms = 1000,
      lsp_format = "fallback",
    },
  },
}
