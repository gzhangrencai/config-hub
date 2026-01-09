import client, { AuditLog } from './client'

export interface AuditLogQuery {
  page?: number
  page_size?: number
  action?: string
  resource_type?: string
  start_time?: string
  end_time?: string
}

export const auditApi = {
  list: (projectId: number, query?: AuditLogQuery) =>
    client.get<{ logs: AuditLog[]; total: number }>(`/projects/${projectId}/audit-logs`, {
      params: query,
    }),

  export: (projectId: number, query?: AuditLogQuery) =>
    client.get(`/projects/${projectId}/audit-logs/export`, {
      params: query,
      responseType: 'blob',
    }),
}
