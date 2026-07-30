import { getEntryList } from '@/plugin/dict/api/entry'

/**
 * 字典格式（参照 buchang 项目）
 * { languages: string[], entries: Record<中文, Record<语言key, 译文>> }
 */

const LANG_MAP = {
  english: 'english',
  russian: 'russian',
  arabic: 'arabic',
  indonesian: 'indonesian'
}

// 中文别名（同义词回退）
const CHINESE_ALIASES = {
  '成分': ['成份'],
  '成份': ['成分'],
  '面料': ['面布'],
  '里料': ['里布'],
}

let cachedDict = null
let cachedDictName = ''

/** 从后端加载完整字典（使用已有的分页接口，避免权限问题） */
export async function fetchDictionary(dictName) {
  if (cachedDict && cachedDictName === dictName) {
    console.log('[字典] 命中缓存:', dictName, Object.keys(cachedDict.entries).length, '条')
    return cachedDict
  }
  console.log('[字典] 开始加载:', dictName)
  try {
    const res = await getEntryList({ dictName, page: 1, pageSize: 99999 })
    console.log('[字典] API 返回:', res.code, 'data keys:', res.data ? Object.keys(res.data) : 'no data')
    if (res.code !== 0) { console.error('[字典] 加载失败:', res); return null }
    // GVA 分页返回: res.data = { list: [...], total: N, page: 1, pageSize: 99999 }
    const rows = res.data?.list || (Array.isArray(res.data) ? res.data : [])
    console.log('[字典] 获取到', rows.length, '条记录')
    const entries = {}
    const languages = new Set()
    for (const row of rows) {
      const entry = {}
      if (row.english) { entry.english = row.english; languages.add('english') }
      if (row.russian) { entry.russian = row.russian; languages.add('russian') }
      if (row.arabic) { entry.arabic = row.arabic; languages.add('arabic') }
      if (row.indonesian) { entry.indonesian = row.indonesian; languages.add('indonesian') }
      entries[row.chinese] = entry
    }
    console.log('[字典] 构建完成:', Object.keys(entries).length, '条词条, 样例:', Object.keys(entries).slice(0, 5))
    cachedDict = { languages: [...languages], entries }
    cachedDictName = dictName
    return cachedDict
  } catch(e) {
    console.error('[字典] 网络错误:', e)
    return null
  }
}

/** 清除缓存 */
export function clearDictionaryCache() {
  cachedDict = null
  cachedDictName = ''
}

/** 字典查找（含别名回退、括号变体） — 参照 buchang lookupTranslation */
function lookup(dict, chinese, field) {
  const trimmed = chinese.trim()
  if (!trimmed) return null

  // 构建候选 key 列表：原文、全角括号包裹、半角括号包裹、别名
  const keys = new Set([
    trimmed,
    `（${trimmed}）`,
    `(${trimmed})`,
    ...(CHINESE_ALIASES[trimmed] ?? []),
    ...(CHINESE_ALIASES[`（${trimmed}）`] ?? []),
  ])

  const queryHasParen = hasOuterParen(trimmed)
  for (const key of keys) {
    const direct = dict.entries[key]?.[field]?.trim()
    if (direct) return normalizeLookupValue(direct, queryHasParen)
    for (const alias of CHINESE_ALIASES[key] ?? []) {
      const value = dict.entries[alias]?.[field]?.trim()
      if (value) return normalizeLookupValue(value, queryHasParen)
    }
  }
  return null
}

function hasOuterParen(s) {
  const t = String(s || '').trim()
  if (t.length < 2) return false
  const first = t[0]
  const last = t[t.length - 1]
  return (first === '（' || first === '(') && (last === '）' || last === ')')
}

function stripOuterParen(s) {
  let t = String(s || '').trim()
  while (hasOuterParen(t)) {
    t = t.slice(1, -1).trim()
  }
  return t
}

function normalizeLookupValue(v, queryHasParen) {
  if (queryHasParen) return v
  return stripOuterParen(v)
}

/** 拆分中文文本（括号内容、中文块、非中文块） */
function splitText(text) {
  const re = /（[^）]*）|[\u4e00-\u9fff]+|[^\u4e00-\u9fff]+/g
  const m = text.match(re)
  return m || [text]
}

// 内联括号匹配
const WHOLE_PAREN_RE = /^[（(]([^（）()]+)[）)]$/
const INLINE_PAREN_RE = /[（(]([^（）()]+)[）)]/g
const HAS_PAREN_RE = /[（(][^（）()]+[）)]/

