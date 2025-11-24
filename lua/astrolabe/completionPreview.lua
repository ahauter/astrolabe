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
local completion = ""
local line = 0
local M = {}
local extmark_table = {}

local function clear_completion()
  local ext_id = table.remove(extmark_table)
  while ext_id ~= nil do
    vim.api.nvim_buf_del_extmark(0, ns, ext_id)
    ext_id = table.remove(extmark_table)
  end
end


local function update_state()
  local current_line = vim.api.nvim_buf_get_lines(
    0, line, line + 1, false
  )
  log.debug("Current line: ", current_line)

  current_line = current_line[1]
  if #current_line >= #completion then
    log.info("Completion is too short!")
    clear_completion()
    completion = ""
    return
  end

  for pos = 1, #current_line do
    local completion_char = completion:sub(pos, pos)
    local actual_char = current_line:sub(pos, pos)

    if completion_char ~= actual_char then
      log.info("Compleion does not match typed value")
      clear_completion()
      completion = ""
      return
    end
  end
  local remaining_line = completion:sub(#current_line + 1, -1)
  table.insert(extmark_table, vim.api.nvim_buf_set_extmark(
    0, ns, line, #current_line, {
      virt_text = { { remaining_line, ghost_text } },
      virt_text_pos = 'overlay'
    }
  ))
end

function M.SetCompletion(comp, l)
  completion = comp
  log.debug(l)
  line = l
  update_state()
end

function M.AcceptCompletion()
  log.debug("Accept Completion")
  if not completion then
    log.debug("No valild completion")
    return
  end
  vim.api.nvim_buf_set_lines(
    0, line, line, false, { completion }
  )

  clear_completion()
  completion = ""
end

function M.ClearCompletion()
  completion = ""
  clear_completion()
end

return M
