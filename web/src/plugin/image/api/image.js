import service from '@/utils/request'

export const createImage = (data) => {
  return service({
    url: '/image/createImage',
    method: 'post',
    data
  })
}

export const deleteImage = (params) => {
  return service({
    url: '/image/deleteImage',
    method: 'delete',
    params
  })
}

export const deleteImageByIds = (params) => {
  return service({
    url: '/image/deleteImageByIds',
    method: 'delete',
    params
  })
}

export const updateImage = (data) => {
  return service({
    url: '/image/updateImage',
    method: 'put',
    data
  })
}

export const findImage = (params) => {
  return service({
    url: '/image/findImage',
    method: 'get',
    params
  })
}

export const getImageList = (params) => {
  return service({
    url: '/image/getImageList',
    method: 'get',
    params
  })
}
