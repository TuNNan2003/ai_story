import { useState, useEffect, useRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import './InspirationMarkdownView.css'

function InspirationMarkdownView({ content, isLoading }) {
  const [displayedContent, setDisplayedContent] = useState('')
  const prevContentRef = useRef('')
  const currentIndexRef = useRef(0)
  const timeoutIdRef = useRef(null)

  useEffect(() => {
    if (!content) {
      setDisplayedContent('')
      prevContentRef.current = ''
      currentIndexRef.current = 0
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
      return
    }

    // 如果内容在增长（流式更新），继续打字新部分
    if (content.startsWith(prevContentRef.current)) {
      const newTextStart = prevContentRef.current.length
      // 如果当前显示的内容索引小于新内容的起始位置，更新到新内容的起始位置
      if (currentIndexRef.current < newTextStart) {
        currentIndexRef.current = newTextStart
        setDisplayedContent(content.slice(0, currentIndexRef.current))
      }
      // 注意：不要在这里更新prevContentRef.current，它应该保持为上一次完整显示的内容
      // 只有在打字完成或内容完全显示时才更新prevContentRef.current
    } else if (content === prevContentRef.current) {
      // 内容完全相同，不需要重新开始
      return
    } else {
      // 新内容，重新开始
      // v1.2: 如果新内容不是流式更新（isLoading为false），立即显示完整内容
      if (!isLoading) {
        setDisplayedContent(content)
        currentIndexRef.current = content.length
        prevContentRef.current = content
        if (timeoutIdRef.current) {
          clearTimeout(timeoutIdRef.current)
          timeoutIdRef.current = null
        }
        return
      }
      // 否则重新开始打字机效果
      setDisplayedContent('')
      currentIndexRef.current = 0
      prevContentRef.current = ''
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
    }

    // 继续打字
    const type = () => {
      if (currentIndexRef.current < content.length) {
        setDisplayedContent(content.slice(0, currentIndexRef.current + 1))
        currentIndexRef.current++
        timeoutIdRef.current = setTimeout(type, 20) // 打字机速度
      } else {
        // 打字完成，更新prevContentRef
        prevContentRef.current = content
        timeoutIdRef.current = null
      }
    }

    // 如果还有未显示的文本，开始打字
    if (currentIndexRef.current < content.length && !timeoutIdRef.current) {
      type()
    } else if (currentIndexRef.current >= content.length) {
      // 文本已完整显示
      setDisplayedContent(content)
      prevContentRef.current = content
    }

    return () => {
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
    }
  }, [content, isLoading])

  // 当内容加载完成时，立即显示完整内容
  useEffect(() => {
    if (!isLoading && content && currentIndexRef.current < content.length) {
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
      setDisplayedContent(content)
      currentIndexRef.current = content.length
      prevContentRef.current = content
    }
  }, [isLoading, content])

  return (
    <div className="inspiration-markdown-view">
      <div className="inspiration-markdown-content">
        {displayedContent ? (
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {displayedContent}
          </ReactMarkdown>
        ) : (
          <div className="inspiration-markdown-empty">
            {isLoading ? '等待AI响应...' : '暂无内容'}
          </div>
        )}
      </div>
    </div>
  )
}

export default InspirationMarkdownView

