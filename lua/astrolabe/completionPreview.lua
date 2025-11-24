local log_path = "/home/austin/astrologs/client.nvim.astro.log"
local ghost_text = "CmpGhostText"
local ns = vim.api.nvim_create_namespace("astrolabe")
local log = require('plenary.log').new({
  plugin = "astrolabe",
  level = "debug",
  use_console = "sync",
  use_file = true,
  outfile = log_path
})
local completion = "hello world"
local line = 0
local M = {}
local extmark_table = {}


local function update_state()
  local current_line = vim.api.nvim_buf_get_lines(
    0, line, line, false
  )
  log.debug(current_line)

  if #current_line >= #completion then
    return
  end

  for pos = 1, #current_line do
    local completion_char = completion:sub(pos, pos)
    local actual_char = current_line:sub(pos, pos)

    if completion_char ~= actual_char then
      --maybe remove the completion..
    end
  end
  local remaining_line = completion:sub(#current_line + 1, -1)
  table.insert(extmark_table, vim.api.nvim_buf_set_extmark(
    vim.api.nvim_buf_get_name(0),
    ns, line, #current_line, {
      virt_text = { { remaining_line, ghost_text } },
      virt_text_pos = 'overlay'
    }
  ))
end

local function clear_completion()
  local ext_id = table.remove(extmark_table)
  local current_bfnr = vim.api.nvim_buf_get_name(0)
  while ext_id ~= nil do
    vim.api.nvim_buf_del_extmark(current_bfnr, ns, ext_id)
    ext_id = table.remove(extmark_table)
  end
end

function M.SetCompletion(comp, l)
  completion = comp
  line = l
  update_state()
end

function M.ClearCompletion()
  completion = ""
  clear_completion()
end

return M
