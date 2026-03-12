# 🌌 Aura - Modern SSH Manager for macOS

**Aura** is a premium, high-performance SSH connection manager and TUI dashboard designed exclusively for macOS. It transforms the way you interact with your servers by combining native macOS security with a stunning neon-dark aesthetic.

---

## ✨ Key Features

### 🖥️ **Aura Dashboard (Interactive TUI)**
Launch a full-screen interactive dashboard with `./aura dash`. Navigate your hosts, see real-time health status, and connect with a single keypress.

### 🩺 **Real-time Health Monitoring**
Stop guessing if your servers are up. Aura performs background TCP health checks and displays status indicators (🟢/🔴) in both the dashboard and host lists.

### 🔐 **Native macOS Security**
- **Keychain Integration**: Securely store passwords and private keys in the macOS System Keychain.
- **Biometric Auth**: Protect connections with **Touch ID** or system password fallback.
- **Zero Dependencies**: Pure Go implementation using native PTYs—no more `sshpass`.

### 🎨 **Aura Eye Candy (Themes)**
Extreme personalization with dark-mode optimized themes:
- `aura` (Premium Purple/Cyan)
- `deep-space` (Deep Black/Purple)
- `catppuccin` (Mocha)
- `tokyo-night` (Neon balanced)
- `nord` (Professional Frost)

### 🚀 **DevOps Snippets**
Execute common DevOps tasks instantly. Includes pre-loaded snippets for **Docker**, **Kubernetes**, and `systemctl`.
- `./aura snippets run my-server docker-ps`

### 📢 **Smart Notifications**
Get native macOS notifications with high-quality audio feedback when batch tasks (`aura at @tag`) are completed.

---

## 🛠️ Installation

### Prerequisites
- macOS (Intel or Apple Silicon)
- `go 1.14+` (for building from source)

### Build and Install
```bash
git clone https://github.com/chelo/aura.git
cd aura
go build -o aura ./cmd/cmd.go
sudo cp ./aura /usr/local/bin/aura
```

---

## 📖 Basic Usage

| Command | Description |
| --- | --- |
| `aura dash` | Launch the interactive TUI Dashboard |
| `aura ls` | List all hosts with health status indicators |
| `aura add` | Interactive flow to add/edit a host |
| `aura to <host>` | Connect to a specific host (uses Touch ID if saved) |
| `aura at @tag "ls"` | Execute a command across multiple hosts by tag |
| `aura theme <name>` | Change the global aesthetic |
| `aura snippets ls` | View saved DevOps command snippets |

---

## 💜 Credits
Aura is a modern fork and complete rewrite of `sshed`, optimized for the macOS ecosystem.

Developed with 💜 for DevOps and SREs who spend their lives in the terminal.
