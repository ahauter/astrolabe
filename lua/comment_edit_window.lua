popup = require('plenary.popup')

background_buffer = nil
foreground_buffer = nil
local M = {}

function M.SetBuffer()
  if foreground_buffer == nil then
    return
  end
  vim.api.nvim_buf_set_lines(foreground_buffer, 0, 0, false, { "#############", "Loading", "#############" })
end

function M.MakePopup()
  if foreground_buffer ~= nil then
    return
  end
  foreground_buffer = vim.api.nvim_create_buf(true, true)
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
  return win_id
end

return M
