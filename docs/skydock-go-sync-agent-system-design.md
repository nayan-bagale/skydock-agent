# SkyDock Folder Sync Agent - Go Roadmap & System Design

## 1. Project Goal

Build a **Go-based folder synchronization agent** for SkyDock, similar in concept to Google Drive.

The agent will run in the background, watch a local SkyDock folder, synchronize changes with the SkyDock backend, maintain local sync state, handle retries and conflicts, and communicate with the SkyDock desktop application.

### High-level architecture

```text
Local SkyDock Folder
        │
        │ filesystem changes
        ▼
┌──────────────────────┐
│   SkyDock Go Agent   │
│                      │
│ watcher              │
│ sync engine          │
│ local state DB       │
│ upload/download      │
│ conflict resolver    │
│ retry queue          │
└──────────┬───────────┘
           │
           │ HTTPS
           ▼
┌──────────────────────┐
│   SkyDock Backend    │
│                      │
│ file metadata        │
│ object storage       │
│ sync API             │
└──────────────────────┘
```

Eventually:

```text
                    SkyDock Cloud
                         │
              ┌──────────┴──────────┐
              │                     │
          Web Client             Backend
                                    │
                               ┌────┴────┐
                               │   S3    │
                               └─────────┘
                                    ▲
                                    │
                              HTTPS/WebSocket
                                    │
                          ┌─────────┴─────────┐
                          │                   │
                    Mac Sync Agent      Windows Agent
                          │                   │
                      ~/SkyDock            C:\SkyDock
```

---

# 2. Don't Start With "Sync"

A naive implementation might be:

```text
File changed
    ↓
Upload file
```

That breaks quickly.

A real sync system needs to answer:

- Was the change local or remote?
- Is this a new file?
- Was it modified?
- Was it deleted?
- Was it renamed?
- Did the same file change on another computer?
- What happens if the network disappears?
- What happens if the agent crashes halfway through?
- What happens with a very large file?
- What happens if thousands of files change simultaneously?

Build the system incrementally.

---

# 3. Core Agent Architecture

Recommended components:

```text
                    ┌────────────────────────────┐
                    │        Sync Agent          │
                    │                            │
Filesystem ────────▶│ ┌────────────────────────┐ │
                    │ │   File Watcher         │ │
                    │ └──────────┬─────────────┘ │
                    │            ▼               │
                    │ ┌────────────────────────┐ │
                    │ │ Change Detector         │ │
                    │ └──────────┬─────────────┘ │
                    │            ▼               │
                    │ ┌────────────────────────┐ │
                    │ │ Sync Queue              │ │
                    │ └──────────┬─────────────┘ │
                    │            ▼               │
                    │ ┌────────────────────────┐ │
                    │ │ Sync Engine             │ │
                    │ └───────┬────────┬────────┘ │
                    │         │        │          │
                    │         ▼        ▼          │
                    │      Upload   Download      │
                    │         │        │           │
                    │         └───┬────┘           │
                    │             ▼                │
                    │      Remote API              │
                    │                              │
                    │ ┌────────────────────────┐   │
                    │ │ Local State DB          │   │
                    │ └────────────────────────┘   │
                    └──────────────────────────────┘
```

---

# 4. Local Folder

Suppose the user selects:

```text
~/SkyDock
```

The folder might contain:

```text
~/SkyDock
├── Documents/
├── Photos/
├── Music/
└── project.zip
```

The agent's internal state should **not** live inside the synced folder.

Use a separate application-data directory:

```text
~/.skydock/
├── agent.db
├── logs/
├── cache/
├── temp/
└── config.json
```

On Windows, use the appropriate application-data directory rather than hard-coding a Unix path.

### Why?

If the database is inside the sync folder, the agent could start syncing its own state.

That creates a feedback loop:

```text
agent.db changes
    ↓
watcher detects change
    ↓
upload agent.db
    ↓
remote change
    ↓
download agent.db
    ↓
agent.db changes
    ↓
...
```

---

# 5. Local State Database

Use **SQLite** for the agent's persistent state.

The database should remember what the agent knows about each file and what synchronization operations are pending.

A starting schema:

