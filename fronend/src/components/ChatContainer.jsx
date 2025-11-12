import { useRef, useEffect, useState } from 'react'
import MessageList from './MessageList'
import EnhancedInputArea from './EnhancedInputArea'
import InspirationMarkdownView from './InspirationMarkdownView'
import './ChatContainer.css'

function ChatContainer({ messages, onSendMessage, isLoading, models, selectedModel, onModelChange, enableTypewriter = true, onLoadMore, canLoadMore, isLoadingMore, shouldScrollToBottom = false, isInspirationMode = false, currentWorkId = null, isModifyOriginal = false, onModifyOriginalChange, onViewDocument }) {
  const containerRef = useRef(null)
  const [prevScrollHeight, setPrevScrollHeight] = useState(0)
  const [prevScrollTop, setPrevScrollTop] = useState(0)
  const [isUserScrolling, setIsUserScrolling] = useState(false)
  const scrollTimeoutRef = useRef(null)
  const isTriggeredRef = useRef(false)
  const lastMessageIdsRef = useRef([])
  const isRestoringScrollRef = useRef(false)

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    // 如果正在加载更多历史消息，不处理自动滚动
    if (isLoadingMore) {
      return
    }

    // 如果正在恢复滚动位置，不自动滚动
    if (isRestoringScrollRef.current) {
      return
    }

    const currentMessageIds = messages.map(m => m.id)
    const lastMessageIds = lastMessageIdsRef.current

    // 判断是否在列表末尾添加了新消息（而不是在开头加载历史消息）
    let isNewMessageAtEnd = false
    if (currentMessageIds.length > lastMessageIds.length) {
      // 检查是否在末尾添加（新消息的ID不在旧列表中）
      const newIds = currentMessageIds.filter(id => !lastMessageIds.includes(id))
      if (newIds.length > 0) {
        // 检查新消息是否在列表末尾
        const lastNewId = newIds[newIds.length - 1]
        const lastCurrentId = currentMessageIds[currentMessageIds.length - 1]
        if (lastNewId === lastCurrentId) {
          // 新消息在末尾，说明是添加新消息，不是加载历史
          isNewMessageAtEnd = true
        }
      }
    }

    // 检查是否是流式响应更新（消息ID列表没有变化，但最后一条消息的内容在更新）
    // 或者最后一条消息是AI响应且正在加载中
    const isStreamingUpdate = currentMessageIds.length === lastMessageIds.length && 
                              currentMessageIds.length > 0 &&
                              messages[messages.length - 1]?.role === 'assistant' &&
                              (isLoading || messages[messages.length - 1]?.content.length > 0)

    // 更新消息ID列表
    lastMessageIdsRef.current = currentMessageIds

    // 检查是否在底部附近（距离底部100px以内）
    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 100

    // 如果是在末尾添加新消息，或者流式响应更新，则自动滚动到底部
    // 这样确保用户发送新消息或AI响应时，页面会自动跟随
    if (isNewMessageAtEnd) {
      // 如果是新消息添加到末尾，总是滚动到底部（用户发送消息时）
      requestAnimationFrame(() => {
        if (containerRef.current) {
          containerRef.current.scrollTop = containerRef.current.scrollHeight
          // 再次尝试，确保滚动到底部
          requestAnimationFrame(() => {
            if (containerRef.current) {
              containerRef.current.scrollTop = containerRef.current.scrollHeight
            }
          })
        }
      })
    } else if (isStreamingUpdate) {
      // 如果是流式响应更新，且用户在底部附近或没有主动滚动，则滚动到底部
      if (isNearBottom || !isUserScrolling) {
        // 使用requestAnimationFrame确保DOM更新完成后再滚动
        requestAnimationFrame(() => {
          if (containerRef.current) {
            containerRef.current.scrollTop = containerRef.current.scrollHeight
            // 流式响应时，可能需要多次尝试才能完全滚动到底部
            requestAnimationFrame(() => {
              if (containerRef.current) {
                containerRef.current.scrollTop = containerRef.current.scrollHeight
                // 流式响应时，内容不断增长，需要持续滚动
                setTimeout(() => {
                  if (containerRef.current) {
                    containerRef.current.scrollTop = containerRef.current.scrollHeight
                  }
                }, 50)
              }
            })
          }
        })
      }
    }
  }, [messages, isLoadingMore, isUserScrolling, isLoading])

  // 处理滚动事件，检测用户是否滚动到顶部
  useEffect(() => {
    const container = containerRef.current
    if (!container || !canLoadMore) return

    const handleScroll = () => {
      // 记录当前滚动位置
      const currentScrollTop = container.scrollTop
      setPrevScrollTop(currentScrollTop)

      // 检测用户是否在滚动
      // 只有在用户主动滚动时才设置为true（而不是系统自动滚动）
      // 通过检查滚动位置是否在底部附近来判断是否是用户主动滚动
      const isNearBottomForScroll = container.scrollHeight - currentScrollTop - container.clientHeight < 100
      if (!isNearBottomForScroll) {
        // 用户滚动到了非底部位置，说明是用户主动滚动
        setIsUserScrolling(true)
        if (scrollTimeoutRef.current) {
          clearTimeout(scrollTimeoutRef.current)
        }
        scrollTimeoutRef.current = setTimeout(() => {
          setIsUserScrolling(false)
        }, 500) // 延长超时时间，避免快速滚动时误判
      } else {
        // 用户在底部附近，可能是自动滚动，重置用户滚动状态
        setIsUserScrolling(false)
        if (scrollTimeoutRef.current) {
          clearTimeout(scrollTimeoutRef.current)
        }
      }

      // 如果滚动到顶部附近（距离顶部50px以内），自动加载更多
      // 使用isTriggeredRef防止重复触发
      if (currentScrollTop <= 50 && canLoadMore && !isLoadingMore && !isTriggeredRef.current) {
        isTriggeredRef.current = true
        const scrollHeightBefore = container.scrollHeight
        setPrevScrollHeight(scrollHeightBefore)
        onLoadMore()
        // 加载完成后重置标志
        setTimeout(() => {
          isTriggeredRef.current = false
        }, 1000)
      }
    }

    container.addEventListener('scroll', handleScroll)
    return () => {
      container.removeEventListener('scroll', handleScroll)
      if (scrollTimeoutRef.current) {
        clearTimeout(scrollTimeoutRef.current)
      }
    }
  }, [canLoadMore, isLoadingMore, onLoadMore])

  // 当加载更多完成后，保持滚动位置
  useEffect(() => {
    if (containerRef.current && isLoadingMore === false && prevScrollHeight > 0 && prevScrollTop > 0) {
      const container = containerRef.current
      const newScrollHeight = container.scrollHeight
      const scrollDifference = newScrollHeight - prevScrollHeight
      
      // 设置标志，防止自动滚动逻辑干扰
      isRestoringScrollRef.current = true
      
      // 调整滚动位置，保持用户看到的顶部内容不变
      // 使用requestAnimationFrame确保DOM更新完成
      requestAnimationFrame(() => {
        if (containerRef.current) {
          containerRef.current.scrollTop = prevScrollTop + scrollDifference
          
          // 重置标志，使用setTimeout确保滚动完成后才重置
          setTimeout(() => {
            isRestoringScrollRef.current = false
            setPrevScrollHeight(0)
            setPrevScrollTop(0)
          }, 100)
        }
      })
    }
  }, [isLoadingMore, prevScrollHeight, prevScrollTop])

  // 处理初始加载或需要滚动到底部的情况
  useEffect(() => {
    if (shouldScrollToBottom && containerRef.current && messages.length > 0) {
      const scrollToBottom = () => {
        const container = containerRef.current
        if (!container) return
        
        // 强制滚动到底部
        container.scrollTop = container.scrollHeight
        
        // 多次尝试确保滚动到底部（处理异步渲染的情况）
        requestAnimationFrame(() => {
          if (containerRef.current) {
            containerRef.current.scrollTop = containerRef.current.scrollHeight
            // 再次尝试，确保完全滚动到底部
            requestAnimationFrame(() => {
              if (containerRef.current) {
                containerRef.current.scrollTop = containerRef.current.scrollHeight
              }
            })
          }
        })
      }
      
      // 使用多重延迟确保DOM完全渲染
      // 第一次延迟：等待DOM更新
      requestAnimationFrame(() => {
        // 第二次延迟：确保内容完全渲染
        setTimeout(() => {
          scrollToBottom()
          // 再次延迟尝试，确保完全渲染后滚动
          setTimeout(() => {
            scrollToBottom()
          }, 100)
        }, 100)
      })
    }
  }, [shouldScrollToBottom, messages])

  // 处理首次加载对话时的滚动（当消息从空变为有内容时）
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    // 如果消息为空，重置消息ID列表，以便下次加载时能正确识别为首次加载
    if (messages.length === 0) {
      lastMessageIdsRef.current = []
      return
    }

    const currentMessageIds = messages.map(m => m.id)
    const lastMessageIds = lastMessageIdsRef.current

    // 如果是首次加载（从空消息列表变为有消息，且lastMessageIds为空）
    const isFirstLoad = lastMessageIds.length === 0 && currentMessageIds.length > 0

    // 如果正在加载更多历史消息，不处理首次加载滚动
    if (isLoadingMore) {
      // 但仍需要更新消息ID列表
      lastMessageIdsRef.current = currentMessageIds
      return
    }

    // 如果正在恢复滚动位置，不处理首次加载滚动
    if (isRestoringScrollRef.current) {
      // 但仍需要更新消息ID列表
      lastMessageIdsRef.current = currentMessageIds
      return
    }

    // 如果是首次加载，应该滚动到底部
    if (isFirstLoad) {
      const scrollToBottom = () => {
        if (containerRef.current) {
          const container = containerRef.current
          container.scrollTop = container.scrollHeight
        }
      }
      
      // 使用多重延迟确保DOM完全渲染
      requestAnimationFrame(() => {
        setTimeout(() => {
          scrollToBottom()
          // 再次尝试，确保完全渲染后滚动
          requestAnimationFrame(() => {
            scrollToBottom()
            // 最后一次尝试，确保滚动到底部
            setTimeout(() => {
              scrollToBottom()
            }, 100)
          })
        }, 150)
      })
    }

    // 更新消息ID列表（放在最后，确保首次加载判断在更新之前）
    lastMessageIdsRef.current = currentMessageIds
  }, [messages, isLoadingMore])

  // 获取最后一条AI响应的内容（用于右侧Markdown渲染）
  const lastAssistantMessage = messages.filter(m => m.role === 'assistant').slice(-1)[0]
  const lastAssistantContent = lastAssistantMessage?.content || ''
  const isLastMessageComplete = !isLoading && lastAssistantMessage && lastAssistantMessage.content.length > 0

  // 在灵感模式下，如果还没有对话，显示初始界面
  const hasConversation = messages.length > 0

  return (
    <div className={`chat-container ${isInspirationMode && hasConversation ? 'inspiration-mode' : ''}`}>
      {isInspirationMode && hasConversation ? (
        <div className="inspiration-layout">
          <div className="inspiration-left">
            <div className="chat-messages" ref={containerRef}>
              <MessageList
                messages={messages}
                isLoading={isLoading}
                enableTypewriter={false}
                onLoadMore={onLoadMore}
                canLoadMore={canLoadMore}
                isLoadingMore={isLoadingMore}
                isInspirationMode={true}
                isLastMessageComplete={isLastMessageComplete}
                onViewDocument={onViewDocument}
              />
            </div>
            <EnhancedInputArea
              onSendMessage={onSendMessage}
              isLoading={isLoading}
              models={models}
              selectedModel={selectedModel}
              onModelChange={onModelChange}
              hasMessages={messages.length > 0}
              isInspirationMode={true}
              isModifyOriginal={isModifyOriginal}
              onModifyOriginalChange={onModifyOriginalChange}
            />
          </div>
          <div className="inspiration-right">
            <InspirationMarkdownView
              content={lastAssistantContent}
              isLoading={isLoading}
            />
          </div>
        </div>
      ) : (
        <>
          <div className="chat-messages" ref={containerRef}>
            <MessageList
              messages={messages}
              isLoading={isLoading}
              enableTypewriter={enableTypewriter}
              onLoadMore={onLoadMore}
              canLoadMore={canLoadMore}
              isLoadingMore={isLoadingMore}
            />
          </div>
          <EnhancedInputArea
            onSendMessage={onSendMessage}
            isLoading={isLoading}
            models={models}
            selectedModel={selectedModel}
            onModelChange={onModelChange}
            hasMessages={messages.length > 0}
          />
        </>
      )}
    </div>
  )
}

export default ChatContainer

