# calbridge
Calbridge is a tool that acts as a bridge between a CalDAV server and mail servers (SMTP, IMAP). It facilitates the synchronization of calendar events with email systems, enabling seamless communication and scheduling.

# Features
- Synchronize calendar events from CalDAV.
- Read calendar events from your caldav server and sends invitations using email.
- Read calendar invites from emails using IMAP and add those to your caldav server.
- Multi user support.
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
2. Or download and use the pre-built binaries from github release page of this repository.

# Usage
1. Invoke `calbridge` binary.
> [!NOTE]
> If it's your first time using calbridge, a sample config file will be created for you in your home directory. Update that with your caldav, smtp and imap details
