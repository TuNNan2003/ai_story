import Message from './Message'
import './MessageList.css'

function MessageList({ messages, isLoading, enableTypewriter = true, onLoadMore, canLoadMore = false, isLoadingMore = false, isInspirationMode = false, isLastMessageComplete = false, onViewDocument }) {
  if (messages.length === 0) {
    return null
  }

  // 在灵感模式下，只显示用户消息
  const displayMessages = isInspirationMode 
    ? messages.filter(m => m.role === 'user')
    : messages

  if (displayMessages.length === 0) {
    return null
  }

  return (
    <div className="message-list">
      {canLoadMore && (
        <div className="load-more-container">
          {isLoadingMore ? (
            <div className="load-more-indicator">
              <div className="load-more-spinner">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <span className="load-more-text">加载中...</span>
            </div>
          ) : (
            <div className="load-more-hint">向上滚动加载更多</div>
          )}
        </div>
      )}
      {displayMessages.map((message, displayIndex) => {
        // 在灵感模式下，需要找到该用户消息对应的AI响应
        let assistantMessage = null
        let assistantIndex = -1
        
        if (isInspirationMode) {
          // 找到该用户消息在原始消息列表中的位置
          const originalIndex = messages.findIndex(m => m.id === message.id)
          // 检查下一条消息是否是AI响应
          if (originalIndex >= 0 && originalIndex < messages.length - 1) {
            const nextMessage = messages[originalIndex + 1]
            if (nextMessage.role === 'assistant') {
              assistantMessage = nextMessage
              assistantIndex = originalIndex + 1
            }
          }
        }
        
        // 在灵感模式下，检查用户消息后是否有完成的AI响应
        const showCompleteIndicator = isInspirationMode && 
          assistantMessage && 
          (assistantIndex === messages.length - 1 ? isLastMessageComplete : true)
        
        return (
          <div key={message.id}>
            <Message message={message} enableTypewriter={enableTypewriter} onAddToStory={message.onAddToStory} />
            {showCompleteIndicator && (
              <div className="message-complete-section">
                <div className="message-complete-indicator">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                    <path
                      d="M13 4L6 11L3 8"
                      stroke="rgba(16, 163, 127, 0.8)"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <span>我已完成创作，你可以提出后续问题或者要求进行改动</span>
                </div>
                {assistantMessage && onViewDocument && (
                  <div 
                    className="message-view-document-btn"
                    onClick={() => {
                      const documentId = assistantMessage.documentId || assistantMessage.id
                      if (documentId && assistantMessage.content) {
                        onViewDocument(documentId, assistantMessage.content)
                      }
                    }}
                  >
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                      <path
                        d="M14 2H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M10 2v4h4M4 10h8M4 14h8"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                    <span>查看此次创作内容</span>
                  </div>
                )}
              </div>
            )}
          </div>
        )
      })}
      {isLoading && !isInspirationMode && messages[messages.length - 1]?.role === 'assistant' && (
        <div className="typing-indicator">
          <span></span>
          <span></span>
          <span></span>
        </div>
      )}
    </div>
  )
}

export default MessageList

