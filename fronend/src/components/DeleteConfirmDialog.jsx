import { useState, useEffect } from 'react'
import './DeleteConfirmDialog.css'

function DeleteConfirmDialog({ isOpen, onConfirm, onCancel, conversationTitle }) {
  const [dontShowAgain, setDontShowAgain] = useState(false)

  useEffect(() => {
    // 检查localStorage中是否设置了不再提醒
    const saved = localStorage.getItem('dontShowDeleteConfirm')
    if (saved === 'true') {
      setDontShowAgain(true)
      // 如果设置了不再提醒，直接确认
      if (isOpen) {
        onConfirm()
      }
    }
  }, [isOpen, onConfirm])

  if (!isOpen) return null

  const handleConfirm = () => {
    if (dontShowAgain) {
      localStorage.setItem('dontShowDeleteConfirm', 'true')
    } else {
      localStorage.removeItem('dontShowDeleteConfirm')
    }
    onConfirm()
  }

  const handleCancel = () => {
    onCancel()
  }

  return (
    <div className="delete-dialog-overlay" onClick={handleCancel}>
      <div className="delete-dialog-content" onClick={(e) => e.stopPropagation()}>
        <div className="delete-dialog-icon">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
            <path
              d="M12 9v4m0 4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
        <h3 className="delete-dialog-title">删除对话</h3>
        <p className="delete-dialog-message">
          确定要删除对话 <span className="delete-dialog-title-highlight">"{conversationTitle}"</span> 吗？
          <br />
          此操作无法撤销。
        </p>
        <div className="delete-dialog-checkbox">
          <label>
            <input
              type="checkbox"
              checked={dontShowAgain}
              onChange={(e) => setDontShowAgain(e.target.checked)}
            />
            <span>不再提醒</span>
          </label>
        </div>
        <div className="delete-dialog-actions">
          <button className="delete-dialog-cancel-btn" onClick={handleCancel}>
            取消
          </button>
          <button className="delete-dialog-confirm-btn" onClick={handleConfirm}>
            删除
          </button>
        </div>
      </div>
    </div>
  )
}

export default DeleteConfirmDialog

