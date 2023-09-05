package main

import (
	"fmt"
	"loggit/lib/db"
	"loggit/lib/ui"
	"os"
	"time"
)

func showMenu(action string) {
	if action == "" {
		menu := ui.NewMenu("Select an action:")
		menu.AddItem("Add a new log", "add")
		menu.AddItem("View logs", "view")
		menu.AddItem("Exit", "exit")
		action = menu.Select()
	}
	switch action {
	case "add":
		addLog()
		break
	case "view":
		viewLogs()
		break
	case "exit":
	default:
		os.Exit(0)
	}
}

func main() {
	args := os.Args[1:]

	var action string
	if len(args) > 0 {
		action = args[0]
		showMenu(action)
	} else {
		showMenu("")
	}

}

func addLog() {
	ui.ClearScreen()
	customer := ui.NewPrompt("Customer:").Get()
	hours := ui.NewFloatPrompt("Hours:").Get()
	date := ui.NewPrompt("Date:").WithDefault(getTodayDateStr()).Get()
	description := ui.NewPrompt("Description:").Get()

	db.GetDb().AddLog(db.Log{
		Customer:    customer,
		Hours:       hours,
		Date:        date,
		Description: description,
	})

	showMenu("")
}

func getTodayDateStr() string {
	// get today's date
	date := time.Now()
	year, month, day := date.Date()
	return fmt.Sprintf("%02d-%02d-%d", day, month, year)
}

func viewLogs() {
	ui.ClearScreen()
	logs := db.GetDb().GetLogs()
	table := ui.NewTable()
	table.AddColumn("Customer")
	table.AddColumn("Hours")
	table.AddColumn("Date")
	table.AddColumn("Description")
	for _, logItem := range logs {
		table.AddRow([]string{
			logItem.Customer,
			fmt.Sprintf("%.2f", logItem.Hours),
			logItem.Date,
			logItem.Description,
		})
	}
	table.Render()

	showMenu("")
}
