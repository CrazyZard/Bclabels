import service from '@/utils/request'

export const uploadAndTranslate = (data) => {
  return service({
    url: '/translist/uploadAndTranslate',
    method: 'post',
    data,
    headers: { 'Content-Type': 'multipart/form-data' },
    donNotShowLoading: true
  })
}

export const deleteJob = (params) => {
  return service({
    url: '/translist/deleteJob',
    method: 'delete',
    params
  })
}

export const deleteJobByIds = (params) => {
  return service({
    url: '/translist/deleteJobByIds',
    method: 'delete',
    params
  })
}

export const findJob = (params) => {
  return service({
    url: '/translist/findJob',
    method: 'get',
    params
  })
}

export const getJobList = (params) => {
  return service({
    url: '/translist/getJobList',
    method: 'get',
    params
  })
}

export const retranslate = (params) => {
  return service({
    url: '/translist/retranslate',
    method: 'post',
    params
  })
}

export const exportExcel = (params) => {
  return service({
    url: '/translist/exportExcel',
    method: 'get',
    params,
    responseType: 'blob'
  })
}
