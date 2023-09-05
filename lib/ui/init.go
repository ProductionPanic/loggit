package ui

import (
	"bufio"
	"fmt"
	"github.com/pkg/term"
	"log"
	"loggit/lib/db"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func StyleMap() map[string]string {
	return map[string]string{
		"red":       "\033[31;1m",
		"blue":      "\033[34;1m",
		"green":     "\033[32;1m",
		"yellow":    "\033[33;1m",
		"black":     "\033[30m",
		"white":     "\033[37m",
		"cyan":      "\033[36m",
		"magenta":   "\033[35m",
		"gray":      "\033[30;1m",
		"lightGray": "\033[37;1m",
		"bgWhite":   "\033[47m",
		"bgBlack":   "\033[40m",
		"bgBlue":    "\033[44m",
		"bold":      "\033[1m",
		"reset":     "\033[0m",
		"underline": "\033[4m",
		"blink":     "\033[5m",
		"reverse":   "\033[7m",
	}
}

var keys = map[byte]string{
	13: "enter",
	32: "space",
	27: "esc",
	91: "arrow",
}
var up = byte(65)
var down = byte(66)
var right = byte(67)
var left = byte(68)

type CursorController struct {
}

func Cursor() *CursorController {
	return &CursorController{}
}

func (c *CursorController) hide() {
	fmt.Print("\033[?25l")
}

func (c *CursorController) show() {
	fmt.Print("\033[?25h")
}

func Parse(input string) string {
	// parses strings like this:
	// hello [red,bold]world[reset,blue]![reset]
	// into this:
	// hello \033[31;1mworld\033[0m\033[34m!\033[0m
	var output string
	var currentStyle string
	styling := false
	var styleMap = StyleMap()
	var prevChar byte
	for i := 0; i < len(input); i++ {
		if input[i] == '[' && prevChar != '/' {
			currentStyle = ""
			styling = true
		} else if (input[i] == '[' || input[i] == ']') && prevChar == '/' {
			// remove previous slash
			output = output[:len(output)-1]
			output += string(input[i])
		} else if input[i] == ']' && prevChar != '/' {
			styling = false
			styles := strings.Split(currentStyle, ",")
			for _, style := range styles {
				trimmedStyle := strings.TrimSpace(style)
				output += styleMap[trimmedStyle]
			}
			currentStyle = ""
		} else {
			if !styling {
				output += string(input[i])
			} else {
				currentStyle += string(input[i])
			}
		}
		prevChar = input[i]
	}
	return output

}

func Print(input string) {
	fmt.Print(Parse(input))
}

func Println(input string) {
	fmt.Println(Parse(input))
}

func GetInput(prompt string) string {
	Print(prompt)
	time.Sleep(100 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		line := scanner.Text()
		return line
	}
	return ""
}

type BasePrompt struct {
	Prompt       string
	RenderPrompt string
	defaultValue string
}

func (b *BasePrompt) WithDefault(defaultValue string) *BasePrompt {
	b.RenderPrompt = fmt.Sprintf("[green][bold]%s[reset] [lightGray]/[%s/][reset]", b.Prompt, defaultValue)
	b.defaultValue = defaultValue
	return b
}

func (b *BasePrompt) Get() string {
	input := GetInput(b.RenderPrompt)
	if b.defaultValue != "" && input == "" {
		return b.defaultValue
	}
	return input
}

type FloatPrompt struct {
	BasePrompt
}

func (f *FloatPrompt) Get() float32 {
	input := GetInput(f.RenderPrompt)
	if f.defaultValue != "" && input == "" {
		input = f.defaultValue
	}
	var output float32
	float, err := strconv.ParseFloat(input, 32)
	if err != nil {
		log.Fatal(err)
	}

	output = float32(float)
	return output
}

func NewFloatPrompt(prompt string) *FloatPrompt {
	return &FloatPrompt{
		BasePrompt: BasePrompt{
			Prompt:       prompt,
			RenderPrompt: "[green,bold]" + prompt + "[reset]",
		},
	}
}

func (p *FloatPrompt) WithDefault(defaultValue string) *FloatPrompt {
	p.RenderPrompt = fmt.Sprintf("[green][bold]%s[reset] [lightGray]/[%s/][reset]", p.Prompt, defaultValue)
	p.defaultValue = defaultValue
	return p
}

func NewPrompt(prompt string) *BasePrompt {
	return &BasePrompt{
		Prompt:       prompt,
		RenderPrompt: "[green,bold]" + prompt + "[reset]",
	}
}

type TextPrompt struct {
	BasePrompt
}

type Menu struct {
	Prompt    string
	CursorPos int
	Selected  []int
	MenuItems []*MenuItem
}

type MenuItem struct {
	Text  string
	Value string
}

func GetKey() byte {
	t, _ := term.Open("/dev/tty")
	err := term.RawMode(t)
	if err != nil {
		log.Fatal(err)
	}

	readBytes := make([]byte, 3)
	_, err = t.Read(readBytes)

	t.Restore()
	t.Close()

	if readBytes[0] == 27 && readBytes[1] == 91 {
		// This means it's an arrow key, return the third byte
		return readBytes[2]
	} else {
		// Non-arrow keys, return the first byte
		return readBytes[0]
	}
}

func NewMenu(prompt string) *Menu {
	return &Menu{
		Prompt:    prompt,
		MenuItems: make([]*MenuItem, 0),
		Selected:  make([]int, 0),
	}
}

func (m *Menu) AddItem(text string, value string) *Menu {
	menuItem := &MenuItem{
		Text:  text,
		Value: value,
	}
	m.MenuItems = append(m.MenuItems, menuItem)
	return m
}

func (m *Menu) render(redraw bool, isMulti bool) string {
	if redraw {
		// move cursor up
		for i := 0; i < len(m.MenuItems)+1; i++ {
			// clear line
			fmt.Printf("\033[K")
			// move cursor up
			fmt.Printf("\033[1A")
		}
	}
	var output string
	output += m.Prompt + "\n"
	for i, item := range m.MenuItems {
		cursor := " "
		if i == m.CursorPos {
			cursor = ">"
		}
		if isMulti && contains(m.Selected, i) {
			output += cursor + " /[x/] [reset][bold][cyan]" + item.Text + "[reset]\n"
		} else if isMulti {
			output += cursor + " /[ /] [reset][bold][cyan]" + item.Text + "[reset]\n"
		} else if i == m.CursorPos {
			output += "[blue]> [reset][bold][cyan]" + item.Text + "[reset]\n"
		} else {
			output += "  [reset][bold][cyan]" + item.Text + "[reset]\n"
		}
	}
	return output
}

func (m *Menu) Select() string {
	Cursor().hide()
	Print(m.render(false, false))
	var key byte
	for {
		Print(m.render(true, false))
		key = GetKey()
		if key == up {
			if m.CursorPos > 0 {
				m.CursorPos--
			}
		} else if key == down {
			if m.CursorPos < len(m.MenuItems)-1 {
				m.CursorPos++
			}
		} else if key == 13 {
			Cursor().show()
			return m.MenuItems[m.CursorPos].Value
		}
	}
	Cursor().show()
	return ""
}

func (m *Menu) MultiSelect() []string {
	Cursor().hide()
	Print(m.render(false, true))
	m.Selected = make([]int, 0)
	for {
		Print(m.render(true, true))
		key := m.getKey()
		if key == up {
			m.CursorPos = max(m.CursorPos-1, 0)
		} else if key == down {
			m.CursorPos = min(m.CursorPos+1, len(m.MenuItems)-1)
		} else if key /** space */ == 32 {
			if !contains(m.Selected, m.CursorPos) {
				m.Selected = append(m.Selected, m.CursorPos)
			} else {
				m.Selected = remove(m.Selected, m.CursorPos)
			}
		} else if key == 13 { // enter
			var values []string
			for _, i := range m.Selected {
				values = append(values, m.MenuItems[i].Value)
			}
			Cursor().show()
			return values
		}
	}
	Cursor().show()
	return []string{}
}

func (m *Menu) getKey() byte {
	t, _ := term.Open("/dev/tty")
	err := term.RawMode(t)
	if err != nil {
		log.Fatal(err)
	}

	var read int
	readBytes := make([]byte, 3)
	read, err = t.Read(readBytes)

	err = t.Restore()
	if err != nil {
		return 0
	}
	err = t.Close()
	if err != nil {
		return 0
	}

	if read == 3 {
		if _, ok := keys[readBytes[2]]; ok {
			return readBytes[2]
		}
	} else {
		return readBytes[0]
	}

	return 0
}

func contains(arr []int, val int) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}

