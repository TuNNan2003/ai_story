import { useState, useRef, useEffect } from 'react'
import './EnhancedInputArea.css'

function EnhancedInputArea({ onSendMessage, isLoading, models, selectedModel, onModelChange, hasMessages, isInspirationMode = false, isModifyOriginal = false, onModifyOriginalChange }) {
  const [input, setInput] = useState('')
  const textareaRef = useRef(null)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = textareaRef.current.scrollHeight + 'px'
    }
  }, [input])

  const handleSubmit = (e) => {
    e.preventDefault()
    if (input.trim() && !isLoading) {
      onSendMessage(input.trim())
      setInput('')
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
      }
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!isLoading && input.trim()) {
        handleSubmit(e)
      }
    }
  }

  return (
    <div className={`enhanced-input-area ${hasMessages ? 'has-messages' : 'no-messages'}`}>
      {!hasMessages && (
        <div className="input-welcome-text">开启一个故事</div>
      )}
      <form onSubmit={handleSubmit} className="enhanced-input-form">
        <div className="input-top-bar">
          <select
            value={selectedModel}
            onChange={(e) => onModelChange(e.target.value)}
            className="model-select-inline"
            disabled={isLoading}
          >
            {models.map((model) => (
              <option key={model.id} value={model.id}>
                {model.name}
              </option>
            ))}
          </select>
          {isInspirationMode && hasMessages && (
            <label className="modify-original-checkbox">
              <input
                type="checkbox"
                checked={isModifyOriginal}
                onChange={(e) => onModifyOriginalChange && onModifyOriginalChange(e.target.checked)}
                disabled={isLoading}
              />
              <span>在原文上修改</span>
            </label>
          )}
        </div>
        <div className="input-bottom-bar">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={hasMessages ? "输入你的想法或问题..." : ""}
            rows={1}
            className="input-textarea"
          />
          <button
            type="submit"
            disabled={!input.trim() || isLoading}
            className="input-button"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M.5 1.163L1.31.463 13.221 8.04 1.31 15.617l-.81.7L15 8.04v.96l-14.5-7.837Z"
                fill="currentColor"
              />
            </svg>
          </button>
        </div>
      </form>
    </div>
  )
}

export default EnhancedInputArea

