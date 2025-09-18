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
	fileType types.FileType
	filePath string

	fileContent []string

	cursorX uint
	cursorY uint

	visualCursorX uint
	visualCursorY uint

	windowWidth  int
	windowHeight int
}

func (s *EditScreen) updateCursorPositions(x, y, visualX, visualY uint) {
	s.updateCursorXPositions(x, visualX)
	s.updateCursorYPositions(y, visualY)
}
func (s *EditScreen) updateCursorXPositions(x, visualX uint) {
	s.cursorX = x
	s.visualCursorX = visualX
}
func (s *EditScreen) updateCursorYPositions(y, visualY uint) {
	s.cursorY = y
	s.visualCursorY = visualY
}

func (s *EditScreen) updateCursorPosition(x, y uint) {
	s.updateCursorXPosition(x)
	s.updateCursorYPosition(y)
}
func (s *EditScreen) updateCursorXPosition(x uint) {
	s.updateCursorXPositions(x, x)
}
func (s *EditScreen) updateCursorYPosition(y uint) {
	s.updateCursorYPositions(y, y)
}

func Create(path *string) EditScreen {
	var filePath string
	var fileContent []string
	var fileType types.FileType

	if path != nil {
		filePath = *path
		fileType = types.FileTypePersistent

		file, err := os.ReadFile(*path)
		if err != nil {
			if os.IsNotExist(err) {
				_, err := os.Create(*path)
				if err != nil {
					panic(err)
				}
				file, err = os.ReadFile(*path)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}

		fileStr := string(file)

		fileContent = strings.Split(fileStr, "\n")
	} else {
		filePath = ""
		fileType = types.FileTypeTemporary

		fileContent = make([]string, 0)
	}

	return EditScreen{
		filePath: filePath,
		fileType: fileType,

		fileContent: fileContent,

		cursorX: 0,
		cursorY: 0,

		visualCursorX: 0,
		visualCursorY: 0,

		windowWidth:  0,
		windowHeight: 0,
	}
}

func (s EditScreen) Init() tea.Cmd {
	return nil
}

func (s EditScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			{
				if s.fileType == types.FileTypePersistent {
					stat, err := os.Stat(s.filePath)
					if err != nil {
						panic(err)
					}

					fileText := strings.Join(s.fileContent, "\n")

					err = os.WriteFile(s.filePath, []byte(fileText), stat.Mode())
					if err != nil {
						panic(err)
					}
				}

				return s, tea.Quit
			}
		case tea.KeyUp:
			{
				if s.cursorY > 0 {
					s.updateCursorYPosition(s.cursorY - 1)
					lineWidth := uint(len(s.fileContent[s.cursorY]))
					s.updateCursorXPositions(s.cursorX, math.ClampUintMinMax(s.cursorX, 0, lineWidth))
				}
			}
		case tea.KeyDown:
			{
				if s.cursorY < uint(len(s.fileContent)) {
					s.updateCursorYPosition(s.cursorY + 1)

					if s.cursorY != uint(len(s.fileContent)) {
						lineWidth := uint(len(s.fileContent[s.cursorY]))
						s.updateCursorXPositions(s.cursorX, math.ClampUintMinMax(s.cursorX, 0, lineWidth))
					} else {
						s.updateCursorXPositions(s.cursorX, 0)
					}
				}
			}
		case tea.KeyLeft:
			//FIXME: wrong movement with visual cursor != cursor
			{
				if s.cursorX > 0 {
					s.updateCursorXPosition(s.cursorX - 1)
				} else {
					if s.cursorY > 0 {
						s.updateCursorYPosition(s.cursorY - 1)
						lineWidth := uint(len(s.fileContent[s.cursorY]))
						s.updateCursorXPosition(lineWidth)
					}
				}
			}
		case tea.KeyRight:
			{
				if s.cursorY != uint(len(s.fileContent)) {
					if s.cursorX < uint(len(s.fileContent[s.cursorY])) {
						s.updateCursorXPosition(s.cursorX + 1)
					} else {
						if s.cursorY < uint(len(s.fileContent)) {
							s.updateCursorPosition(0, s.cursorY+1)
						}
					}
				}
			}
		case tea.KeyTab:
			{
				line := s.fileContent[s.cursorY]
				before, after := line[:s.cursorX], line[s.cursorX:]
				s.fileContent[s.cursorY] = fmt.Sprintf("%s%s%s", before, "    ", after)
				s.updateCursorXPosition(s.cursorX + uint(len("    ")))
			}
		case tea.KeyBackspace:
			{
				if s.cursorX != 0 {
					line := s.fileContent[s.cursorY]
					before, after := line[:s.cursorX-1], line[s.cursorX:]
					s.fileContent[s.cursorY] = fmt.Sprintf("%s%s", before, after)
					s.updateCursorXPosition(s.cursorX - 1)
				} else {
					if s.cursorY > 0 {
						lineBefore, line := s.fileContent[s.cursorY-1], s.fileContent[s.cursorY]
						s.fileContent[s.cursorY-1] = fmt.Sprintf("%s%s", lineBefore, line)
						s.fileContent = slices.Delete(s.fileContent, int(s.cursorY), int(s.cursorY+1))
						s.updateCursorPosition(uint(len(lineBefore)), s.cursorY-1)
					}
				}
			}
		case tea.KeyDelete:
			{
				if s.cursorX != uint(len(s.fileContent[s.cursorY])) {
					line := s.fileContent[s.cursorY]
					before, after := line[:s.cursorX], line[s.cursorX+1:]
					s.fileContent[s.cursorY] = fmt.Sprintf("%s%s", before, after)
				} else {
					if s.cursorY < uint(len(s.fileContent)-1) {
						line, lineAfter := s.fileContent[s.cursorY], s.fileContent[s.cursorY+1]
						s.fileContent[s.cursorY] = fmt.Sprintf("%s%s", line, lineAfter)
						s.fileContent = slices.Delete(s.fileContent, int(s.cursorY+1), int(s.cursorY+2))
					}
				}
			}
		case tea.KeyEnter:
			{
				line := s.fileContent[s.cursorY]
				before, after := line[:s.cursorX], line[s.cursorX:]
				s.fileContent[s.cursorY] = fmt.Sprintf("%s", before)
				s.fileContent = slices.Insert(s.fileContent, int(s.cursorY+1), fmt.Sprintf("%s", after))
				s.updateCursorPosition(0, s.cursorY+1)
			}
		default:
			{
				if s.cursorY == uint(len(s.fileContent)) {
					s.fileContent = slices.Insert(s.fileContent, len(s.fileContent), "")
				}

				line := s.fileContent[s.cursorY]
				before, after := line[:s.cursorX], line[s.cursorX:]
				s.fileContent[s.cursorY] = fmt.Sprintf("%s%s%s", before, string(msg.Runes), after)
				s.updateCursorXPosition(s.cursorX + uint(len(string(msg.Runes))))
			}
		}
	case tea.WindowSizeMsg:
		s.windowWidth = msg.Width
		s.windowHeight = msg.Height
	}

	return s, nil
}

