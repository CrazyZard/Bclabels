import service from '@/utils/request'

export const saveTemplate = (data) => {
  return service({
    url: '/label/saveTemplate',
    method: 'post',
    data
  })
}

export const loadTemplate = (params) => {
  return service({
    url: '/label/loadTemplate',
    method: 'get',
    params
  })
}

export const listTemplate = () => {
  return service({
    url: '/label/listTemplate',
    method: 'get'
  })
}

export const deleteTemplate = (params) => {
  return service({
    url: '/label/deleteTemplate',
    method: 'delete',
    params
  })
}

export const translateText = (params) => {
  return service({
    url: '/label/translateText',
    method: 'get',
    params
  })
}

export const publishTemplate = (params) => {
  return service({
    url: '/label/publishTemplate',
    method: 'post',
    params
  })
}

export const unpublishTemplate = (params) => {
  return service({
    url: '/label/unpublishTemplate',
    method: 'post',
    params
  })
}

export const listPublishedTemplate = () => {
  return service({
    url: '/label/listPublishedTemplate',
    method: 'get'
  })
}

export const downloadBatchTemplate = (params) => {
  return service({
    url: '/label/downloadBatchTemplate',
    method: 'get',
    params,
    responseType: 'blob'
  })
}
