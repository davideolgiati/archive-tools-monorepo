package commons

import (
	"fmt"
	"time"
)

type ui struct {
	lines               map[string]func(string, int, int) int
	id_to_lines         map[string]int
	id_to_format        map[string]string
	lines_id_last_value map[string]string
	current_line        int
	next_line           int
	line_last_update    map[string]int64
}

func New_UI() *ui {
	output := ui{}
	output.current_line = 1
	output.next_line = 1
	output.lines = make(map[string]func(string, int, int) int)
	output.id_to_lines = make(map[string]int)
	output.lines_id_last_value = make(map[string]string)
	output.id_to_format = make(map[string]string)
	output.line_last_update = make(map[string]int64)

	return &output
}

func Register_new_line(line_id string, format string, ui *ui) {
	ui.id_to_lines[line_id] = ui.next_line
	ui.id_to_format[line_id] = format

	custom_fn := func(data string, current_line int, line_number int) int {
		offset := line_number - current_line - 1

		if offset < 0 && current_line > 1 {
			fmt.Printf("\033[%dA", -offset)
		} else if offset > 0 {
			fmt.Printf("\033[%dB", offset)
		}

		fmt.Printf("\033[2K\r%s\n", data)

		return line_number
	}

	ui.lines[line_id] = custom_fn
	ui.lines_id_last_value[line_id] = ""
	ui.line_last_update[line_id] = time.Now().UnixMilli()
	ui.next_line++
}

func Print_to_line(ui *ui, line_id string, a ...any) {
	data := fmt.Sprintf(ui.id_to_format[line_id], a...)
	if data != ui.lines_id_last_value[line_id] && (time.Now().UnixMilli()-ui.line_last_update[line_id]) > 60 {
		line_number := ui.id_to_lines[line_id]
		ui.current_line = ui.lines[line_id](data, ui.current_line, line_number)
		ui.line_last_update[line_id] = time.Now().UnixMilli()
	}
}

func Print_not_registered(ui *ui, format string, a ...any) {
	data := fmt.Sprintf(format, a...)
	offset := ui.next_line - ui.current_line - 1

	if offset < 0 && ui.current_line > 1 {
		fmt.Printf("\033[%dA", -offset)
	} else if offset > 0 {
		fmt.Printf("\033[%dB", offset)
	}

	fmt.Printf("\r%s\n", data)
	ui.next_line++
	ui.current_line = ui.next_line
}

func Close_UI(ui *ui) {
	offset := ui.next_line - ui.current_line

	if offset < 0 {
		fmt.Printf("\033[%dA", -offset)
	} else if offset > 0 {
		fmt.Printf("\033[%dB", offset)
	}

	fmt.Printf("\n")
}
