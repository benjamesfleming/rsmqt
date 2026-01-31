package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/benjamesfleming/rsmqt/lib/rsmq"
	qt "github.com/mappu/miqt/qt6"
)

type Config struct {
	Host string
	Port string
	Pass string
	DB   int
	NS   string
}

var globalCfg = Config{
	Host: "localhost",
	Port: "6379",
	Pass: "",
	DB:   0,
	NS:   "rsmq:",
}

type ConnectWindow struct {
	*qt.QWidget

	hostInput *qt.QLineEdit
	portInput *qt.QLineEdit
	passInput *qt.QLineEdit
	dbInput   *qt.QComboBox
	nsInput   *qt.QLineEdit

	connectBtn *qt.QPushButton
	testBtn    *qt.QPushButton

	onConnect func()
}

func NewConnectWindow(onConnect func()) *ConnectWindow {
	cw := &ConnectWindow{}
	cw.QWidget = qt.NewQWidget2()
	cw.SetWindowTitle("RSMQ Connection")
	cw.SetGeometry(300, 300, 300, 250)
	cw.onConnect = onConnect

	layout := qt.NewQVBoxLayout(cw.QWidget)

	tabs := qt.NewQTabWidget(cw.QWidget)
	layout.AddWidget(tabs.QWidget)

	// Basic Tab
	basicTab := qt.NewQWidget(tabs.QWidget)
	basicForm := qt.NewQFormLayout(basicTab)

	cw.hostInput = qt.NewQLineEdit(basicTab)
	cw.hostInput.SetText(globalCfg.Host)
	basicForm.AddRow3("Host:", cw.hostInput.QWidget)

	cw.portInput = qt.NewQLineEdit(basicTab)
	cw.portInput.SetText(globalCfg.Port)
	basicForm.AddRow3("Port:", cw.portInput.QWidget)

	cw.passInput = qt.NewQLineEdit(basicTab)
	cw.passInput.SetEchoMode(qt.QLineEdit__Password)
	cw.passInput.SetText(globalCfg.Pass)
	basicForm.AddRow3("Password:", cw.passInput.QWidget)

	cw.dbInput = qt.NewQComboBox(basicTab)
	for i := 0; i < 16; i++ {
		cw.dbInput.AddItem(strconv.Itoa(i))
	}
	cw.dbInput.SetCurrentIndex(globalCfg.DB)
	basicForm.AddRow3("DB:", cw.dbInput.QWidget)

	cw.nsInput = qt.NewQLineEdit(basicTab)
	cw.nsInput.SetText(globalCfg.NS)
	basicForm.AddRow3("Namespace:", cw.nsInput.QWidget)

	basicTab.SetLayout(basicForm.QLayout)
	tabs.AddTab(basicTab, "Basic")

	// Advanced Tab
	advTab := qt.NewQWidget(tabs.QWidget)
	advForm := qt.NewQFormLayout(advTab)

	sshHost := qt.NewQLineEdit(advTab)
	advForm.AddRow3("SSH Host:", sshHost.QWidget)

	sshPort := qt.NewQLineEdit(advTab)
	sshPort.SetText("22")
	advForm.AddRow3("SSH Port:", sshPort.QWidget)

	sshUser := qt.NewQLineEdit(advTab)
	advForm.AddRow3("SSH User:", sshUser.QWidget)

	sshPass := qt.NewQLineEdit(advTab)
	sshPass.SetEchoMode(qt.QLineEdit__Password)
	advForm.AddRow3("SSH Key/Pass:", sshPass.QWidget)

	advTab.SetLayout(advForm.QLayout)
	tabs.AddTab(advTab, "Advanced")

	// Buttons
	btnLayout := qt.NewQHBoxLayout(nil)
	cw.testBtn = qt.NewQPushButton3("Test Connection")
	cw.connectBtn = qt.NewQPushButton3("Connect")

	btnLayout.AddWidget(cw.testBtn.QWidget)
	btnLayout.AddWidget(cw.connectBtn.QWidget)
	layout.AddLayout(btnLayout.QLayout)

	cw.connectBtn.OnClicked(func() {
		globalCfg.Host = cw.hostInput.Text()
		globalCfg.Port = cw.portInput.Text()
		globalCfg.Pass = cw.passInput.Text()
		globalCfg.DB = cw.dbInput.CurrentIndex()
		globalCfg.NS = cw.nsInput.Text()

		if cw.onConnect != nil {
			cw.onConnect()
		}
	})

	cw.testBtn.OnClicked(func() {
		fmt.Println("Test Connection clicked (Stub)")
	})

	return cw
}

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

	// Actions
	actDisconnect *qt.QAction
	actNewQueue   *qt.QAction
	actDelQueue   *qt.QAction
	actSendMsg    *qt.QAction
	actEditQueue  *qt.QAction
	actClearQueue *qt.QAction
	actDelMsg     *qt.QAction
}

