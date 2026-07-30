import opentype from 'opentype.js'

const FONT_URLS = {
  zh: '/fonts/FZ.TTF',
  latin: '/fonts/GO.TTF',
  arabic: '/fonts/ARIAL.TTF',
  MiSans: '/fonts/MiSans-Regular.ttf',
  FZTH: '/fonts/FZLTTHPRO--GB1-4_0.OTF',
  FZZH: '/fonts/FZLTZHUNHPRO--GB1-4_0.OTF',
  GOTHAM: '/fonts/GOTHAM-BOOK_0.OTF',
  FZCH: '/fonts/FZLTCHPRO--GB1-4_0.OTF',
}

/** @type {Record<string, opentype.Font>} */
const fontCache = {}

/** 自定义错误：字体缺失，用于阻断导出 */
export class FontMissingError extends Error {
  constructor(message) {
    super(message)
    this.name = 'FontMissingError'
  }
}

async function loadFonts() {
  const missing = Object.entries(FONT_URLS).filter(([k]) => !fontCache[k])
  if (!missing.length) return
  await Promise.all(missing.map(async ([role, url]) => {
    try {
      const resp = await fetch(url)
      if (!resp.ok) return
      const buffer = await resp.arrayBuffer()
      fontCache[role] = opentype.parse(buffer)
    } catch (e) {
      console.warn('[svgTextToPaths] 字体加载失败:', url, e)
    }
  }))
}

function parseCoord(val, fallback) {
  const n = parseFloat(val)
  return isNaN(n) ? fallback : n
}

/**
 * 收集 text 元素中的文本行（含 tspan 定位与 dy 偏移）
 *
 * SVG 规范：
 * - 有 y/dtspan 写了 y → 绝对定位换行
 * - 有 dy → 相对上一行偏移（大多数编辑器导出多行文本用的正是 dy）
 * - 没有 y 也没有 dy → 同一基线水平接续
 */
function collectRuns(textEl, baseSize) {
  const runs = []
  const ox = parseCoord(textEl.getAttribute('x'), 0)
  const oy = parseCoord(textEl.getAttribute('y'), 0)

  const tspans = textEl.querySelectorAll('tspan')
  if (tspans.length > 0) {
    let cx = ox, cy = oy
    for (const ts of tspans) {
      const txt = ts.textContent || ''
      const tx = ts.hasAttribute('x') ? parseCoord(ts.getAttribute('x'), cx) : cx
      // y 绝对定位 > dy 相对偏移 > 保持上一行
      let ty = cy
      if (ts.hasAttribute('y')) {
        ty = parseCoord(ts.getAttribute('y'), cy)
      } else if (ts.hasAttribute('dy')) {
        ty = cy + parseCoord(ts.getAttribute('dy'), 0)
      }
      const sz = ts.hasAttribute('font-size') ? parseCoord(ts.getAttribute('font-size'), baseSize) : baseSize
      if (txt.trim()) {
        runs.push({ text: txt, x: tx, y: ty, fontSize: sz })
      }
      // x 水平接续；y 推进（支持 dy 换行）
      cx = tx
      cy = ty
    }
  } else {
    // 无 tspan：按 \n 拆行，每行自动推进一个标准行高（120% 字号）
    const rawLines = (textEl.textContent || '').split('\n')
    const lineHeight = baseSize * 1.2
    for (let i = 0; i < rawLines.length; i++) {
      const line = rawLines[i]
      if (line.trim()) {
        runs.push({ text: line, x: ox, y: oy + i * lineHeight, fontSize: baseSize })
      }
    }
  }

  return runs
}

// ========== 字符缺失检测 ==========

/**
 * 检测文本中哪些字符在指定字体中不存在
 * @returns {string[]} 缺失的字符列表
 */
function getMissingChars(text, font) {
  if (!text || !font) return []
  const seen = new Set()
  const missing = new Set()
  for (const ch of text) {
    if (!ch.trim() || seen.has(ch)) continue
    seen.add(ch)
    const glyph = font.charToGlyph(ch)
    if (!glyph || glyph.name === '.notdef') {
      missing.add(ch)
    }
  }
  return [...missing]
}

/**
 * 从已加载的字体中，为整段文本找到覆盖率最高的字体
 * 优先使用 MiSans（字符覆盖率最高），其次按原有策略（阿拉伯文→ARIAL，中文→FZ，其余→GO）
 * @returns {{ font: opentype.Font | null, missing: string[] }}
 */