func remove(arr []int, val int) []int {
	var output []int
	for _, item := range arr {
		if item != val {
			output = append(output, item)
		}
	}
	return output
}

type Table struct {
	Columns        []string
	DisplayHeaders bool
	Rows           [][]string
	selectedRow    int
}

func NewTable() *Table {
	return &Table{
		Columns: make([]string, 0),
		Rows:    make([][]string, 0),
	}
}

func (t *Table) AddColumn(column string) *Table {
	t.Columns = append(t.Columns, column)
	return t
}

func (t *Table) AddRow(row []string) *Table {
	t.Rows = append(t.Rows, row)
	return t
}

func (t *Table) Render() {
	Cursor().hide()
	t.render()
	for {
		key := GetKey()
		// move cursor up
		for i := 0; i < len(t.Rows)+2; i++ {
			// move cursor up
			fmt.Printf("\033[1A")
			// clear line
			fmt.Printf("\033[K")
		}
		if key == up {
			t.selectedRow = max(t.selectedRow-1, 0)
		} else if key == down {
			t.selectedRow = min(t.selectedRow+1, len(t.Rows)-1)
		} else if key == 13 {
			Cursor().show()
			return
		} else if key == 100 { // d
			ClearScreen()
			var confirmed bool
			Println("[red,bold]Are you sure you want to delete this log? [blue]/[y/n/][reset]")
			for {
				key := GetKey()
				if key == 121 { // y
					confirmed = true
					break
				} else if key == 110 { // n
					confirmed = false
					break
				}
			}
			if confirmed {
				db.GetDb().RemoveLog(t.selectedRow)
				t.Rows = t.removeIndex(t.selectedRow)
				ClearScreen()
				Println("[green,bold]Log deleted![reset]")
				time.Sleep(1 * time.Second)
				ClearScreen()
			} else {
				ClearScreen()
			}

		} else if key == 101 { // e
			ClearScreen()
			logItem := db.GetDb().GetLogs()[t.selectedRow]
			customer := NewPrompt("Customer:").WithDefault(logItem.Customer).Get()

			hours := NewFloatPrompt("Hours:").WithDefault(fmt.Sprintf("%f", logItem.Hours)).Get()

			date := NewPrompt("Date:").WithDefault(logItem.Date).Get()

			description := NewPrompt("Description:").WithDefault(logItem.Description).Get()

			db.GetDb().UpdateLog(t.selectedRow, db.Log{
				Customer:    customer,
				Hours:       hours,
				Date:        date,
				Description: description,
			})
			t.Rows[t.selectedRow] = []string{customer, fmt.Sprintf("%f", hours), date, description}
			ClearScreen()
		} else if key == 98 { // b
			Cursor().show()
			return
		}
		t.render()
	}
	Cursor().show()
}

