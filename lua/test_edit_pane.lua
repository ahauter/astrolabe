local test_edit_buffer = nil
local window_id = nil

local M = {}

function M.SetBuffer(lines)
  if test_edit_buffer == nil then
    return
  end
  local num_lines = vim.api.nvim_buf_line_count(test_edit_buffer)
  vim.api.nvim_buf_set_lines(test_edit_buffer, 0, num_lines + 1, false, lines)
end

function M.SetBufferName(name)
  print(test_edit_buffer)
  print(name)
  vim.api.nvim_buf_set_name(test_edit_buffer, name)
end

--TODO Maybe add a vertical split in buffer
function M.MakePopup()
  if test_edit_buffer ~= nil then
    return
  end
  test_edit_buffer = vim.api.nvim_create_buf(true, true)
  if instruction_buffer == nil then
    local instruction_buffer = vim.api.nvim_create_buf(true, true)
  end
  --TODO logic for comments under function header
  local insert_line = vim.fn.getpos("'<")[2] - 1
  local width = 100
  local height = 40
  local borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" }
  vim.cmd("vsplit")
  local win_id = vim.api.nvim_get_current_win()
  vim.api.nvim_win_set_buf(win_id, test_edit_buffer)
  window_id = win_id
  vim.api.nvim_buf_set_keymap(test_edit_buffer,
    "n", "q",
    ":<C-u>call v:lua.CloseTestWindow()<CR>",
    { silent = true })
  vim.api.nvim_buf_set_keymap(test_edit_buffer,
    "n", "<esc>",
    ":<C-u>call v:lua.CloseTestWindow()<CR>",
    { silent = true })
  return win_id
end

function CloseTestWindow()
  vim.api.nvim_win_close(window_id, true)
  window_id = nil
  test_edit_buffer = nil
end

return M