func NewRSMQTMainWindow(onDisconnect func()) *RSMQTMainWindow {
	mw := &RSMQTMainWindow{}
	mw.QMainWindow = qt.NewQMainWindow2()
	mw.SetWindowTitle("RSMQ UI")
	mw.SetStyleSheet("background-color: #f1f2f6")
	mw.SetGeometry(100, 100, 1000, 700)

	// Actions
	mw.actDisconnect = qt.NewQAction5("Disconnect", mw.QObject)
	mw.actDisconnect.OnTriggered(func() {
		if onDisconnect != nil {
			onDisconnect()
		}
	})

	mw.actNewQueue = qt.NewQAction5("New Queue", mw.QObject)
	mw.actNewQueue.OnTriggered(func() { fmt.Println("Action: New Queue") })

	mw.actEditQueue = qt.NewQAction5("Edit Queue", mw.QObject)
	mw.actEditQueue.OnTriggered(func() { fmt.Println("Action: Edit Queue") })
	mw.actEditQueue.SetEnabled(false)

	mw.actDelQueue = qt.NewQAction5("Delete Queue", mw.QObject)
	mw.actDelQueue.OnTriggered(func() { fmt.Println("Action: Delete Queue") })
	mw.actDelQueue.SetEnabled(false)

	mw.actClearQueue = qt.NewQAction5("Clear Queue", mw.QObject)
	mw.actClearQueue.OnTriggered(func() { fmt.Println("Action: Clear Queue") })
	mw.actClearQueue.SetEnabled(false)

	mw.actSendMsg = qt.NewQAction5("Send Message", mw.QObject)
	mw.actSendMsg.OnTriggered(func() { fmt.Println("Action: Send Message") })
	mw.actSendMsg.SetEnabled(false)

	mw.actDelMsg = qt.NewQAction5("Delete Message", mw.QObject)
	mw.actDelMsg.OnTriggered(func() { fmt.Println("Action: Delete Message") })
	mw.actDelMsg.SetEnabled(false)

	// Menu Bar
	mb := mw.MenuBar()

	fileMenu := mb.AddMenuWithTitle("File")
	fileMenu.AddAction(mw.actDisconnect)

	queueMenu := mb.AddMenuWithTitle("Queue")
	queueMenu.AddAction(mw.actNewQueue)
	queueMenu.AddSeparator()
	queueMenu.AddAction(mw.actEditQueue)
	queueMenu.AddAction(mw.actClearQueue)
	queueMenu.AddAction(mw.actDelQueue)

	msgMenu := mb.AddMenuWithTitle("Message")
	msgMenu.AddAction(mw.actSendMsg)
	msgMenu.AddAction(mw.actDelMsg)

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
	mw.msgTableView.SetSelectionMode(qt.QAbstractItemView__SingleSelection)
	mw.msgTableView.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	mw.msgTableView.SetStyleSheet("QTableView { background-color: white; } QTableView::item:selected { background-color: #f5f5f5; color: black; } QTableView::item:focus { background-color: #0078d7; color: white; }")

	splitter.AddWidget(mw.msgTableView.QWidget)

	// Set initial splitter sizes
	splitter.SetStretchFactor(0, 1)
	splitter.SetStretchFactor(1, 3)

	// Initialize Client
	addr := globalCfg.Host + ":" + globalCfg.Port
	mw.client = rsmq.NewClient(addr, globalCfg.Pass, globalCfg.DB, globalCfg.NS)

	// Signals
	mw.queueListView.SelectionModel().OnSelectionChanged(func(selected, deselected *qt.QItemSelection) {
		indexes := mw.queueListView.SelectionModel().SelectedIndexes()
		hasSelection := len(indexes) > 0

		mw.actEditQueue.SetEnabled(hasSelection)
		mw.actDelQueue.SetEnabled(hasSelection)
		mw.actClearQueue.SetEnabled(hasSelection)
		mw.actSendMsg.SetEnabled(hasSelection)

		if !hasSelection {
			mw.statsModel.SetRowCount(0)
			mw.msgModel.SetRowCount(0)
			return
		}

		idx := indexes[0]
		qname := idx.Data().ToString()
		mw.UpdateQueueData(qname)
	})

	mw.msgTableView.SelectionModel().OnSelectionChanged(func(selected, deselected *qt.QItemSelection) {
		mw.actDelMsg.SetEnabled(mw.msgTableView.SelectionModel().HasSelection())
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

	var connectWindow *ConnectWindow
	var mainWindow *RSMQTMainWindow

	connectWindow = NewConnectWindow(func() {
		mainWindow = NewRSMQTMainWindow(func() {
			mainWindow.Close()
			connectWindow.Show()
		})
		mainWindow.Show()
		connectWindow.Close()
	})
	connectWindow.Show()

	qt.QApplication_Exec()
}