```sql
CREATE TABLE files (
    id TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    size BIGINT NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    checksum TEXT,
    remote_version INTEGER,
    is_directory BOOLEAN NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT FALSE
);
```

Eventually, separate the concepts more cleanly.

### Files

```text
files
-----------------------------------------
file_id
relative_path
type
size
mtime
checksum
remote_version
local_version
deleted
```

### Sync operations

```text
sync_operations
-----------------------------------------
id
file_id
operation
status
retry_count
created_at
updated_at
```

### Sync roots

```text
sync_roots
-----------------------------------------
id
local_path
remote_folder_id
enabled
```

The database becomes the agent's persistent memory.

---

# 6. Filesystem Watcher

Do not continuously scan the entire folder just to detect changes.

Use a filesystem watcher.

Flow:

```text
User changes file
       │
       ▼
OS filesystem event
       │
       ▼
Go watcher
       │
       ▼
Normalize event
       │
       ▼
Sync queue
```

Typical event types:

```text
CREATE
WRITE
REMOVE
RENAME
```

However, one user action can produce many events:

```text
WRITE
WRITE
WRITE
WRITE
```

Therefore, don't immediately upload on every event.

---

# 7. Debouncing

Implement a debounce mechanism.

Example:

```text
14:02:01 WRITE
14:02:01 WRITE
14:02:02 WRITE
14:02:02 WRITE

          ↓ debounce

14:02:04

UPLOAD FILE
```

Conceptually:

```text
filesystem event
       │
       ▼
   debounce map
       │
       ▼
wait until quiet
       │
       ▼
queue sync task
```

Example model:

```go
type PendingChange struct {
    Path       string
    LastEvent  time.Time
}
```

You can later improve this into a centralized scheduler rather than creating an uncontrolled number of timers.

---

# 8. Never Trust Filesystem Events Alone

Filesystem watchers are useful but should not be treated as the only source of truth.

Events can be missed because of:

- application crashes
- watcher limitations
- OS behavior
- large bursts of changes
- temporary watcher failures

Use two mechanisms:

```text
REALTIME

filesystem event
      ↓
quick sync
```

and:

```text
PERIODIC

reconciliation scan
      ↓
compare filesystem
      ↓
compare local state DB
      ↓
repair inconsistencies
```

For example, perform a reconciliation scan every few minutes.

This gives you:

> Event-driven sync + periodic reconciliation

That combination is much more robust.

---

# 9. Upload Pipeline

When a local file changes:

```text
File changed
    ↓
Debounce
    ↓
Calculate metadata
    ↓
Calculate checksum
    ↓
Create sync operation
    ↓
Upload
    ↓
Update DB
    ↓
Mark synced
```

Conceptually:

```text
                    Upload
                      │
                      ▼
               ┌─────────────┐
               │ Read file   │
               └──────┬──────┘
                      ▼
                checksum
                      │
                      ▼
               Upload object
                      │
                      ▼
             update metadata
                      │
                      ▼
             mark operation done
```

---

# 10. File Hashing

For large files, calculate hashes as a stream rather than loading the entire file into memory.

Example:

```go
hash := sha256.New()

_, err := io.Copy(hash, file)
```

You can also avoid unnecessary hashing.

A basic strategy:

```text
size changed?
    YES → content changed

mtime changed?
    YES → maybe changed → hash

neither changed?
    probably unchanged
```

The reconciliation algorithm can become more sophisticated later.

---

# 11. Remote Metadata API

The backend needs a remote representation of the file tree.

For example:

```http
GET /sync/state
```

Possible response:

```json
{
  "cursor": "abc123",
  "files": [
    {
      "id": "f123",
      "path": "Documents/test.pdf",
      "size": 12043,
      "hash": "...",
      "version": 14
    }
  ]
}
```

However, don't repeatedly download the complete file tree.

Eventually, use incremental synchronization.

For example:

```http
GET /sync/changes?cursor=abc123
```

Response:

```json
{
  "next_cursor": "abc999",
  "changes": [
    {
      "type": "modified",
      "file_id": "f123"
    },
    {
      "type": "deleted",
      "file_id": "f456"
    }
  ]
}
```

The cursor-based approach is one of the most important design decisions in the system.

---

# 12. Sync Algorithm

The agent needs to reason about **three states**:

