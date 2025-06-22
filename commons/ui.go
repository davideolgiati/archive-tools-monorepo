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

type ui struct {
	lines             map[string]*line
	currentLineNumber int
	nextLineNumber    int
	mutex             sync.Mutex
	silent            bool
}

func NewUI() *ui {
	output := ui{}
	output.currentLineNumber = 0
	output.nextLineNumber = 1
	output.lines = make(map[string]*line)
	output.silent = false

	return &output
}

func (ui *ui) ToggleSilence() {
	ui.silent = !ui.silent
}

func (ui *ui) AddNewNamedLine(lineID string, format string) {
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

func (ui *ui) UpdateNamedLine(lineID string, a ...any) {
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

func (ui *ui) Println(format string, a ...any) {
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

func (ui *ui) Close() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	offset := ui.nextLineNumber - ui.currentLineNumber

	moveCursor(offset)
}

func (ui *ui) goToLine(line int) {
	offset := line - ui.currentLineNumber
	moveCursor(offset)
	ui.currentLineNumber = line
}

func (ui *ui) printToNamedLine(data string, lineNumber int) {
	ui.goToLine(lineNumber)
	fmt.Printf("\033[2K\r%s", data)
}

func moveCursor(n int) {
	if n == 0 {
		return
	} else if n < 0 {
		fmt.Printf("\033[%dA", -n)
	} else {
		fmt.Printf("\033[%dB", n)
	}
}
