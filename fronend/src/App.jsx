import { useState, useEffect } from 'react'
import ChatContainer from './components/ChatContainer'
import ConversationHistory from './components/ConversationHistory'
import './App.css'

function App() {
  const [messages, setMessages] = useState([])
  const [selectedModel, setSelectedModel] = useState('openai')
  const [models, setModels] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [currentConversationId, setCurrentConversationId] = useState(null)
  const [conversations, setConversations] = useState([])
  const [isHistoryCollapsed, setIsHistoryCollapsed] = useState(false)
  const [isHistoryView, setIsHistoryView] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [hasMoreMessages, setHasMoreMessages] = useState(false)
  const [shouldScrollToBottom, setShouldScrollToBottom] = useState(false)
  const [historyWidth, setHistoryWidth] = useState(() => {
    // 从localStorage读取保存的宽度，默认260px
    const saved = localStorage.getItem('conversationHistoryWidth')
    return saved ? parseInt(saved, 10) : 260
  })

  useEffect(() => {
    fetchModels()
    fetchConversations()
  }, [])

  const fetchModels = async () => {
    try {
      const response = await fetch('/api/models')
      const data = await response.json()
      setModels(data.models)
      if (data.models.length > 0) {
        setSelectedModel(data.models[0].id)
      }
    } catch (error) {
      console.error('Failed to fetch models:', error)
    }
  }

  const fetchConversations = async () => {
    try {
      const response = await fetch('/api/conversations?page=1&page_size=50')
      const data = await response.json()
      setConversations(data.conversations || [])
    } catch (error) {
      console.error('Failed to fetch conversations:', error)
    }
  }

  const handleNewChat = async () => {
    try {
      // 判断当前对话是否为默认标题，如果是则自动命名
      if (currentConversationId && messages.length > 0) {
        const currentConversation = conversations.find(c => c.id === currentConversationId)
        if (currentConversation && currentConversation.title === '新对话') {
          // 收集当前对话的所有用户输入
          const userInputs = messages
            .filter(msg => msg.role === 'user')
            .map(msg => msg.content)

          if (userInputs.length > 0) {
            try {
              // 调用智能命名接口生成标题
              const titleResponse = await fetch('/api/conversations/generate-title', {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                  user_inputs: userInputs,
                }),
              })

              if (titleResponse.ok) {
                const titleData = await titleResponse.json()
                // 使用生成的标题重命名旧对话
                await handleRenameConversation(currentConversationId, titleData.title)
              }
            } catch (error) {
              console.error('Failed to generate title for old conversation:', error)
            }
          }
        }
      }

      // 创建新对话
      const response = await fetch('/api/conversations/new', {
        method: 'POST',
      })

      if (response.ok) {
        const data = await response.json()
        setCurrentConversationId(data.id)
        setMessages([])
        setIsHistoryView(false) // 新对话，不是历史视图
        fetchConversations()
      }
    } catch (error) {
      console.error('Failed to create new conversation:', error)
    }
  }

  const handleRenameConversation = async (conversationId, newTitle) => {
    try {
      const response = await fetch(`/api/conversations/${conversationId}/title`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title: newTitle,
        }),
      })
      if (response.ok) {
        fetchConversations() // 刷新对话列表
      }
    } catch (error) {
      console.error('Failed to rename conversation:', error)
    }
  }

  const handleDeleteConversation = async (conversationId) => {
    try {
      const response = await fetch(`/api/conversations/${conversationId}`, {
        method: 'DELETE',
      })
      if (response.ok) {
        // 如果删除的是当前对话，清空消息并重置
        if (conversationId === currentConversationId) {
          setCurrentConversationId(null)
          setMessages([])
          setIsHistoryView(false)
        }
        fetchConversations() // 刷新对话列表
      }
    } catch (error) {
      console.error('Failed to delete conversation:', error)
    }
  }

  const handleSelectConversation = async (conversationId) => {
    // 如果点击的是当前对话，不发起请求
    if (conversationId === currentConversationId) {
      return
    }

    setCurrentConversationId(conversationId)
    
    // 清空之前的消息，避免显示旧数据
    setMessages([])
    
    try {
      // 第一步：请求文档ID列表接口，获取这个对话的最新10个文档的ID
      // 后端返回的ID列表已经按时间正序排列（最早的在前，最新的在后）
      const idsResponse = await fetch(`/api/documents/ids?conversation_id=${conversationId}&limit=10`)
      if (!idsResponse.ok) {
        throw new Error('Failed to get document IDs')
      }

      const idsData = await idsResponse.json()
      const documentIDs = idsData.document_ids || []

      if (documentIDs.length === 0) {
        setMessages([])
        setIsHistoryView(true)
        setHasMoreMessages(false)
        return
      }

      // 第二步：并发请求文档管理模块，同时发送多个请求，每个请求获取一个文档的详细信息
      // 注意：这里只获取最新10条文档，不会获取全部对话
      const documentPromises = documentIDs.map(id =>
        fetch(`/api/documents/${id}`)
          .then(response => {
            if (!response.ok) {
              throw new Error(`Failed to get document ${id}`)
            }
            return response.json()
          })
          .catch(error => {
            console.error(`Error fetching document ${id}:`, error)
            return null
          })
      )

      const documentResults = await Promise.all(documentPromises)
      
      // 过滤掉失败的结果，并按ID顺序排序（保持时间顺序）
      // documentIDs已经按时间正序排列（最早的在前），所以documents也需要按此顺序排列
      const documents = documentResults
        .filter(doc => doc !== null)
        .sort((a, b) => {
          // 按照documentIDs的顺序排序，确保时间顺序正确
          const indexA = documentIDs.indexOf(a.id)
          const indexB = documentIDs.indexOf(b.id)
          return indexA - indexB
        })

      // 转换为消息格式
      // 消息列表顺序：最早的在前（顶部），最新的在后（底部）
      const conversationMessages = documents.map((doc) => ({
        id: doc.id,
        role: doc.role,
        content: doc.content,
      }))

      // 设置消息（只显示最新10条）
      setMessages(conversationMessages)
      setIsHistoryView(true) // 加载历史对话后，设置为历史视图（不使用打字机效果）
      
      // 判断是否还有更多文档（如果返回的文档数量等于limit，说明可能还有更多）
      setHasMoreMessages(documentIDs.length >= 10)
      
      // 触发滚动到底部，显示最新消息（最新的消息在列表底部）
      // 使用多重延迟确保DOM完全渲染后再滚动
      requestAnimationFrame(() => {
        setTimeout(() => {
          setShouldScrollToBottom(true)
          // 再次延迟重置标志，确保滚动完成
          setTimeout(() => {
            setShouldScrollToBottom(false)
          }, 300)
        }, 150)
      })
    } catch (error) {
      console.error('Failed to load conversation:', error)
      setMessages([]) // 出错时清空消息
    }
  }

  // 加载更早的文档（翻页功能）
  const loadOlderMessages = async () => {
    if (messages.length === 0 || !currentConversationId || isLoadingMore) return

    // 获取当前页面最早的文档ID
    const earliestDocId = messages[0].id

    setIsLoadingMore(true)
    try {
      // 第一步：获取比earliestDocId更早的文档ID列表
      const idsResponse = await fetch(
        `/api/documents/ids?conversation_id=${currentConversationId}&before_id=${earliestDocId}&limit=10`
      )

      if (!idsResponse.ok) {
        throw new Error('Failed to get document IDs')
      }

      const idsData = await idsResponse.json()
      if (idsData.document_ids.length === 0) {
        setHasMoreMessages(false) // 没有更多文档了
        return
      }

      // 第二步：并发请求文档管理模块，同时发送多个请求，每个请求获取一个文档的详细信息
      const documentPromises = idsData.document_ids.map(id =>
        fetch(`/api/documents/${id}`)
          .then(response => {
            if (!response.ok) {
              throw new Error(`Failed to get document ${id}`)
            }
            return response.json()
          })
          .catch(error => {
            console.error(`Error fetching document ${id}:`, error)
            return null
          })
      )

      const documentResults = await Promise.all(documentPromises)
      
      // 过滤掉失败的结果，并按ID顺序排序（保持时间顺序）
      const documents = documentResults
        .filter(doc => doc !== null)
        .sort((a, b) => {
          // 按照documentIDs的顺序排序
          const indexA = idsData.document_ids.indexOf(a.id)
          const indexB = idsData.document_ids.indexOf(b.id)
          return indexA - indexB
        })

      const olderMessages = documents.map((doc) => ({
        id: doc.id,
        role: doc.role,
        content: doc.content,
      }))

      // 将更早的消息插入到列表前面（按顺序渲染）
      setMessages((prev) => [...olderMessages, ...prev])
      
      // 如果返回的文档数量小于limit，说明没有更多了
      if (idsData.document_ids.length < 10) {
        setHasMoreMessages(false)
      }
    } catch (error) {
      console.error('Failed to load older messages:', error)
    } finally {
      setIsLoadingMore(false)
    }
  }

  const getOrCreateConversationId = async () => {
    // 如果已有对话ID，直接返回
    if (currentConversationId) {
      return currentConversationId
    }
    // 如果没有对话ID，创建新对话
    // 注意：这里不刷新列表，因为新对话还没有标题，等到第一次发送消息后再刷新
    try {
      const response = await fetch('/api/conversations/new', {
        method: 'POST',
      })
      if (response.ok) {
        const data = await response.json()
        setCurrentConversationId(data.id)
        return data.id
      }
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
    return null
  }

  const handleSendMessage = async (message) => {
    if (!message.trim() || isLoading) return

    // 检查是否是新对话（第一次发送消息）
    const isNewConversation = !currentConversationId

    // 发送新消息时，切换到非历史视图（使用打字机效果）
    setIsHistoryView(false)

    // 首先获取或创建对话ID
    const conversationId = await getOrCreateConversationId()
    if (!conversationId) {
      console.error('Failed to get conversation ID')
      return
    }

    const userMessage = {
      id: Date.now(),
      role: 'user',
      content: message,
    }

    setMessages((prev) => [...prev, userMessage])
    setIsLoading(true)

    const assistantMessage = {
      id: Date.now() + 1,
      role: 'assistant',
      content: '',
    }
    setMessages((prev) => [...prev, assistantMessage])

    try {
      // 前端仅携带本次用户请求的内容
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          conversation_id: conversationId,
          model: selectedModel,
          messages: [
            {
              role: 'user',
              content: message,
            },
          ],
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to get response')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let fullContent = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        fullContent += chunk

        setMessages((prev) => {
          const newMessages = [...prev]
          const lastMessage = newMessages[newMessages.length - 1]
          if (lastMessage.role === 'assistant') {
            lastMessage.content = fullContent
          }
          return newMessages
        })
      }
      // 只有在新对话第一次发送消息时才刷新列表（更新标题）
      // 已存在的对话不需要刷新，因为用户已经在当前对话中
      if (isNewConversation) {
        await fetchConversations()
      }
    } catch (error) {
      console.error('Error:', error)
      setMessages((prev) => {
        const newMessages = [...prev]
        const lastMessage = newMessages[newMessages.length - 1]
        if (lastMessage.role === 'assistant') {
          lastMessage.content = '抱歉，发生了错误：' + error.message
        }
        return newMessages
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="app">
      <ConversationHistory
        conversations={conversations}
        currentConversationId={currentConversationId}
        onSelectConversation={handleSelectConversation}
        isCollapsed={isHistoryCollapsed}
        onToggleCollapse={() => setIsHistoryCollapsed(!isHistoryCollapsed)}
        onRenameConversation={handleRenameConversation}
        onDeleteConversation={handleDeleteConversation}
        width={historyWidth}
        onWidthChange={(newWidth) => {
          setHistoryWidth(newWidth)
          localStorage.setItem('conversationHistoryWidth', newWidth.toString())
        }}
      />
      <div className="main-content">
        <div className="new-chat-button-container">
          <button onClick={handleNewChat} className="new-chat-button-top">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path
                d="M8 1v14M1 8h14"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            </svg>
            新对话
          </button>
        </div>
        <ChatContainer
          messages={messages}
          onSendMessage={handleSendMessage}
          isLoading={isLoading}
          models={models}
          selectedModel={selectedModel}
          onModelChange={setSelectedModel}
          enableTypewriter={!isHistoryView}
          onLoadMore={loadOlderMessages}
          canLoadMore={isHistoryView && hasMoreMessages}
          isLoadingMore={isLoadingMore}
          shouldScrollToBottom={shouldScrollToBottom}
        />
      </div>
    </div>
  )
}

export default App

