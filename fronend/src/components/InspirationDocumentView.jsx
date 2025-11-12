import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import './InspirationDocumentView.css'

function InspirationDocumentView({ isOpen, content, onClose }) {
  if (!isOpen) return null

  return (
    <div className="inspiration-document-view-overlay" onClick={onClose}>
      <div className="inspiration-document-view" onClick={(e) => e.stopPropagation()}>
        <div className="inspiration-document-view-header">
          <h3 className="inspiration-document-view-title">此次创作内容</h3>
          <button
            className="inspiration-document-view-close-btn"
            onClick={onClose}
            title="关闭"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path
                d="M12 4l-8 8M4 4l8 8"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            </svg>
          </button>
        </div>
        <div className="inspiration-document-view-body">
          <div className="inspiration-document-view-content">
            {content ? (
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {content}
              </ReactMarkdown>
            ) : (
              <div className="inspiration-document-view-empty">暂无内容</div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default InspirationDocumentView

