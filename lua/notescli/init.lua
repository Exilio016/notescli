local pickers = require("telescope.pickers")
local finders = require("telescope.finders")
local conf = require("telescope.config").values

local function file_exists(name)
	local f = io.open(name, "r")
	if f ~= nil then
		io.close(f)
		return true
	else
		return false
	end
end

local function get_project_path()
	local file = debug.getinfo(1).source:sub(2)
	local path = file:match("(.*" .. "/" .. ").*/.*/.*")
	return path
end

local setup = function()
	local path = get_project_path()
	vim.print(path)
	if not file_exists(path .. "notescli") then
		vim.print("Building notescli...")
		vim.print(vim.fn.system("cd " .. path .. "; go build"))
	end
	local executable = path .. "notescli"
end

return { setup = setup }
