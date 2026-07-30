import service from '@/utils/request'

export const createEntry = (data) => {
  return service({
    url: '/entry/createEntry',
    method: 'post',
    data
  })
}

export const deleteEntry = (params) => {
  return service({
    url: '/entry/deleteEntry',
    method: 'delete',
    params
  })
}

export const deleteEntryByIds = (params) => {
  return service({
    url: '/entry/deleteEntryByIds',
    method: 'delete',
    params
  })
}

export const updateEntry = (data) => {
  return service({
    url: '/entry/updateEntry',
    method: 'put',
    data
  })
}

export const findEntry = (params) => {
  return service({
    url: '/entry/findEntry',
    method: 'get',
    params
  })
}

export const getEntryList = (params) => {
  return service({
    url: '/entry/getEntryList',
    method: 'get',
    params
  })
}

export const importExcel = (data) => {
  return service({
    url: '/entry/importExcel',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const loadDictionary = (params) => {
  return service({
    url: '/entry/loadDictionary',
    method: 'get',
    params
  })
}
