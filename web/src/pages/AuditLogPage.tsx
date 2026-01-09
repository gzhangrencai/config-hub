import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Table, Select, DatePicker, Space, Typography, Breadcrumb, Spin, Tag, Button } from 'antd'
import { DownloadOutlined } from '@ant-design/icons'
import { auditApi, AuditLogQuery } from '../api/audit'
import { projectApi } from '../api/projects'
import type { Project, AuditLog } from '../api/client'
import dayjs from 'dayjs'

const { Title } = Typography
const { RangePicker } = DatePicker

export default function AuditLogPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [query, setQuery] = useState<AuditLogQuery>({ page: 1, page_size: 20 })

  const fetchData = async () => {
    if (!projectId) return
    setLoading(true)
    try {
      const [projectRes, logsRes] = await Promise.all([
        projectApi.get(Number(projectId)),
        auditApi.list(Number(projectId), query),
      ])
      setProject(projectRes.data.project)
      setLogs(logsRes.data.logs || [])
      setTotal(logsRes.data.total || 0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [projectId, query])

  const handleExport = async () => {
    if (!projectId) return
    const response = await auditApi.export(Number(projectId), query)
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `audit-logs-${dayjs().format('YYYY-MM-DD')}.csv`)
    document.body.appendChild(link)
    link.click()
    link.remove()
  }

  const actionMap: Record<string, { color: string; text: string }> = {
    create: { color: 'green', text: '创建' },
    read: { color: 'blue', text: '读取' },
    update: { color: 'orange', text: '更新' },
    delete: { color: 'red', text: '删除' },
    release: { color: 'purple', text: '发布' },
    rollback: { color: 'cyan', text: '回滚' },
    gray_release: { color: 'magenta', text: '灰度发布' },
  }

  const resourceTypeMap: Record<string, string> = {
    project: '项目',
    config: '配置',
    key: '密钥',
    release: '发布',
  }

  const columns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: string) => (
        <Tag color={actionMap[action]?.color || 'default'}>
          {actionMap[action]?.text || action}
        </Tag>
      ),
    },
    {
      title: '资源类型',
      dataIndex: 'resource_type',
      key: 'resource_type',
      width: 100,
      render: (type: string) => resourceTypeMap[type] || type,
    },
    {
      title: '资源名称',
      dataIndex: 'resource_name',
      key: 'resource_name',
      render: (name: string) => name || '-',
    },
    {
      title: '操作者',
      key: 'operator',
      width: 120,
      render: (_: unknown, record: AuditLog) => {
        if (record.user_id) return `用户 #${record.user_id}`
        if (record.access_key_id) return `密钥 #${record.access_key_id}`
        return '-'
      },
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 140,
    },
  ]

  if (loading && logs.length === 0) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <a onClick={() => navigate('/projects')}>项目列表</a> },
          { title: <a onClick={() => navigate(`/projects/${projectId}`)}>{project?.name}</a> },
          { title: '审计日志' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        title="审计日志"
        extra={
          <Button icon={<DownloadOutlined />} onClick={handleExport}>
            导出
          </Button>
        }
      >
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="操作类型"
            allowClear
            style={{ width: 120 }}
            onChange={(value) => setQuery({ ...query, action: value, page: 1 })}
            options={[
              { value: 'create', label: '创建' },
              { value: 'read', label: '读取' },
              { value: 'update', label: '更新' },
              { value: 'delete', label: '删除' },
              { value: 'release', label: '发布' },
            ]}
          />
          <Select
            placeholder="资源类型"
            allowClear
            style={{ width: 120 }}
            onChange={(value) => setQuery({ ...query, resource_type: value, page: 1 })}
            options={[
              { value: 'project', label: '项目' },
              { value: 'config', label: '配置' },
              { value: 'key', label: '密钥' },
              { value: 'release', label: '发布' },
            ]}
          />
          <RangePicker
            onChange={(dates) => {
              if (dates) {
                setQuery({
                  ...query,
                  start_time: dates[0]?.toISOString(),
                  end_time: dates[1]?.toISOString(),
                  page: 1,
                })
              } else {
                setQuery({ ...query, start_time: undefined, end_time: undefined, page: 1 })
              }
            }}
          />
        </Space>

        <Table
          columns={columns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            current: query.page,
            pageSize: query.page_size,
            total,
            onChange: (page, pageSize) => setQuery({ ...query, page, page_size: pageSize }),
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>
    </div>
  )
}
