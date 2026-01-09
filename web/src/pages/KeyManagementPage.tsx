import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Table, Button, Modal, Form, Input, Checkbox, Space, Typography, Breadcrumb, Spin, Tag, message, DatePicker, Popconfirm, Alert
} from 'antd'
import { PlusOutlined, DeleteOutlined, ReloadOutlined, CopyOutlined } from '@ant-design/icons'
import { keyApi, CreateKeyRequest } from '../api/keys'
import { projectApi } from '../api/projects'
import type { Project, ProjectKey } from '../api/client'
import dayjs from 'dayjs'

const { Text, Paragraph } = Typography

export default function KeyManagementPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [keys, setKeys] = useState<ProjectKey[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [newSecretKey, setNewSecretKey] = useState<string | null>(null)
  const [form] = Form.useForm()

  const fetchData = async () => {
    if (!projectId) return
    try {
      const [projectRes, keysRes] = await Promise.all([
        projectApi.get(Number(projectId)),
        keyApi.list(Number(projectId)),
      ])
      setProject(projectRes.data.project)
      setKeys(keysRes.data.keys || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [projectId])

  const handleCreate = async (values: { name: string; permissions: string[]; expires_at?: dayjs.Dayjs }) => {
    if (!projectId) return
    setCreating(true)
    try {
      const permissions = {
        read: values.permissions.includes('read'),
        write: values.permissions.includes('write'),
        delete: values.permissions.includes('delete'),
        release: values.permissions.includes('release'),
        admin: values.permissions.includes('admin'),
        decrypt: values.permissions.includes('decrypt'),
      }
      const data: CreateKeyRequest = {
        name: values.name,
        permissions,
        expires_at: values.expires_at?.toISOString(),
      }
      const { data: res } = await keyApi.create(Number(projectId), data)
      setNewSecretKey(res.secret_key)
      setModalOpen(false)
      form.resetFields()
      fetchData()
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (keyId: number) => {
    await keyApi.delete(keyId)
    message.success('密钥已删除')
    fetchData()
  }

  const handleRegenerate = async (keyId: number) => {
    const { data } = await keyApi.regenerate(keyId)
    setNewSecretKey(data.secret_key)
    fetchData()
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
  }

  const parsePermissions = (permStr: string) => {
    try {
      return JSON.parse(permStr)
    } catch {
      return {}
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Access Key',
      dataIndex: 'access_key',
      key: 'access_key',
      render: (key: string) => (
        <Space>
          <Text code copyable={{ text: key }}>{key}</Text>
        </Space>
      ),
    },
    {
      title: '权限',
      dataIndex: 'permissions',
      key: 'permissions',
      render: (perms: string) => {
        const p = parsePermissions(perms)
        const tags = []
        if (p.read) tags.push(<Tag key="read" color="green">读取</Tag>)
        if (p.write) tags.push(<Tag key="write" color="blue">写入</Tag>)
        if (p.delete) tags.push(<Tag key="delete" color="orange">删除</Tag>)
        if (p.release) tags.push(<Tag key="release" color="purple">发布</Tag>)
        if (p.admin) tags.push(<Tag key="admin" color="red">管理</Tag>)
        if (p.decrypt) tags.push(<Tag key="decrypt" color="cyan">解密</Tag>)
        return <Space wrap>{tags}</Space>
      },
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>{active ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '过期时间',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (t: string | null) => t ? dayjs(t).format('YYYY-MM-DD') : '永不过期',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: ProjectKey) => (
        <Space>
          <Popconfirm
            title="确认重新生成"
            description="重新生成后，旧的 Secret Key 将失效"
            onConfirm={() => handleRegenerate(record.id)}
          >
            <Button type="link" size="small" icon={<ReloadOutlined />}>
              重新生成
            </Button>
          </Popconfirm>
          <Popconfirm
            title="确认删除"
            description="删除后，使用此密钥的应用将无法访问"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
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
          { title: <a onClick={() => navigate(`/projects/${projectId}`)}>{project?.name}</a> },
          { title: '密钥管理' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        title="API 密钥"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            创建密钥
          </Button>
        }
      >
        <Table columns={columns} dataSource={keys} rowKey="id" pagination={false} />
      </Card>

      <Modal title="创建密钥" open={modalOpen} onCancel={() => setModalOpen(false)} footer={null}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreate}
          initialValues={{ permissions: ['read'] }}
        >
          <Form.Item name="name" label="密钥名称" rules={[{ required: true, message: '请输入密钥名称' }]}>
            <Input placeholder="如: production-server" />
          </Form.Item>
          <Form.Item name="permissions" label="权限">
            <Checkbox.Group
              options={[
                { label: '读取', value: 'read' },
                { label: '写入', value: 'write' },
                { label: '删除', value: 'delete' },
                { label: '发布', value: 'release' },
                { label: '管理', value: 'admin' },
                { label: '解密', value: 'decrypt' },
              ]}
            />
          </Form.Item>
          <Form.Item name="expires_at" label="过期时间">
            <DatePicker placeholder="留空表示永不过期" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={creating}>创建</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="密钥创建成功"
        open={!!newSecretKey}
        onCancel={() => setNewSecretKey(null)}
        footer={<Button type="primary" onClick={() => setNewSecretKey(null)}>我已保存</Button>}
      >
        <Alert
          type="warning"
          message="请立即保存 Secret Key"
          description="Secret Key 只会显示一次，关闭后将无法再次查看。"
          style={{ marginBottom: 16 }}
        />
        <Paragraph>
          <Text strong>Secret Key:</Text>
        </Paragraph>
        <Input.TextArea
          value={newSecretKey || ''}
          readOnly
          rows={2}
          style={{ fontFamily: 'monospace' }}
        />
        <Button
          icon={<CopyOutlined />}
          onClick={() => copyToClipboard(newSecretKey || '')}
          style={{ marginTop: 8 }}
        >
          复制
        </Button>
      </Modal>
    </div>
  )
}
