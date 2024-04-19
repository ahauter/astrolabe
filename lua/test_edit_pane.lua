local test_edit_buffer = nil
local instruction_buffer = nil
local window_id = nil
local second_window_id = nil

local M = {}

function M.SetBuffer(lines)
  if test_edit_buffer == nil then
    return
  end
  local num_lines = vim.api.nvim_buf_line_count(test_edit_buffer)
  vim.api.nvim_buf_set_lines(test_edit_buffer, 0, num_lines + 1, false, lines)
end

function M.AddInstructions(lines)
  if instruction_buffer == nil then
    return
  end
  local num_lines = vim.api.nvim_buf_line_count(instruction_buffer)
  vim.api.nvim_buf_set_lines(instruction_buffer, 0, num_lines + 1, false, lines)
end

function M.MakePopup()
  if test_edit_buffer ~= nil then
    return
  end
  test_edit_buffer = vim.api.nvim_create_buf(true, true)
  if instruction_buffer == nil then
    instruction_buffer = vim.api.nvim_create_buf(true, true)
  end
  --TODO logic for comments under function header
  insert_line = vim.fn.getpos("'<")[2] - 1
  local width = 120
  local height = 40
  local borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" }
  local second_wind_id, win = popup.create(instruction_buffer, {
    title = "Astrolabe",
    highlight = "Astrolabe",
    focusable = false,
    line = math.floor(((vim.o.lines - height) / 2) - 1),
    col = math.floor((vim.o.columns - width) / 2),
    minwidth = width,
    minheight = 5,
    borderchars = borderchars,
  })
  local win_id, win = popup.create(test_edit_buffer, {
    line = math.floor(((vim.o.lines - height) / 2) + 5),
    col = math.floor((vim.o.columns - width) / 2),
    minwidth = width,
    minheight = height - 5,
    borderchars = borderchars,
  })
  window_id = win_id
  second_window_id = second_wind_id
  M.AddInstructions({
    "q to Quit",
    "a to Append to to <test_file>",
    "c to Create new test file"
  })
  M.SetBuffer({
    "q to Quit",
    "a to Append to to <test_file>",
    "c to Create new test file"
  })
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
  vim.api.nvim_win_close(second_window_id, true)
  window_id = nil
  test_edit_buffer = nil
end

return M
