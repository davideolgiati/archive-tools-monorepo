package commons

import (
	"fmt"
	"sync"
	"time"
)

type line struct {
	lastUpdate       time.Time
	format           string
	lineNumber       int
	currentLineValue string
}

type UI struct {
	lines             map[string]*line
	currentLineNumber int
	nextLineNumber    int
	mutex             sync.Mutex
	silent            bool
}

func NewUI() *UI {
	output := UI{}
	output.currentLineNumber = 0
	output.nextLineNumber = 1
	output.lines = make(map[string]*line)
	output.silent = false

	return &output
}

func (ui *UI) ToggleSilence() {
	ui.silent = !ui.silent
}

func (ui *UI) AddNewNamedLine(lineID string, format string) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	newLine := &line{
		lastUpdate:       time.Now(),
		lineNumber:       ui.nextLineNumber,
		format:           format,
		currentLineValue: "",
	}

	ui.goToLine(ui.nextLineNumber)
	fmt.Println("")

	ui.lines[lineID] = newLine
	ui.nextLineNumber++
}

func (ui *UI) UpdateNamedLine(lineID string, a ...any) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	currentLine := ui.lines[lineID]

	if time.Since(currentLine.lastUpdate).Milliseconds() < 16 {
		return
	}

	data := fmt.Sprintf(currentLine.format, a...)

	if data == currentLine.currentLineValue {
		return
	}

	lineNumber := currentLine.lineNumber
	ui.printToNamedLine(data, lineNumber)

	currentLine.lastUpdate = time.Now()
	currentLine.currentLineValue = data
	ui.lines[lineID] = currentLine
}

func (ui *UI) Println(format string, a ...any) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	data := fmt.Sprintf(format, a...)
	ui.goToLine(ui.nextLineNumber)
	fmt.Printf("\r%s\n", data)

	ui.nextLineNumber++
}

func (ui *UI) Close() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	offset := ui.nextLineNumber - ui.currentLineNumber

	moveCursor(offset)
}

func (ui *UI) goToLine(line int) {
	offset := line - ui.currentLineNumber
	moveCursor(offset)
	ui.currentLineNumber = line
}

func (ui *UI) printToNamedLine(data string, lineNumber int) {
	ui.goToLine(lineNumber)
	fmt.Printf("\033[2K\r%s", data)
}

func moveCursor(n int) {
	switch {
	case n == 0:
		return
	case n < 0:
		fmt.Printf("\033[%dA", -n)
	default:
		fmt.Printf("\033[%dB", n)
	}
}