func (s EditScreen) View() string {
	headerStart, headerText, headerEnd := "\u001B[47;100m", "TinyText", "\u001B[0m"
	if s.windowWidth < len(headerText) {
		return ""
	}
	headerSpacing := strings.Repeat(" ", s.windowWidth-len(headerText))
	headerSpacingLeft, headerSpacingRight := headerSpacing[:len(headerSpacing)/2], headerSpacing[len(headerSpacing)/2:]
	str := fmt.Sprintf("\n%s\n", fmt.Sprintf("%s%s%s%s%s", headerStart, headerSpacingLeft, headerText, headerSpacingRight, headerEnd))

	for y, l := range s.fileContent {
		var line string

		if s.visualCursorY == uint(y) {
			isFinal := s.visualCursorX == uint(len(l))
			var before, char, after string

			if !isFinal {
				before, char, after = l[:s.visualCursorX], l[s.visualCursorX:s.visualCursorX+1], l[s.visualCursorX+1:]
			} else {
				before, char, after = l, " ", ""
			}

			line = fmt.Sprintf("%s\u001B[47;30m%s\u001B[0m%s", before, char, after)
		} else {
			line = l
		}

		str += fmt.Sprintf("%s\n", line)
	}

	if s.visualCursorY == uint(len(s.fileContent)) {
		str += fmt.Sprintf("\u001B[47;30m%s\u001B[0m", " ")
	}

	return str
}
