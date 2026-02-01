# RSMQT Project Instructions

## Project Overview
RSMQT (Redis Simple Message Queue Terminal) is a desktop GUI application for managing RSMQ (Redis Simple Message Queue) instances. It is built using Go and the MIQT (Qt bindings) library.

## Tech Stack
- **Language**: Go 1.24+
- **GUI Framework**: Qt 6 via [MIQT](https://github.com/mappu/miqt)
- **Redis Client**: `go-redis`
- **SSH**: `golang.org/x/crypto/ssh`

## Build & Development
**CRITICAL**: Always use `make` to build the project. Do **NOT** use `go build` directly.
- The project requires specific CGO flags (`CGO_CXXFLAGS`) and linker flags to compile correctly with the C++ Qt bindings.
- **Build Command**: `make`
- **Output**: `build/rsmqt`

## Architecture
### 1. User Interface (`main.go`)
- Contains the entire UI implementation including `RSMQTMainWindow`, `ConnectWindow`, and various dialogs (`QueueDialog`, `SendMessageDialog`).
- **MIQT Patterns**:
    - Widgets are wrapped in Go structs embedding the MIQT type (e.g., `type ConnectWindow struct { *qt.QWidget ... }`).
    - Signals are handled via Go closures (e.g., `btn.OnClicked(func() { ... })`).
    - **Note**: Pay close attention to MIQT constructor signatures (e.g., `NewQPushButton3` vs `NewQPushButton`). Use the `miqt-developer-guide` skill for reference.

### 2. Backend Library (`lib/rsmq/`)
- **`rsmq.go`**: 
    - Wraps `go-redis` to implement the RSMQ protocol.
    - **Strict Compliance**: The ID generation and parsing logic must match the standard Node.js RSMQ implementation (Base36 microsecond timestamp + 22 random chars).
    - `QueueStats` and `Message` structs use `time.Time` for timestamp fields.
- **`ssh.go`**:
    - Implements SSH tunneling logic.
    - Provides `DialSSH` to create a `net.Conn` dialer function that routes Redis traffic through an SSH tunnel.
    - Supports Password, Private Key, and **Encrypted Private Key** (via interactive passphrase prompt) authentication.

## Implemented Features
1.  **Connection Manager**:
    - Connect to Redis instances.
    - **SSH Tunneling**: Support for connecting via SSH jump hosts with password or key-based auth.
2.  **Queue Management**:
    - List queues.
    - Create new queues (configurable VT, Delay, MaxSize).
    - Delete queues.
    - **Clear Queue**: Removes all messages/stats without deleting the queue configuration.
3.  **Message Management**:
    - List messages in a table (ID, Sent, Visible, RC, Body).
    - Send new messages.
    - Delete individual messages.
4.  **Real-time Stats**:
    - View queue attributes (Hidden messages, Total sent/recv).

## Coding Guidelines
- **UI Changes**: When modifying `main.go`, ensure signal handlers are thread-safe (MIQT signals run on the main thread).
- **Destructive Actions**: Always wrap destructive actions (Delete/Clear) in a `QMessageBox_Question` confirmation dialog.
- **Planning**: For complex features, use "Plan Mode" (`[[PLAN]]`) to draft a `plan.md` before implementation.