```text
             ┌─────────┐
             │ Previous│
             │  state  │
             └────┬────┘
                  │
          ┌───────┴────────┐
          ▼                ▼
       Local              Remote
       state              state
```

Or conceptually:

```text
             Local
               │
               ▼
         Local State DB
               ▲
               │
             Remote
```

The engine needs to determine:

```text
Local changed?
Remote changed?
Both changed?
Neither changed?
```

### Basic decision matrix

| Local | Remote | Action |
|---|---|---|
| No | No | Nothing |
| Yes | No | Upload |
| No | Yes | Download |
| Yes | Yes | Conflict resolution |
| Deleted | No | Delete remotely |
| No | Deleted | Delete locally |

This decision process is the heart of the sync engine.

---

# 13. Conflicts

Example:

Computer A has:

```text
report.txt
"Hello Nayan"
```

Computer B changes the same file:

```text
report.txt
"Hello World"
```

Both changes reach the server.

The agent must not silently overwrite one version.

A simple first implementation can create a conflict copy:

```text
report.txt
report (Nayan's Mac).conflicted.txt
```

Later, implement smarter conflict handling.

Possible future strategies:

- last-write-wins
- explicit conflict copies
- version history
- text merge for supported file types
- user-selected resolution

For a first version, **preserving both versions is safer than silently losing data**.

---

# 14. Rename Detection

Suppose:

```text
hello.txt
```

becomes:

```text
hello-world.txt
```

The watcher might report:

```text
REMOVE hello.txt
CREATE hello-world.txt
```

You need to distinguish:

```text
delete + new file
```

from:

```text
rename
```

A basic fallback strategy:

```text
old file checksum
        │
        ▼
new file checksum
        │
        ▼
same content?
        │
       YES
        │
        ▼
probably rename
```

Where possible, use the native rename information supplied by the filesystem watcher.

---

# 15. Prevent Sync Loops

This is one of the most important problems.

Suppose a remote file is downloaded:

```text
Server
  ↓
Agent
  ↓
write local file
```

The filesystem watcher sees:

```text
FILE CREATED!
```

The agent may then incorrectly decide:

```text
UPLOAD IT!
```

You can get:

```text
download
   ↓
watch event
   ↓
upload
   ↓
download
   ↓
upload
   ↓
...
```

### Solution: track event origin

For example:

```go
type ChangeSource int

const (
    SourceUnknown ChangeSource = iota
    SourceUser
    SourceSync
)
```

When the agent writes a remote file:

```text
Agent writing file
       ↓
mark expected filesystem event
       ↓
write file
       ↓
watcher event arrives
       ↓
ignore expected agent-originated event
```

This needs careful implementation because filesystem events are asynchronous and can arrive later than expected.

---

# 16. Download Pipeline

Downloads should use temporary files.

Do not directly write into the final file:

```text
~/SkyDock/movie.mp4
```

Instead:

```text
~/SkyDock/.skydock-temp/movie.mp4.part
```

Then:

```text
download complete
       ↓
checksum valid
       ↓
atomic rename
       ↓
movie.mp4
```

This prevents users from seeing partially downloaded files.

The flow becomes:

```text
Remote change
      ↓
Download queue
      ↓
Temporary file
      ↓
Verify checksum
      ↓
Atomic rename
      ↓
Update DB
```

---

# 17. Large File Support

Do not design the agent around uploading an entire large file as one operation.

For example, a 5 GB file should not become:

```text
5 GB file
   ↓
RAM
   ↓
upload
```

Use chunks:

```text
File
 │
 ├── chunk 1
 ├── chunk 2
 ├── chunk 3
 ├── chunk 4
 └── ...
```

For example:

```text
10 MB chunks
```

Then:

```text
Upload chunk 1
Upload chunk 2
Upload chunk 3
...
```

If the network fails at chunk 348:

```text
resume from chunk 349
```

rather than restarting the entire upload.

This provides resumable uploads.

---

# 18. Concurrency Architecture

Go is particularly useful here.

Eventually, the agent might have:

```text
                   Event Dispatcher
                         │
              ┌──────────┼───────────┐
              ▼          ▼           ▼
           Upload     Download    Delete
            Queue       Queue       Queue
              │          │           │
          ┌───┴───┐   ┌──┴───┐   ┌───┴───┐
          ▼       ▼   ▼      ▼   ▼       ▼
        Worker Worker Worker Worker Worker Worker
```

