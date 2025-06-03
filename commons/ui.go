package commons

import (
	"fmt"
	"sync"
	"time"
)

type line struct {
	last_update  time.Time
	format      string
	line_number int
}

type ui struct {
	lines        map[string]line
	current_line int
	next_line    int
	mutex        sync.Mutex
	silent       bool
}

func New_UI() *ui {
	output := ui{}
	output.current_line = 0
	output.next_line = 1
	output.lines = make(map[string]line)
	output.silent = false

	return &output
}

func (ui *ui) Toggle_silence() {
	ui.silent = !ui.silent
}

func (ui *ui) Register_line(line_id string, format string) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	new_line := line{
		last_update: time.Now(),
		line_number: ui.next_line,
		format:      format,
	}

	offset := ui.next_line - ui.current_line - 1

	if offset > 0 {
		move_cursor(offset)
	}

	fmt.Println("")

	ui.lines[line_id] = new_line
	ui.current_line = ui.next_line
	ui.next_line++
}

func (ui *ui) Update_line(line_id string, a ...any) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	current_line := ui.lines[line_id]

	if ui.silent {
		return
	}

	if time.Since(current_line.last_update).Milliseconds() < 16 {
		return
	}

	data := fmt.Sprintf(current_line.format, a...)

	line_number := current_line.line_number
	ui.current_line = update_line(data, ui.current_line, line_number)
	current_line.last_update = time.Now()

	ui.lines[line_id] = current_line
}

func (ui *ui) Print_not_registered(format string, a ...any) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	data := fmt.Sprintf(format, a...)
	offset := ui.next_line - ui.current_line - 1

	if offset > 0 {
		move_cursor(offset)
	}

	fmt.Printf("\r%s\n", data)
	ui.current_line = ui.next_line
	ui.next_line++
}

func (ui *ui) Close() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.silent {
		return
	}

	offset := ui.next_line - ui.current_line

	move_cursor(offset)
}

func move_cursor(n int) {
	if n < 0 {
		fmt.Printf("\033[%dA", -n)
	} else {
		fmt.Printf("\033[%dB", n)
	}

}

func update_line(data string, current_line int, line_number int) int {
	offset := (line_number - current_line)

	if offset != 0 {
		move_cursor(offset)
	}

	fmt.Printf("\033[2K\r%s", data)

	return line_number
}
