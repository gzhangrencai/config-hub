import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Row, Col, Button, Modal, Form, Input, Empty, Spin, Typography, Tag, Space } from 'antd'
import { PlusOutlined, SettingOutlined, FolderOutlined } from '@ant-design/icons'
import { projectApi } from '../api/projects'
import type { Project } from '../api/client'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const { Title, Text, Paragraph } = Typography

export default function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const fetchProjects = async () => {
    try {
      const { data } = await projectApi.list()
      setProjects(data.projects || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchProjects()
  }, [])

  const handleCreate = async (values: { name: string; description?: string }) => {
    setCreating(true)
    try {
      await projectApi.create(values)
      setModalOpen(false)
      form.resetFields()
      fetchProjects()
    } finally {
      setCreating(false)
    }
  }

  const accessModeMap: Record<string, { color: string; text: string }> = {
    public: { color: 'green', text: '公开' },
    private: { color: 'orange', text: '私有' },
    protected: { color: 'blue', text: '受保护' },
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={4} style={{ margin: 0 }}>项目列表</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          创建项目
        </Button>
      </div>

      {projects.length === 0 ? (
        <Card>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无项目"
          >
            <Button type="primary" onClick={() => setModalOpen(true)}>
              创建第一个项目
            </Button>
          </Empty>
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          {projects.map((project) => (
            <Col xs={24} sm={12} lg={8} xl={6} key={project.id}>
              <Card
                hoverable
                onClick={() => navigate(`/projects/${project.id}`)}
                actions={[
                  <SettingOutlined
                    key="settings"
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/projects/${project.id}/settings`)
                    }}
                  />,
                ]}
              >
                <Card.Meta
                  avatar={
                    <div
                      style={{
                        width: 48,
                        height: 48,
                        borderRadius: 8,
                        background: '#1890ff',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <FolderOutlined style={{ fontSize: 24, color: '#fff' }} />
                    </div>
                  }
                  title={
                    <Space>
                      <Text strong>{project.name}</Text>
                      <Tag color={accessModeMap[project.access_mode]?.color || 'default'}>
                        {accessModeMap[project.access_mode]?.text || project.access_mode}
                      </Tag>
                    </Space>
                  }
                  description={
                    <>
                      <Paragraph
                        ellipsis={{ rows: 2 }}
                        style={{ marginBottom: 8, minHeight: 44 }}
                        type="secondary"
                      >
                        {project.description || '暂无描述'}
                      </Paragraph>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        更新于 {dayjs(project.updated_at).fromNow()}
                      </Text>
                    </>
                  }
                />
              </Card>
            </Col>
          ))}
        </Row>
      )}

      <Modal
        title="创建项目"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="name"
            label="项目名称"
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>
          <Form.Item name="description" label="项目描述">
            <Input.TextArea rows={3} placeholder="请输入项目描述（可选）" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={creating}>
                创建
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
