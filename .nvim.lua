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

function Create_Comment()
  client = vim.lsp.get_client_by_id(id)
  local vstart = vim.fn.getpos("'<")

  local vend = vim.fn.getpos("'>")

  local line_start = vstart[2]
  local line_end = vend[2]

  -- or use api.nvim_buf_get_lines
  local lines = get_visual_selection()
  print("Got Lines")
  client.request("workspace/executeCommand", {
    command = "create_comment", arguments = { { lines = lines }, }
  })
end

vim.keymap.set('v', '<leader>c', ":<C-u>call v:lua.Create_Comment()<CR>")


start_lsp()
