# 🌌 Aura - Modern SSH Manager for macOS

[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-Donate-orange?style=flat-square&logo=buymeacoffee)](https://www.buymeacoffee.com/tissoni)

**Aura** is a high-performance SSH connection manager and TUI dashboard designed exclusively for macOS. It transforms the way you interact with your servers by combining native macOS security with a stunning dark aesthetic.

---

## ✨ Key Features

### 🖥️ **Aura Dashboard (Interactive TUI)**
Launch a full-screen interactive dashboard with `./aura dash`. Navigate your hosts, see real-time health status, connect with a single keypress, or **ping** servers instantly.
- **🔍 Advanced Fuzzy Search**: Find servers by alias, IP, or tags (e.g., `prod`, `web`, `db`).
- **🏓 Live Ping**: Press `p` in the dashboard to check ICMP latency to the selected host.

### 🩺 **Real-time Health Monitoring**
Stop guessing if your servers are up. Aura performs background TCP health checks and displays status indicators (🟢/🔴) in both the dashboard and host lists.

### 🔐 **Native macOS Security**
- **Keychain Integration**: Securely store passwords and private keys in the macOS System Keychain.
- **Biometric Auth**: Protect connections with **Touch ID** or system password fallback.
- **SSH Key Deployment**: Easily deploy your public keys to remote servers. Generate new keypairs (ED25519) or use existing ones and append them to `authorized_keys` automatically.

### 🎨 **Aura Eye Candy (Themes)**
Extreme personalization with dark-mode optimized themes:
- `aura` (Purple/Cyan)
- `deep-space` (Deep Black/Purple)
- `catppuccin` (Mocha)
- `tokyo-night` (Neon balanced)
- `nord` (Professional Frost)

### 🚀 **DevOps Snippets & SCP**
Execute common tasks instantly and transfer files securely. Includes snippets for **Docker**, **Kubernetes**, and `systemctl`.
- **📁 Interactive File Picker**: Run `aura scp` without arguments to launch a visual picker for local files and easy remote path selection.
- `./aura snippets run my-server docker-ps`
- `./aura scp local-file.txt my-server:/tmp/`

### 📢 **Smart Notifications & Monitoring**
Get native macOS notifications and real-time health indicators (🟢/🔴) for all your hosts.

### 📦 **Encrypted Backup & Restore**
Keep your configuration and secrets safe. Aura can export all data to an **AES-GCM** encrypted package with a master password of your choice.

### 🛠️ **Security Utilities**
Generate strong random passwords or memorable passphrases directly from the CLI.

---

## 🛠️ Installation

### Homebrew (Recommended)
```bash
brew install tissoni/tap/aura
```

### Manual / Build from source

#### Prerequisites
- macOS (Intel or Apple Silicon)
- `go 1.25+`

#### Build and Install
```bash
git clone https://github.com/tissoni/aura-ssh-manager.git
cd aura-ssh-manager
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
| `aura to <host>` | Connect to a specific host (uses Touch ID) |
| `aura scp` | **NEW**: Launch interactive file picker for transfers |
| `aura deploy-key` | **NEW**: Generate/deploy SSH keys to remote servers |
| `aura scp <src> <dst>` | Secure file transfer using Aura config |
| `aura backup [file]` | Create an encrypted backup of all data |
| `aura restore [file]` | Restore data from an encrypted backup |
| `aura snippets ls` | View saved DevOps command snippets |
| `aura utils gen-pwd` | Generate a strong random password |

---

## 💜 Credits
Aura is based on the excellent work of [sshed](https://github.com/trntv/sshed), but heavily modified and optimized to function **exclusively on macOS**, leveraging native features like Keychain, Touch ID, and System Notifications.

Developed with 💜 for DevOps and SREs who spend their lives in the terminal.
