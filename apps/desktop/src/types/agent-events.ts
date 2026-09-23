/** Keep in sync with apps/agent-service/internal/zmq/events.go */

export const EventGetRecentActivity = 'GET_RECENT_ACTIVITY'

export interface FileRecord {
  ID: number
  Path: string
  Name: string
  Size: number
  ModifiedAt: string
  IsDirectory: boolean
  Inode: number
  Device: number
  Checksum: string
  RemoteID: string
  SyncStatus: string
  LastSeenAt: string
  CreatedAt: string
  UpdatedAt: string
}

export interface RecentActivityPayload {
  ok: boolean
  records: FileRecord[]
}
