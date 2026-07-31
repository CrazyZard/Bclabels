import { jsPDF } from 'jspdf'
import { svg2pdf } from 'svg2pdf.js'
import { convertSvgTextToPaths, FontMissingError } from './svgTextToPaths'

// ========== Constants ==========
const SVG_NS = 'http://www.w3.org/2000/svg'
const MM_PER_PT = 25.4 / 72

const FONT_SPECS = {
  'FZLTXIHJW--GB1-0': { url: '/fonts/FZ.TTF' },
  CenturyGothic:     { url: '/fonts/GO.TTF' },
  ArialMT:           { url: '/fonts/ARIAL.TTF' },
  'MiSans-Regular':  { url: '/fonts/MiSans-Regular.ttf' },
  'FZLTTHPRO--GB1-4':    { url: '/fonts/FZLTTHPRO--GB1-4_0.OTF' },
  'FZLTZHUNHPRO--GB1-4': { url: '/fonts/FZLTZHUNHPRO--GB1-4_0.OTF' },
  'Gotham-Book':         { url: '/fonts/GOTHAM-BOOK_0.OTF' },
  'FZLTCHPRO--GB1-4':    { url: '/fonts/FZLTCHPRO--GB1-4_0.OTF' },
}

// Map legacy short names to PostScript names (used by template data)
const FONT_PS_MAP = {
  FZ:     'FZLTXIHJW--GB1-0',
  GO:     'CenturyGothic',
  ARIAL:  'ArialMT',
  MiSans: 'MiSans-Regular',
  FZTH:   'FZLTTHPRO--GB1-4',
  FZZH:   'FZLTZHUNHPRO--GB1-4',
  GOTHAM: 'Gotham-Book',
  FZCH:   'FZLTCHPRO--GB1-4',
}
const DEFAULT_FONT = 'FZLTXIHJW--GB1-0'

// VFS keys (must match PostScript name = addFileToVFS first arg)
const VFS_KEYS = {
  zh:     'FZLTXIHJW--GB1-0',
  latin:  'CenturyGothic',
  arabic: 'ArialMT',
}

function fontPsName(shortName) {
  return FONT_PS_MAP[shortName] || shortName || DEFAULT_FONT
}

// ========== Font loading ==========
const fontBase64Cache = {}

function arrayBufferToBase64(buffer) {
  let binary = ''
  const bytes = new Uint8Array(buffer)
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return btoa(binary)
}

async function loadFontBase64(key) {
  if (fontBase64Cache[key]) return fontBase64Cache[key]
  const spec = FONT_SPECS[key]
  if (!spec) return null
  try {
    const resp = await fetch(spec.url)
    if (!resp.ok) {
      console.error(`[PDF] font fetch failed: ${spec.url} status=${resp.status}`)
      return null
    }
    const buffer = await resp.arrayBuffer()
    const base64 = arrayBufferToBase64(buffer)
    fontBase64Cache[key] = base64
    console.log(`[PDF] font loaded: ${key} (${spec.url}) size=${buffer.byteLength}`)
    return base64
  } catch (e) {
    console.error(`[PDF] font fetch error: ${spec.url}`, e)
    return null
  }
}

