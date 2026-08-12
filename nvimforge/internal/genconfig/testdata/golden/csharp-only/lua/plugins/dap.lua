
return {
  "mfussenegger/nvim-dap",
  dependencies = { "rcarriga/nvim-dap-ui", "nvim-neotest/nvim-nio" },
  config = function()
    local dap = require("dap")
    local dapui = require("dapui")






    dap.adapters.coreclr = {
      type = "executable",
      command = vim.fn.stdpath("data") .. "/mason/bin/netcoredbg",
      args = { "--interpreter=vscode" },
    }
    dap.configurations.cs = {
      {
        type = "coreclr",
        name = "Launch",
        request = "launch",
        justMyCode = false,
        program = function()
          -- Debug output lands in bin/Debug/<tfm>/<Project>.dll, which is
          -- tedious to type: offer the first match as the default.
          local root = vim.fs.root(0, function(name)
            return name:match("%.csproj$") ~= nil
          end) or vim.fn.getcwd()
          local dlls = vim.fn.glob(root .. "/bin/Debug/*/*.dll", true, true)
          return vim.fn.input("Path to dll: ", dlls[1] or (root .. "/bin/Debug/"), "file")
        end,
      },
    }

    dapui.setup()
    dap.listeners.after.event_initialized["dapui_config"] = function()
      dapui.open()
    end
    dap.listeners.before.event_terminated["dapui_config"] = function()
      dapui.close()
    end
    dap.listeners.before.event_exited["dapui_config"] = function()
      dapui.close()
    end
  end,
}

