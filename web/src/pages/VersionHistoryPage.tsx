import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Timeline, Button, Space, Typography, Breadcrumb, Spin, Modal, message, Tag, Empty } from 'antd'
import { RollbackOutlined, DiffOutlined, UserOutlined, ClockCircleOutlined } from '@ant-design/icons'
import ReactDiffViewer from 'react-diff-viewer-continued'
import { configApi } from '../api/configs'
import type { Config, ConfigVersion } from '../api/client'
import dayjs from 'dayjs'

const { Title, Text } = Typography

export default function VersionHistoryPage() {
  const { configId } = useParams<{ configId: string }>()
  const navigate = useNavigate()
  const [config, setConfig] = useState<Config | null>(null)
  const [versions, setVersions] = useState<ConfigVersion[]>([])
  const [loading, setLoading] = useState(true)
  const [diffModalOpen, setDiffModalOpen] = useState(false)
  const [diffData, setDiffData] = useState<{ old: string; new: string; v1: number; v2: number } | null>(null)
  const [selectedVersions, setSelectedVersions] = useState<number[]>([])

  const fetchData = async () => {
    if (!configId) return
    try {
      const [configRes, versionsRes] = await Promise.all([
        configApi.get(Number(configId)),
        configApi.listVersions(Number(configId)),
      ])
      setConfig(configRes.data.config)
      setVersions(versionsRes.data.versions || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [configId])

  const handleRollback = async (version: number) => {
    if (!configId) return
    Modal.confirm({
      title: '确认回滚',
      content: `确定要回滚到版本 v${version} 吗？这将创建一个新版本。`,
      onOk: async () => {
        await configApi.rollback(Number(configId), version)
        message.success('回滚成功')
        fetchData()
      },
    })
  }

  const handleCompare = async () => {
    if (selectedVersions.length !== 2 || !configId) return
    const [v1, v2] = selectedVersions.sort((a, b) => a - b)
    try {
      const [ver1Res, ver2Res] = await Promise.all([
        configApi.getVersion(Number(configId), v1),
        configApi.getVersion(Number(configId), v2),
      ])
      setDiffData({
        old: ver1Res.data.version.content,
        new: ver2Res.data.version.content,
        v1,
        v2,
      })
      setDiffModalOpen(true)
    } catch {
      message.error('获取版本内容失败')
    }
  }

  const toggleVersionSelect = (version: number) => {
    setSelectedVersions((prev) => {
      if (prev.includes(version)) {
        return prev.filter((v) => v !== version)
      }
      if (prev.length >= 2) {
        return [prev[1], version]
      }
      return [...prev, version]
    })
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <a onClick={() => navigate('/projects')}>项目列表</a> },
          { title: <a onClick={() => navigate(`/projects/${config?.project_id}`)}>项目详情</a> },
          { title: <a onClick={() => navigate(`/configs/${configId}`)}>{config?.name}</a> },
          { title: '版本历史' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        title={
          <Space>
            <Title level={5} style={{ margin: 0 }}>版本历史</Title>
            <Text type="secondary">共 {versions.length} 个版本</Text>
          </Space>
        }
        extra={
          <Button
            icon={<DiffOutlined />}
            onClick={handleCompare}
            disabled={selectedVersions.length !== 2}
          >
            对比选中版本 ({selectedVersions.length}/2)
          </Button>
        }
      >
        {versions.length === 0 ? (
          <Empty description="暂无版本记录" />
        ) : (
          <Timeline
            items={versions.map((v) => ({
              color: selectedVersions.includes(v.version) ? 'blue' : 'gray',
              children: (
                <Card
                  size="small"
                  style={{
                    cursor: 'pointer',
                    border: selectedVersions.includes(v.version) ? '2px solid #1890ff' : undefined,
                  }}
                  onClick={() => toggleVersionSelect(v.version)}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Space direction="vertical" size={4}>
                      <Space>
                        <Tag color="blue">v{v.version}</Tag>
                        <Text strong>{v.commit_message || '无提交说明'}</Text>
                      </Space>
                      <Space size={16}>
                        <Text type="secondary">
                          <UserOutlined /> {v.author}
                        </Text>
                        <Text type="secondary">
                          <ClockCircleOutlined /> {dayjs(v.created_at).format('YYYY-MM-DD HH:mm:ss')}
                        </Text>
                        <Text type="secondary" copyable={{ text: v.commit_hash }}>
                          {v.commit_hash?.substring(0, 8)}
                        </Text>
                      </Space>
                    </Space>
                    <Button
                      type="link"
                      icon={<RollbackOutlined />}
                      onClick={(e) => {
                        e.stopPropagation()
                        handleRollback(v.version)
                      }}
                      disabled={v.version === config?.current_version}
                    >
                      回滚
                    </Button>
                  </div>
                </Card>
              ),
            }))}
          />
        )}
      </Card>

      <Modal
        title={`版本对比: v${diffData?.v1} → v${diffData?.v2}`}
        open={diffModalOpen}
        onCancel={() => setDiffModalOpen(false)}
        footer={null}
        width={1000}
      >
        {diffData && (
          <div style={{ maxHeight: '60vh', overflow: 'auto' }}>
            <ReactDiffViewer
              oldValue={diffData.old}
              newValue={diffData.new}
              splitView={true}
              leftTitle={`v${diffData.v1}`}
              rightTitle={`v${diffData.v2}`}
            />
          </div>
        )}
      </Modal>
    </div>
  )
}
