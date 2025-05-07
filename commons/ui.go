package commons

import "fmt"

type ui struct {
	lines        map[string]func(string, int, int) int
	id_to_lines  map[string]int
	current_line int
	next_line    int
}

func New_UI() *ui {
	output := ui{}
	output.current_line = 1
	output.next_line = 1
	output.lines = make(map[string]func(string, int, int) int)
	output.id_to_lines = make(map[string]int)

	return &output
}

func Register_new_line(line_id string, ui *ui) {
	ui.id_to_lines[line_id] = ui.next_line

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
	ui.next_line++
}

func Print_to_line(ui *ui, line_id string, format string, a ...any) {
	data := fmt.Sprintf(format, a...)
	line_number := ui.id_to_lines[line_id]
	ui.current_line = ui.lines[line_id](data, ui.current_line, line_number)
	ui.next_line = ui.current_line + 1
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
