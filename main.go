package main

import (
	"net"
	"os"
	"strconv"
	"strings"
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

	SSHEnabled  bool
	SSHHost     string
	SSHPort     string
	SSHUser     string
	SSHAuthType string // "password" or "key"
	SSHPass     string
	SSHKeyPath  string
	SSHKeyPassphrase string
}

var globalCfg = Config{
	Host: "localhost",
	Port: "6379",
	Pass: "",
	DB:   0,
	NS:   "rsmq:",

	SSHEnabled:  false,
	SSHHost:     "",
	SSHPort:     "22",
	SSHUser:     "",
	SSHAuthType: "password",
	SSHPass:     "",
	SSHKeyPath:  "",
}

type ConnectWindow struct {
	*qt.QWidget

	hostInput *qt.QLineEdit
	portInput *qt.QLineEdit
	passInput *qt.QLineEdit
	dbInput   *qt.QComboBox
	nsInput   *qt.QLineEdit

	sshEnabledCheck  *qt.QCheckBox
	sshHostInput     *qt.QLineEdit
	sshPortInput     *qt.QLineEdit
	sshUserInput     *qt.QLineEdit
	sshAuthTypeCombo *qt.QComboBox
	sshPassInput     *qt.QLineEdit
	sshKeyPathInput  *qt.QLineEdit
	sshKeyBrowseBtn  *qt.QPushButton
	sshContainer     *qt.QWidget

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
	advLayout := qt.NewQVBoxLayout(advTab)

	cw.sshEnabledCheck = qt.NewQCheckBox(advTab)
	cw.sshEnabledCheck.SetText("Use SSH Tunnel")
	cw.sshEnabledCheck.SetChecked(globalCfg.SSHEnabled)
	advLayout.AddWidget(cw.sshEnabledCheck.QWidget)

	cw.sshContainer = qt.NewQWidget(advTab)
	sshLayout := qt.NewQVBoxLayout(cw.sshContainer)
	sshLayout.SetContentsMargins(0, 0, 0, 0)

	// Helper to create rows
	createRow := func(label string, widget *qt.QWidget) *qt.QWidget {
		row := qt.NewQWidget(cw.sshContainer)
		l := qt.NewQHBoxLayout(row)
		l.SetContentsMargins(0, 0, 0, 0)
		lbl := qt.NewQLabel(row)
		lbl.SetText(label)
		lbl.SetFixedWidth(80)
		l.AddWidget(lbl.QWidget)
		l.AddWidget(widget)
		return row
	}

	cw.sshHostInput = qt.NewQLineEdit(cw.sshContainer)
	cw.sshHostInput.SetText(globalCfg.SSHHost)
	sshLayout.AddWidget(createRow("SSH Host:", cw.sshHostInput.QWidget))

	cw.sshPortInput = qt.NewQLineEdit(cw.sshContainer)
	cw.sshPortInput.SetText(globalCfg.SSHPort)
	sshLayout.AddWidget(createRow("SSH Port:", cw.sshPortInput.QWidget))

	cw.sshUserInput = qt.NewQLineEdit(cw.sshContainer)
	cw.sshUserInput.SetText(globalCfg.SSHUser)
	sshLayout.AddWidget(createRow("SSH User:", cw.sshUserInput.QWidget))

	cw.sshAuthTypeCombo = qt.NewQComboBox(cw.sshContainer)
	cw.sshAuthTypeCombo.AddItem("Password")
	cw.sshAuthTypeCombo.AddItem("Private Key")
	if globalCfg.SSHAuthType == "key" {
		cw.sshAuthTypeCombo.SetCurrentIndex(1)
	} else {
		cw.sshAuthTypeCombo.SetCurrentIndex(0)
	}
	sshLayout.AddWidget(createRow("Auth Type:", cw.sshAuthTypeCombo.QWidget))

	// Password Row
	cw.sshPassInput = qt.NewQLineEdit(cw.sshContainer)
	cw.sshPassInput.SetEchoMode(qt.QLineEdit__Password)
	cw.sshPassInput.SetText(globalCfg.SSHPass)
	passRow := createRow("Password:", cw.sshPassInput.QWidget)
	sshLayout.AddWidget(passRow)

	// Key Row
	keyWidget := qt.NewQWidget(cw.sshContainer)
	keyLayout := qt.NewQHBoxLayout(keyWidget)
	keyLayout.SetContentsMargins(0, 0, 0, 0)
	cw.sshKeyPathInput = qt.NewQLineEdit(keyWidget)
	cw.sshKeyPathInput.SetText(globalCfg.SSHKeyPath)
	cw.sshKeyBrowseBtn = qt.NewQPushButton3("Browse")
	keyLayout.AddWidget(cw.sshKeyPathInput.QWidget)
	keyLayout.AddWidget(cw.sshKeyBrowseBtn.QWidget)

	keyRow := qt.NewQWidget(cw.sshContainer)
	keyRowLayout := qt.NewQHBoxLayout(keyRow)
	keyRowLayout.SetContentsMargins(0, 0, 0, 0)
	keyLbl := qt.NewQLabel(keyRow)
	keyLbl.SetText("Private Key:")
	keyLbl.SetFixedWidth(80)
	keyRowLayout.AddWidget(keyLbl.QWidget)
	keyRowLayout.AddWidget(keyWidget)
	sshLayout.AddWidget(keyRow)

	cw.sshContainer.SetLayout(sshLayout.QLayout)
	advLayout.AddWidget(cw.sshContainer)
	advLayout.AddStretch()

	advTab.SetLayout(advLayout.QLayout)
	tabs.AddTab(advTab, "Advanced")

	// SSH Logic
	updateSSHState := func() {
		enabled := cw.sshEnabledCheck.IsChecked()
		cw.sshContainer.SetEnabled(enabled)

		isKey := cw.sshAuthTypeCombo.CurrentIndex() == 1
		if isKey {
			passRow.Hide()
			keyRow.Show()
		} else {
			passRow.Show()
			keyRow.Hide()
		}
	}
	cw.sshEnabledCheck.OnToggled(func(checked bool) { updateSSHState() })
	cw.sshAuthTypeCombo.OnCurrentIndexChanged(func(index int) { updateSSHState() })
	updateSSHState() // Initial state

	cw.sshKeyBrowseBtn.OnClicked(func() {
		filename := qt.QFileDialog_GetOpenFileName4(cw.QWidget, "Select Private Key", "", "All Files (*)")
		if filename != "" {
			cw.sshKeyPathInput.SetText(filename)
		}
	})

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

		globalCfg.SSHEnabled = cw.sshEnabledCheck.IsChecked()
		globalCfg.SSHHost = cw.sshHostInput.Text()
		globalCfg.SSHPort = cw.sshPortInput.Text()
		globalCfg.SSHUser = cw.sshUserInput.Text()
		if cw.sshAuthTypeCombo.CurrentIndex() == 1 {
			globalCfg.SSHAuthType = "key"
		} else {
			globalCfg.SSHAuthType = "password"
		}
		globalCfg.SSHPass = cw.sshPassInput.Text()
		globalCfg.SSHKeyPath = cw.sshKeyPathInput.Text()

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

		sshEnabled := cw.sshEnabledCheck.IsChecked()
		sshHost := cw.sshHostInput.Text()
		sshPort := cw.sshPortInput.Text()
		sshUser := cw.sshUserInput.Text()
		sshAuthType := "password"
		if cw.sshAuthTypeCombo.CurrentIndex() == 1 {
			sshAuthType = "key"
		}
		sshPass := cw.sshPassInput.Text()
		sshKeyPath := cw.sshKeyPathInput.Text()
		// We use globalCfg.SSHKeyPassphrase for test if available, or prompt?
		// For test, we might want to start fresh or use what's in global if it matches?
		// Let's assume empty passphrase initially for test.
		sshKeyPassphrase := globalCfg.SSHKeyPassphrase

		var dialer func(string, string) (net.Conn, error)
		var err error

		if sshEnabled {
			// Helper to try dial
			tryDial := func(passphrase string) error {
				dialer, err = rsmq.DialSSH(rsmq.SSHConfig{
					Host:       sshHost,
					Port:       sshPort,
					User:       sshUser,
					AuthType:   sshAuthType,
					Password:   sshPass,
					KeyPath:    sshKeyPath,
					Passphrase: passphrase,
				})
				return err
			}

			err = tryDial(sshKeyPassphrase)
			if err != nil && strings.Contains(err.Error(), "passphrase") && sshAuthType == "key" {
				// Prompt for passphrase
				var ok bool
				text := qt.QInputDialog_GetText4(cw.QWidget, "SSH Key Passphrase", "Enter passphrase for private key:", qt.QLineEdit__Password, "", &ok)
				if ok && text != "" {
					sshKeyPassphrase = text
					err = tryDial(text)
					if err == nil {
						// Save the successful passphrase globally so Connect works
						globalCfg.SSHKeyPassphrase = text
					}
				}
			}
		}

		var toolTip string
		if err != nil {
			toolTip = "❌ SSH Error: " + err.Error()
		} else {
			testAddr := testHost + ":" + testPort
			client := rsmq.NewClientWithDialer(testAddr, testPass, testDB, testNS, dialer)
			err = client.TestConnection()

			if err != nil {
				toolTip = "❌ Redis Error: " + err.Error()
			} else {
				toolTip = "✅ Connection Successful"
			}
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
	var dialer func(string, string) (net.Conn, error)
	if globalCfg.SSHEnabled {
		// Try to dial with existing passphrase
		var err error
		dialer, err = rsmq.DialSSH(rsmq.SSHConfig{
			Host:       globalCfg.SSHHost,
			Port:       globalCfg.SSHPort,
			User:       globalCfg.SSHUser,
			AuthType:   globalCfg.SSHAuthType,
			Password:   globalCfg.SSHPass,
			KeyPath:    globalCfg.SSHKeyPath,
			Passphrase: globalCfg.SSHKeyPassphrase,
		})
		
		// If failed due to passphrase, prompt
		if err != nil && strings.Contains(err.Error(), "passphrase") && globalCfg.SSHAuthType == "key" {
			var ok bool
			text := qt.QInputDialog_GetText4(mw.QWidget, "SSH Key Passphrase", "Enter passphrase for private key:", qt.QLineEdit__Password, "", &ok)
			if ok && text != "" {
				globalCfg.SSHKeyPassphrase = text
				dialer, err = rsmq.DialSSH(rsmq.SSHConfig{
					Host:       globalCfg.SSHHost,
					Port:       globalCfg.SSHPort,
					User:       globalCfg.SSHUser,
					AuthType:   globalCfg.SSHAuthType,
					Password:   globalCfg.SSHPass,
					KeyPath:    globalCfg.SSHKeyPath,
					Passphrase: globalCfg.SSHKeyPassphrase,
				})
			}
		}

		if err != nil {
			qt.QMessageBox_Critical(mw.QWidget, "Connection Error", "Failed to establish SSH tunnel: "+err.Error())
			// We should probably fail gracefully, but NewRSMQTMainWindow returns *RSMQTMainWindow.
			// Maybe just return nil or let the client be broken?
			// If client is broken, operations will fail.
		}
	}

	addr := globalCfg.Host + ":" + globalCfg.Port
	mw.client = rsmq.NewClientWithDialer(addr, globalCfg.Pass, globalCfg.DB, globalCfg.NS, dialer)

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
