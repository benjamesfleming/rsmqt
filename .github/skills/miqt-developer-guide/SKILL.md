---
name: miqt-developer-guide
description: A comprehensive guide for advanced developers building Qt applications in Go using MIQT. Use this when writing Go code with Qt/MIQT bindings.
---

# MIQT Developer Skills Guide

A comprehensive guide for advanced developers building Qt applications in Go using MIQT (MIT-licensed Qt bindings). This guide assumes familiarity with Go and Qt fundamentals.

## Table of Contents

1. [Architecture & Design Principles](#architecture--design-principles)
2. [Widget Hierarchy & Lifecycle](#widget-hierarchy--lifecycle)
3. [Event Handling & Signals/Slots](#event-handling--signalsslots)
4. [Custom Models & Views (MVC)](#custom-models--views-mvc)
5. [Threading & Concurrency](#threading--concurrency)
6. [Performance Optimization](#performance-optimization)
7. [Styling & Theming](#styling--theming)
8. [QML Integration](#qml-integration)
9. [Advanced Patterns](#advanced-patterns)
10. [Edge Cases & Troubleshooting](#edge-cases--troubleshooting)

---

## Architecture & Design Principles

### MIQT Philosophy

MIQT is a straightforward CGO binding of Qt 5.15/6.4+ API that maintains Qt's traditional object hierarchy while providing Go's idioms. Key design decisions:

**1. CGO-Based Binding**
- Direct C++ interop; no wrapper abstractions reduce overhead
- Compile-time C++ knowledge required; builds require Qt development toolchain
- Binary size ~6MB stripped (use `-ldflags="-s -w"` and `upx` for further compression)
- Platform support: Linux, Windows, macOS, Android, FreeBSD

**2. Qt Object Model in Go**
- Embedded pointers mimic Qt's inheritance (e.g., `*qt.QMainWindow` embeds `*qt.QWidget`)
- Memory management: Qt parents own children; avoid double-deletion
- Method receivers are pointers; no implicit copies
- Constructor variants use numeric suffixes: `NewQPushButton()`, `NewQPushButton2(text)`, `NewQPushButton3(text, parent)`

**3. Go Idioms Applied**
- Signal slots use Go closures/callbacks via `OnSignalName()` methods
- Error handling through Go's standard patterns (no exceptions)
- Goroutines for concurrent work; special `mainthread` package for UI thread safety

### Design Patterns for MIQT

**Pattern 1: Embedded Types for Inheritance**
```go
type MyCustomWidget struct {
    *qt.QWidget
    label *qt.QLabel
    data  map[string]interface{}
}

func NewMyCustomWidget() *MyCustomWidget {
    w := &MyCustomWidget{
        QWidget: qt.NewQWidget(),
        data:    make(map[string]interface{}),
    }
    w.setupUI()
    return w
}

func (w *MyCustomWidget) setupUI() {
    layout := qt.NewQVBoxLayout2(w.QWidget)
    w.label = qt.NewQLabel(w.QWidget)
    layout.AddWidget(w.label.QWidget)
    w.SetLayout(layout.QLayout)
}
```

**Pattern 2: Lifecycle Management with Defer**
```go
// Use defer for cleanup in short-lived objects
widget := qt.NewQWidget2()
defer widget.Delete()

// For long-lived objects (owned by parent), skip defer
mainWindow := qt.NewQMainWindow2()
// parent relationship established, no explicit Delete needed
```

---

## Widget Hierarchy & Lifecycle

### The Widget Tree

MIQT follows Qt's parent-child hierarchy where parents manage children lifetimes:

```go
window := qt.NewQMainWindow2()                    // Top-level widget
centralWidget := qt.NewQWidget(window.QWidget)   // Parent: window
layout := qt.NewQVBoxLayout2(centralWidget)      // Parent: centralWidget
button := qt.NewQPushButton5("Click", centralWidget)  // Parent: centralWidget

// When window is destroyed, all children destroyed automatically
```

**Key Rules:**
1. **Ownership**: Parent widgets delete children in destructor
2. **Visibility**: Hidden parents hide all children (unless explicitly overridden)
3. **Focus Chain**: Tabulation order follows widget insertion order
4. **Geometry**: Child coordinates relative to parent

### Widget Lifecycle Events

MIQT provides hooks into critical lifecycle events:

```go
widget := qt.NewQWidget2()

// Show event: fired when widget becomes visible
// Use for deferred initialization
widget.OnShowEvent(func(event *qt.QShowEvent) {
    // Initialize expensive resources
})

// Hide event: fired when widget hidden
widget.OnHideEvent(func(event *qt.QHideEvent) {
    // Cleanup or pause operations
})

// Close event: fired on close request (interceptable)
widget.OnCloseEvent(func(event *qt.QCloseEvent) {
    // Perform cleanup, validate state before closing
    // event.Accept() to allow close, event.Ignore() to prevent
})

// Resize event: fired after geometry changes
widget.OnResizeEvent(func(event *qt.QResizeEvent) {
    // Recalculate layout or adjust child positions
})
```

### Layout System

Layouts manage child widget geometry automatically:

```go
// Vertical stacking
vbox := qt.NewQVBoxLayout2(parent)
vbox.AddWidget(button1.QWidget)
vbox.AddWidget(button2.QWidget)
vbox.SetSpacing(10)        // Pixels between items
vbox.SetContentsMargins(5, 5, 5, 5)  // Margin around layout
parent.SetLayout(vbox.QLayout)

// Horizontal stacking
hbox := qt.NewQHBoxLayout2(parent)
hbox.AddLayout(vbox.QLayout, 1)  // Weight for stretchable space
hbox.AddWidget(button3.QWidget, 0)

// Grid layout
grid := qt.NewQGridLayout2(parent)
grid.AddWidget(label.QWidget, 0, 0)      // Row 0, Col 0
grid.AddWidget(input.QWidget, 0, 1)      // Row 0, Col 1
grid.SetColumnStretch(1, 1)  // Column 1 expands to fill space
parent.SetLayout(grid.QLayout)
```

---

## Event Handling & Signals/Slots

### Signal/Slot Mechanism

MIQT maps Qt's signal-slot mechanism to Go closures:

```go
button := qt.NewQPushButton3("Click me")

// OnClicked fires when button clicked
var clickCount int
button.OnClicked(func() {
    clickCount++
    button.SetText(fmt.Sprintf("Clicked %d times", clickCount))
})

// OnPressed fires on press; OnReleased on release
button.OnPressed(func() { fmt.Println("Button pressed") })
button.OnReleased(func() { fmt.Println("Button released") })
```

**Common Signals by Widget Type:**

| Widget | Signal | Callback Signature |
|--------|--------|-------------------|
| QPushButton | OnClicked | `func()` |
| QLineEdit | OnTextChanged | `func(text string)` |
| QComboBox | OnCurrentIndexChanged | `func(index int)` |
| QSlider | OnValueChanged | `func(value int)` |
| QCheckBox | OnToggled | `func(checked bool)` |
| QTimer | OnTimeout | `func()` |

### Custom Event Handling

Override virtual methods for lower-level events:

```go
type CustomWidget struct {
    *qt.QWidget
}

func NewCustomWidget() *CustomWidget {
    w := &CustomWidget{QWidget: qt.NewQWidget2()}
    
    // Mouse press event
    w.OnMousePressEvent(func(event *qt.QMouseEvent) {
        pos := event.Pos()
        fmt.Printf("Mouse pressed at: %d, %d\n", pos.X(), pos.Y())
    })
    
    // Key press event
    w.OnKeyPressEvent(func(event *qt.QKeyEvent) {
        key := event.Key()
        if key == int(qt.Key_Escape) {
            w.Close()
        }
    })
    
    // Paint event (custom drawing)
    w.OnPaintEvent(func(event *qt.QPaintEvent) {
        painter := qt.NewQPainter(w.QWidget)
        defer painter.Delete()
        
        painter.DrawText4(10, 10, "Custom text")
        painter.End()
    })
    
    return w
}
```

### Timer-Based Operations

```go
// Single-shot timer
timer := qt.NewQTimer2(window.QObject)
timer.SetSingleShot(true)  // Fire once only
timer.Start(1000)  // 1000ms delay
timer.OnTimeout(func() {
    fmt.Println("Timer fired once")
})

// Periodic timer
ticker := qt.NewQTimer2(window.QObject)
ticker.Start(500)  // Fire every 500ms
var count int
ticker.OnTimeout(func() {
    count++
    if count >= 10 {
        ticker.Stop()  // Stop after 10 firings
    }
})
```

---

## Custom Models & Views (MVC)

### QAbstractListModel Pattern

Implement custom data models for list views:

```go
type StringListModel struct {
    *qt.QAbstractListModel
    items []string
}

func NewStringListModel(items []string) *StringListModel {
    m := &StringListModel{
        QAbstractListModel: qt.NewQAbstractListModel(),
        items:              items,
    }
    
    // RowCount: required; returns number of items
    m.OnRowCount(func(parent *qt.QModelIndex) int {
        return len(m.items)
    })
    
    // Data: required; returns data for display/edit roles
    m.OnData(func(idx *qt.QModelIndex, role int) *qt.QVariant {
        if !idx.IsValid() || idx.Row() >= len(m.items) {
            return qt.NewQVariant()
        }
        
        switch qt.ItemDataRole(role) {
        case qt.DisplayRole, qt.EditRole:
            return qt.NewQVariant14(m.items[idx.Row()])
        case qt.BackgroundRole:
            // Alternate row colors
            if idx.Row()%2 == 0 {
                return qt.NewQVariant15(qt.NewQBrush6(qt.NewQColor3(240, 240, 240)))
            }
        }
        return qt.NewQVariant()
    })
    
    // SetData: optional; handles editing
    m.OnSetData(func(idx *qt.QModelIndex, value *qt.QVariant, role int) bool {
        if idx.IsValid() && role == qt.EditRole {
            m.items[idx.Row()] = value.ToString()
            return true
        }
        return false
    })
    
    // Flags: optional; controls item editability
    m.OnFlags(func(idx *qt.QModelIndex) qt.ItemFlag {
        return qt.ItemIsEnabled | qt.ItemIsSelectable | qt.ItemIsEditable
    })
    
    return m
}

// Method to add items and update view
func (m *StringListModel) AddItem(item string) {
    newRow := len(m.items)
    m.BeginInsertRows(nil, newRow, newRow)
    m.items = append(m.items, item)
    m.EndInsertRows()
}

func (m *StringListModel) RemoveItem(row int) {
    if row >= 0 && row < len(m.items) {
        m.BeginRemoveRows(nil, row, row)
        m.items = append(m.items[:row], m.items[row+1:]...)
        m.EndRemoveRows()
    }
}

func (m *StringListModel) Clear() {
    m.BeginRemoveRows(nil, 0, len(m.items)-1)
    m.items = []string{}
    m.EndRemoveRows()
}
```

**Using the Model:**
```go
view := qt.NewQListView2()
model := NewStringListModel([]string{"Item 1", "Item 2", "Item 3"})
view.SetModel(model.QAbstractItemModel)

// React to selection changes
view.SetSelectionMode(qt.SingleSelection)
view.OnDoubleClicked(func(idx *qt.QModelIndex) {
    if idx.IsValid() {
        fmt.Printf("Double clicked: %s\n", model.items[idx.Row()])
    }
})
```

### QAbstractTableModel Pattern

For 2D data:

```go
type TableData struct {
    *qt.QAbstractTableModel
    rows    int
    cols    int
    data    map[string]interface{}
}

func NewTableData(rows, cols int) *TableData {
    m := &TableData{
        QAbstractTableModel: qt.NewQAbstractTableModel(),
        rows:               rows,
        cols:               cols,
        data:               make(map[string]interface{}),
    }
    
    m.OnRowCount(func(parent *qt.QModelIndex) int {
        return m.rows
    })
    
    m.OnColumnCount(func(parent *qt.QModelIndex) int {
        return m.cols
    })
    
    m.OnData(func(idx *qt.QModelIndex, role int) *qt.QVariant {
        if !idx.IsValid() {
            return qt.NewQVariant()
        }
        
        if role == qt.DisplayRole {
            key := fmt.Sprintf("%d:%d", idx.Row(), idx.Column())
            if val, ok := m.data[key]; ok {
                return qt.NewQVariant14(fmt.Sprint(val))
            }
        }
        return qt.NewQVariant()
    })
    
    m.OnHeaderData(func(section int, orientation qt.Orientation, role int) *qt.QVariant {
        if role == qt.DisplayRole {
            if orientation == qt.Horizontal {
                return qt.NewQVariant14(fmt.Sprintf("Col %d", section))
            } else {
                return qt.NewQVariant14(fmt.Sprintf("Row %d", section))
            }
        }
        return qt.NewQVariant()
    })
    
    return m
}

func (m *TableData) SetCell(row, col int, value interface{}) {
    key := fmt.Sprintf("%d:%d", row, col)
    m.data[key] = value
    idx := qt.NewQModelIndex()
    m.DataChanged(idx, idx)
}
```

---

## Threading & Concurrency

### Main Thread Safety

MIQT UI operations must run on the main thread. Use `mainthread` package for goroutine-safe UI updates:

```go
import "github.com/mappu/miqt/qt6/mainthread"

// Unsafe: goroutine directly calling Qt
go func() {
    label.SetText("Updated")  // May crash or deadlock
}()

// Safe: marshal UI updates to main thread
go func() {
    for i := 0; i < 100; i++ {
        mainthread.Wait(func() {
            label.SetText(fmt.Sprintf("Count: %d", i))
        })
        time.Sleep(100 * time.Millisecond)
    }
}()
```

### Complete Threading Example

```go
package main

import (
    "fmt"
    "os"
    "time"

    qt "github.com/mappu/miqt/qt6"
    "github.com/mappu/miqt/qt6/mainthread"
)

func main() {
    qt.NewQApplication(os.Args)

    window := qt.NewQMainWindow2()
    window.SetWindowTitle("Threading Example")

    widget := qt.NewQWidget(window.QWidget)
    layout := qt.NewQVBoxLayout2(widget)

    label := qt.NewQLabel(widget)
    label.SetText("Waiting...")
    layout.AddWidget(label.QWidget)

    button := qt.NewQPushButton3("Start Work", widget)
    layout.AddWidget(button.QWidget)

    window.SetCentralWidget(widget)

    button.OnClicked(func() {
        button.SetDisabled(true)
        
        // Long-running work in goroutine
        go func() {
            for i := 0; i < 10; i++ {
                time.Sleep(500 * time.Millisecond)
                
                // Safe UI update
                mainthread.Wait(func() {
                    label.SetText(fmt.Sprintf("Processing: %d/10", i+1))
                })
            }
            
            // Re-enable button
            mainthread.Wait(func() {
                button.SetDisabled(false)
                label.SetText("Done!")
            })
        }()
    })

    window.Show()
    qt.QApplication_Exec()
}
```

### Thread Pool Pattern

For many concurrent operations:

```go
type WorkerPool struct {
    workers  int
    tasks    chan Task
    results  chan Result
    stopChan chan struct{}
}

type Task struct {
    ID   int
    Data interface{}
}

type Result struct {
    TaskID int
    Value  interface{}
    Error  error
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    p := &WorkerPool{
        workers:  numWorkers,
        tasks:    make(chan Task, numWorkers*2),
        results:  make(chan Result),
        stopChan: make(chan struct{}),
    }
    
    for i := 0; i < numWorkers; i++ {
        go p.worker()
    }
    
    return p
}

func (p *WorkerPool) worker() {
    for {
        select {
        case task, ok := <-p.tasks:
            if !ok {
                return
            }
            // Process task
            result := Result{
                TaskID: task.ID,
                Value:  processTask(task.Data),
            }
            
            mainthread.Wait(func() {
                // Update UI with result
                _ = result
            })
            
        case <-p.stopChan:
            return
        }
    }
}

func (p *WorkerPool) Submit(task Task) {
    select {
    case p.tasks <- task:
    case <-p.stopChan:
    }
}

func (p *WorkerPool) Stop() {
    close(p.stopChan)
    close(p.tasks)
}

func processTask(data interface{}) interface{} {
    // CPU-intensive work
    return data
}
```

---

## Performance Optimization

### Binary Size Optimization

Default MIQT binaries are large. Reduce with:

```bash
# Build with strip and discard
go build -ldflags="-s -w" .           # 6MB

# Further compress with UPX
upx --best output                    # ~2MB
upx --lzma output                    # ~1.4MB
```

### Render Performance

**Limit Redraws:**
```go
// Batch updates to reduce paint events
model.BeginResetModel()
for i := 0; i < 1000; i++ {
    // Add items
}
model.EndResetModel()  // Single redraw, not 1000
```

**Viewport Optimization:**
```go
view := qt.NewQListView2()
// Set uniform item heights for efficient rendering
view.SetUniformItemSizes(true)

// Avoid animated transitions when adding many items
view.SetLayoutMode(qt.BatchLayoutMode)
```

**Custom Paint Optimization:**
```go
type OptimizedWidget struct {
    *qt.QWidget
    cachedPixmap *qt.QPixmap
    dirty        bool
}

func (w *OptimizedWidget) OnPaintEvent(event *qt.QPaintEvent) {
    if w.dirty {
        // Render to pixmap once
        w.cachedPixmap = qt.NewQPixmap2(w.Width(), w.Height())
        painter := qt.NewQPainter(w.cachedPixmap)
        w.renderContent(painter)
        painter.End()
        w.dirty = false
    }
    
    // Draw cached pixmap
    painter := qt.NewQPainter(w.QWidget)
    painter.DrawPixmap5(0, 0, w.cachedPixmap)
    painter.End()
}

func (w *OptimizedWidget) renderContent(painter *qt.QPainter) {
    // Custom drawing here
}
```

### Memory Management

**Avoid Circular References:**
```go
// Bad: closes reference prevents GC
widget.OnDestroyed(func() {
    _ = widget  // Circular reference
})

// Good: capture only needed state
ptr := widget
widget.OnDestroyed(func() {
    ptr.SetStyleSheet("")  // Use only if needed
})
```

**Explicit Cleanup:**
```go
// For temporary objects without parent
pixmap := qt.NewQPixmap2(width, height)
defer pixmap.Delete()

brush := qt.NewQBrush6(color)
defer brush.Delete()
```

---

## Styling & Theming

### Qt Stylesheets

Apply CSS-like styling to widgets:

```go
// Single widget styling
button := qt.NewQPushButton3("Styled Button")
button.SetStyleSheet(`
    QPushButton {
        background-color: #3498db;
        color: white;
        border-radius: 5px;
        padding: 8px 16px;
        font-weight: bold;
    }
    QPushButton:hover {
        background-color: #2980b9;
    }
    QPushButton:pressed {
        background-color: #1c5a8f;
    }
`)

// Application-wide stylesheet
app := qt.QApplication_Instance()
app.SetStyleSheet(`
    QWidget { background-color: #ecf0f1; }
    QPushButton { padding: 5px; }
    QLabel { color: #2c3e50; }
`)
```

### Theme System

```go
type Theme struct {
    PrimaryColor    *qt.QColor
    SecondaryColor  *qt.QColor
    TextColor       *qt.QColor
    BackgroundColor *qt.QColor
}

func NewDarkTheme() *Theme {
    return &Theme{
        PrimaryColor:    qt.NewQColor5(52, 73, 94),
        SecondaryColor:  qt.NewQColor5(44, 62, 80),
        TextColor:       qt.NewQColor5(236, 240, 241),
        BackgroundColor: qt.NewQColor5(25, 25, 25),
    }
}

func ApplyTheme(window *qt.QWidget, theme *Theme) {
    stylesheet := fmt.Sprintf(`
        QWidget {
            background-color: %s;
            color: %s;
        }
        QPushButton {
            background-color: %s;
            color: %s;
            border: none;
            padding: 6px 12px;
            border-radius: 3px;
        }
        QPushButton:hover {
            background-color: %s;
        }
    `,
        theme.BackgroundColor.Name(),
        theme.TextColor.Name(),
        theme.PrimaryColor.Name(),
        theme.TextColor.Name(),
        theme.SecondaryColor.Name(),
    )
    window.SetStyleSheet(stylesheet)
}
```

---

## QML Integration

### Loading QML Files

Embed Qt Quick UI in Go with QML:

```go
import (
    "fmt"
    "os"

    qt "github.com/mappu/miqt/qt6"
    qml "github.com/mappu/miqt/qt6/qml"
)

func main() {
    app := qt.NewQApplication(os.Args)

    engine := qml.NewQQmlApplicationEngine()
    
    // Load QML file from filesystem
    engine.Load(qml.NewQUrl2("qrc:/main.qml"))
    
    // Check for errors
    if len(engine.Errors()) > 0 {
        for _, err := range engine.Errors() {
            fmt.Printf("QML Error: %s\n", err.ToString())
        }
        os.Exit(1)
    }

    qt.QApplication_Exec()
}
```

### Exposing Go Objects to QML

```go
type DataModel struct {
    *qt.QObject
    data []*qt.QVariant
}

func NewDataModel() *DataModel {
    m := &DataModel{QObject: qt.NewQObject()}
    
    // Expose method to QML
    m.OnMetaCallEvent(func(arg *qt.QMetaCallEvent) {
        // Handle method calls from QML
    })
    
    return m
}

// In main:
engine := qml.NewQQmlApplicationEngine()
context := engine.RootContext()
model := NewDataModel()
context.SetContextProperty("goModel", model.QObject)

engine.Load(...)
```

---

## Advanced Patterns

### Custom Delegate Pattern

Render custom cells in item views:

```go
type CustomDelegate struct {
    *qt.QAbstractItemDelegate
}

func NewCustomDelegate() *CustomDelegate {
    d := &CustomDelegate{
        QAbstractItemDelegate: qt.NewQAbstractItemDelegate(),
    }
    
    d.OnPaint(func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, 
                   idx *qt.QModelIndex) {
        // Custom rendering for each item
        if !idx.IsValid() {
            return
        }
        
        painter.FillRect2(option.Rect(), option.Palette().Highlight())
        
        text := idx.Data(qt.DisplayRole).ToString()
        painter.DrawText4(
            option.Rect().X() + 5,
            option.Rect().Y() + option.Rect().Height()/2,
            text,
        )
    })
    
    d.OnSizeHint(func(option *qt.QStyleOptionViewItem, 
                      idx *qt.QModelIndex) *qt.QSize {
        return qt.NewQSize2(100, 30)
    })
    
    return d
}
```

### Context Menu Implementation

```go
widget.OnContextMenuEvent(func(event *qt.QContextMenuEvent) {
    menu := qt.NewQMenu(widget)
    defer menu.Delete()
    
    actionCut := menu.AddAction("Cut")
    actionCopy := menu.AddAction("Copy")
    actionPaste := menu.AddAction("Paste")
    
    menu.AddSeparator()
    actionSettings := menu.AddAction("Settings")
    
    // Execute menu at cursor position
    action := menu.Exec(event.GlobalPos(), nil)
    
    if action == actionCut {
        fmt.Println("Cut triggered")
    } else if action == actionCopy {
        fmt.Println("Copy triggered")
    } else if action == actionSettings {
        fmt.Println("Settings triggered")
    }
})
```

### File Dialog Integration

```go
button.OnClicked(func() {
    fileName := qt.QFileDialog_GetOpenFileName(
        window.QWidget,
        "Open File",
        "",
        "Text Files (*.txt);;All Files (*)",
        nil,
        qt.ReadOnly,
    )
    
    if fileName != "" {
        fmt.Printf("Selected: %s\n", fileName)
    }
})
```

### Async Task Queue

```go
type TaskQueue struct {
    tasks chan func()
    done  chan struct{}
}

func NewTaskQueue(workerCount int) *TaskQueue {
    tq := &TaskQueue{
        tasks: make(chan func(), 10),
        done:  make(chan struct{}),
    }
    
    for i := 0; i < workerCount; i++ {
        go func() {
            for task := range tq.tasks {
                task()
            }
        }()
    }
    
    return tq
}

func (tq *TaskQueue) Submit(task func()) {
    tq.tasks <- task
}

func (tq *TaskQueue) Stop() {
    close(tq.tasks)
    close(tq.done)
}
```

---

## Edge Cases & Troubleshooting

### Common Issues & Solutions

**1. Widget Not Appearing**
```go
// Common mistake: not calling Show()
widget := qt.NewQWidget2()
// Missing: widget.Show()

// Fix:
widget.Show()

// For QMainWindow, also set central widget
mainWindow := qt.NewQMainWindow2()
centralWidget := qt.NewQWidget(mainWindow.QWidget)
mainWindow.SetCentralWidget(centralWidget)
mainWindow.Show()
```

**2. Memory Leaks from Cycles**
```go
// Problematic: closure captures parent
parent := qt.NewQWidget2()
child := qt.NewQPushButton3("Click", parent)
child.OnClicked(func() {
    parent.Update()  // Keeps parent alive in closure
})

// Better: use weak reference or limit scope
parent := qt.NewQWidget2()
child := qt.NewQPushButton3("Click", parent)
// Parent already owns child; just update self
child.OnClicked(func() {
    child.SetText("Clicked!")
})
```

**3. Crashes from Double Delete**
```go
// Unsafe: parent owns child, explicit delete causes crash
parent := qt.NewQWidget2()
child := qt.NewQLabel(parent)
child.Delete()  // Parent will also delete
// Crash on parent destruction

// Safe: let parent manage lifetime
parent := qt.NewQWidget2()
child := qt.NewQLabel(parent)
// No explicit Delete; parent handles cleanup
```

**4. UI Freezes from Long Operations**
```go
// Bad: blocks UI thread
button.OnClicked(func() {
    time.Sleep(5 * time.Second)  // Freezes entire app
    label.SetText("Done")
})

// Good: use goroutine + mainthread marshal
button.OnClicked(func() {
    go func() {
        time.Sleep(5 * time.Second)
        mainthread.Wait(func() {
            label.SetText("Done")
        })
    }()
})
```

**5. Signal Slot Not Triggering**
```go
// Ensure callback is set AFTER widget creation
button := qt.NewQPushButton3("Click")

// This works
button.OnClicked(func() { fmt.Println("Clicked") })

// Common mistakes:
// - Callback set, then widget hidden: still fires
// - Multiple callbacks: only last one fires (overwrite previous)
// - Wrong signal type: OnPressed vs OnClicked

// If multiple handlers needed, wrap in custom handler:
var handlers []func()
button.OnClicked(func() {
    for _, h := range handlers {
        h()
    }
})

handlers = append(handlers, func() { fmt.Println("Handler 1") })
handlers = append(handlers, func() { fmt.Println("Handler 2") })
```

**6. QML Loading Issues**
```go
// qrc:/ paths require Qt resource system setup
// Alternatively, use file:// for filesystem paths
engine := qml.NewQQmlApplicationEngine()

// File path (requires .qml file to exist)
engine.Load(qml.NewQUrl2("file:///path/to/main.qml"))

// Or embed resources at compile time using Qt's rcc tool
```

**7. Platform-Specific Crashes**

- **Linux/Wayland**: Some decorations differ; test with `-platform wayland`
- **macOS**: Thread-safety issues; always use `mainthread.Wait()`
- **Windows**: DPI scaling can affect geometry; use logical pixels

### Debugging Tips

**1. Print Widget Tree**
```go
func printWidgetTree(w *qt.QWidget, depth int) {
    indent := strings.Repeat("  ", depth)
    fmt.Printf("%s%s (%dx%d @ %d,%d)\n",
        indent,
        w.MetaObject().ClassName(),
        w.Width(), w.Height(),
        w.X(), w.Y(),
    )
    
    // Iterate children
    children := w.Children()
    for i := 0; i < children.Length(); i++ {
        if child, ok := children.At(i).(*qt.QWidget); ok {
            printWidgetTree(child, depth+1)
        }
    }
}
```

**2. Monitor Signal Emissions**
```go
// Wrap callbacks to track calls
func TrackedCallback(name string, fn func()) func() {
    return func() {
        fmt.Printf("[SIGNAL] %s\n", name)
        fn()
    }
}

button.OnClicked(TrackedCallback("button.OnClicked", func() {
    label.SetText("Clicked")
}))
```

**3. Check Object Lifetimes**
```go
widget := qt.NewQWidget2()
widget.OnDestroyed(func() {
    fmt.Println("Widget destroyed!")
})

// Verify timely cleanup
// Widget should be destroyed when parent deleted or window closed
```

---

## Best Practices Summary

1. **Parent-Child Relationships**: Always establish clear ownership; let parents manage lifetimes
2. **Main Thread Safety**: Use `mainthread.Wait()` for all UI updates from goroutines
3. **Model Updates**: Batch operations with `Begin/EndInsert/RemoveRows` for efficiency
4. **Resource Cleanup**: Defer Delete() for temporary objects; rely on parent for long-lived ones
5. **Event Handling**: Prefer signal callbacks (`OnClicked`) over virtual method overrides when possible
6. **Threading**: Isolate CPU work in goroutines; marshal UI updates back to main thread
7. **Styling**: Use stylesheets for appearance; avoid hardcoded colors/fonts
8. **Error Handling**: Check for empty results from dialogs and model operations
9. **Performance**: Use layouts for dynamic UIs; cache pixmaps for expensive drawing
10. **Testing**: Test on target platforms early; Qt behavior varies across Windows/macOS/Linux

---

## Further Resources

- **MIQT Repository**: https://github.com/mappu/miqt
- **Qt Documentation**: https://doc.qt.io/qt-6/
- **Go-Qt Bindings**: https://pkg.go.dev/github.com/mappu/miqt/qt6
- **Example Applications**: https://github.com/mappu/miqt/tree/master/examples
