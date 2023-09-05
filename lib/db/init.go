package db

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func getDbFolder() string {
	var home string
	if os.Getenv("HOME") != "" {
		home = os.Getenv("HOME")
	} else {
		home = os.Getenv("USERPROFILE")
	}

	return home + "/.loggit"
}

func getDbFile() string {
	return getDbFolder() + "/db.json"
}

func EnsureDb() {
	if _, err := os.Stat(getDbFolder()); os.IsNotExist(err) {
		err := os.Mkdir(getDbFolder(), 0777)
		if err != nil {
			return
		}
	}
	if _, err := os.Stat(getDbFile()); os.IsNotExist(err) {
		_, err := os.Create(getDbFile())
		if err != nil {
			return
		}
	}
}

type DB struct {
	Logs []Log
}

type Log struct {
	Customer    string  `json:"customer"`
	Hours       float32 `json:"hours"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
}

func (d *DB) AddLog(log Log) {
	d.Logs = append(d.Logs, log)
	d.save()
}

func (d *DB) GetLogs() []Log {
	d.load()
	return d.Logs
}

func (d *DB) save() {
	var data []byte
	data, err := json.Marshal(d)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(getDbFile(), data, 0777)
	if err != nil {
		return
	}
}

func (d *DB) RemoveLog(index int) {
	d.load()
	d.Logs = append(d.Logs[:index], d.Logs[index+1:]...)
	d.save()
}

func (d *DB) UpdateLog(index int, log Log) {
	d.load()
	d.Logs[index] = log
	d.save()
}

func (d *DB) load() {
	data, err := ioutil.ReadFile(getDbFile())
	if err != nil {
		return
	}
	err = json.Unmarshal(data, d)
	if err != nil {
		return
	}

	if d.Logs == nil {
		d.Logs = []Log{}
	}
}

var db *DB

func GetDb() *DB {
	if db == nil {
		db = &DB{}
		db.load()
	}
	return db
}
