const viewModules = import.meta.glob('../view/**/*.vue')
const pluginModules = import.meta.glob('../plugin/**/*.vue')

export const asyncRouterHandle = (asyncRouter) => {
  for (let i = asyncRouter.length - 1; i >= 0; i--) {
    const item = asyncRouter[i]
    // 先递归处理子路由
    if (item.children) {
      asyncRouterHandle(item.children)
    }
    // 将字符串 component 转为动态 import 函数
    if (item.component && typeof item.component === 'string') {
      item.meta.path = '/src/' + item.component
      if (item.component.split('/')[0] === 'view') {
        item.component = dynamicImport(viewModules, item.component)
      } else if (item.component.split('/')[0] === 'plugin') {
        item.component = dynamicImport(pluginModules, item.component)
      }
    }
    // 删除组件为 null 且没有子路由的无效项，防止 Vue Router 崩溃
    if (item.component === null && (!item.children || item.children.length === 0)) {
      console.warn(`[asyncRouter] 移除无效路由 "${item.name}"：找不到组件文件`)
      asyncRouter.splice(i, 1)
    }
  }
}

function dynamicImport(dynamicViewsModules, component) {
  const keys = Object.keys(dynamicViewsModules)
  const matchKeys = keys.filter((key) => {
    const k = key.replace('../', '')
    return k === component
  })
  const matchKey = matchKeys[0]

  if (!matchKey) {
    console.warn(`[asyncRouter] 找不到组件文件: ${component}，该菜单项将被跳过`)
    return null
  }

  return dynamicViewsModules[matchKey]
}
