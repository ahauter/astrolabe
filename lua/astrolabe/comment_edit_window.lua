popup = require('plenary.popup')
local log = require('plenary.log')
insert_line = nil
foreground_buffer = nil
instruction_buffer = nil
window_id = nil
second_window_id = nil
local M = {}

local log_path = vim.loop.cwd() .. "/selected.astro.log"
print(log_path)
-- Create a custom logger
local log = require('plenary.log').new({
  plugin = "my_plugin",
  level = "debug",
  use_console = "sync",
  use_file = true,
  outfile = log_path
})

M.file_buffer = nil

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
  vim.api.nvim_buf_set_lines(instruction_buffer, 0, num_lines + 1, false, lines)
end

function InsertComment()
  comment_lines = vim.api.nvim_buf_get_lines(
    foreground_buffer,
    0,
    vim.api.nvim_buf_line_count(foreground_buffer) - 1,
    false
  )
  log.info("InsertComment: ")
  log.info(table.concat(comment_lines, "\n"))
  -- a new line to the comment does not get inserted ..
  vim.api.nvim_buf_set_lines(M.file_buffer, insert_line, insert_line, false, comment_lines)
  CloseWindow()
end

function CloseWindow()
  if window_id == nil then
    return
  end
  vim.api.nvim_win_close(window_id, true)
  vim.api.nvim_win_close(second_window_id, true)
  window_id = nil
  foreground_buffer = nil
  M.file_buffer = nil
end

function M.MakePopup()
  if foreground_buffer ~= nil then
    return
  end
  M.file_buffer = vim.api.nvim_win_get_buf(0)
  print(M.file_buffer)
  foreground_buffer = vim.api.nvim_create_buf(true, true)
  if instruction_buffer == nil then
    instruction_buffer = vim.api.nvim_create_buf(true, true)
  end
  --TODO logic for comments under function header
  insert_line = vim.fn.getpos("'<")[2] - 1
  local width = 80
  local height = 20
  local borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" }
  local second_wind_id, win = popup.create(instruction_buffer, {
    title = "Astrolabe",
    highlight = "Astrolabe",
    focusable = false,
    line = math.floor(((vim.o.lines - height) / 2) - 1),
    col = math.floor((vim.o.columns - width) / 2),
    minwidth = width,
    minheight = 5,
    maxheight = 5,
    borderchars = borderchars,
  })
  local win_id, win = popup.create(foreground_buffer, {
    line = math.floor(((vim.o.lines - height) / 2) + 5),
    col = math.floor((vim.o.columns - width) / 2),
    minwidth = width,
    minheight = 15,
    borderchars = borderchars,
  })
  window_id = win_id
  second_window_id = second_wind_id
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "q",
    ":<C-u>call v:lua.CloseWindow()<CR>",
    { silent = true })
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "<esc>",
    ":<C-u>call v:lua.CloseWindow()<CR>",
    { silent = true })
  return win_id
end

function M.AllowSaving()
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "x",
    ":<C-u>call v:lua.InsertComment()<CR>",
    { silent = true })
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "r",
    ":<C-u>call v:lua.GenerateComment()<CR>",
    { silent = true })
  vim.api.nvim_buf_set_keymap(foreground_buffer,
    "n", "t",
    ":<C-u>call v:lua.CreateTests()<CR>",
    { silent = true })
end

function M.GetCommentLines()
  if foreground_buffer == nil then
    return {}
  end
  comment_lines = vim.api.nvim_buf_get_lines(
    foreground_buffer,
    0,
    vim.api.nvim_buf_line_count(foreground_buffer) - 1,
    false
  )
  return comment_lines
end

function M.DisableSaving()
  vim.api.nvim_buf_del_keymap(foreground_buffer, "n", "x")
end

return M
