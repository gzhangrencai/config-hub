import client, { Project, Config } from './client'

export const projectApi = {
  list: () => client.get<{ projects: Project[] }>('/projects'),
  
  get: (id: number) => client.get<{ project: Project }>(`/projects/${id}`),
  
  create: (data: { name: string; description?: string; access_mode?: string }) =>
    client.post<{ project: Project }>('/projects', data),
  
  update: (id: number, data: Partial<Project>) =>
    client.put<{ project: Project }>(`/projects/${id}`, data),
  
  delete: (id: number) => client.delete(`/projects/${id}`),
  
  listConfigs: (projectId: number) =>
    client.get<{ configs: Config[] }>(`/projects/${projectId}/configs`),
  
  uploadConfig: (projectId: number, data: {
    name: string
    content: string
    file_type?: string
    namespace?: string
    environment?: string
    message?: string
  }) => client.post<{ config: Config }>(`/projects/${projectId}/configs`, data),
}