func (t *Table) removeIndex(i int) [][]string {
	return append(t.Rows[:i], t.Rows[i+1:]...)
}

func (t *Table) render() {
	var output string
	var columnMaxWidthMap = make(map[int]int)
	padding := 2
	for i, column := range t.Columns {
		var curmax int
		for _, row := range t.Rows {
			curmax = max(curmax, len(row[i]))
		}
		curmax = max(curmax, len(column))
		columnMaxWidthMap[i] = curmax + padding
	}
	for i, column := range t.Columns {
		output += fmt.Sprintf("[bgBlack,bold,blue]%s%s%s[reset]", strings.Repeat(" ", columnMaxWidthMap[i]-len(column)), column, strings.Repeat(" ", padding))
	}
	output += "\n"
	for i, row := range t.Rows {
		style := "[bgBlack,white]"
		if t.selectedRow == i {
			style = "[bgWhite,black]"
		}
		for i, column := range row {
			output += fmt.Sprintf(style+"%s%s%s[reset]", strings.Repeat(" ", columnMaxWidthMap[i]-len(column)), column, strings.Repeat(" ", padding))
		}
		output += "\n"
	}
	Println(output)
	// move up one line
	fmt.Printf("\033[1A")
	totalWidth := 0
	for _, width := range columnMaxWidthMap {
		totalWidth += width
	}
	// clear line
	fmt.Printf("\033[K")
	msg := "D to delete, E to edit, Enter to exit, B to go back"
	whitespace := strings.Repeat(" ", totalWidth/2-len(msg)/2)
	Println("[bgBlack,green]" + whitespace + msg + whitespace + "[reset]")
}

func ClearScreen() {
	switch oss := runtime.GOOS; oss {
	case "darwin", "linux":
		// Escape sequence to clear screen for UNIX-based systems
		fmt.Println("\033[H\033[2J")
	case "windows":
		// For Windows, use cls command
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Printf("Operative system %s not supported, the screen cannot be cleared automatically", oss)
	}
}
