package commons

import (
	"fmt"
	"sync"
	"time"
)

type line struct {
	last_update      time.Time
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

func (ui *ui) AddNewNamedLine(line_id string, format string) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	new_line := &line{
		last_update:      time.Now(),
		lineNumber:       ui.nextLineNumber,
		format:           format,
		currentLineValue: "",
	}

	ui.goToLine(ui.nextLineNumber)
	fmt.Println("")

	ui.lines[line_id] = new_line
	ui.nextLineNumber++
}

func (ui *ui) UpdateNamedLine(line_id string, a ...any) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	current_line := ui.lines[line_id]

	if time.Since(current_line.last_update).Milliseconds() < 16 {
		return
	}

	data := fmt.Sprintf(current_line.format, a...)

	if data == current_line.currentLineValue {
		return
	}

	line_number := current_line.lineNumber
	ui.printToNamedLine(data, line_number)

	current_line.last_update = time.Now()
	current_line.currentLineValue = data
	ui.lines[line_id] = current_line
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

func (ui *ui) printToNamedLine(data string, line_number int) {
	ui.goToLine(line_number)
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
