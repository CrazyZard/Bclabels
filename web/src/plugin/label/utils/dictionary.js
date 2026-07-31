import { translateBatch as apiTranslateBatch } from '@/plugin/dict/api/entry'

/**
 * 前端翻译客户端：调用后端统一引擎，本地仅做结果缓存。
 * 保留中文换行与缩进对齐由后端保证。
 */

const cache = new Map() // `${dictName}\0${lang}\0${text}` → translated

function cacheKey(dictName, lang, text) {
  return `${dictName}\0${lang}\0${text}`
}

/** 清除缓存（切换字典时调用） */
export function clearDictionaryCache() {
  cache.clear()
}

/** 同步读取已缓存译文；未命中返回 null */
export function getCachedTranslation(dictName, chinese, lang) {
  if (!dictName || chinese == null || chinese === '' || !lang) return null
  const hit = cache.get(cacheKey(dictName, lang, chinese))
  return hit === undefined ? null : hit
}

/**
 * 同步查询（供 PDF 导出等已预取场景）
 * @param {string} dictName
 * @param {string} chinese
 * @param {string} lang
 */
export function translateText(dictName, chinese, lang) {
  if (!dictName || !chinese || !lang) return chinese || ''
  const cached = getCachedTranslation(dictName, chinese, lang)
  return cached !== null ? cached : chinese
}

/**
 * 批量向后端请求翻译并写入缓存
 * @param {string} dictName
 * @param {{ text: string, langs: string[] }[]} items
 */
export async function translateMany(dictName, items) {
  if (!dictName || !items?.length) return

  const needMap = new Map() // text → Set(langs)
  for (const it of items) {
    const text = it.text
    if (text == null || text === '' || !String(text).trim()) continue
    const langs = (it.langs || []).filter(Boolean)
    if (!langs.length) continue
    let set = needMap.get(text)
    if (!set) {
      set = new Set()
      needMap.set(text, set)
    }
    for (const lang of langs) {
      if (!cache.has(cacheKey(dictName, lang, text))) set.add(lang)
    }
  }

  const need = []
  for (const [text, langSet] of needMap) {
    if (langSet.size) need.push({ text, langs: [...langSet] })
  }
  if (!need.length) return

  const res = await apiTranslateBatch({ dictName, items: need })
  if (res.code !== 0) {
    console.error('[翻译] 批量失败:', res)
    throw new Error(res.msg || '批量翻译失败')
  }
  const list = res.data?.items || []
  for (const item of list) {
    const translations = item.translations || {}
    for (const [lang, translated] of Object.entries(translations)) {
      cache.set(cacheKey(dictName, lang, item.text), translated)
    }
  }
}

/**
 * 构造供 PDF 使用的 lookup
 */
export function makeTranslateLookup(dictName) {
  return (source, lang) => {
    const cached = getCachedTranslation(dictName, source, lang)
    let text = cached !== null ? cached : source
    text = stripBidiMarks(text)
    // PDF：阿语逻辑 %84.8 转成绘制用的 84.8%
    if (lang === 'arabic') {
      text = text.replace(/%(\d+(?:\.\d+)?)/g, '$1%')
    }
    return text
  }
}

/** 去掉双向控制符，便于前端用 HTML 重新包一层 */
export function stripBidiMarks(s) {
  return String(s || '').replace(/[\u200E\u200F\u202A-\u202E\u2066-\u2069\u2060]/g, '')
}

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

/**
 * 阿语预览：把 %84.8 / 84.8% 都规范成从左往右的 84.8% 显示
 *（后端阿语逻辑序为 %84.8，避免 Excel/双向里看起来像百分号在左）
 */
export function formatArabicDisplayHtml(text) {
  const plain = stripBidiMarks(text)
  const escaped = escapeHtml(plain)
  // 先处理逻辑序 %84.8，再处理普通 84.8%
  return escaped
    .replace(/%(\d+(?:\.\d+)?)/g, '<span dir="ltr" style="direction:ltr;unicode-bidi:bidi-override;display:inline">$1%</span>')
    .replace(/(?<![%\d])(\d+(?:\.\d+)?)%/g, '<span dir="ltr" style="direction:ltr;unicode-bidi:bidi-override;display:inline">$1%</span>')
}

/** @deprecated 已改为后端翻译，保留空实现避免旧调用报错 */
export async function fetchDictionary(dictName) {
  if (!dictName) return null
  return { dictName, entries: {} }
}