/** 半角括号包裹；译文已自带括号时先剥掉，避免 ((xxx)) */
function wrapInParen(translated) {
  return `(${stripOuterParen(translated)})`
}

export function translateText(dict, chinese, lang) {
  if (!dict || !chinese) return chinese
  const field = LANG_MAP[lang]
  if (!field) return chinese

  const trimmed = chinese.trim()
  if (!trimmed) return ''

  // 按行分割，逐行翻译后合并（保留换行）
  const lines = trimmed.split('\n')
  return lines.map(line => translateLine(dict, line, field)).join('\n')
}

function translateLine(dict, chinese, field) {
  const trimmed = chinese.trim()
  if (!trimmed) return ''

  // 1. 被全角括号完整包裹 → 翻译内容后用半角括号包裹
  const wholeParen = trimmed.match(WHOLE_PAREN_RE)
  if (wholeParen) {
    const inner = wholeParen[1].trim()
    const translated = lookup(dict, inner, field)
    if (translated) return wrapInParen(translated)
    return `[${inner}]`
  }

  // 2. 含内联括号 → 分段翻译
  if (HAS_PAREN_RE.test(trimmed)) {
    let result = ''
    let lastIndex = 0
    let match
    // 重置正则状态后逐个匹配
    const re = new RegExp(INLINE_PAREN_RE.source, 'g')
    while ((match = re.exec(trimmed)) !== null) {
      const before = trimmed.slice(lastIndex, match.index)
      if (before) result += translatePlain(dict, before, field)
      const inner = match[1].trim()
      const translated = lookup(dict, inner, field)
      result += translated ? wrapInParen(translated) : `[${inner}]`
      lastIndex = re.lastIndex
    }
    const tail = trimmed.slice(lastIndex)
    if (tail) result += translatePlain(dict, tail, field)
    return result
  }

  // 3. 纯文本 → 冒号键值对或直接查字典
  return translatePlain(dict, trimmed, field)
}

function translatePlain(dict, chinese, field) {
  const trimmed = chinese.trim()
  if (!trimmed) return ''

  // 冒号分隔的键值对：“面料：70%棉 30%腈纶”（参照 buchang translateKeyValueLine）
  const colonIdx = trimmed.search(/[：:]/)
  if (colonIdx > 0) {
    const label = trimmed.slice(0, colonIdx).trim()
    const value = trimmed.slice(colonIdx + 1).trim()
    const translatedLabel = lookup(dict, label, field)
    const displayLabel = translatedLabel ? `${translatedLabel}:` : `${label}:`
    const translatedValue = value ? translateCompositionValue(dict, value, field) ?? translatePlainText(dict, value, field) : ''
    return translatedValue ? `${displayLabel} ${translatedValue}` : displayLabel
  }

  const exact = lookup(dict, trimmed, field)
  if (exact) return exact

  // 成分翻译
  const compResult = translateCompositionValue(dict, trimmed, field)
  if (compResult !== null) return compResult

  return translatePlainText(dict, trimmed, field)
}

/** 纯文本逐词翻译 */
function translatePlainText(dict, chinese, field) {
  const parts = splitText(chinese)
  return parts.map(p => {
    if (!/[\u4e00-\u9fff]/.test(p)) return p
    const t = lookup(dict, p.trim(), field)
    return t !== null ? t : p
  }).join('')
}

/**
 * 成分值翻译：匹配 "百分比% 中文材质名" 模式（参照 buchang translateCompositionValue）
 * 示例： "70.3%棉 29.7%腈纶" → "70.3% Cotton 29.7% Acrylic"
 * 返回 null 表示不是成分格式，需要上层继续处理
 */
function translateCompositionValue(dict, value, field) {
  const pctNameRe = /(\d+(?:\.\d+)?)\s*[%％]\s*([\u4e00-\u9fff][\u4e00-\u9fff\w]*)/g

  let hasMatch = false
  const result = value.replace(pctNameRe, (_match, pct, name) => {
    hasMatch = true
    const t = lookup(dict, name.trim(), field)
    if (t && !t.startsWith('[')) {
      return `${pct}% ${t}`
    }
    return `${pct}%${name.trim()}`
  })

  if (hasMatch) return result

  // 纯百分比值（如 "90％" → "90%"）
  const pctOnlyRe = /(\d+(?:\.\d+)?)\s*[%％]/g
  const pctOnlyResult = value.replace(pctOnlyRe, '$1%')
  if (pctOnlyResult !== value) return pctOnlyResult

  // 无百分比模式：不是成分格式，返回 null
  return null
}
