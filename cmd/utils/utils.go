package utils

func TruncateContent(width, height int, content string) string {
	contentHeight := 0
	truncateSize := len(content)
	currentWidth := 0
	for index, c := range content {
		currentWidth++
		if c == '\n' {
			contentHeight++
			// currentWidth = 0 - Removed this line as it seems go-fzf preview is not reseting the width count after \n
		}
		if currentWidth >= width {
			contentHeight++
			currentWidth = 0
		}
		if contentHeight == height {
			truncateSize = index - 1
			break
		}
	}
	return content[0:truncateSize] //snippets[i].content[:truncateSize]
}
