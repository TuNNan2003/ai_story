import MarkdownRenderer from './MarkdownRenderer'
import './Message.css'

function Message({ message }) {
  const isUser = message.role === 'user'

  return (
    <div className={`message ${isUser ? 'message-user' : 'message-assistant'}`}>
      <div className="message-content">
        <div className="message-text">
          {isUser ? (
            <div className="user-text">{message.content}</div>
          ) : (
            <MarkdownRenderer text={message.content} speed={20} />
          )}
        </div>
      </div>
    </div>
  )
}

export default Message