Do not create unlimited goroutines.

Use bounded worker pools.

Example:

```text
Upload workers:   3
Download workers: 3
Hash workers:     2
```

Tune these values later based on real workloads.

---

# 19. Suggested Go Project Structure

Start with a structure like:

```text
skydock-agent/
│
├── cmd/
│   └── skydock-agent/
│       └── main.go
│
├── internal/
│
│   ├── agent/
│   │   ├── agent.go
│   │   └── lifecycle.go
│   │
│   ├── watcher/
│   │   ├── watcher.go
│   │   └── events.go
│   │
│   ├── sync/
│   │   ├── engine.go
│   │   ├── planner.go
│   │   ├── conflict.go
│   │   └── reconcile.go
│   │
│   ├── queue/
│   │   ├── queue.go
│   │   └── worker.go
│   │
│   ├── storage/
│   │   ├── sqlite.go
│   │   ├── files.go
│   │   └── operations.go
│   │
│   ├── api/
│   │   ├── client.go
│   │   ├── files.go
│   │   └── sync.go
│   │
│   ├── transfer/
│   │   ├── upload.go
│   │   ├── download.go
│   │   └── chunks.go
│   │
│   ├── filesystem/
│   │   ├── filesystem.go
│   │   ├── darwin.go
│   │   ├── linux.go
│   │   └── windows.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   └── logging/
│       └── logging.go
│
├── migrations/
├── go.mod
└── README.md
```

Do **not** create all these packages on day one. Let the architecture grow as responsibilities become clear.

---

# 20. Development Milestones

## Milestone 1: Filesystem Watcher

Build:

```text
Go Agent
   ↓
watch ~/SkyDock
   ↓
print filesystem events
```

Example:

```text
[CREATE] Documents/test.txt
[WRITE]  Documents/test.txt
[DELETE] Documents/old.txt
```

No cloud synchronization yet.

### Learn

- filesystem APIs
- goroutines
- channels
- context cancellation
- graceful shutdown

---

## Milestone 2: Local State

Add:

```text
Filesystem
    ↓
Watcher
    ↓
SQLite
```

Now the agent knows:

```text
What exists
What changed
What was last synchronized
```

### Learn

- SQLite
- transactions
- database schema design
- persistence

---

## Milestone 3: Upload

Add:

```text
local file
   ↓
hash
   ↓
POST metadata
   ↓
upload to S3
```

The backend should store metadata separately from the actual file object.

---

## Milestone 4: Download

Add:

```text
Server changes
      ↓
Agent polls
      ↓
Download
      ↓
Write temporary file
      ↓
Verify
      ↓
Atomic rename
```

---

## Milestone 5: Bidirectional Sync

Build:

```text
       ┌───────────────┐
       │ Sync Engine   │
       └───────┬───────┘
          ┌────┴────┐
          ▼         ▼
       Local      Remote
```

At this point, the agent should be capable of basic two-way synchronization.

---

## Milestone 6: Reliability

Add:

```text
debouncing
retries
offline queue
reconciliation
conflict detection
```

The agent should continue operating when the network disappears.

Example:

```text
Internet available
       ↓
     Sync

Internet disappears
       ↓
Store operations locally

Internet returns
       ↓
Retry pending operations
```

---

## Milestone 7: Large Files

Add:

```text
chunked uploads
chunked downloads
resume
parallel transfers
```

---

## Milestone 8: Background Service

Turn the agent into a real OS background service:

```text
macOS   → launchd
Linux   → systemd
Windows → Windows Service
```

---

## Milestone 9: Electron Integration

The architecture should be:

```text
Electron
  = UI

Go Agent
  = synchronization brain

Backend
  = cloud metadata + storage
```

Electron communicates with the agent through local IPC.

For example:

```text
Electron
   │
   │ IPC
   ▼
Go Agent
```

The Electron UI can show:

