package main

import (
	"tinytext/internal/screen/edit_screen"

	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
)

var Args struct {
	FilePath string `arg:"positional,required"`
}

func main() {
	arg.MustParse(&Args)

	screen := tea.NewProgram(edit_screen.Create(&Args.FilePath), tea.WithAltScreen())
	if _, err := screen.Run(); err != nil {
		panic(err)
	}
}
