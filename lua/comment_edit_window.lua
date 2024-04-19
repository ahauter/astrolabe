popup = require('plenary.popup')

insert_line = nil
background_buffer = nil
foreground_buffer = nil
instruction_buffer = nil
window_id = nil
local M = {}

function M.SetBuffer(lines)
  if foreground_buffer == nil then
    return
  end
  local num_lines = vim.api.nvim_buf_line_count(foreground_buffer)
  vim.api.nvim_buf_set_lines(foreground_buffer, 0, num_lines + 1, false, lines)
end

function M.AddInstructions(lines)
  if instruction_buffer == nil then
    return
  end
  local num_lines = vim.api.nvim_buf_line_count(instruction_buffer)
  if num_lines > 2 then
    return
  end
  vim.api.nvim_buf_set_lines(foreground_buffer, num_lines + 1, num_lines + 1, false, lines)
end

function InsertComment()
  comment_lines = vim.api.nvim_buf_get_lines(
    foreground_buffer,
    0,
    vim.api.nvim_buf_line_count(foreground_buffer) - 1,
    false
  )
  vim.api.nvim_buf_set_lines(background_buffer, insert_line, insert_line, false, comment_lines)
  CloseWindow()
end

function CloseWindow()
  if window_id == nil then
    return
  end
  vim.api.nvim_win_close(window_id, true)
  window_id = nil
  foreground_buffer = nil
  background_buffer = nil
end

function M.MakePopup()
  if foreground_buffer ~= nil then
    return
  end
  background_buffer = vim.api.nvim_win_get_buf(0)
  foreground_buffer = vim.api.nvim_create_buf(true, true)
  if instruction_buffer == nil then
    instruction_buffer = vim.api.nvim_create_buf(true, true)
  end
  --TODO logic for comments under function header
  insert_line = vim.fn.getpos("'<")[2] - 1
  local width = 80
  local height = 20
  local borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" }
  local win_id, win = popup.create(foreground_buffer, {
    title = "Astrolabe",
    highlight = "Astrolabe",
    line = math.floor(((vim.o.lines - height) / 2) - 1),
    col = math.floor((vim.o.columns - width) / 2),
    minwidth = width,
    minheight = height,
    borderchars = borderchars,
  })
  local win_id, win = popup.create(instruction_buffer, {
    title = "Astrolabe",
    highlight = "Astrolabe",
    line = math.floor(((vim.o.lines - height) / 2) - 1),
    col = math.floor((vim.o.columns - width) / 2),
    minwidth = width,
    minheight = height,
    borderchars = borderchars,
  })
  window_id = win_id
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "q",
    ":<C-u>call v:lua.CloseWindow()<CR>",
    { silent = true })
  return win_id
end

function M.AllowSaving()
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "x",
    ":<C-u>call v:lua.InsertComment()<CR>",
    { silent = true })
end

function M.DisableSaving()
  vim.api.nvim_buf_del_keymap(foreground_buffer, "n", "x")
end

return M
