import { useState } from 'react'
import './WorkDocumentList.css'

function WorkDocumentList({ documents, onNewDocument, onRenameDocument, onDeleteDocument, onSelectDocument }) {
  const [editingId, setEditingId] = useState(null)
  const [editTitle, setEditTitle] = useState('')

  const handleRenameClick = (e, doc) => {
    e.stopPropagation()
    setEditingId(doc.id)
    setEditTitle(doc.title)
  }

  const handleRenameSubmit = async (e, docId) => {
    e.stopPropagation()
    if (editTitle.trim() && editTitle.trim() !== documents.find(d => d.id === docId)?.title) {
      await onRenameDocument(docId, editTitle.trim())
    }
    setEditingId(null)
    setEditTitle('')
  }

  const handleRenameCancel = (e) => {
    e.stopPropagation()
    setEditingId(null)
    setEditTitle('')
  }

  const handleKeyDown = (e, docId) => {
    if (e.key === 'Enter') {
      handleRenameSubmit(e, docId)
    } else if (e.key === 'Escape') {
      handleRenameCancel(e)
    }
  }

  return (
    <div className="work-document-list">
      <div className="work-document-list-header">
        <h3 className="work-document-list-title">文档列表</h3>
        <button onClick={onNewDocument} className="new-document-button" title="新建文档">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path
              d="M8 1v14M1 8h14"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
        </button>
      </div>
      <div className="work-document-list-content">
        {documents.length === 0 ? (
          <div className="work-document-empty">暂无文档</div>
        ) : (
          documents.map((doc) => (
            <div
              key={doc.id}
              className="work-document-item"
              onClick={() => onSelectDocument && onSelectDocument(doc)}
            >
              {editingId === doc.id ? (
                <div className="work-document-item-edit" onClick={(e) => e.stopPropagation()}>
                  <input
                    type="text"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    onKeyDown={(e) => handleKeyDown(e, doc.id)}
                    onBlur={(e) => handleRenameSubmit(e, doc.id)}
                    className="work-document-item-edit-input"
                    autoFocus
                  />
                </div>
              ) : (
                <>
                  <div className="work-document-item-content">
                    <div className="work-document-item-title" title={doc.title}>{doc.title || '未命名文档'}</div>
                    <div className="work-document-item-time">
                      {new Date(doc.updated_at).toLocaleDateString('zh-CN')}
                    </div>
                  </div>
                  <div className="work-document-item-actions">
                    <button
                      className="work-document-item-rename-btn"
                      onClick={(e) => handleRenameClick(e, doc)}
                      title="重命名"
                    >
                      <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                        <path
                          d="M11.333 2.667a1.414 1.414 0 0 1 2 2L5.333 12l-2.666.667L3.333 10l8-8z"
                          stroke="currentColor"
                          strokeWidth="1.5"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                      </svg>
                    </button>
                    <button
                      className="work-document-item-delete-btn"
                      onClick={(e) => {
                        e.stopPropagation()
                        onDeleteDocument && onDeleteDocument(doc.id)
                      }}
                      title="删除"
                    >
                      <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                        <path
                          d="M4 4l8 8M12 4l-8 8"
                          stroke="currentColor"
                          strokeWidth="1.5"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                      </svg>
                    </button>
                  </div>
                </>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default WorkDocumentList