function findBestFont(text) {
  const fontKeys = Object.keys(fontCache)
  if (!fontKeys.length) return { font: null, missing: [] }

  // 排序：MiSans 优先（字符覆盖率最高），然后是中文系列、阿拉伯文、拉丁系列
  const sorted = [...fontKeys].sort((a, b) => {
    const order = { MiSans: 0, zh: 1, FZTH: 1, FZZH: 1, FZCH: 1, arabic: 2, latin: 3, GOTHAM: 3 }
    return (order[a] ?? 9) - (order[b] ?? 9)
  })

  for (const key of sorted) {
    const font = fontCache[key]
    if (!font) continue
    const missing = getMissingChars(text, font)
    if (missing.length === 0) return { font, missing: [] }
  }

  // 所有字体都有缺失 → 返回覆盖率最高的
  let best = null
  let leastMissing = Infinity
  for (const key of fontKeys) {
    const font = fontCache[key]
    if (!font) continue
    const missing = getMissingChars(text, font)
    if (missing.length < leastMissing) {
      leastMissing = missing.length
      best = { font, missing }
    }
  }
  return best || { font: null, missing: [] }
}

// ========== 路径生成辅助 ==========

/**
 * 用指定字体将文本转为 <path>，处理 text-anchor 修正
 */
function textToPath(runText, runX, runY, fontSize, font, anchor, fill, doc) {
  let adjustedX = runX
  if (anchor !== 'start') {
    const textWidth = font.getAdvanceWidth(runText, fontSize)
    if (anchor === 'middle') {
      adjustedX = runX - textWidth / 2
    } else if (anchor === 'end') {
      adjustedX = runX - textWidth
    }
  }
  const pathObj = font.getPath(runText, adjustedX, runY, fontSize)
  const pathEl = doc.createElementNS('http://www.w3.org/2000/svg', 'path')
  pathEl.setAttribute('d', pathObj.toPathData(2))
  pathEl.setAttribute('fill', fill)
  return pathEl
}

// ========== 主转曲函数 ==========

/**
 * 将 SVG 内所有 <text> 元素转为 <path>（参照 buchang 转曲方案）
 * 确保 svg2pdf 渲染时不会因为缺字体而乱码
 *
 * @throws {FontMissingError} 当存在字符在所有已加载字体中都找不到时抛出
 */
export async function convertSvgTextToPaths(svgEl, { skipEditable = false } = {}) {
  await loadFonts()
  const loadedKeys = Object.keys(fontCache)
  if (loadedKeys.length === 0) {
    throw new FontMissingError('所有字体文件加载失败，无法导出PDF。请检查 /fonts/ 目录下的字体文件是否存在。')
  }

  const doc = svgEl.ownerDocument
  const textEls = [...svgEl.querySelectorAll('text')]

  /** @type {Array<{ text: string, missing: string[] }>} */
  const missingEntries = []

  for (const textEl of textEls) {
    try {
      // 文字组件标记为可编辑，跳过转曲 → svg2pdf 用 Identity-H 字体渲染
      if (skipEditable && textEl.getAttribute('data-editable') === 'true') continue
      const baseSize = parseCoord(textEl.getAttribute('font-size'), 12)
      const fill = textEl.getAttribute('fill') || '#000'
      const fullText = textEl.textContent || ''
      const anchor = textEl.getAttribute('text-anchor') || 'start'

      if (!fullText.trim()) {
        textEl.remove()
        continue
      }

      // 为该 text 元素的完整文本找到最优字体
      const { font, missing } = findBestFont(fullText)
      if (!font) continue

      // 记录缺失字符
      if (missing.length > 0) {
        const preview = fullText.length > 30 ? fullText.slice(0, 30) + '...' : fullText
        missingEntries.push({ text: preview, missing })
      }

      const runs = collectRuns(textEl, baseSize)
      const group = doc.createElementNS('http://www.w3.org/2000/svg', 'g')

      for (const run of runs) {
        // 每个 run 可能包含不同语种的文本，重新找最优字体
        const runBest = findBestFont(run.text)
        const runFont = runBest.font || font

        try {
          const pathEl = textToPath(run.text, run.x, run.y, run.fontSize, runFont, anchor, fill, doc)
          group.appendChild(pathEl)
        } catch {
          // 单个 run 转曲失败，用整个 text 的最优字体回退
          try {
            const pathEl = textToPath(run.text, run.x, run.y, run.fontSize, font, anchor, fill, doc)
            group.appendChild(pathEl)
          } catch (e2) {
            console.warn('[svgTextToPaths] 单条文本转曲失败:', run.text.slice(0, 20), e2)
          }
        }
      }

      textEl.replaceWith(group)
    } catch (e) {
      console.warn('[svgTextToPaths] text 元素处理失败:', e)
    }
  }

  // 存在缺失字符 → 抛出错误，阻止导出
  if (missingEntries.length > 0) {
    const lines = missingEntries.map(({ text, missing }) =>
      `  "${text}" 缺少字符: [ ${missing.join(' ')} ]`
    )
    throw new FontMissingError(
      `以下文本中的字符在已加载的字体（${loadedKeys.join('、')}）中均不存在，导出已停止：\n${lines.join('\n')}\n\n请确保字体文件能覆盖这些字符，或更换包含这些字符的字体文件。`
    )
  }
}
