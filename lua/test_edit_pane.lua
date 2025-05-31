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

function M.MakePopup(cur_buf)
  test_edit_buffer = vim.api.nvim_create_buf(true, true)
  local buf_type   = vim.api.nvim_buf_get_option(cur_buf, 'buftype')
  local file_type  = vim.api.nvim_buf_get_option(cur_buf, 'filetype')
  vim.api.nvim_buf_set_option(test_edit_buffer, 'buftype', buf_type)
  vim.api.nvim_buf_set_option(test_edit_buffer, 'filetype', file_type)
  --TODO logic for comments under function header
  vim.cmd("vsplit")
  local win_id = vim.api.nvim_get_current_win()
  vim.api.nvim_win_set_buf(win_id, test_edit_buffer)
  window_id = win_id
  return win_id
end

return M
