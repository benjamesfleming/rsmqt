package main

import (
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
		cw.testBtn.SetEnabled(false)
		cw.testBtn.Repaint() // Ensure UI updates

		// Mock config for testing
		testHost := cw.hostInput.Text()
		testPort := cw.portInput.Text()
		testPass := cw.passInput.Text()
		testDB := cw.dbInput.CurrentIndex()
		testNS := cw.nsInput.Text()

		testAddr := testHost + ":" + testPort
		client := rsmq.NewClient(testAddr, testPass, testDB, testNS)

		err := client.TestConnection()

		var toolTip string
		if err != nil {
			toolTip = "❌ Error: " + err.Error()
		} else {
			toolTip = "✅ Connection Successful"
		}
		cw.testBtn.SetToolTip(toolTip)

		// Force tooltip to show immediately
		qt.QToolTip_ShowText(qt.QCursor_Pos().OperatorMinusAssign(qt.NewQPoint2(0, 25)), cw.testBtn.ToolTip())

		cw.testBtn.SetEnabled(true)
	})

	return cw
}

type QueueDialog struct {
	*qt.QDialog
	Name    *qt.QLineEdit
	Vt      *qt.QSpinBox
	Delay   *qt.QSpinBox
	MaxSize *qt.QSpinBox
}

func NewQueueDialog(parent *qt.QWidget, title string, isEdit bool) *QueueDialog {
	qd := &QueueDialog{}
	qd.QDialog = qt.NewQDialog(parent)
	qd.SetWindowTitle(title)

	layout := qt.NewQFormLayout(qd.QWidget)

	qd.Name = qt.NewQLineEdit(qd.QWidget)
	if isEdit {
		qd.Name.SetReadOnly(true)
	}
	layout.AddRow3("Name:", qd.Name.QWidget)

	qd.Vt = qt.NewQSpinBox(qd.QWidget)
	qd.Vt.SetRange(0, 999999)
	qd.Vt.SetValue(30)
	layout.AddRow3("Visibility Timeout (s):", qd.Vt.QWidget)

	qd.Delay = qt.NewQSpinBox(qd.QWidget)
	qd.Delay.SetRange(0, 999999)
	qd.Delay.SetValue(0)
	layout.AddRow3("Delay (s):", qd.Delay.QWidget)

	qd.MaxSize = qt.NewQSpinBox(qd.QWidget)
	qd.MaxSize.SetRange(1024, 65536*100)
	qd.MaxSize.SetValue(65536)
	layout.AddRow3("Max Message Size (bytes):", qd.MaxSize.QWidget)

	btns := qt.NewQDialogButtonBox(qd.QWidget)
	btns.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	layout.AddWidget(btns.QWidget)

	btns.OnAccepted(qd.Accept)
	btns.OnRejected(qd.Reject)

	return qd
}

type SendMessageDialog struct {
	*qt.QDialog
	Message *qt.QTextEdit
}

func NewSendMessageDialog(parent *qt.QWidget) *SendMessageDialog {
	smd := &SendMessageDialog{}
	smd.QDialog = qt.NewQDialog(parent)
	smd.SetWindowTitle("Send Message")
	smd.SetMinimumSize2(400, 300)

	layout := qt.NewQVBoxLayout(smd.QWidget)

	smd.Message = qt.NewQTextEdit(smd.QWidget)
	smd.Message.SetStyleSheet("background-color: white;")
	layout.AddWidget(smd.Message.QWidget)

	btns := qt.NewQDialogButtonBox(smd.QWidget)
	btns.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	layout.AddWidget(btns.QWidget)

	btns.OnAccepted(smd.Accept)
	btns.OnRejected(smd.Reject)

	return smd
}

