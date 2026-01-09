import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Table, Button, Modal, Form, Input, Select, Space, Typography, Breadcrumb, Spin, Empty, Tag, message, Upload
} from 'antd'
import { PlusOutlined, SettingOutlined, KeyOutlined, FileTextOutlined, UploadOutlined, CopyOutlined } from '@ant-design/icons'
import { projectApi } from '../api/projects'
import type { Project, Config } from '../api/client'
import dayjs from 'dayjs'

const { Title, Text } = Typography
const { Dragger } = Upload

export default function ProjectDetailPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [configs, setConfigs] = useState<Config[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [form] = Form.useForm()

  const fetchData = async () => {
    if (!projectId) return
    try {
      const [projectRes, configsRes] = await Promise.all([
        projectApi.get(Number(projectId)),
        projectApi.listConfigs(Number(projectId)),
      ])
      setProject(projectRes.data.project)
      setConfigs(configsRes.data.configs || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [projectId])

  const handleCreate = async (values: { name: string; content: string; file_type: string; namespace?: string; environment?: string }) => {
    if (!projectId) return
    setCreating(true)
    try {
      await projectApi.uploadConfig(Number(projectId), values)
      setModalOpen(false)
      form.resetFields()
      fetchData()
      message.success('配置创建成功')
    } finally {
      setCreating(false)
    }
  }

  const handleFileUpload = (file: File) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const content = e.target?.result as string
      const fileType = file.name.endsWith('.yaml') || file.name.endsWith('.yml') ? 'yaml' : 'json'
      form.setFieldsValue({
        name: file.name.replace(/\.(json|yaml|yml)$/, ''),
        content,
        file_type: fileType,
      })
    }
    reader.readAsText(file)
    return false
  }

  const copyApiExample = (config: Config) => {
    const example = `curl -X GET "http://localhost:8080/api/v1/config?name=${config.name}" \\
  -H "X-Access-Key: YOUR_ACCESS_KEY"`
    navigator.clipboard.writeText(example)
    message.success('API 示例已复制')
  }

  const columns = [
    {
      title: '配置名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Config) => (
        <Space>
          <FileTextOutlined />
          <a onClick={() => navigate(`/configs/${record.id}`)}>{name}</a>
        </Space>
      ),
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
      render: (ns: string) => ns || '-',
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      render: (env: string) => {
        const envMap: Record<string, string> = {
          dev: 'blue',
          test: 'orange',
          staging: 'purple',
          prod: 'red',
        }
        return env ? <Tag color={envMap[env] || 'default'}>{env}</Tag> : '-'
      },
    },
    {
      title: '类型',
      dataIndex: 'file_type',
      key: 'file_type',
      render: (type: string) => <Tag>{type.toUpperCase()}</Tag>,
    },
    {
      title: '版本',
      dataIndex: 'current_version',
      key: 'current_version',
      render: (v: number) => `v${v}`,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Config) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/configs/${record.id}`)}>
            编辑
          </Button>
          <Button type="link" size="small" onClick={() => navigate(`/configs/${record.id}/versions`)}>
            历史
          </Button>
          <Button type="link" size="small" icon={<CopyOutlined />} onClick={() => copyApiExample(record)}>
            API
          </Button>
        </Space>
      ),
    },
  ]

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <a onClick={() => navigate('/projects')}>项目列表</a> },
          { title: project?.name || '项目详情' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Title level={4} style={{ margin: 0 }}>{project?.name}</Title>
            <Text type="secondary">{project?.description || '暂无描述'}</Text>
          </div>
          <Space>
            <Button icon={<KeyOutlined />} onClick={() => navigate(`/projects/${projectId}/keys`)}>
              密钥管理
            </Button>
            <Button icon={<SettingOutlined />} onClick={() => navigate(`/projects/${projectId}/settings`)}>
              设置
            </Button>
          </Space>
        </div>
      </Card>

      <Card
        title="配置列表"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            新建配置
          </Button>
        }
      >
        {configs.length === 0 ? (
          <Empty description="暂无配置">
            <Button type="primary" onClick={() => setModalOpen(true)}>创建第一个配置</Button>
          </Empty>
        ) : (
          <Table columns={columns} dataSource={configs} rowKey="id" pagination={false} />
        )}
      </Card>

      <Modal title="新建配置" open={modalOpen} onCancel={() => setModalOpen(false)} footer={null} width={600}>
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ file_type: 'json' }}>
          <Dragger beforeUpload={handleFileUpload} showUploadList={false} style={{ marginBottom: 16 }}>
            <p className="ant-upload-drag-icon"><UploadOutlined /></p>
            <p className="ant-upload-text">点击或拖拽文件上传</p>
            <p className="ant-upload-hint">支持 JSON、YAML 格式</p>
          </Dragger>
          <Form.Item name="name" label="配置名称" rules={[{ required: true, message: '请输入配置名称' }]}>
            <Input placeholder="如: app-config" />
          </Form.Item>
          <Form.Item name="file_type" label="文件类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'json', label: 'JSON' }, { value: 'yaml', label: 'YAML' }]} />
          </Form.Item>
          <Form.Item name="namespace" label="命名空间">
            <Input placeholder="可选，如: application" />
          </Form.Item>
          <Form.Item name="environment" label="环境">
            <Select allowClear placeholder="可选" options={[
              { value: 'dev', label: '开发环境' },
              { value: 'test', label: '测试环境' },
              { value: 'staging', label: '预发布环境' },
              { value: 'prod', label: '生产环境' },
            ]} />
          </Form.Item>
          <Form.Item name="content" label="配置内容" rules={[{ required: true, message: '请输入配置内容' }]}>
            <Input.TextArea rows={8} placeholder='{"key": "value"}' style={{ fontFamily: 'monospace' }} />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={creating}>创建</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
