package main

import (
	"os"
	"strconv"
	"time"

	"github.com/benjamesfleming/rsmqt/lib/rsmq"
	qt "github.com/mappu/miqt/qt6"
)

type RSMQTMainWindow struct {
	*qt.QMainWindow

	client *rsmq.Client

	// Left
	queueListView  *qt.QListView
	queueListModel *qt.QStringListModel

	// Right Top
	statsTableView *qt.QTableView
	statsModel     *qt.QStandardItemModel

	// Right Bottom
	msgTableView *qt.QTableView
	msgModel     *qt.QStandardItemModel
}

func NewRSMQTMainWindow() *RSMQTMainWindow {
	mw := &RSMQTMainWindow{}
	mw.QMainWindow = qt.NewQMainWindow2()
	mw.SetWindowTitle("RSMQ UI")
	mw.SetStyleSheet("background-color: #f1f2f6")
	mw.SetGeometry(100, 100, 1000, 700)

	// Central Widget
	central := qt.NewQWidget(mw.QWidget)
	mw.SetCentralWidget(central)

	layout := qt.NewQHBoxLayout(central)

	// Main Splitter
	splitter := qt.NewQSplitter4(qt.Horizontal, central)
	layout.AddWidget(splitter.QWidget)

	// Left Pane: Splitter Vertical
	leftSplitter := qt.NewQSplitter4(qt.Vertical, splitter.QWidget)
	splitter.AddWidget(leftSplitter.QWidget)

	// Left Top: Queue List
	mw.queueListView = qt.NewQListView(leftSplitter.QWidget)
	mw.queueListModel = qt.NewQStringListModel()
	mw.queueListView.SetModel(mw.queueListModel.QAbstractItemModel)
	mw.queueListView.SetStyleSheet("background-color: white")
	leftSplitter.AddWidget(mw.queueListView.QWidget)

	// Left Bottom: Metadata
	mw.statsTableView = qt.NewQTableView(leftSplitter.QWidget)
	mw.statsModel = qt.NewQStandardItemModel()
	mw.statsModel.SetHorizontalHeaderLabels([]string{"Attribute", "Value"})
	mw.statsTableView.SetModel(mw.statsModel.QAbstractItemModel)
	mw.statsTableView.HorizontalHeader().SetStretchLastSection(false)
	mw.statsTableView.HorizontalHeader().SetSectionResizeMode2(0, qt.QHeaderView__Stretch)
	mw.statsTableView.VerticalHeader().SetVisible(false)
	mw.statsTableView.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	mw.statsTableView.SetSelectionMode(qt.QAbstractItemView__SingleSelection)
	mw.statsTableView.SetStyleSheet("background-color: white")

	leftSplitter.AddWidget(mw.statsTableView.QWidget)
	leftSplitter.SetStretchFactor(0, 1)
	leftSplitter.SetStretchFactor(1, 0)

	// Right Pane: Items
	mw.msgTableView = qt.NewQTableView(splitter.QWidget)
	mw.msgModel = qt.NewQStandardItemModel()
	mw.msgModel.SetHorizontalHeaderLabels([]string{"ID", "Sent At", "Visible At", "Read Count", "Message"})
	mw.msgTableView.SetModel(mw.msgModel.QAbstractItemModel)
	mw.msgTableView.HorizontalHeader().SetStretchLastSection(true)
	mw.msgTableView.SetStyleSheet("background-color: white")

	splitter.AddWidget(mw.msgTableView.QWidget)

	// Set initial splitter sizes
	splitter.SetStretchFactor(0, 1)
	splitter.SetStretchFactor(1, 3)

	// Initialize Client
	mw.client = rsmq.NewClient("localhost:6379", "rsmq:")

	// Signals
	mw.queueListView.SelectionModel().OnCurrentChanged(func(current, previous *qt.QModelIndex) {
		if !current.IsValid() {
			return
		}
		qname := current.Data().ToString()
		mw.UpdateQueueData(qname)
	})

	mw.RefreshQueues()

	return mw
}

func (mw *RSMQTMainWindow) RefreshQueues() {
	queues, err := mw.client.ListQueues()
	if err != nil {
		// Log or show error?
		return
	}
	mw.queueListModel.SetStringList(queues)
}

func (mw *RSMQTMainWindow) UpdateQueueData(qname string) {
	// Stats
	stats, err := mw.client.GetQueueStats(qname)
	mw.statsModel.SetRowCount(0)
	if err == nil {
		data := [][2]string{
			{"Visibility Timeout", strconv.Itoa(stats.Vt)},
			{"Delay", strconv.Itoa(stats.Delay)},
			{"Max Size", strconv.Itoa(stats.MaxSize)},
			{"Total Received", strconv.FormatUint(stats.TotalRecv, 10)},
			{"Total Sent", strconv.FormatUint(stats.TotalSent, 10)},
			{"Messages (Visible)", strconv.FormatInt(stats.Msgs, 10)},
			{"Messages (Hidden)", strconv.FormatInt(stats.HiddenMsgs, 10)},
		}
		for _, row := range data {
			items := []*qt.QStandardItem{
				qt.NewQStandardItem2(row[0]),
				qt.NewQStandardItem2(row[1]),
			}
			mw.statsModel.AppendRow(items)
		}
	} else {
		items := []*qt.QStandardItem{
			qt.NewQStandardItem2("Error"),
			qt.NewQStandardItem2("Could not fetch stats"),
		}
		mw.statsModel.AppendRow(items)
	}

	// Messages
	msgs, err := mw.client.ListMessages(qname)
	if err == nil {
		mw.msgModel.SetRowCount(0)
		for _, m := range msgs {
			items := []*qt.QStandardItem{
				qt.NewQStandardItem2(m.ID),
				qt.NewQStandardItem2(time.UnixMicro(m.Sent).Format(time.DateTime)),
				qt.NewQStandardItem2(time.UnixMilli(m.VisibleAt).Format(time.DateTime)),
				qt.NewQStandardItem2(strconv.Itoa(m.Rc)),
				qt.NewQStandardItem2(m.Body),
			}
			mw.msgModel.AppendRow(items)
		}
	}
}

func main() {
	qt.NewQApplication(os.Args)

	window := NewRSMQTMainWindow()
	window.Show()

	qt.QApplication_Exec()
}
