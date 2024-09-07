# calbridge
Calbridge is a tool that acts as a bridge between a CalDAV server and mail servers (SMTP, IMAP). It facilitates the synchronization of calendar events with email systems, enabling seamless communication and scheduling.

# Features
- Synchronize calendar events from CalDAV.
- Send calendar invites through SMTP.
- Read calendar invites from IMAP.
- Multi user support
- (Planned) Run continuously
- (Planned) Handle all users concurrently


# Installation
1. Build and run locally. Script below puts the binary in your `/tmp` directory. Move it to any of your PATH dirs.
    ```sh
    git clone https://github.com/yourusername/calbridge.git
    cd calbridge
    go mod tidy
    go build -o /tmp/calbridge cmd/main.go
    ```
2. (Coming soon) download from github release page of this repository

# Usage
1. Using `config.example.json` file, create a `config.json` file. Make sure to provide correct connection details of your caldav, smtp and imap.
2. Invoke `calbridge` binary
