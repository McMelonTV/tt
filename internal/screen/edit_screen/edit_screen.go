package edit_screen

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"tinytext/pkg/types"
	"tinytext/pkg/util/math"

	tea "github.com/charmbracelet/bubbletea"
)

type EditScreen struct {
	fileType *types.FileType
	filePath string

	fileContent []string

	cursorX uint
	cursorY uint

	windowWidth  int
	windowHeight int
}

func Create(filePath string) EditScreen {
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				panic(err)
			}
			file, err = os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	fileStr := string(file)

	fileContent := strings.Split(fileStr, "\n")

	return EditScreen{
		filePath: filePath,

		fileContent: fileContent,

		cursorX: 0,
		cursorY: 0,

		windowWidth:  0,
		windowHeight: 0,
	}
}

func (m EditScreen) Init() tea.Cmd {
	return nil
}

func (m EditScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			{
				stat, err := os.Stat(m.filePath)
				if err != nil {
					panic(err)
				}

				fileText := strings.Join(m.fileContent, "\n")

				err = os.WriteFile(m.filePath, []byte(fileText), stat.Mode())
				if err != nil {
					panic(err)
				}

				return m, tea.Quit
			}
		case tea.KeyUp:
			{
				if m.cursorY > 0 {
					m.cursorY--
					lineWidth := uint(len(m.fileContent[m.cursorY]))
					m.cursorX = math.ClampUintMinMax(m.cursorX, 0, lineWidth)
				}
			}
		case tea.KeyDown:
			{
				if m.cursorY < uint(len(m.fileContent)-1) {
					m.cursorY++
					lineWidth := uint(len(m.fileContent[m.cursorY]))
					m.cursorX = math.ClampUintMinMax(m.cursorX, 0, lineWidth)
				}
			}
		case tea.KeyLeft:
			{
				if m.cursorX > 0 {
					m.cursorX--
				} else {
					if m.cursorY > 0 {
						m.cursorY--
						lineWidth := uint(len(m.fileContent[m.cursorY]))
						m.cursorX = lineWidth
					}
				}
			}
		case tea.KeyRight:
			{
				if m.cursorX < uint(len(m.fileContent[m.cursorY])) {
					m.cursorX++
				} else {
					if m.cursorY < uint(len(m.fileContent)-1) {
						m.cursorY++
						m.cursorX = 0
					}
				}
			}
		case tea.KeyTab:
			{
				line := m.fileContent[m.cursorY]
				before, after := line[:m.cursorX], line[m.cursorX:]
				m.fileContent[m.cursorY] = fmt.Sprintf("%s%s%s", before, "    ", after)
				m.cursorX += uint(len("    "))
			}
		case tea.KeyBackspace:
			{
				if m.cursorX != 0 {
					line := m.fileContent[m.cursorY]
					before, after := line[:m.cursorX-1], line[m.cursorX:]
					m.fileContent[m.cursorY] = fmt.Sprintf("%s%s", before, after)
					m.cursorX -= 1
				} else {
					if m.cursorY > 0 {
						lineBefore, line := m.fileContent[m.cursorY-1], m.fileContent[m.cursorY]
						m.fileContent[m.cursorY-1] = fmt.Sprintf("%s%s", lineBefore, line)
						m.fileContent = slices.Delete(m.fileContent, int(m.cursorY), int(m.cursorY+1))
						m.cursorX = uint(len(lineBefore))
						m.cursorY -= 1
					}
				}
			}
		case tea.KeyDelete:
			{
				if m.cursorX != uint(len(m.fileContent[m.cursorY])) {
					line := m.fileContent[m.cursorY]
					before, after := line[:m.cursorX], line[m.cursorX+1:]
					m.fileContent[m.cursorY] = fmt.Sprintf("%s%s", before, after)
				} else {
					if m.cursorY < uint(len(m.fileContent)-1) {
						line, lineAfter := m.fileContent[m.cursorY], m.fileContent[m.cursorY+1]
						m.fileContent[m.cursorY] = fmt.Sprintf("%s%s", line, lineAfter)
						m.fileContent = slices.Delete(m.fileContent, int(m.cursorY+1), int(m.cursorY+2))
					}
				}
			}
		case tea.KeyEnter:
			{
				line := m.fileContent[m.cursorY]
				before, after := line[:m.cursorX], line[m.cursorX:]
				m.fileContent[m.cursorY] = fmt.Sprintf("%s", before)
				m.fileContent = slices.Insert(m.fileContent, int(m.cursorY+1), fmt.Sprintf("%s", after))
				m.cursorY += 1
				m.cursorX = 0
			}
		default:
			{
				line := m.fileContent[m.cursorY]
				before, after := line[:m.cursorX], line[m.cursorX:]
				m.fileContent[m.cursorY] = fmt.Sprintf("%s%s%s", before, string(msg.Runes), after)
				m.cursorX += uint(len(string(msg.Runes)))
			}
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}

	return m, nil
}

func (m EditScreen) View() string {
	headerStart, headerText, headerEnd := "\u001B[47;100m", "TinyText", "\u001B[0m"
	if m.windowWidth < len(headerText) {
		return ""
	}
	headerSpacing := strings.Repeat(" ", m.windowWidth-len(headerText))
	headerSpacingLeft, headerSpacingRight := headerSpacing[:len(headerSpacing)/2], headerSpacing[len(headerSpacing)/2:]
	s := fmt.Sprintf("\n%s\n", fmt.Sprintf("%s%s%s%s%s", headerStart, headerSpacingLeft, headerText, headerSpacingRight, headerEnd))

	for y, l := range m.fileContent {
		var line string

		if m.cursorY == uint(y) {
			isFinal := m.cursorX == uint(len(l))
			var before, char, after string

			if !isFinal {
				before, char, after = l[:m.cursorX], l[m.cursorX:m.cursorX+1], l[m.cursorX+1:]
			} else {
				before, char, after = l, " ", ""
			}

			line = fmt.Sprintf("%s\u001B[47;30m%s\u001B[0m%s", before, char, after)
		} else {
			line = l
		}

		s += fmt.Sprintf("%s\n", line)
	}

	return s
}
