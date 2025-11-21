-- Create a custom logger
local log_path = "/home/austin/astrologs/client.nvim.astro.log"
local log = require('plenary.log').new({
  plugin = "astrolabe",
  level = "debug",
  use_console = "sync",
  use_file = true,
  outfile = log_path
})
local commentWindow = require("astrolabe.comment_edit_window")
local testWindow = require("astrolabe.test_edit_pane")
local cur_buffer = nil
local LSP_NAME = "Astrolabe"

local function getLspClient(name)
  local clients = vim.lsp.get_active_clients() -- get clients for current buffer
  local client

  for _, c in ipairs(clients) do
    if c.name == name then
      client = c
    end
  end

  if not client then
    return nil
  end
  return client.id
end

local function getBufferByName(name)
  for _, buf in ipairs(vim.api.nvim_list_bufs()) do
    local buf_name = vim.api.nvim_buf_get_name(buf)
    if buf_name == name then
      return buf
    end
  end
  return -1
end

function start_lsp()
  log.info("Starting LSP!")
  local id = getLspClient(LSP_NAME)
  if id ~= nil then
    log.info("Astrolabe LSP started")
    return
  end
  id = vim.lsp.start({
    name = LSP_NAME,
    cmd = { 'lsp' },
    root_dir = vim.loop.cwd(),
  })
  log.info(string.format("client id =  %q", id))
end

start_lsp()

local function attach_lsp(args)
  local id = getLspClient(LSP_NAME)
  log.debug("Attaching lsp")
  if id == nil then
    log.debug("No lsp running")
    return
  end
  vim.lsp.buf_attach_client(args.buffer, id);
  vim.lsp.inline_completion.enable()
  cur_buffer = args.buffer
  log.debug(string.format("cur_buffer = %q", cur_buffer))
end

vim.api.nvim_create_autocmd("BufNew", { callback = attach_lsp });
vim.api.nvim_create_autocmd("BufEnter", { callback = attach_lsp, });

local function stop_lsp()
  local id = getLspClient(LSP_NAME)
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


Lines = {}
File_type = ""
function CreateComment()
  -- or use api.nvim_buf_get_lines
  Lines = get_visual_selection()
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
  File_type = vim.api.nvim_buf_get_option(commentWindow.file_buffer, 'filetype')
  GenerateComment()
end

function GenerateComment()
  local id = getLspClient(LSP_NAME)
  local client = vim.lsp.get_client_by_id(id)
  if client == nil then
    log.error("Client is nil, exiting")
    return
  end
  client.request("workspace/executeCommand", {
      command = "create_comment", arguments = { File_type, Lines }
    },
    function(err, result, ctx, config)
      if err ~= nil then
        print(string.format("Error generating comment: %q", err))
        commentWindow.SetBuffer({ "#############", "Error", "#############" })
        return
      end
      local comment = {}
      for line in result:gmatch("([^\n]*)\n?") do
        table.insert(comment, line)
      end
      commentWindow.SetBuffer(comment)
      commentWindow.AllowSaving()
    end)
end

function CreateTests()
  local comment = commentWindow.GetCommentLines()
  local comment = table.concat(comment, '\n')
  local file_name = vim.api.nvim_buf_get_name(commentWindow.file_buffer)
  local file_type = vim.api.nvim_buf_get_option(commentWindow.file_buffer, 'filetype')
  InsertComment()
  testWindow.MakePopup(0)
  testWindow.SetBuffer({
    "################################################################################",
    "################################### LOADING ###################################",
    "################################################################################"
  })
  resp = client.request("workspace/executeCommand", {
      command = "create_tests", arguments = { comment, file_name, file_type }
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
      local tests_output = {}
      local test_file_path = ""
      for line in result:gmatch("([^\n]*)\n?") do
        local name_start_ind, name_end_ind = string.find(line, "__astro_test_file_path__=")
        local indicator_start_ind, indicator_end_ind = string.find(line, "```")
        if name_start_ind ~= nil and name_start_ind >= 0 then
          test_file_path = string.sub(line, name_end_ind + 1, -1)
          if getBufferByName(test_file_path) > 0 then
            print("Buffer exists")
          end
          testWindow.SetBufferName(test_file_path)
        elseif not (indicator_start_ind ~= nil and indicator_start_ind >= 0) then
          table.insert(tests_output, line)
        end
      end
      testWindow.SetBuffer(tests_output)
    end)
end

vim.keymap.set('v', '<leader>c', ":<C-u>call v:lua.CreateComment()<CR>")


--
-- Fetches and prints language server protocol (LSP) completions at the current cursor position.
--
-- This function retrieves completion suggestions from the LSP server named "Astrolabe" and prints each suggestion's label.
--
-- @function get_lsp_completions
-- @return void
-- @throws Error if there is an issue with the LSP request or if the LSP client is not available.
--
-- @example
-- get_lsp_completions()
-- -- This will print the completion suggestions at the current cursor position.
--
-- @note
-- - The function uses the `vim.lsp.util.make_position_params()` to get the current cursor position.
-- - If the LSP client is not available, the function will not perform any actions.
-- - If there are no completion items, it will print "No items".
local function get_lsp_completions()
  print("LM completions")
  local client_id = getLspClient(LSP_NAME)
  local params = vim.lsp.util.make_position_params() -- current cursor position
  local client = vim.lsp.get_client_by_id(client_id)
  if client then
    client.request(
      "textDocument/completion",
      params,
      function(err, result, ctx, config)
        if err then
          print("LSP completion error:", err)
          return
        end

        -- `result` can be an array or CompletionList
        local items = result and result.items or result

        if not items then
          print("No items")
          return
        end

        -- Now you have the completion items as Lua tables
        for _, item in ipairs(items) do
          print("Completion:", item.label)
        end
      end
    )
  end
end

vim.api.nvim_create_autocmd("InsertEnter", {
  callback = get_lsp_completions,
})