type RSMQTMainWindow struct {
	*qt.QMainWindow

	client *rsmq.Client

	currentQueueStats *rsmq.QueueStats

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
	mw.actNewQueue.OnTriggered(func() {
		dlg := NewQueueDialog(mw.QWidget, "New Queue", false)
		if dlg.Exec() == int(qt.QDialog__Accepted) {
			name := dlg.Name.Text()
			vt := dlg.Vt.Value()
			delay := dlg.Delay.Value()
			maxsize := dlg.MaxSize.Value()

			err := mw.client.CreateQueue(name, vt, delay, maxsize)
			if err != nil {
				qt.QMessageBox_Critical(mw.QWidget, "Error", err.Error())
			} else {
				mw.RefreshQueues()
			}
		}
	})

	mw.actDelQueue = qt.NewQAction5("Delete Queue", mw.QObject)
	mw.actDelQueue.OnTriggered(func() {
		if mw.currentQueueStats == nil {
			return
		}
		qname := mw.currentQueueStats.Name
		ret := qt.QMessageBox_Question(mw.QWidget, "Confirm Delete", "Are you sure you want to delete queue '"+qname+"'?")
		if ret == qt.QMessageBox__Yes {
			err := mw.client.DeleteQueue(qname)
			if err != nil {
				qt.QMessageBox_Critical(mw.QWidget, "Error", err.Error())
			} else {
				mw.currentQueueStats = nil
				mw.statsModel.SetRowCount(0)
				mw.msgModel.SetRowCount(0)
				mw.RefreshQueues()
			}
		}
	})
	mw.actDelQueue.SetEnabled(false)

	mw.actClearQueue = qt.NewQAction5("Clear Queue", mw.QObject)
	mw.actClearQueue.OnTriggered(func() {
		if mw.currentQueueStats == nil {
			return
		}
		qname := mw.currentQueueStats.Name
		ret := qt.QMessageBox_Question(mw.QWidget, "Confirm Clear", "Are you sure you want to clear queue '"+qname+"'? This will delete all messages.")
		if ret == qt.QMessageBox__Yes {
			err := mw.client.ClearQueue(qname)
			if err != nil {
				qt.QMessageBox_Critical(mw.QWidget, "Error", "Failed to clear queue: "+err.Error())
			} else {
				mw.UpdateQueueData(qname)
			}
		}
	})
	mw.actClearQueue.SetEnabled(false)

	mw.actSendMsg = qt.NewQAction5("Send Message", mw.QObject)
	mw.actSendMsg.OnTriggered(func() {
		if mw.currentQueueStats == nil {
			return
		}
		dlg := NewSendMessageDialog(mw.QWidget)
		if dlg.Exec() == int(qt.QDialog__Accepted) {
			msg := dlg.Message.ToPlainText()
			err := mw.client.SendMessage(mw.currentQueueStats.Name, msg)
			if err != nil {
				qt.QMessageBox_Critical(mw.QWidget, "Error", err.Error())
			} else {
				mw.UpdateQueueData(mw.currentQueueStats.Name)
			}
		}
	})
	mw.actSendMsg.SetEnabled(false)

	mw.actDelMsg = qt.NewQAction5("Delete Message", mw.QObject)
	mw.actDelMsg.OnTriggered(func() {
		if mw.currentQueueStats == nil {
			return
		}
		// Get selected message ID
		indexes := mw.msgTableView.SelectionModel().SelectedIndexes()
		if len(indexes) == 0 {
			return
		}

		// Use the row of the first selected item to get the ID from column 0
		row := indexes[0].Row()
		idIdx := mw.msgModel.Index(row, 0, qt.NewQModelIndex())
		id := mw.msgModel.Data(idIdx, int(qt.DisplayRole)).ToString()

		err := mw.client.DeleteMessage(mw.currentQueueStats.Name, id)
		if err != nil {
			qt.QMessageBox_Critical(mw.QWidget, "Error", err.Error())
		} else {
			mw.UpdateQueueData(mw.currentQueueStats.Name)
		}
	})
	mw.actDelMsg.SetEnabled(false)

	// Menu Bar
	mb := mw.MenuBar()

	fileMenu := mb.AddMenuWithTitle("File")
	fileMenu.AddAction(mw.actDisconnect)

	queueMenu := mb.AddMenuWithTitle("Queue")
	queueMenu.AddAction(mw.actNewQueue)
	queueMenu.AddSeparator()
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
	mw.statsTableView.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	mw.statsTableView.SetStyleSheet("QTableView { background-color: white; } QTableView::item:selected { background-color: #f5f5f5; color: black; } QTableView::item:focus { background-color: #0078d7; color: white; }")

	leftSplitter.AddWidget(mw.statsTableView.QWidget)
	leftSplitter.SetStretchFactor(0, 6)
	leftSplitter.SetStretchFactor(1, 4)

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
		mw.currentQueueStats = stats
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
		mw.currentQueueStats = nil
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
				qt.NewQStandardItem2(m.Sent.Format(time.DateTime)),
				qt.NewQStandardItem2(m.VisibleAt.Format(time.DateTime)),
				qt.NewQStandardItem2(strconv.Itoa(m.Rc)),
				qt.NewQStandardItem2(m.Body),
			}
			mw.msgModel.AppendRow(items)
		}
	}
}

func main() {
	app := qt.NewQApplication(os.Args)
	app.SetStyleSheet("QToolTip { background-color: #333; color: white; padding: 2px; }")

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
