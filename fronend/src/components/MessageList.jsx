import Message from './Message'
import './MessageList.css'

function MessageList({ messages, isLoading, enableTypewriter = true, onLoadMore, canLoadMore = false, isLoadingMore = false, isInspirationMode = false, isLastMessageComplete = false }) {
  if (messages.length === 0) {
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
      {messages.map((message, index) => {
        // 在灵感模式下，检查用户消息后是否有完成的AI响应
        const showCompleteIndicator = isInspirationMode && 
          message.role === 'user' && 
          index < messages.length - 1 && 
          messages[index + 1]?.role === 'assistant' &&
          (index + 1 === messages.length - 1 ? isLastMessageComplete : true)
        
        return (
          <div key={message.id}>
            <Message message={message} enableTypewriter={enableTypewriter} onAddToStory={message.onAddToStory} />
            {showCompleteIndicator && (
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
                <span>AI已返回内容</span>
              </div>
            )}
          </div>
        )
      })}
      {isLoading && messages[messages.length - 1]?.role === 'assistant' && (
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

