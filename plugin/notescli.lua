if vim.g.loaded_notescli then
	return
end
vim.g.loaded_notescli = true

local notescli = require("notescli")
notescli.setup()
