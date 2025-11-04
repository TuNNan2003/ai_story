import Message from './Message'
import './MessageList.css'

function MessageList({ messages, isLoading, enableTypewriter = true }) {
  if (messages.length === 0) {
    return null
  }

  return (
    <div className="message-list">
      {messages.map((message) => (
        <Message key={message.id} message={message} enableTypewriter={enableTypewriter} />
      ))}
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