async function registerFont(doc, psName) {
  if (doc.getFontList()[psName]) return true
  const base64 = await loadFontBase64(psName)
  if (!base64) {
    console.warn(`[PDF] font load failed: ${psName}`)
    return false
  }
  try {
    // addFileToVFS: one call per font file (VFS key = PostScript name)
    doc.addFileToVFS(psName, base64)

    // addFont: register PRIMARY font name + ALL aliases (matching buchang)
    // The 1st arg (PostScript name) MUST match the VFS key; 2nd arg is the font name alias
    if (psName === VFS_KEYS.latin) {
      doc.addFont(psName, 'CenturyGothic', 'normal', 'Identity-H')
      doc.addFont(psName, 'Century Gothic', 'normal', 'Identity-H')
      doc.addFont(psName, 'GO', 'normal', 'Identity-H')
    } else if (psName === VFS_KEYS.zh) {
      doc.addFont(psName, 'FZLTXIHJW--GB1-0', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZLanTingHeiS-L-GB', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZ', 'normal', 'Identity-H')
    } else if (psName === VFS_KEYS.arabic) {
      doc.addFont(psName, 'ArialMT', 'normal', 'Identity-H')
      doc.addFont(psName, 'ARIAL', 'normal', 'Identity-H')
    } else if (psName === 'FZLTTHPRO--GB1-4') {
      doc.addFont(psName, 'FZLTTHPRO--GB1-4', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZLTTHPro', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZTH', 'normal', 'Identity-H')
    } else if (psName === 'FZLTZHUNHPRO--GB1-4') {
      doc.addFont(psName, 'FZLTZHUNHPRO--GB1-4', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZLTZHUNHPro', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZZH', 'normal', 'Identity-H')
    } else if (psName === 'FZLTCHPRO--GB1-4') {
      doc.addFont(psName, 'FZLTCHPRO--GB1-4', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZLTCHPro', 'normal', 'Identity-H')
      doc.addFont(psName, 'FZCH', 'normal', 'Identity-H')
    } else if (psName === 'Gotham-Book') {
      doc.addFont(psName, 'Gotham-Book', 'normal', 'Identity-H')
      doc.addFont(psName, 'Gotham Book', 'normal', 'Identity-H')
      doc.addFont(psName, 'GOTHAM', 'normal', 'Identity-H')
    } else {
      doc.addFont(psName, psName, 'normal', 'Identity-H')
    }

    doc.setFont(psName)
    doc.setFontSize(10)
    console.log(`[PDF] font registered: ${psName}`)
    return true
  } catch (e) {
    console.error(`[PDF] font register failed: ${psName}`, e)
    return false
  }
}

// ========== Utility ==========
async function rasterizeImage(url) {
  try {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    await new Promise((resolve, reject) => {
      img.onload = resolve
      img.onerror = () => reject(new Error('Image load failed'))
      img.src = url
    })
    const canvas = document.createElement('canvas')
    canvas.width = img.naturalWidth || 100
    canvas.height = img.naturalHeight || 100
    const ctx = canvas.getContext('2d')
    ctx.drawImage(img, 0, 0)
    return {
      w: canvas.width,
      h: canvas.height,
      dataURL: canvas.toDataURL('image/png'),
    }
  } catch (e) {
    console.error('[PDF] image load failed:', url, e)
    return null
  }
}

function isSvgUrl(url) {
  if (!url) return false
  return url.endsWith('.svg') || url.includes('/svg')
}

/** svg2pdf compatibility normalization */
function normalizeSvgForSvg2pdf(svgEl) {
  // currentColor → black (svg2pdf cannot resolve)
  for (const el of svgEl.querySelectorAll('[fill="currentColor"]')) {
    el.setAttribute('fill', '#000000')
  }
  for (const el of svgEl.querySelectorAll('[stroke="currentColor"]')) {
    el.setAttribute('stroke', '#000000')
  }
  // remove <style> blocks (svg2pdf cannot parse CSS), inline class styles
  const styleEls = svgEl.querySelectorAll('style')
  if (styleEls.length) {
    const rules = []
    for (const s of styleEls) {
      try {
        const sheet = s.sheet || s.styleSheet
        if (sheet) {
          for (const rule of sheet.cssRules || []) {
            if (rule.selectorText) {
              rules.push({ selector: rule.selectorText, style: rule.style.cssText })
            }
          }
        }
      } catch { /* cross-origin, ignore */ }
      s.remove()
    }
    // inline to matching elements
    if (rules.length) {
      for (const { selector, style: cssText } of rules) {
        const targets = svgEl.querySelectorAll(selector)
        for (const t of targets) {
          const existingStyle = t.getAttribute('style') || ''
          t.setAttribute('style', existingStyle + ';' + cssText)
        }
      }
    }
  }
  // clip-path / mask not supported by svg2pdf → remove
  for (const el of svgEl.querySelectorAll('[clip-path]')) {
    el.removeAttribute('clip-path')
  }
  for (const el of svgEl.querySelectorAll('[mask]')) {
    el.removeAttribute('mask')
  }
}

/** download SVG, curve text, return documentElement */
async function loadSvgWithTextAsPaths(url) {
  try {
    const resp = await fetch(url)
    if (!resp.ok) return null
    const svgText = await resp.text()
    const svgDoc = new DOMParser().parseFromString(svgText, 'image/svg+xml')
    const svgEl = svgDoc.documentElement
    if (!svgEl || svgEl.tagName !== 'svg') return null

    // compatibility normalization
    normalizeSvgForSvg2pdf(svgEl)

    // has <text> without embedded fonts → error, ask user to curve
    if (svgEl.querySelector('text')) {
      throw new FontMissingError(`SVG not curved: ${url} contains text elements. Please convert text to paths in Illustrator/Inkscape first.`)
    }

    await convertSvgTextToPaths(svgEl)
    return svgEl
  } catch (e) {
    if (e.name === 'FontMissingError') throw e
    console.warn('[PDF] SVG load/curve failed:', url, e)
    return null
  }
}

/** get SVG viewBox or natural size */
function getSvgSize(svgEl) {
  const vb = svgEl.getAttribute('viewBox')
  if (vb) {
    const parts = vb.split(/[\s,]+/).map(Number)
    return { w: parts[2] || 100, h: parts[3] || 100 }
  }
  return {
    w: parseFloat(svgEl.getAttribute('width')) || 100,
    h: parseFloat(svgEl.getAttribute('height')) || 100,
  }
}

// ========== SVG 构建 ==========

/**
 * Build a single-page composite SVG with all elements as <text>/<image>/<rect>
 */
async function buildPageSvg(elements, config, opts = {}) {
  const { labelWidth, labelHeight, headSeam, marginLR } = config
  const { doc, translateInfo, frontElements, isBack } = opts
  const contentWidth = labelWidth - 2 * marginLR

  // 创建主 SVG 容器（动态适配模板 viewBox，毫米制）
  const svg = document.createElementNS(SVG_NS, 'svg')
  svg.setAttribute('viewBox', `0 0 ${labelWidth} ${labelHeight}`)
  svg.setAttribute('width', `${labelWidth}mm`)
  svg.setAttribute('height', `${labelHeight}mm`)

  let y = headSeam
  for (const el of elements) {
    const width = Math.min(el.width || contentWidth, contentWidth)
    const height = el.height || 10

    if (el.type === 'text') {
      await appendTextToSvg(svg, el, marginLR, y, width, height, doc, translateInfo, frontElements, isBack)
    } else if (el.type === 'image') {
      await appendImageToSvg(svg, el, marginLR, y, width, height)
    } else if (el.type === 'table') {
      appendTableToSvg(svg, el, marginLR, y, width, height, doc)
    }

    y += height
  }

  return svg
}

// ========== Text → SVG ==========
const LANG_STYLE = {
  english: { font: 'CenturyGothic', dir: 'ltr' },
  russian: { font: 'CenturyGothic', dir: 'ltr' },
  arabic: { font: 'ArialMT', dir: 'ltr' },
  indonesian: { font: 'CenturyGothic', dir: 'ltr' },
}

/** split text into lines using jsPDF measurement */
function splitTextToLines(doc, text, fontKey, fontSize, maxWidth) {
  if (!text) return []
  try { doc.setFont(fontKey) } catch { /* measurement fallback */ }
  doc.setFontSize(fontSize)
  return doc.splitTextToSize(text, maxWidth)
}

/** create SVG <text> element (baseline positioning for opentype.js curve conversion)
 *  font-size uses mm unitless value for correct opentype.js scaling */
function createSvgText(text, x, baselineY, fontSizePt, fontFamily, fill, alignment, maxWidth, letterSpacingPt) {
  const textEl = document.createElementNS(SVG_NS, 'text')
  textEl.setAttribute('font-size', `${fontSizePt * MM_PER_PT}`) // pt → mm user units
  textEl.setAttribute('font-family', fontPsName(fontFamily))
  textEl.setAttribute('fill', fill || '#000000')
  textEl.setAttribute('text-anchor', alignment === 'center' ? 'middle' : alignment === 'right' ? 'end' : 'start')
  textEl.setAttribute('data-editable', 'true')
  if (letterSpacingPt) {
    textEl.setAttribute('letter-spacing', `${letterSpacingPt * MM_PER_PT}`)
  }

  const anchorX = alignment === 'center' ? x + maxWidth / 2
    : alignment === 'right' ? x + maxWidth
    : x

  textEl.setAttribute('x', anchorX)
  textEl.setAttribute('y', baselineY)
  textEl.textContent = text
  return textEl
}

/** baseline offset (top-of-text → baseline), mm */
function baselineOffset(fontSizePt) {
  return fontSizePt * MM_PER_PT * 0.8
}

/** text block total height */
function calcTextBlockHeight(lines, fontSize, lineHeight) {
  const lh = lineHeight || 1.0
  return lines.length * fontSize * MM_PER_PT * lh
}

/** valign start Y */
function calcValignY(y, containerH, blockH, valign) {
  if (valign === 'middle') return y + (containerH - blockH) / 2
  if (valign === 'bottom') return y + containerH - blockH
  return y
}

async function appendTextToSvg(svg, el, x, y, width, height, doc, translateInfo, frontElements, isBack) {
  const fontSize = el.fontSize || 5
  const valign = el.valign || 'middle'
  const lineHFactor = el.lineHeight || 1.0
  const lineH = fontSize * MM_PER_PT * lineHFactor // 单行高度 mm（含行间距）
  const fontKey = fontPsName(el.fontFamily)
  const alignment = el.alignment || 'left'

  // backside translation — match by key (Excel header), render all langKeys
  if (isBack && el.langKeys?.length && translateInfo?.lookup) {
    const frontEl = frontElements?.find(e => el.key ? e.key === el.key : e.id === el.id)
    const source = frontEl?.text
    if (source && source.trim() && source.trim() !== 'text') {
      appendBackTranslationSvg(svg, el, x, y, width, height, doc, translateInfo, source, fontSize, valign, lineH)
      return
    }
  }

  // front side / no-translation backside: plain or rich text
  const fill = el.color || '#000000'
  const html = el.html || ''
  const plainText = el.text || ''

  if (html && html !== '<br>') {
    appendRichTextToSvg(svg, el, x, y, width, height, valign, alignment, fill)
    return
  }

  if (!plainText) return

  const letterSpacingPt = el.letterSpacing || 0
  const lines = splitTextToLines(doc, plainText, fontKey, fontSize, width)
  const blockH = calcTextBlockHeight(lines, fontSize, lineHFactor)
  const topY = calcValignY(y, height, blockH, valign)
  const baseY = topY + baselineOffset(fontSize)

  for (let i = 0; i < lines.length; i++) {
    const textEl = createSvgText(lines[i], x, baseY + i * lineH, fontSize, fontKey, fill, alignment, width, letterSpacingPt)
    svg.appendChild(textEl)
  }
}

/** Parse rich HTML into segments and create SVG <text> with <tspan> children */
function appendRichTextToSvg(svg, el, x, y, width, height, valign, alignment, fill) {
  const html = el.html || ''
  const defaultFontSize = (el.fontSize || 5) * MM_PER_PT
  const defaultFontFamily = fontPsName(el.fontFamily || 'FZLTXIHJW--GB1-0')

  // parse HTML into segments: [{ text, fontSize, fontFamily, bold }]
  const parser = new DOMParser()
  const doc = parser.parseFromString(`<root>${html}</root>`, 'text/html')
  const root = doc.body

  function collectSegments(node, inheritedFs, inheritedFf, inheritedBold, inheritedLs) {
    const segments = []
    if (node.nodeType === 3) { // text node
      const t = node.textContent
      if (t) segments.push({ text: t, fontSize: inheritedFs, fontFamily: inheritedFf, bold: inheritedBold, letterSpacing: inheritedLs })
    } else if (node.nodeType === 1) {
      if (node.tagName === 'BR') {
        segments.push({ text: '\n', fontSize: inheritedFs, fontFamily: inheritedFf, bold: inheritedBold, letterSpacing: inheritedLs })
      } else {
        const fs = inheritedFs
        let ff = inheritedFf
        let bw = inheritedBold
        let lsp = inheritedLs
        // parse style
        const style = node.getAttribute('style') || ''
        if (style.includes('font-weight:bold') || style.includes('font-weight:700') || style.includes('font-weight: 700')) bw = true
        if (style.includes('font-weight:normal') || style.includes('font-weight:400')) bw = false

        // font-family from style — extract first value and convert to PostScript name
        const ffMatch = style.match(/font-family:\s*([^;]+)/)
        if (ffMatch) {
          const raw = ffMatch[1].replace(/['"]/g, '').split(',')[0].trim()
          ff = fontPsName(raw) || ff
        }
        // font-family from old font tag
        const faceAttr = node.getAttribute('face')
        if (faceAttr) ff = fontPsName(faceAttr) || ff

        // font-size from style (keep in pt)
        let fsPt = null
        const fsMatch = style.match(/font-size:\s*([^;]+)/)
        if (fsMatch) {
          const val = fsMatch[1].trim()
          if (val.endsWith('pt')) fsPt = parseFloat(val)
          else if (val.endsWith('px')) fsPt = parseFloat(val) * 0.75
        }

        // letter-spacing from style (parse pt → mm)
        const lsMatch = style.match(/letter-spacing:\s*([^;]+)/)
        if (lsMatch) {
          const val = lsMatch[1].trim()
          if (val.endsWith('pt')) lsp = parseFloat(val) * MM_PER_PT
          else if (val.endsWith('px')) lsp = parseFloat(val) * MM_PER_PT * 0.75
          else if (val === 'normal' || val === '0') lsp = null
        }

        const childFs = fsPt !== null ? fsPt * MM_PER_PT : fs
        // is bold node
        if (node.tagName === 'B' || node.tagName === 'STRONG') bw = true
        if (node.tagName === 'FONT' && !fsPt) {
          // old font size
          const sizeAttr = node.getAttribute('size')
          if (sizeAttr) {
            const sz = parseInt(sizeAttr)
            const sizeMap = { 1: 8, 2: 10, 3: 12, 4: 14, 5: 16, 6: 20, 7: 24 }
            const pt = sizeMap[sz] || (sz * 3.5)
            segments.push(...collectSegments(node, fs, ff, bw, lsp)) // won't work fully, let's handle differently
            // Actually, for old font tag we can't override size here. Let's use the style style.
          }
        }

        for (const child of node.childNodes) {
          segments.push(...collectSegments(child, childFs, ff, bw, lsp))
        }
      }
    }
    return segments
  }

  const segments = []
  for (const child of root.childNodes) {
    segments.push(...collectSegments(child, defaultFontSize, defaultFontFamily, false, null))
  }

  if (!segments.length) return

  // compute total width of all segments for width estimation
  let totalWidth = 0
  for (const seg of segments) {
    if (seg.text === '\n') continue
    totalWidth += seg.text.length * seg.fontSize * 0.55 // rough char width estimate
  }

  // simple wrapping: split on \n
  const lines = []
  let current = []
  for (const seg of segments) {
    if (seg.text === '\n') {
      lines.push(current)
      current = []
    } else {
      current.push(seg)
    }
  }
  if (current.length) lines.push(current)

  if (!lines.length) return

  const lineHFactor = el.lineHeight || 1.2
  const lineH = Math.max(...segments.map(s => s.fontSize)) * lineHFactor
  const blockHeight = lines.length * lineH
  const topY = valign === 'middle' ? y + (height - blockHeight) / 2
    : valign === 'bottom' ? y + height - blockHeight
    : y

  for (let li = 0; li < lines.length; li++) {
    const lineSeg = lines[li]
    if (!lineSeg.length) continue

    const textEl = document.createElementNS(SVG_NS, 'text')
    textEl.setAttribute('text-anchor', 'start')
    textEl.setAttribute('data-editable', 'true')
    textEl.setAttribute('fill', fill)

    const baseY = topY + baselineOffset(defaultFontSize / MM_PER_PT) + li * lineH
    textEl.setAttribute('y', baseY)

    let cx = x
    const lineW = lineSeg.reduce((w, s) => w + s.text.length * s.fontSize * 0.55, 0)
    if (alignment === 'center') cx = x + (width - lineW) / 2
    else if (alignment === 'right') cx = x + width - lineW

    textEl.setAttribute('x', cx)

    for (const seg of lineSeg) {
      const tsp = document.createElementNS(SVG_NS, 'tspan')
      tsp.setAttribute('font-size', seg.fontSize)
      tsp.setAttribute('font-family', seg.fontFamily)
      if (seg.bold) tsp.setAttribute('font-weight', 'bold')
      if (seg.letterSpacing) tsp.setAttribute('letter-spacing', seg.letterSpacing)
      tsp.textContent = seg.text
      textEl.appendChild(tsp)
    }

    svg.appendChild(textEl)
  }
}

/** backside translation text — render all langKeys stacked, separated by blank line */
function appendBackTranslationSvg(svg, el, x, y, width, height, doc, translateInfo, source, baseFontSize, valign, lineH) {
  const langs = el.langKeys
  if (!langs || !langs.length) return

  // collect per-language line groups
  const groups = []
  for (const lang of langs) {
    const translated = translateInfo.lookup(source, lang)
    const text = (translated && translated !== source) ? translated : source
    const style = LANG_STYLE[lang] || { font: 'CenturyGothic', dir: 'ltr' }
    const lines = splitTextToLines(doc, text, style.font, baseFontSize, width)
    const isRtl = style.dir === 'rtl'
    const align = isRtl ? 'right' : (el.alignment || 'left')
    groups.push({ lines, font: style.font, align, lang })
  }

  // total block height: sum of all line groups + blank-line gaps between groups
  let totalLines = 0
  for (const g of groups) totalLines += g.lines.length
  totalLines += Math.max(0, groups.length - 1)  // blank line between groups

  const blockH = totalLines * lineH
  const topY = calcValignY(y, height, blockH, valign)
  let cursorY = topY + baselineOffset(baseFontSize)

  for (let gi = 0; gi < groups.length; gi++) {
    const { lines, font, align } = groups[gi]
    for (let i = 0; i < lines.length; i++) {
      const textEl = createSvgText(lines[i], x, cursorY, baseFontSize, font, '#000000', align, width, 0)
      svg.appendChild(textEl)
      cursorY += lineH
    }
    // blank line gap between language groups
    if (gi < groups.length - 1) cursorY += lineH
  }
}

// ========== Image → SVG ==========
/** external image: SVG vector-first (fail if cannot embed), non-SVG rasterize */
async function appendImageToSvg(svg, el, x, y, width, height) {
  if (!el.src) return
  const url = el.src
  const align = el.alignment || 'left'
  const valign = el.valign || 'middle'

  if (isSvgUrl(url)) {
    const svgEl = await loadSvgWithTextAsPaths(url)
    if (!svgEl) throw new Error(`SVG vector embed failed: ${url}`)
    const svgSize = getSvgSize(svgEl)
    const scale = Math.min(width / svgSize.w, height / svgSize.h)
    const drawW = svgSize.w * scale
    const drawH = svgSize.h * scale

    let drawX = x
    if (align === 'center') drawX = x + (width - drawW) / 2
    else if (align === 'right') drawX = x + width - drawW

    let drawY = y
    if (valign === 'middle') drawY = y + (height - drawH) / 2
    else if (valign === 'bottom') drawY = y + height - drawH

    const group = document.createElementNS(SVG_NS, 'g')
    group.setAttribute('transform', `translate(${drawX}, ${drawY}) scale(${scale}, ${scale})`)
    while (svgEl.firstChild) {
      group.appendChild(svgEl.firstChild)
    }
    svg.appendChild(group)
    return
  }

  // non-SVG (PNG/JPEG): rasterize
  const raster = await rasterizeImage(url)
  if (!raster || !raster.w || !raster.h) return

  const scaleW = width / (raster.w * MM_PER_PT)
  const scaleH = height / (raster.h * MM_PER_PT)
  const scale = Math.min(scaleW, scaleH)
  const drawW = raster.w * MM_PER_PT * scale
  const drawH = raster.h * MM_PER_PT * scale

  let drawX = x
  if (align === 'center') drawX = x + (width - drawW) / 2
  else if (align === 'right') drawX = x + width - drawW

  let drawY = y
  if (valign === 'middle') drawY = y + (height - drawH) / 2
  else if (valign === 'bottom') drawY = y + height - drawH

  const imageEl = document.createElementNS(SVG_NS, 'image')
  imageEl.setAttribute('href', raster.dataURL)
  imageEl.setAttribute('x', drawX)
  imageEl.setAttribute('y', drawY)
  imageEl.setAttribute('width', drawW)
  imageEl.setAttribute('height', drawH)
  svg.appendChild(imageEl)
}

// ========== Table → SVG ==========
function appendTableToSvg(svg, el, x, y, width, height, doc) {
  const cells = el.cells
  if (!cells || !cells.length) return

  const colWidth = el.colWidth || 4.4
  const rowHeightVal = el.rowHeight || 2.2
  const rowCount = cells.length
  const colCount = Math.max(...cells.map(r => r.length), 0)
  const tableW = colWidth * colCount
  const tableH = rowHeightVal * rowCount

  let tableX = x
  if (el.alignment === 'center') tableX = x + (width - tableW) / 2
  else if (el.alignment === 'right') tableX = x + width - tableW

  let tableY = y
  if (el.valign === 'middle') tableY = y + (height - tableH) / 2
  else if (el.valign === 'bottom') tableY = y + height - tableH

  const showBorder = el.showBorder !== false
  const fontSize = el.fontSize || 5
  const fontKey = fontPsName(el.fontFamily)

  const tableGroup = document.createElementNS(SVG_NS, 'g')

  for (let ri = 0; ri < rowCount; ri++) {
    const row = cells[ri] || []
    for (let ci = 0; ci < colCount; ci++) {
      const cell = row[ci] || { value: '', textAlign: 'center' }
      const cx = tableX + ci * colWidth
      const cy = tableY + ri * rowHeightVal

      // border
      if (showBorder) {
        const rect = document.createElementNS(SVG_NS, 'rect')
        rect.setAttribute('x', cx)
        rect.setAttribute('y', cy)
        rect.setAttribute('width', colWidth)
        rect.setAttribute('height', rowHeightVal)
        rect.setAttribute('fill', 'none')
        rect.setAttribute('stroke', '#000000')
        rect.setAttribute('stroke-width', '0.3')
        tableGroup.appendChild(rect)
      }

      // text (multi-line support)
      const cellText = cell.value || ''
      const cellAlign = cell.textAlign || el.alignment || 'center'
      const lines = splitTextToLines(doc, cellText, fontKey, fontSize, colWidth - 0.5)
      const lineH = fontSize * MM_PER_PT
      const textBlockH = lines.length * lineH
      const textTopY = cy + (rowHeightVal - textBlockH) / 2
      const textBaseY = textTopY + baselineOffset(fontSize)

      for (let li = 0; li < lines.length; li++) {
        const textEl = document.createElementNS(SVG_NS, 'text')
        textEl.setAttribute('font-size', `${fontSize * MM_PER_PT}`)
        textEl.setAttribute('font-family', fontKey)
        textEl.setAttribute('fill', '#000000')

        const padX = 0.5
        const anchorX = cellAlign === 'center'
          ? cx + colWidth / 2
          : cellAlign === 'right'
            ? cx + colWidth - padX
            : cx + padX
        textEl.setAttribute('x', anchorX)
        textEl.setAttribute('y', textBaseY + li * lineH)
        textEl.setAttribute('text-anchor', cellAlign === 'center' ? 'middle' : cellAlign === 'right' ? 'end' : 'start')
        textEl.textContent = lines[li]
        tableGroup.appendChild(textEl)
      }
    }
  }

  svg.appendChild(tableGroup)
}

// ========== Core Export ==========

/** Normalize SVG font-family attributes to match registered PDF font names (matching buchang's prepareSvgForEditablePdf).
 *  svg2pdf reads font-family and calls doc.setFont() — the value must be an exact PostScript name from getFontList(). */
function normalizeSvgFontsForPdf(svgEl) {
  const texts = svgEl.querySelectorAll('text, tspan')
  for (const el of texts) {
    const attr = el.getAttribute('font-family')
    if (attr) {
      const normalized = fontPsName(attr.replace(/['"]/g, '').split(',')[0].trim())
      el.setAttribute('font-family', normalized)
    }
    // strip inline font-family style (svg2pdf prioritizes style over attribute)
    const style = el.getAttribute('style')
    if (style && /font-family/i.test(style)) {
      const cleaned = style.split(';').filter(p => !/^\s*font-family/i.test(p)).join(';')
      if (cleaned.trim()) el.setAttribute('style', cleaned)
      else el.removeAttribute('style')
    }
  }
}

/**
 * Export label PDF (Identity-H fonts + svg2pdf direct text rendering)
 *
 * Strategy:
 *   - text elements → keep <text>, svg2pdf renders with registered Identity-H fonts (editable in PDF)
 *   - table text → opentype.js to paths (no font dependency)
 *   - external SVG → vector-first embed (curve + normalize -> svg2pdf), fallback rasterize
 *   - external bitmap → rasterize embed
 *
 * @param {Object} cfg
 * @param {Array}  cfg.frontElements
 * @param {Array}  cfg.backElements
 * @param {Object} cfg.config          { labelWidth, labelHeight, headSeam, marginLR }
 * @param {Object} cfg.translateInfo   { lookup(source, lang), translateLangs }
 * @param {String} cfg.fileName
 */
export async function exportLabelPDF({
  frontElements,
  backElements,
  config,
  translateInfo,
  fileName = 'label.pdf',
  existingDoc = null,
}) {
  const { labelWidth, labelHeight } = config

  const isAppend = !!existingDoc
  const doc = existingDoc || new jsPDF({
    orientation: labelWidth > labelHeight ? 'landscape' : 'portrait',
    unit: 'mm',
    format: [labelWidth, labelHeight],
  })

  if (!isAppend) {
    // register fonts to jsPDF (Identity-H encoding for svg2pdf <text> rendering)
    const fontResults = await Promise.all(Object.keys(FONT_SPECS).map(async k => {
      const ok = await registerFont(doc, k)
      console.log(`[PDF] font ${k}: ${ok ? 'OK' : 'FAILED'}`)
      return { name: k, ok }
    }))
    const failedFonts = fontResults.filter(f => !f.ok)
    if (failedFonts.length) {
      console.error('[PDF] 字体加载失败:', failedFonts.map(f => f.name).join(', '))
    }
    const fontList = doc.getFontList()
    console.log('[PDF] registered fonts:', JSON.stringify(fontList))
    if (!fontList[VFS_KEYS.zh] || !fontList[VFS_KEYS.latin]) {
      console.error('[PDF] 关键字体未注册! zh=', Boolean(fontList[VFS_KEYS.zh]), 'latin=', Boolean(fontList[VFS_KEYS.latin]))
    }
  }

  const pages = [{ elements: frontElements, isBack: false }]
  if (backElements && backElements.length) {
    pages.push({ elements: backElements, isBack: true })
  }

  try {
    for (let pi = 0; pi < pages.length; pi++) {
      if (pi > 0 || isAppend) {
        doc.addPage([labelWidth, labelHeight], labelWidth > labelHeight ? 'landscape' : 'portrait')
      }

      const { elements, isBack } = pages[pi]
      const pageSvg = await buildPageSvg(elements, config, {
        doc,
        translateInfo: isBack ? translateInfo : null,
        frontElements: isBack ? frontElements : null,
        isBack,
      })

      // table/SVG text → paths; text elements → skip (keep <text> for svg2pdf)
      await convertSvgTextToPaths(pageSvg, { skipEditable: true })

      // normalize SVG font-family for svg2pdf compatibility (matching buchang's prepareSvgForEditablePdf)
      normalizeSvgFontsForPdf(pageSvg)

      // svg2pdf renders paths + text elements (with registered Identity-H fonts)
      await svg2pdf(pageSvg, doc, { x: 0, y: 0 })
    }

    return doc
  } catch (e) {
    console.error('[PDF] 导出失败:', e)
    throw e
  }
}
