import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Form, Input, Select, Button, Space, Typography, Breadcrumb, Spin, message, Popconfirm, Divider } from 'antd'
import { SaveOutlined, DeleteOutlined } from '@ant-design/icons'
import { projectApi } from '../api/projects'
import type { Project } from '../api/client'

const { Title, Text } = Typography

export default function ProjectSettingsPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [form] = Form.useForm()

  const fetchProject = async () => {
    if (!projectId) return
    try {
      const { data } = await projectApi.get(Number(projectId))
      setProject(data.project)
      form.setFieldsValue(data.project)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchProject()
  }, [projectId])

  const handleSave = async (values: Partial<Project>) => {
    if (!projectId) return
    setSaving(true)
    try {
      await projectApi.update(Number(projectId), values)
      message.success('保存成功')
      fetchProject()
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!projectId) return
    try {
      await projectApi.delete(Number(projectId))
      message.success('项目已删除')
      navigate('/projects')
    } catch {
      // 错误已在拦截器中处理
    }
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <a onClick={() => navigate('/projects')}>项目列表</a> },
          { title: <a onClick={() => navigate(`/projects/${projectId}`)}>{project?.name}</a> },
          { title: '设置' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card title="基本信息">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            name="name"
            label="项目名称"
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>

          <Form.Item name="description" label="项目描述">
            <Input.TextArea rows={3} placeholder="请输入项目描述" />
          </Form.Item>

          <Form.Item name="access_mode" label="访问模式">
            <Select
              options={[
                { value: 'private', label: '私有 - 仅授权用户可访问' },
                { value: 'protected', label: '受保护 - 需要密钥访问' },
                { value: 'public', label: '公开 - 所有人可读取' },
              ]}
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
              保存更改
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="危险操作" style={{ marginTop: 16 }}>
        <Space direction="vertical">
          <div>
            <Title level={5} style={{ color: '#ff4d4f', margin: 0 }}>删除项目</Title>
            <Text type="secondary">删除项目后，所有配置和版本历史将被永久删除，此操作不可恢复。</Text>
          </div>
          <Popconfirm
            title="确认删除"
            description="确定要删除此项目吗？此操作不可恢复。"
            onConfirm={handleDelete}
            okText="确认删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button danger icon={<DeleteOutlined />}>
              删除项目
            </Button>
          </Popconfirm>
        </Space>
      </Card>
    </div>
  )
}