```text
┌────────────────────────────────────────────┐
│              Agent Dashboard               │
├────────────────────────────────────────────┤
│                                            │
│  Agent: MacBook                            │
│  Status: ● Online                          │
│  Version: 1.2.0                            │
│                                            │
│  CPU       ███████░░░  72%                 │
│  Memory    █████░░░░░  51%                 │
│  Disk      ████████░░  81%                 │
│                                            │
│  [Refresh] [Restart Agent] [Update]        │
│                                            │
└────────────────────────────────────────────┘
```

---

# 21. Final Architecture

The eventual SkyDock architecture could look like this:

```text
                         SkyDock Cloud
                              │
                    ┌─────────┴──────────┐
                    │                    │
              Metadata API             S3
                    │                    │
                    └─────────┬──────────┘
                              │
                             HTTPS
                              │
                 ┌────────────▼────────────┐
                 │      Go Sync Agent      │
                 │                         │
                 │  ┌───────────────────┐  │
                 │  │ Filesystem Watcher│  │
                 │  └─────────┬─────────┘  │
                 │            ▼            │
                 │  ┌───────────────────┐  │
                 │  │ Change Detector   │  │
                 │  └─────────┬─────────┘  │
                 │            ▼            │
                 │  ┌───────────────────┐  │
                 │  │ Sync Planner      │  │
                 │  └──────┬───────┬────┘  │
                 │         │       │        │
                 │         ▼       ▼        │
                 │      Upload  Download    │
                 │         │       │        │
                 │         └───┬───┘        │
                 │             ▼            │
                 │       Transfer Manager   │
                 │                          │
                 │  ┌───────────────────┐   │
                 │  │ SQLite State DB   │   │
                 │  └───────────────────┘   │
                 │                          │
                 └────────────┬─────────────┘
                              │
                              │ IPC
                              ▼
                    ┌──────────────────┐
                    │ SkyDock Desktop  │
                    │ Electron UI      │
                    └──────────────────┘
```

---

# 22. Recommended Learning Roadmap

Build and learn in this order:

```text
Go Basics
   │
   ▼
Filesystem Watcher
   │
   ▼
Context + Goroutines
   │
   ▼
Channels + Worker Pool
   │
   ▼
Debouncing
   │
   ▼
SQLite / Persistence
   │
   ▼
HTTP Client
   │
   ▼
Upload
   │
   ▼
Download
   │
   ▼
Remote Change Detection
   │
   ▼
Bidirectional Sync
   │
   ▼
Conflict Resolution
   │
   ▼
Retry + Offline Queue
   │
   ▼
Reconciliation
   │
   ▼
Chunked Transfers
   │
   ▼
IPC
   │
   ▼
OS Background Service
   │
   ▼
Production Sync Agent
```

---

# 23. First Version to Build

Do **not** attempt the entire architecture immediately.

The first reliable version should only do:

```text
~/SkyDock
   ↓
filesystem watcher
   ↓
debounced event queue
   ↓
SQLite state
   ↓
upload changed file
```

Once that works reliably, add the reverse direction:

```text
Server
   ↓
remote changes
   ↓
download
   ↓
local file
```

Then add conflict handling, retries, reconciliation, chunking, IPC, and background-service support.

This keeps the project manageable while still teaching the important parts of Go and distributed systems.

---

# 24. Key Design Principle

The clean separation should be:

```text
┌──────────────────────────────┐
│ Electron                     │
│                              │
│ UI / Settings / Status       │
└──────────────┬───────────────┘
               │
               │ IPC
               ▼
┌──────────────────────────────┐
│ Go Sync Agent                │
│                              │
│ Filesystem                   │
│ Sync Engine                  │
│ Queue                        │
│ SQLite                       │
│ Transfers                    │
│ Retry / Reconciliation       │
└──────────────┬───────────────┘
               │
               │ HTTPS
               ▼
┌──────────────────────────────┐
│ SkyDock Backend              │
│                              │
│ Auth                         │
│ Metadata                     │
│ Sync API                     │
│ Object Storage               │
└──────────────────────────────┘
```

**Electron should be the UI. The Go agent should be the synchronization brain. The backend should be the cloud source of truth and object-storage layer.**

This also means:

```text
Electron crashes
     ↓
Go agent keeps syncing
     ↓
Electron starts again
     ↓
UI reconnects to agent
```

That is a much stronger architecture for a Google Drive-style desktop sync application.
