import client, { Config, ConfigVersion } from './client'

export const configApi = {
  get: (id: number) =>
    client.get<{ config: Config; content: string; version: number }>(`/configs/${id}`),

  update: (id: number, data: { content: string; message?: string }) =>
    client.put<{ version: ConfigVersion }>(`/configs/${id}`, data),

  delete: (id: number) => client.delete(`/configs/${id}`),

  // 版本管理
  listVersions: (configId: number) =>
    client.get<{ versions: ConfigVersion[] }>(`/configs/${configId}/versions`),

  getVersion: (configId: number, version: number) =>
    client.get<{ version: ConfigVersion }>(`/configs/${configId}/versions/${version}`),

  diff: (configId: number, v1: number, v2: number) =>
    client.get<{ diff: { changes: Array<{ type: string; path: string; old_value: unknown; new_value: unknown }> } }>(
      `/configs/${configId}/diff?v1=${v1}&v2=${v2}`
    ),

  rollback: (configId: number, version: number) =>
    client.post<{ version: ConfigVersion }>(`/configs/${configId}/rollback/${version}`),

  // Schema 管理
  getSchema: (configId: number) =>
    client.get<{ schema: string }>(`/configs/${configId}/schema`),

  updateSchema: (configId: number, schema: string) =>
    client.put(`/configs/${configId}/schema`, { schema }),

  generateSchema: (configId: number) =>
    client.post<{ schema: string }>(`/configs/${configId}/schema/generate`),

  // 发布管理
  release: (configId: number, data: { environment: string; version?: number }) =>
    client.post(`/configs/${configId}/release`, data),

  listReleases: (configId: number) =>
    client.get<{ releases: Array<{ id: number; version: number; environment: string; status: string; released_at: string }> }>(
      `/configs/${configId}/releases`
    ),

  // 灰度发布
  grayRelease: (configId: number, data: {
    environment: string
    version?: number
    rule_type: 'percentage' | 'client_id' | 'ip_range'
    percentage?: number
    client_ids?: string[]
    ip_ranges?: string[]
  }) => client.post(`/configs/${configId}/gray-release`, data),

  // 环境对比
  compare: (configId: number, source: string, target: string) =>
    client.get(`/configs/${configId}/compare?source=${source}&target=${target}`),

  sync: (configId: number, data: { source_env: string; target_env: string; keys?: string[] }) =>
    client.post(`/configs/${configId}/sync`, data),
}
