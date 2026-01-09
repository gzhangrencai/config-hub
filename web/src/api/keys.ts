import client, { ProjectKey } from './client'

export interface CreateKeyRequest {
  name: string
  permissions: {
    read: boolean
    write: boolean
    delete: boolean
    release: boolean
    admin: boolean
    decrypt: boolean
  }
  ip_whitelist?: string[]
  expires_at?: string
}

export const keyApi = {
  list: (projectId: number) =>
    client.get<{ keys: ProjectKey[] }>(`/projects/${projectId}/keys`),

  create: (projectId: number, data: CreateKeyRequest) =>
    client.post<{ key: ProjectKey; secret_key: string }>(`/projects/${projectId}/keys`, data),

  update: (keyId: number, data: Partial<CreateKeyRequest>) =>
    client.put<{ key: ProjectKey }>(`/keys/${keyId}`, data),

  delete: (keyId: number) => client.delete(`/keys/${keyId}`),

  regenerate: (keyId: number) =>
    client.post<{ key: ProjectKey; secret_key: string }>(`/keys/${keyId}/regenerate`),
}
