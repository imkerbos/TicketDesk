import request from '@/utils/request'
import type { Attachment } from '@/types/attachment'

/**
 * 上传附件
 */
export const uploadAttachment = (issueKey: string, file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<Attachment>(`/issues/${issueKey}/attachments`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 获取附件列表
 */
export const listAttachments = (issueKey: string) => {
  return request.get<Attachment[]>(`/issues/${issueKey}/attachments`)
}

/**
 * 删除附件
 */
export const deleteAttachment = (issueKey: string, attachmentId: number) => {
  return request.delete(`/issues/${issueKey}/attachments/${attachmentId}`)
}

/**
 * 获取附件下载URL
 */
export const getAttachmentDownloadUrl = (issueKey: string, attachmentId: number) => {
  return `/api/v1/issues/${issueKey}/attachments/${attachmentId}/download`
}
