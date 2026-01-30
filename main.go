package main

import (
	"os"

	qt "github.com/mappu/miqt/qt6"
)

type RSMQTMainWindow struct {
	*qt.QMainWindow
}

func NewRSMQTMainWindow() *RSMQTMainWindow {
	var self RSMQTMainWindow

	self.QMainWindow = qt.NewQMainWindow2()

	self.SetWindowTitle("RSMQT")
	self.SetGeometry(100, 100, 1000, 600)

	return &self
}

func main() {
	qt.NewQApplication(os.Args)

	window := NewRSMQTMainWindow()
	window.Show()

	qt.QApplication_Exec()
}
