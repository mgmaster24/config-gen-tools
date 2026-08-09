
return {
  "mfussenegger/nvim-dap",
  dependencies = { "rcarriga/nvim-dap-ui", "nvim-neotest/nvim-nio" },
  config = function()
    local dap = require("dap")
    local dapui = require("dapui")

    dap.adapters.codelldb = {
      type = "server",
      port = "${port}",
      executable = {
        command = vim.fn.stdpath("data") .. "/mason/bin/codelldb",
        args = { "--port", "${port}" },
      },
    }

    local codelldb_config = {
      name = "Launch",
      type = "codelldb",
      request = "launch",
      cwd = "${workspaceFolder}",
      stopOnEntry = false,
      program = function()
        return vim.fn.input("Path to executable: ", vim.fn.getcwd() .. "/", "file")
      end,
    }


    dap.configurations.rust = { codelldb_config }


    dap.configurations.c = { codelldb_config }
    dap.configurations.cpp = { codelldb_config }


    dap.adapters.delve = {
      type = "server",
      port = "${port}",
      executable = { command = "dlv", args = { "dap", "-l", "127.0.0.1:${port}" } },
    }
    dap.configurations.go = {
      { type = "delve", name = "Debug", request = "launch", program = "${file}" },
    }


    dap.adapters.python = {
      type = "executable",
      command = vim.fn.stdpath("data") .. "/mason/packages/debugpy/venv/bin/python",
      args = { "-m", "debugpy.adapter" },
    }
    dap.configurations.python = {
      { type = "python", request = "launch", name = "Launch file", program = "${file}" },
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

