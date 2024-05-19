commentWindow = require("lua.comment_edit_window")
testWindow = require("lua.test_edit_pane")

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


function CreateComment()
  client = vim.lsp.get_client_by_id(id)
  -- or use api.nvim_buf_get_lines
  local lines = get_visual_selection()
  commentWindow.MakePopup()
  commentWindow.AddInstructions({
    "q Quit",
    "x Add comment to buffer",
    "r Regenerate comments",
    "t Write tests given comments"
  })
  commentWindow.SetBuffer({
    "################################################################################",
    "################################### LOADING ###################################",
    "################################################################################"
  })
  resp = client.request("workspace/executeCommand", {
      command = "create_comment", arguments = { lines }
    },
    function(err, result, ctx, config)
      if err ~= nil then
        print("Error generating comment: " .. err)
        commentWindow.SetBuffer({ "#############", "Error", "#############" })
        return
      end
      comment = {}
      for line in result:gmatch("([^\n]*)\n?") do
        table.insert(comment, line)
      end
      commentWindow.SetBuffer(comment)
      commentWindow.AllowSaving()
    end)
end

function CreateTests()
  comment = commentWindow.GetCommentLines()
  comment = table.concat(comment, '\n')
  file_name = vim.api.nvim_buf_get_name(commentWindow.file_buffer)
  InsertComment()
  testWindow.MakePopup()
  testWindow.SetBuffer({
    "################################################################################",
    "################################### LOADING ###################################",
    "################################################################################"
  })
  resp = client.request("workspace/executeCommand", {
      command = "create_tests", arguments = { comment, file_name }
    },
    function(err, result, ctx, config)
      if err ~= nil then
        testWindow.SetBuffer({
          "################################################################################",
          "###################################  ERROR  ####################################",
          "################################################################################"
        })
        print("Error generating tests: " .. err)
        return
      end
      tests_output = {}
      test_file_path = ""
      for line in result:gmatch("([^\n]*)\n?") do
        start_ind, end_ind = string.find(line, "__astro_test_file_path__=")
        if start_ind ~= nil and start_ind >= 0 then
          print(line)
        end
        table.insert(tests_output, line)
      end
      testWindow.SetBuffer(tests_output)
    end)
end

vim.keymap.set('v', '<leader>c', ":<C-u>call v:lua.CreateComment()<CR>")

start_lsp()
