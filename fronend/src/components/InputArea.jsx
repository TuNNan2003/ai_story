import { useState, useRef, useEffect } from 'react'
import './InputArea.css'

function InputArea({ onSendMessage, isLoading }) {
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
    <div className="input-area">
      <form onSubmit={handleSubmit} className="input-form">
        <textarea
          ref={textareaRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入你的想法或问题..."
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
      </form>
    </div>
  )
}

export default InputArea

