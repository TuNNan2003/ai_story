import { useState, useEffect, useRef, useCallback } from 'react'
import CryptoJS from 'crypto-js'
import ChatContainer from './components/ChatContainer'
import ConversationHistory from './components/ConversationHistory'
import WorkHistory from './components/WorkHistory'
import WorkDocumentList from './components/WorkDocumentList'
import StoryEditDialog from './components/StoryEditDialog'
import StoryViewDialog from './components/StoryViewDialog'
import InspirationDocumentView from './components/InspirationDocumentView'
import './App.css'

function App() {
  const [messages, setMessages] = useState([])
  const [selectedModel, setSelectedModel] = useState('openai')
  const [models, setModels] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [currentConversationId, setCurrentConversationId] = useState(null)
  const [conversations, setConversations] = useState([])
  const [stories, setStories] = useState([])
  const [isHistoryCollapsed, setIsHistoryCollapsed] = useState(false)
  const [storyEditDialogOpen, setStoryEditDialogOpen] = useState(false)
  const [storyEditContent, setStoryEditContent] = useState('')
  const [storyEditTitle, setStoryEditTitle] = useState('')
  const [storyEditDocumentId, setStoryEditDocumentId] = useState(null)
  const [storyEditStoryId, setStoryEditStoryId] = useState(null)
  const [storyViewDialogOpen, setStoryViewDialogOpen] = useState(false)
  const [selectedStory, setSelectedStory] = useState(null)
  const [isHistoryView, setIsHistoryView] = useState(false)
  const [isInspirationMode, setIsInspirationMode] = useState(false)
  const [works, setWorks] = useState([])
  const [currentWorkId, setCurrentWorkId] = useState(null)
  const [workDocuments, setWorkDocuments] = useState([])
  const [isModifyOriginal, setIsModifyOriginal] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [hasMoreMessages, setHasMoreMessages] = useState(false)
  const [shouldScrollToBottom, setShouldScrollToBottom] = useState(false)
  const [inspirationDocumentViewOpen, setInspirationDocumentViewOpen] = useState(false)
  const [inspirationDocumentContent, setInspirationDocumentContent] = useState('')
  const [rightPanelContent, setRightPanelContent] = useState(null) // 右侧面板显示的内容（null表示显示最新AI响应）
  const [historyWidth, setHistoryWidth] = useState(() => {
    // 从localStorage读取保存的宽度，默认260px
    const saved = localStorage.getItem('conversationHistoryWidth')
    return saved ? parseInt(saved, 10) : 260
  })
  
  // 用户ID：从localStorage获取或生成新的，确保多次打开网页时ID一致
  const [userId, setUserId] = useState(() => {
    const saved = localStorage.getItem('grandma_user_id')
    if (saved) {
      return saved
    }
    // 生成新的用户ID（使用时间戳和随机数）
    const newUserId = 'user_' + Date.now() + '_' + Math.random().toString(36).substring(2, 15)
    localStorage.setItem('grandma_user_id', newUserId)
    return newUserId
  })
  
  // 用于跟踪当前的流式响应读取器，以便在切换对话时取消
  const readerRef = useRef(null)
  const abortControllerRef = useRef(null)
  const currentConversationIdRef = useRef(null)
  const rightPanelContentRef = useRef(null)
  const currentWorkIdRef = useRef(null) // v1.3: 用于在流式响应中跟踪currentWorkId
  const isSendingMessageRef = useRef(false) // v1.3: 用于跟踪是否正在发送消息，避免在发送消息时加载历史消息
  // v1.3: 使用ref存储函数，避免初始化顺序问题
  const loadWorkMessagesRef = useRef(null)
  const loadWorkMessagesWithCallbackRef = useRef(null)
  
  // 同步currentConversationId到ref，以便在异步函数中使用最新值
  useEffect(() => {
    currentConversationIdRef.current = currentConversationId
  }, [currentConversationId])
  
  // 同步rightPanelContent到ref，以便在异步函数中使用最新值
  useEffect(() => {
    rightPanelContentRef.current = rightPanelContent
  }, [rightPanelContent])

  // v1.3: 同步currentWorkId到ref，以便在异步函数中使用最新值
  useEffect(() => {
    currentWorkIdRef.current = currentWorkId
  }, [currentWorkId])

  useEffect(() => {
    fetchModels()
    fetchConversations()
    fetchStories()
    fetchWorks()
  }, [])

  // 当灵感模式开启时，获取works；当选择work时，获取documents
  useEffect(() => {
    if (isInspirationMode) {
      fetchWorks()
      // v1.3: 如果已有currentWorkId，加载该创作的消息（在loadWorkMessages定义后调用）
      // 注意：loadWorkMessages在下面定义，但函数会被提升，所以可以调用
    } else {
      // v1.3: 关闭灵感模式时，清空消息和右侧面板
      setMessages([])
      setRightPanelContent(null)
      setCurrentWorkId(null)
    }
  }, [isInspirationMode])
  
  // v1.3: 当灵感模式开启且currentWorkId变化时，加载该创作的消息
  useEffect(() => {
    // 如果正在发送消息，不要加载历史消息（避免覆盖正在发送的消息）
    if (isSendingMessageRef.current) {
      return
    }
    
    if (isInspirationMode && currentWorkId) {
      // 如果loadWorkMessagesWithCallback已经定义，使用它；否则只使用loadWorkMessages
      if (loadWorkMessagesWithCallbackRef.current) {
        loadWorkMessagesWithCallbackRef.current(currentWorkId)
      } else if (loadWorkMessagesRef.current) {
        loadWorkMessagesRef.current(currentWorkId)
      }
    } else if (isInspirationMode && !currentWorkId) {
      // 如果没有创作，且消息列表为空，不需要清空（避免在发送消息时误清空）
      // 注意：当用户第一次进入灵感模式并发送消息时，会先创建work，然后添加消息
      // 如果在这里清空消息，可能会在消息添加之前清空，导致hasConversation为false
      // 所以，我们只在明确需要清空时才清空（比如用户切换了模式或选择了其他创作）
      // 这里暂时不清空，让handleSendMessage来处理
      // 但是，如果用户切换了模式或选择了其他创作，消息应该已经被清空了
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isInspirationMode, currentWorkId])

  useEffect(() => {
    if (currentWorkId) {
      fetchWorkDocuments(currentWorkId)
    }
  }, [currentWorkId])

  // 辅助函数：为URL添加用户ID参数
  const addUserIdToUrl = (url) => {
    const separator = url.includes('?') ? '&' : '?'
    return `${url}${separator}user_id=${encodeURIComponent(userId)}`
  }

  // v1.3: 加载创作的历史消息（从WorkDocument）
  // 必须在所有useEffect之前定义，因为useEffect会引用它
  // 注意：这个函数不包含handleAddToStory，因为它还没定义
  const loadWorkMessages = useCallback(async (workId) => {
    try {
      // 获取该创作的所有文档（按时间正序）
      const response = await fetch(addUserIdToUrl(`/api/works/${workId}/documents`))
      if (!response.ok) {
        throw new Error('Failed to get work documents')
      }

      const data = await response.json()
      const documents = data.documents || []

      // 过滤出有role的文档（对话消息），并按时间正序排列
      const conversationDocs = documents
        .filter(doc => doc.role === 'user' || doc.role === 'assistant')
        .sort((a, b) => {
          return new Date(a.created_at) - new Date(b.created_at)
        })

      // 转换为消息格式
      // 注意：这里不能直接使用handleAddToStory，因为它还没定义
      // 我们会在后面通过loadWorkMessagesWithCallback添加onAddToStory
      const workMessages = conversationDocs.map((doc) => ({
        id: doc.id,
        documentId: doc.id,
        role: doc.role,
        content: doc.content,
        // onAddToStory会在后面通过loadWorkMessagesWithCallback添加
      }))

      // 设置消息
      setMessages(workMessages)
      setIsHistoryView(true)
      
      // 如果最后一条消息是AI响应，自动在右侧显示
      if (workMessages.length > 0) {
        const lastMessage = workMessages[workMessages.length - 1]
        if (lastMessage.role === 'assistant' && lastMessage.content) {
          setRightPanelContent(lastMessage.content)
        }
      }
      
      setIsLoading(false)
      setHasMoreMessages(false) // WorkDocument暂时不支持翻页
      
      // 触发滚动到底部
      requestAnimationFrame(() => {
        setTimeout(() => {
          setShouldScrollToBottom(true)
          setTimeout(() => {
            setShouldScrollToBottom(false)
          }, 300)
        }, 150)
      })
    } catch (error) {
      console.error('Failed to load work messages:', error)
      setMessages([])
    }
  }, [userId])

  // 将loadWorkMessages保存到ref，以便在useEffect中使用
  useEffect(() => {
    loadWorkMessagesRef.current = loadWorkMessages
  }, [loadWorkMessages])

  const fetchModels = async () => {
    try {
      // models接口不需要用户ID，因为模型列表是全局的
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
      const response = await fetch(addUserIdToUrl('/api/conversations?page=1&page_size=50'))
      const data = await response.json()
      setConversations(data.conversations || [])
    } catch (error) {
      console.error('Failed to fetch conversations:', error)
    }
  }

  const fetchStories = async () => {
    try {
      const response = await fetch(addUserIdToUrl('/api/stories?guid=default'))
      const data = await response.json()
      setStories(data.stories || [])
    } catch (error) {
      console.error('Failed to fetch stories:', error)
    }
  }

  const fetchWorks = async () => {
    try {
      const response = await fetch(addUserIdToUrl('/api/works'))
      const data = await response.json()
      setWorks(data.works || [])
    } catch (error) {
      console.error('Failed to fetch works:', error)
    }
  }

  const fetchWorkDocuments = async (workId) => {
    try {
      const response = await fetch(addUserIdToUrl(`/api/works/${workId}/documents`))
      const data = await response.json()
      setWorkDocuments(data.documents || [])
    } catch (error) {
      console.error('Failed to fetch work documents:', error)
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
                  user_id: userId,
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
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
        }),
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
      const response = await fetch(addUserIdToUrl(`/api/conversations/${conversationId}/title`), {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title: newTitle,
          user_id: userId,
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
      const response = await fetch(addUserIdToUrl(`/api/conversations/${conversationId}`), {
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

  // 获取对话的所有用户输入（用于自动命名）
  const getAllUserInputsForConversation = async (convId) => {
    const userInputs = []
    let beforeID = null
    let hasMore = true

    while (hasMore) {
      try {
        // 获取文档ID列表
        const baseUrl = beforeID
          ? `/api/documents/ids?conversation_id=${convId}&before_id=${beforeID}&limit=100`
          : `/api/documents/ids?conversation_id=${convId}&limit=100`
        
        const idsResponse = await fetch(addUserIdToUrl(baseUrl))
        if (!idsResponse.ok) {
          break
        }

        const idsData = await idsResponse.json()
        const documentIDs = idsData.document_ids || []

        if (documentIDs.length === 0) {
          hasMore = false
          break
        }

        // 并发请求所有文档的详细信息
        const documentPromises = documentIDs.map(id =>
          fetch(addUserIdToUrl(`/api/documents/${id}`))
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
            // 按照documentIDs的顺序排序，确保时间顺序正确
            const indexA = documentIDs.indexOf(a.id)
            const indexB = documentIDs.indexOf(b.id)
            return indexA - indexB
          })
        
        // 收集用户输入（按时间顺序）
        documents.forEach(doc => {
          if (doc.role === 'user') {
            userInputs.push(doc.content)
          }
        })

        // 如果返回的文档数量小于limit，说明没有更多了
        if (documentIDs.length < 100) {
          hasMore = false
        } else {
          // 更新beforeID为最早的文档ID（因为documents是按时间正序排列的）
          beforeID = documentIDs[0]
        }
      } catch (error) {
        console.error('Failed to get user inputs:', error)
        hasMore = false
      }
    }

    return userInputs
  }

  const handleSelectConversation = async (conversationId) => {
    // 如果点击的是当前对话，不发起请求
    if (conversationId === currentConversationId) {
      return
    }

    // 保存旧对话ID，用于后续检查
    const oldConversationId = currentConversationId

    // 如果正在加载中（流式响应进行中），取消当前的请求
    // 注意：不重置isLoading状态，让后端继续保存已接收的内容
    // 切换到新对话后，会加载新对话的消息，不会影响旧对话的保存
    if (isLoading) {
      // 取消流式响应读取器
      if (readerRef.current) {
        try {
          await readerRef.current.cancel()
        } catch (error) {
          console.error('Error canceling reader:', error)
        }
        readerRef.current = null
      }
      // 取消fetch请求
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
        abortControllerRef.current = null
      }
      // 注意：不重置isLoading状态，让后端继续处理
      // 切换对话后，isLoading状态会被新对话的状态覆盖
    }

    // 在切换对话前，检查旧对话是否需要自动命名
    if (oldConversationId) {
      const oldConversation = conversations.find(c => c.id === oldConversationId)
      if (oldConversation && oldConversation.title === '新对话') {
        // 旧对话有默认标题，需要自动命名
        // 获取旧对话的所有用户输入
        try {
          const userInputs = await getAllUserInputsForConversation(oldConversationId)
          if (userInputs.length > 0) {
            // 调用智能命名接口生成标题
            const titleResponse = await fetch('/api/conversations/generate-title', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({
                user_inputs: userInputs,
                user_id: userId,
              }),
            })

            if (titleResponse.ok) {
              const titleData = await titleResponse.json()
              // 使用生成的标题重命名旧对话
              await handleRenameConversation(oldConversationId, titleData.title)
            }
          }
        } catch (error) {
          console.error('Failed to auto-name old conversation:', error)
          // 即使命名失败，也继续切换对话
        }
      }
    }

    setCurrentConversationId(conversationId)
    
    // 清空之前的消息，避免显示旧数据
    setMessages([])
    
    // v1.2: 切换对话时，重置右侧面板，让右侧显示新对话的最新响应
    if (isInspirationMode) {
      setRightPanelContent(null)
    }
    
    try {
      // 第一步：请求文档ID列表接口，获取这个对话的最新10个文档的ID
      // 后端返回的ID列表已经按时间正序排列（最早的在前，最新的在后）
      const idsResponse = await fetch(addUserIdToUrl(`/api/documents/ids?conversation_id=${conversationId}&limit=10`))
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
        fetch(addUserIdToUrl(`/api/documents/${id}`))
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
        documentId: doc.id, // 历史消息的documentId就是id
        role: doc.role,
        content: doc.content,
        onAddToStory: doc.role === 'assistant' ? handleAddToStory : undefined,
      }))

      // 设置消息（只显示最新10条）
      setMessages(conversationMessages)
      setIsHistoryView(true) // 加载历史对话后，设置为历史视图（不使用打字机效果）
      
      // v1.2: 在灵感模式下，如果最后一条消息是AI响应，自动在右侧显示
      if (isInspirationMode) {
        const lastMessage = conversationMessages[conversationMessages.length - 1]
        if (lastMessage && lastMessage.role === 'assistant' && lastMessage.content) {
          setRightPanelContent(lastMessage.content)
        }
      }
      
      // 重置加载状态（切换对话时，无论之前是否在加载，都应该重置）
      setIsLoading(false)
      
      // 判断是否还有更多文档（如果返回的文档数量等于limit，说明可能还有更多）
      setHasMoreMessages(documentIDs.length >= 10)
      
      // 如果切换回的是之前正在生成响应的对话，延迟重新加载以确保获取最新保存的内容
      // 因为后端可能在切换对话时还在保存流式响应的剩余内容
      if (oldConversationId && oldConversationId !== conversationId) {
        // 延迟重新加载，确保后端有时间保存剩余内容
        setTimeout(async () => {
          // 只有在当前对话仍然是目标对话时才重新加载
          if (currentConversationIdRef.current === conversationId) {
            try {
              // 重新获取文档ID列表
              const refreshIdsResponse = await fetch(addUserIdToUrl(`/api/documents/ids?conversation_id=${conversationId}&limit=10`))
              if (refreshIdsResponse.ok) {
                const refreshIdsData = await refreshIdsResponse.json()
                const refreshDocumentIDs = refreshIdsData.document_ids || []
                
                // 如果文档数量发生变化，说明有新的内容被保存，需要重新加载
                if (refreshDocumentIDs.length !== documentIDs.length || 
                    (refreshDocumentIDs.length > 0 && documentIDs.length > 0 && 
                     refreshDocumentIDs[refreshDocumentIDs.length - 1] !== documentIDs[documentIDs.length - 1])) {
                  // 重新加载文档
                  const refreshDocumentPromises = refreshDocumentIDs.map(id =>
                    fetch(addUserIdToUrl(`/api/documents/${id}`))
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
                  
                  const refreshDocumentResults = await Promise.all(refreshDocumentPromises)
                  
                  const refreshDocuments = refreshDocumentResults
                    .filter(doc => doc !== null)
                    .sort((a, b) => {
                      const indexA = refreshDocumentIDs.indexOf(a.id)
                      const indexB = refreshDocumentIDs.indexOf(b.id)
                      return indexA - indexB
                    })
                  
                  const refreshMessages = refreshDocuments.map((doc) => ({
                    id: doc.id,
                    documentId: doc.id, // 历史消息的documentId就是id
                    role: doc.role,
                    content: doc.content,
                    onAddToStory: doc.role === 'assistant' ? handleAddToStory : undefined,
                  }))
                  
                  // 只有在当前对话仍然是目标对话时才更新消息
                  if (currentConversationIdRef.current === conversationId) {
                    setMessages(refreshMessages)
                    setHasMoreMessages(refreshDocumentIDs.length >= 10)
                  }
                }
              }
            } catch (error) {
              console.error('Failed to refresh conversation:', error)
            }
          }
        }, 1000) // 延迟1秒，确保后端有时间保存剩余内容
      }
      
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
        addUserIdToUrl(`/api/documents/ids?conversation_id=${currentConversationId}&before_id=${earliestDocId}&limit=10`)
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
        fetch(addUserIdToUrl(`/api/documents/${id}`))
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
        documentId: doc.id, // 历史消息的documentId就是id
        role: doc.role,
        content: doc.content,
        onAddToStory: doc.role === 'assistant' ? handleAddToStory : undefined,
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
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
        }),
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

  // 计算文本的SHA-256特征值
  // 使用 crypto-js 库，确保在任何环境下都能计算真正的 SHA-256 哈希
  // 这样就能与后端保持一致，避免 hash_mismatch 错误
  const calculateContentHash = async (content) => {
    // 优先使用 crypto.subtle（如果可用），因为它更快且是原生API
    if (typeof crypto !== 'undefined' && crypto.subtle) {
      try {
        const encoder = new TextEncoder()
        const data = encoder.encode(content)
        const hashBuffer = await crypto.subtle.digest('SHA-256', data)
        const hashArray = Array.from(new Uint8Array(hashBuffer))
        const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
        return hashHex
      } catch (error) {
        console.error('Failed to calculate hash with crypto.subtle:', error)
        // 降级到 crypto-js
        return CryptoJS.SHA256(content).toString()
      }
    } else {
      // crypto.subtle 不可用，使用 crypto-js 作为降级方案
      // crypto-js 可以在任何环境下工作，包括非 HTTPS 环境
      return CryptoJS.SHA256(content).toString()
    }
  }

  const handleAddToStory = useCallback(async (documentId) => {
    try {
      // 获取文档内容
      const docResponse = await fetch(addUserIdToUrl(`/api/documents/${documentId}`))
      if (!docResponse.ok) {
        throw new Error('Failed to get document')
      }
      const doc = await docResponse.json()
      
      // 生成标题（使用文档内容的前50个字符）
      let title = doc.content.substring(0, 50).trim()
      if (doc.content.length > 50) {
        title += '...'
      }
      if (!title) {
        title = '未命名故事'
      }

      // 显示编辑对话框（创建模式）
      setStoryEditContent(doc.content)
      setStoryEditTitle(title)
      setStoryEditDocumentId(documentId)
      setStoryEditStoryId(null) // 创建模式，storyId为null
      setStoryEditDialogOpen(true)
    } catch (error) {
      console.error('Failed to add to story:', error)
    }
  }, [userId])

  // v1.3: 更新loadWorkMessages，添加handleAddToStory依赖，并在加载消息后添加onAddToStory回调
  const loadWorkMessagesWithCallback = useCallback(async (workId) => {
    await loadWorkMessages(workId)
    // 加载完成后，更新消息以添加onAddToStory回调
    setMessages((prev) => {
      return prev.map(msg => ({
        ...msg,
        onAddToStory: msg.role === 'assistant' ? handleAddToStory : undefined,
      }))
    })
  }, [loadWorkMessages, handleAddToStory])

  // 将loadWorkMessagesWithCallback保存到ref，以便在useEffect中使用
  useEffect(() => {
    loadWorkMessagesWithCallbackRef.current = loadWorkMessagesWithCallback
  }, [loadWorkMessagesWithCallback])


  const handleSaveStory = async (content, title) => {
    try {
      // 计算内容特征值
      const contentHash = await calculateContentHash(content)

      // 判断是创建还是更新
      const url = storyEditStoryId ? `/api/stories/${storyEditStoryId}` : '/api/stories'
      const method = storyEditStoryId ? 'PUT' : 'POST'

      const requestBody = {
        title: title,
        content: content,
        content_hash: contentHash,
        user_id: userId,
      }

      // 只有在创建时才需要document_id
      if (!storyEditStoryId && storyEditDocumentId) {
        requestBody.document_id = storyEditDocumentId
        requestBody.guid = 'default'
      }

      const response = await fetch(addUserIdToUrl(url), {
        method: method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      })

      const data = await response.json()

      if (response.ok) {
        // 刷新故事列表
        await fetchStories()
        return
      } else {
        // 根据后端返回的错误信息处理
        if (data.error === 'duplicate_story') {
          throw new Error('该故事已存在，请勿重复保存')
        } else if (data.error === 'hash_mismatch') {
          throw new Error('内容验证失败，请重新尝试')
        } else {
          throw new Error(data.error || '保存失败')
        }
      }
    } catch (error) {
      console.error('Failed to save story:', error)
      throw error
    }
  }

  const handleEditStory = async (story) => {
    try {
      // 设置编辑内容
      setStoryEditContent(story.content || '')
      setStoryEditTitle(story.title || '未命名故事')
      setStoryEditStoryId(story.id)
      setStoryEditDocumentId(null) // 编辑时不需要documentId
      setStoryEditDialogOpen(true)
    } catch (error) {
      console.error('Failed to edit story:', error)
    }
  }

  const handleDeleteStory = async (story) => {
    try {
      const response = await fetch(addUserIdToUrl(`/api/stories/${story.id}`), {
        method: 'DELETE',
      })

      if (response.ok) {
        // 刷新故事列表
        await fetchStories()
      } else {
        const data = await response.json()
        throw new Error(data.error || '删除失败')
      }
    } catch (error) {
      console.error('Failed to delete story:', error)
      alert('删除失败：' + error.message)
    }
  }

  const handleSelectStory = async (story) => {
    // 当选择故事时，打开故事查看对话框
    setSelectedStory(story)
    setStoryViewDialogOpen(true)
  }

  // 处理查看灵感模式下的文档
  const handleViewInspirationDocument = (documentId, content) => {
    // v1.2: 不再打开浮层，而是在右侧直接展示文档内容
    setRightPanelContent(content || '')
  }

  // 创作相关处理函数
  const handleNewWork = async () => {
    try {
      const response = await fetch('/api/works', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          title: '新创作',
        }),
      })
      if (response.ok) {
        const work = await response.json()
        await fetchWorks()
        setCurrentWorkId(work.id)
      }
    } catch (error) {
      console.error('Failed to create work:', error)
    }
  }

  const handleSelectWork = async (workId) => {
    // 如果点击的是当前创作，不发起请求
    if (workId === currentWorkId) {
      return
    }

    // v1.3: 切换创作时，清空之前的消息
    setMessages([])
    setRightPanelContent(null)
    
    setCurrentWorkId(workId)
    await fetchWorkDocuments(workId)
    
    // v1.3: 在灵感模式下，加载该创作的历史对话消息
    if (isInspirationMode) {
      await loadWorkMessagesWithCallback(workId)
    }
  }

  const handleRenameWork = async (workId, newTitle) => {
    try {
      const response = await fetch(`/api/works/${workId}/title`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          title: newTitle,
        }),
      })
      if (response.ok) {
        await fetchWorks()
      }
    } catch (error) {
      console.error('Failed to rename work:', error)
    }
  }

  const handleDeleteWork = async (workId) => {
    try {
      const response = await fetch(addUserIdToUrl(`/api/works/${workId}`), {
        method: 'DELETE',
      })
      if (response.ok) {
        await fetchWorks()
        if (currentWorkId === workId) {
          setCurrentWorkId(null)
          setWorkDocuments([])
          // v1.3: 删除当前创作时，清空消息和右侧面板
          if (isInspirationMode) {
            setMessages([])
            setRightPanelContent(null)
          }
        }
      }
    } catch (error) {
      console.error('Failed to delete work:', error)
    }
  }

  // v1.3: WorkDocument相关处理函数
  const handleNewDocument = async () => {
    if (!currentWorkId) return
    try {
      const response = await fetch(`/api/works/${currentWorkId}/documents`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          work_id: currentWorkId,
          title: '新文档',
          content: '',
        }),
      })
      if (response.ok) {
        await fetchWorkDocuments(currentWorkId)
      }
    } catch (error) {
      console.error('Failed to create document:', error)
    }
  }

  const handleRenameDocument = async (docId, newTitle) => {
    try {
      const response = await fetch(`/api/work-documents/${docId}/title`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: userId,
          title: newTitle,
        }),
      })
      if (response.ok) {
        await fetchWorkDocuments(currentWorkId)
      }
    } catch (error) {
      console.error('Failed to rename document:', error)
    }
  }

  const handleDeleteDocument = async (docId) => {
    try {
      const response = await fetch(addUserIdToUrl(`/api/work-documents/${docId}`), {
        method: 'DELETE',
      })
      if (response.ok) {
        await fetchWorkDocuments(currentWorkId)
        // v1.3: 如果删除的是当前显示的消息，清空消息列表
        setMessages((prev) => prev.filter(msg => msg.documentId !== docId))
      }
    } catch (error) {
      console.error('Failed to delete document:', error)
    }
  }

  const handleSelectDocument = async (doc) => {
    // v1.3: 选择文档时，在右侧显示文档内容
    if (isInspirationMode) {
      setRightPanelContent(doc.content || '')
    }
  }

  const handleSendMessage = async (message) => {
    if (!message.trim() || isLoading) return

    // v3: 标记正在发送消息，避免useEffect加载历史消息
    isSendingMessageRef.current = true

    // 检查是否是新对话/创作（第一次发送消息）
    // v1.3: 在灵感模式下检查currentWorkId，否则检查currentConversationId
    const isNewConversation = isInspirationMode ? !currentWorkId : !currentConversationId

    // 发送新消息时，切换到非历史视图（使用打字机效果）
    setIsHistoryView(false)

    // v1.2: 发送新消息时，重置右侧面板为null，这样右侧会自动显示新的AI响应
    if (isInspirationMode) {
      setRightPanelContent(null)
    }

    // v1.3: 在灵感模式下，使用work_id；否则使用conversation_id
    let targetId = null
    if (isInspirationMode) {
      // 灵感模式：确保有currentWorkId
      if (!currentWorkId) {
        // 如果没有创作，创建一个新创作
        try {
          const workResponse = await fetch('/api/works', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              user_id: userId,
              title: '新创作',
            }),
          })
          if (workResponse.ok) {
            const work = await workResponse.json()
            setCurrentWorkId(work.id)
            currentWorkIdRef.current = work.id // 立即更新ref，确保流式响应中可以访问
            await fetchWorks()
            targetId = work.id
          } else {
            console.error('Failed to create work')
            isSendingMessageRef.current = false // 重置标记
            return
          }
        } catch (error) {
          console.error('Failed to create work:', error)
          isSendingMessageRef.current = false // 重置标记
          return
        }
      } else {
        targetId = currentWorkId
        // 确保ref也更新了（虽然useEffect会同步，但这里也更新一下确保一致性）
        if (currentWorkIdRef.current !== currentWorkId) {
          currentWorkIdRef.current = currentWorkId
        }
      }
    } else {
      // 普通模式：获取或创建对话ID
      targetId = await getOrCreateConversationId()
      if (!targetId) {
      console.error('Failed to get conversation ID')
      isSendingMessageRef.current = false // 重置标记
      return
      }
    }

    const userMessage = {
      id: Date.now(),
      role: 'user',
      content: message,
    }

    setMessages((prev) => [...prev, userMessage])
    setIsLoading(true)

    // 使用稳定的临时ID作为key，避免在流式响应过程中更新id导致组件重新挂载
    const tempMessageId = Date.now() + 1
    const assistantMessage = {
      id: tempMessageId,
      role: 'assistant',
      content: '',
      onAddToStory: handleAddToStory,
    }
    setMessages((prev) => [...prev, assistantMessage])

    // 创建AbortController用于取消请求
    const abortController = new AbortController()
    abortControllerRef.current = abortController

    try {
      // 前端仅携带本次用户请求的内容
      // v1.3: 在灵感模式下传递work_id，否则传递conversation_id
      const requestBody = {
          model: selectedModel,
          user_id: userId,
          messages: [
            {
              role: 'user',
              content: message,
            },
          ],
      }
      
      if (isInspirationMode) {
        requestBody.work_id = targetId
      } else {
        requestBody.conversation_id = targetId
      }
      
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
        signal: abortController.signal,
      })

      if (!response.ok) {
        throw new Error('Failed to get response')
      }

      const reader = response.body.getReader()
      readerRef.current = reader
      const decoder = new TextDecoder()
      let fullContent = ''
      let documentID = null // 保存后端返回的真实文档ID
      
      // 定义元数据标记（需要在循环外定义，以便在循环后使用）
      const metadataMarker = '<GRANDMA_METADATA>'
      const metadataEndMarker = '</GRANDMA_METADATA>'

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        // 检查是否还在当前对话/创作（如果切换了，停止处理）
        // v1.3: 在灵感模式下检查currentWorkIdRef，否则检查currentConversationIdRef
        const currentTargetId1 = isInspirationMode ? currentWorkIdRef.current : currentConversationIdRef.current
        if (currentTargetId1 !== targetId) {
          await reader.cancel()
          break
        }

        const chunk = decoder.decode(value, { stream: true })
        
        if (chunk.includes(metadataMarker)) {
          // 提取元数据
          const metadataStart = chunk.indexOf(metadataMarker)
          const metadataEnd = chunk.indexOf(metadataEndMarker)
          
          if (metadataStart !== -1 && metadataEnd !== -1) {
            // 提取元数据JSON
            const metadataJSON = chunk.substring(
              metadataStart + metadataMarker.length,
              metadataEnd
            )
            try {
              const metadata = JSON.parse(metadataJSON)
              if (metadata.document_id) {
                documentID = metadata.document_id
              }
            } catch (e) {
              console.error('Failed to parse metadata:', e)
            }
            
            // 只保留元数据之前的内容
            fullContent += chunk.substring(0, metadataStart)
          } else {
            // 元数据可能跨多个chunk，暂时保留内容
            fullContent += chunk
          }
        } else {
          fullContent += chunk
        }

        // 检查是否还在当前对话/创作（如果切换了，停止更新消息）
        // v1.3: 在灵感模式下检查currentWorkIdRef，否则检查currentConversationIdRef
        const currentTargetId2 = isInspirationMode ? currentWorkIdRef.current : currentConversationIdRef.current
        if (currentTargetId2 === targetId) {
          setMessages((prev) => {
            const newMessages = [...prev]
            const lastMessage = newMessages[newMessages.length - 1]
            if (lastMessage && lastMessage.role === 'assistant') {
              // 移除可能的元数据标记
              let content = fullContent
              if (content.includes(metadataMarker)) {
                const metadataStart = content.indexOf(metadataMarker)
                content = content.substring(0, metadataStart)
              }
              // 只更新content，不更新id（避免key变化导致组件重新挂载）
              // id在循环结束后再更新，这样可以避免组件重新挂载导致打字机效果重新开始
              lastMessage.content = content
              // 确保onAddToStory回调存在
              if (!lastMessage.onAddToStory) {
                lastMessage.onAddToStory = handleAddToStory
              }
              // 注意：不在循环内更新id，避免key变化导致组件重新挂载
            }
            return newMessages
          })
          
          // v1.2: 在灵感模式下，如果右侧面板显示的是最新响应（rightPanelContentRef.current为null）
          // 注意：流式更新时不要设置rightPanelContent，让ChatContainer从lastAssistantContent获取内容
          // 这样可以保持isLoading状态正确，打字机效果才能正常工作
        }
      }
      
      // 流式响应结束后，处理元数据并更新消息ID（如果收到了文档ID）
      // 确保从内容中移除元数据标记，并提取元数据
      let finalContent = fullContent
      if (fullContent.includes(metadataMarker)) {
        const metadataStart = fullContent.indexOf(metadataMarker)
        const metadataEnd = fullContent.indexOf(metadataEndMarker)
        
        if (metadataStart !== -1 && metadataEnd !== -1) {
          // 提取元数据JSON
          const metadataJSON = fullContent.substring(
            metadataStart + metadataMarker.length,
            metadataEnd
          )
          try {
            const metadata = JSON.parse(metadataJSON)
            if (metadata.document_id) {
              documentID = metadata.document_id
            }
          } catch (e) {
            console.error('Failed to parse metadata:', e)
          }
          
          // 移除元数据标记，只保留实际内容
          finalContent = fullContent.substring(0, metadataStart)
        } else {
          // 如果标记不完整，至少移除开始标记
          finalContent = fullContent.substring(0, metadataStart)
        }
      }
      
      // 流式响应结束后，只更新文档ID和回调，不更新内容（避免触发打字机效果重新渲染）
      // 内容已经在循环内更新过了，这里只需要更新元数据
      // 注意：不更新id（因为id是key），而是添加documentId属性，避免key变化导致组件重新挂载
      // v1.3: 在灵感模式下检查currentWorkIdRef，否则检查currentConversationIdRef
      const currentTargetId = isInspirationMode ? currentWorkIdRef.current : currentConversationIdRef.current
      if (currentTargetId === targetId) {
        setMessages((prev) => {
          const lastMessage = prev[prev.length - 1]
          if (!lastMessage || lastMessage.role !== 'assistant') {
            return prev // 不需要更新，返回原数组
          }
          
          // 检查是否需要更新
          const needsUpdateDocumentId = documentID && lastMessage.documentId !== documentID
          const needsUpdateCallback = !lastMessage.onAddToStory
          
          // 如果不需要更新，直接返回原数组（避免触发重新渲染）
          if (!needsUpdateDocumentId && !needsUpdateCallback) {
            return prev
          }
          
          // 需要更新，创建新数组（但保持content和id不变，避免触发打字机效果重新渲染）
          const newMessages = [...prev]
          const updatedLastMessage = { 
            ...newMessages[newMessages.length - 1],
            // 保持content和id引用不变，避免触发打字机效果重新渲染和组件重新挂载
            content: newMessages[newMessages.length - 1].content,
            id: newMessages[newMessages.length - 1].id
          }
          
          // 更新documentId（如果需要）- 不更新id，避免key变化
          if (needsUpdateDocumentId && documentID) {
            updatedLastMessage.documentId = documentID
          }
          
          // 更新回调（如果需要）
          if (needsUpdateCallback) {
            updatedLastMessage.onAddToStory = handleAddToStory
          }
          
          newMessages[newMessages.length - 1] = updatedLastMessage
          return newMessages
        })
      }
      
      // 清理引用
      readerRef.current = null
      abortControllerRef.current = null
      
      // v1.2: 在灵感模式下，流式响应结束后，如果右侧面板显示的是最新响应（rightPanelContentRef.current为null）
      // 注意：不需要设置rightPanelContent，因为ChatContainer会从lastAssistantContent获取内容
      // 这样可以保持打字机效果正常工作
      
      // v1.3: 在灵感模式下，刷新workDocuments列表；在普通模式下，只有在新对话第一次发送消息时才刷新列表
      if (isInspirationMode) {
        // 使用ref获取最新的currentWorkId，因为state可能还没更新
        const workId = currentWorkIdRef.current || currentWorkId
        if (workId) {
          await fetchWorkDocuments(workId)
        }
      } else {
      // 只有在新对话第一次发送消息时才刷新列表（更新标题）
      // 已存在的对话不需要刷新，因为用户已经在当前对话中
      if (isNewConversation) {
        await fetchConversations()
        }
      }
      
      // 注意：不再需要单独更新onAddToStory回调，因为在上面的代码中已经处理过了
    } catch (error) {
      // 如果是AbortError，说明请求被取消了（可能是切换对话），不需要处理错误
      if (error.name === 'AbortError') {
        console.log('Request aborted (conversation switched)')
        // 清理引用
        readerRef.current = null
        abortControllerRef.current = null
        // v3: 重置发送消息标记
        isSendingMessageRef.current = false
        return
      }
      
      console.error('Error:', error)
      // 只有在当前对话/创作时才更新错误消息
      // v1.3: 在灵感模式下检查currentWorkIdRef，否则检查currentConversationIdRef
      const currentTargetId = isInspirationMode ? currentWorkIdRef.current : currentConversationIdRef.current
      if (currentTargetId === targetId) {
        setMessages((prev) => {
          const newMessages = [...prev]
          const lastMessage = newMessages[newMessages.length - 1]
          if (lastMessage && lastMessage.role === 'assistant') {
            lastMessage.content = '抱歉，发生了错误：' + error.message
          }
          return newMessages
        })
      }
    } finally {
      // 只有在当前对话/创作时才重置加载状态
      // v1.3: 在灵感模式下检查currentWorkIdRef，否则检查currentConversationIdRef
      const currentTargetId = isInspirationMode ? currentWorkIdRef.current : currentConversationIdRef.current
      if (currentTargetId === targetId) {
        setIsLoading(false)
      }
      // 清理引用
      readerRef.current = null
      abortControllerRef.current = null
      // v3: 重置发送消息标记，允许useEffect加载历史消息
      isSendingMessageRef.current = false
    }
  }

  return (
    <div className="app">
      {isInspirationMode ? (
        <WorkHistory
          works={works}
          currentWorkId={currentWorkId}
          onSelectWork={handleSelectWork}
          isCollapsed={isHistoryCollapsed}
          onToggleCollapse={() => setIsHistoryCollapsed(!isHistoryCollapsed)}
          onRenameWork={handleRenameWork}
          onDeleteWork={handleDeleteWork}
          width={historyWidth}
          onWidthChange={(newWidth) => {
            setHistoryWidth(newWidth)
            localStorage.setItem('conversationHistoryWidth', newWidth.toString())
          }}
          onNewWork={handleNewWork}
        />
      ) : (
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
          stories={stories}
          onSelectStory={handleSelectStory}
          onEditStory={handleEditStory}
          onDeleteStory={handleDeleteStory}
        />
      )}
      <div className="main-content">
        <div className="main-content-chat">
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
            <div className="inspiration-mode-toggle">
              <label className="toggle-label">
                <input
                  type="checkbox"
                  checked={isInspirationMode}
                  onChange={(e) => setIsInspirationMode(e.target.checked)}
                  className="toggle-input"
                />
                <span className="toggle-slider"></span>
                <span className="toggle-text">灵感模式</span>
              </label>
            </div>
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
          isInspirationMode={isInspirationMode}
          currentWorkId={currentWorkId}
          isModifyOriginal={isModifyOriginal}
          onModifyOriginalChange={setIsModifyOriginal}
          onViewDocument={handleViewInspirationDocument}
          rightPanelContent={isInspirationMode ? rightPanelContent : null}
        />
        </div>
        {isInspirationMode && currentWorkId && (
          <WorkDocumentList
            documents={workDocuments}
            onNewDocument={handleNewDocument}
            onRenameDocument={handleRenameDocument}
            onDeleteDocument={handleDeleteDocument}
            onSelectDocument={handleSelectDocument}
          />
        )}
      </div>
        <StoryEditDialog
        isOpen={storyEditDialogOpen}
        content={storyEditContent}
        title={storyEditTitle}
        storyId={storyEditStoryId}
        onClose={() => {
          setStoryEditDialogOpen(false)
          setStoryEditContent('')
          setStoryEditTitle('')
          setStoryEditDocumentId(null)
          setStoryEditStoryId(null)
        }}
        onSave={handleSaveStory}
      />
      <StoryViewDialog
        isOpen={storyViewDialogOpen}
        story={selectedStory}
        onClose={() => {
          setStoryViewDialogOpen(false)
          setSelectedStory(null)
        }}
      />
      {/* v1.2: 不再使用InspirationDocumentView浮层，改为在右侧直接展示 */}
      {/* <InspirationDocumentView
        isOpen={inspirationDocumentViewOpen}
        content={inspirationDocumentContent}
        onClose={() => {
          setInspirationDocumentViewOpen(false)
          setInspirationDocumentContent('')
        }}
      /> */}
    </div>
  )
}

export default App

