popup = require('plenary.popup')
local id = nil
local cur_buffer = nil


local function attach_lsp(args)
  if id == nil then
    return
  end

  vim.lsp.buf_attach_client(args.buffer, id);
  cur_buffer = args.buffer
end

vim.api.nvim_create_autocmd("BufNew", {
  callback = attach_lsp
});

vim.api.nvim_create_autocmd("BufEnter", {
  callback = attach_lsp,
});

local function start_lsp()
  if id ~= nil then
    return
  end
  id = vim.lsp.start({
    name = 'Code Assistant',
    cmd = { 'go', 'run', 'lsp/server.go' },
    root_dir = vim.loop.cwd(),
  })
end

local function stop_lsp()
  if id == nil then
    return
  end
  if cur_buffer ~= nil then
    vim.lsp.buf_attach_client(cur_buffer, id)
    cur_buffer = nil
  end
  vim.lsp.stop_client(id)
  id = nil
end

function Restart_LSP()
  stop_lsp()
  print("stop_lsp")
  start_lsp()
end

function MakePopup(bufnr)
  local width = 80
  local height = 20
  local borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" }
  local win_id, win = popup.create(bufnr, {
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

local function get_visual_selection()
  local s_start = vim.fn.getpos("'<")
  local s_end = vim.fn.getpos("'>")
  local n_lines = math.abs(s_end[2] - s_start[2]) + 1
  local lines = vim.api.nvim_buf_get_lines(0, s_start[2] - 1, s_end[2], false)
  lines[1] = string.sub(lines[1], s_start[3], -1)
  if n_lines == 1 then
    lines[n_lines] = string.sub(lines[n_lines], 1, s_end[3] - s_start[3] + 1)
  else
    lines[n_lines] = string.sub(lines[n_lines], 1, s_end[3])
  end
  return table.concat(lines, '\n')
end

function InsertComment(bufnr, line, comment_lines)
  vim.api.nvim_buf_set_lines(bufnr, line, line, false, comment_lines)
end

function CloseWindow(w_id)
  vim.api.nvim_win_close(w_id, true)
end

function CreateComment()
  client = vim.lsp.get_client_by_id(id)
  cur_bufner = vim.api.nvim_win_get_buf(0)
  -- or use api.nvim_buf_get_lines
  local s_start = vim.fn.getpos("'<")
  local lines = get_visual_selection()
  local buf = vim.api.nvim_create_buf(true, true)
  win_id = MakePopup(buf)
  vim.api.nvim_buf_set_lines(buf, 0, 0, false, { "#############", "Loading", "#############" })
  resp = client.request("workspace/executeCommand", {
      command = "create_comment", arguments = { lines }
    },
    function(err, result, ctx, config)
      if err ~= nil then
        print("Error generating comment: " .. err)
        vim.api.nvim_buf_set_lines(buf, 0, 3, false, { "#############", "Error", "#############" })
        return
      end
      comment = {}
      for line in result:gmatch("([^\n]*)\n?") do
        print(line)
        table.insert(comment, line)
      end
      vim.api.nvim_buf_set_lines(buf, 0, 3, false, comment)
      vim.api.nvim_buf_set_keymap(
        buf, "n", "x",
        ":<C-u>call v:lua.InsertComment(" .. cur_bufner .. "," .. s_start .. "," .. vim.api.nvim_buf_get_lines("0, 0," .. vim.api.nvim_buf_line_count() .. ", false))",
        { silent = true }
      )
      vim.api.nvim_buf_set_keymap(buf, "n", "q", ":<C-u>call v:lua.CloseWindow(".. w_id ..")", { silent = true })
    end)
end

vim.keymap.set('v', '<leader>c', ":<C-u>call v:lua.CreateComment()<CR>")


start_lsp()
