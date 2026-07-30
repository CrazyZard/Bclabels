---
name: label-pdf-export
description: "标签编辑器 PDF 导出规范：SVG 合成 + 整页转曲 + svg2pdf 高精矢量引擎"
version: 2.0.0
author: BcLabels
tags: [pdf, svg, svg2pdf, opentype, label, export]
---

# 标签编辑器 PDF 导出规范 (全域转曲版)

本规范服务于**可自定义任意尺寸（mm）**的标签设计器。核心策略：将所有画布元素（文字、表格、图片）组装为一张复合 SVG，通过 `opentype.js` 对整张 SVG 内的所有 `<text>` 统一转曲为 `<path>`，再由 `svg2pdf.js` 渲染为 PDF。**转曲后零字体依赖，ai 中随时释放剪切蒙版即可编辑。**

---

## 📦 文件结构

```
web/src/plugin/label/utils/
├── pdfExport.js          # 主导出入口：buildPageSvg → convertSvgTextToPaths → svg2pdf
├── svgTextToPaths.js     # opentype.js 转曲引擎：text → path
└── dictionary.js         # 翻译字典（反面多语言文字）
```

### 依赖

```json
"jspdf": "^4.2.1",
"svg2pdf.js": "^2.7.0",
"opentype.js": "^2.0.0"
```

---

## 🧠 核心架构

```
buildPageSvg (DOM SVG 合成)
    ↓ 每个页面返回一张 <svg viewBox="0 0 W H"> 节点，viewBox 单位为 mm
convertSvgTextToPaths (整页转曲)
    ↓ 遍历所有 <text>，用 opentype.js 生成 <path d="...">
svg2pdf (矢量渲染引擎)
    ↓ 将纯路径 SVG 渲染到 jsPDF 画布
doc.save() → .pdf 文件
```

## 🔑 关键坐标体系

### viewBox 毫米制

```javascript
svg.setAttribute('viewBox', `0 0 ${labelWidth} ${labelHeight}`)
svg.setAttribute('width', `${labelWidth}mm`)
svg.setAttribute('height', `${labelHeight}mm`)
```

- `viewBox` 单位 = mm（和 `labelWidth`/`labelHeight` 1:1 对应）
- 所有内部坐标 `x`、`y` 直接使用 mm 数值，无需 px ↔ mm 换算
- `font-size` 使用 `fontSizePt * MM_PER_PT` 转为 mm 等价 unitless 值（如 5pt → 1.764），确保 opentype.js 缩放正确

### 文字基线定位

```javascript
const baseY = topY + baselineOffset(fontSizePt)
// baselineOffset = fontSizePt * MM_PER_PT * 0.8
```

- `<text>` 的 `y` = 字体基线 (baseline)，opentype.js `getPath()` 期望的 y 正是基线
- 不使用 `dominant-baseline="text-before-edge"`（该属性会让 opentype.js 产生不可控的上下偏移）

### text-anchor 修正

```javascript
// svg2pdf.js 内，根据 text-anchor 修正 x（opentype.js 以 x 为左起点）
const textWidth = font.getAdvanceWidth(text, fontSize)
if (anchor === 'middle') adjustedX = run.x - textWidth / 2
if (anchor === 'end')    adjustedX = run.x - textWidth
```

| `text-anchor` | SVG `x` 含义 | opentype.js 期望 | 修正量 |
|---|---|---|---|
| `start` | 左边缘 | 左边缘 | 0 |
| `middle` | 中点 | 左边缘 | `-textWidth / 2` |
| `end` | 右边缘 | 左边缘 | `-textWidth` |

---

## 📄 pdfExport.js 逐函数说明

### `exportLabelPDF` — 主入口

```javascript
export async function exportLabelPDF({
  frontElements,   // 正面元素数组
  backElements,    // 反面元素数组（可选）
  config,          // { labelWidth, labelHeight, headSeam, marginLR }
  translateInfo,   // { dictionary, translateLangs }
  fileName,
})
```

**流程：**
1. 创建 `jsPDF({ unit: "mm", format: [W, H] })`
2. 注册所有字体到 jsPDF（兜底，转曲成功时不需要）
3. 遍历页面（正面 + 可选反面）
   - `buildPageSvg(elements, config, opts)` → 合成 SVG
   - `convertSvgTextToPaths(pageSvg)` → 整页文字转路径
   - `svg2pdf(pageSvg, doc, { x: 0, y: 0 })` → 渲染到 PDF
4. `doc.save(fileName)`

### `buildPageSvg` — SVG 合成

元素按 `y += height` 逐行堆叠，从 `headSeam` 开始：

| 元素类型 | 处理函数 | 渲染方式 |
|---|---|---|
| `text` | `appendTextToSvg` | `<text>` 元素（基线定位） |
| `image` | `appendImageToSvg` | SVG→`<g>` 矢量嵌入 / 位图→`<image>` |
| `table` | `appendTableToSvg` | `<rect>` 边框 + `<text>` 单元格 |

### `appendTextToSvg` — 文字

- 正面：直接渲染 `el.text` 为 `<text>`
- 反面翻译：翻译文本按语言分段渲染（支持 RTL 阿拉伯语）
- 使用 `splitTextToLines`（jsPDF 测量）拆分为多行
- 垂直对齐 (valign) 计算时先算 `topY`，再加 `baselineOffset` 得到基线 `baseY`

### `appendImageToSvg` — 图片

- SVG 图片：`loadSvgWithTextAsPaths(url)`（独立转曲）→ `getSvgSize` → `scale` → `<g transform="translate(x,y) scale(sx,sy)">`
- 位图：`rasterizeImage` → `<image href="dataURL">`

### `appendTableToSvg` — 表格

- 每个单元格：SVG `<rect>`（边框）+ `<text>`（内容）
- 单元格文字按 `colWidth - 0.5` 宽度自动换行
- 支持行/列对齐和边框开关

---

## 📄 svgTextToPaths.js — opentype.js 转曲引擎

### 字体配置

```javascript
const FONT_URLS = {
  zh:     '/fonts/FZ.TTF',           // 中文字体
  latin:  '/fonts/GO.TTF',           // 拉丁字体
  arabic: '/fonts/ARIAL.TTF',        // 阿拉伯字体
  MiSans: '/fonts/MiSans-Regular.ttf',
}
```

### 文字检测 → 字体选择

```javascript
detectRole(text)
  → 含阿语字符 → 'arabic'
  → 含 CJK 字符 → 'zh'
  → 其他 → 'latin'
```

### `convertSvgTextToPaths(svgEl)`

1. 加载字体（已缓存则跳过）
2. `svgEl.querySelectorAll('text')` 遍历所有 `<text>`
3. 对每个 `<text>`：
   - 读取 `font-size`、`fill`、`text-anchor`
   - `collectRuns` 收集文本行（含 tspan 定位）
   - 根据 `text-anchor` 修正 x 坐标（见上文表格）
   - `font.getPath(text, x, y, fontSize)` → `pathObj.toPathData(2)` → `<path>`
4. `textEl.replaceWith(group)` — 原地替换

### 字体提前预热

`loadFonts()` 在调用 `convertSvgTextToPaths` 时自动触发，字体被 `opentype.parse(buffer)` 后缓存到 `fontCache`，后续调用无需重复加载。

---

## 📄 dictionary.js — 翻译字典

见 `web/src/plugin/label/utils/dictionary.js`，核心函数：

- `fetchDictionary(dictName)` — 从后端加载字典
- `translateText(dict, chinese, lang)` — 中文→目标语言翻译
  - 支持括号包裹、别名回退
  - 支持成分值翻译（如 `70%棉 30%腈纶`）
  - 支持冒号键值对翻译

---

## ⚠️ 关键坑点 & 注意事项

### 1. `font-size` 必须是 mm unitless 值

❌ `<text font-size="5pt">` — opentype.js 读到的 `parseFloat("5pt")` = 5，但 SVG viewBox 是 mm，5mm 字号 ≈ 14pt，文字会变大 2.8 倍

✅ `<text font-size="1.764">` — `5pt * 25.4/72 = 1.764mm`，缩放一致

### 2. text-anchor 必须修正

❌ 直接拿 `<text x="10" text-anchor="middle">` 的 `x=10` 传给 opentype.js 当左起点

✅ 算出文字实际宽度，`x - textWidth / 2` 得到左起点，再传给 opentype.js

### 3. y 是基线不是顶部

❌ 用 `dominant-baseline="text-before-edge"` + `y=topY` — 转曲后位置不对

✅ `y = topY + baselineOffset`，让 opentype.js 拿到正确的基线坐标

### 4. jsPDF 字体注册作为兜底

虽然整页转曲后零字体依赖，但注册字体到 jsPDF 成本很低（已缓存 base64），用于转曲失败的极端兜底场景。

### 5. 字体文件同步

- `FONT_SPECS`（pdfExport.js）和 `FONT_URLS`（svgTextToPaths.js）的字体文件必须同步更新
- 新增字体时两个文件都要加

---

## 🧪 开发调试检查清单

- [ ] 新建任意尺寸模板，导出 PDF 尺寸与模板宽高一致
- [ ] 文字无论左/中/右对齐，转曲后位置不变
- [ ] 表格单元格文字不溢出、不重叠
- [ ] 反面翻译文字按语言正确分段，RTL 语言（阿语）右对齐
- [ ] 嵌入的 SVG 图片内文字已转曲
- [ ] 位图图片可正常渲染
- [ ] AI 中打开 PDF，"释放剪切蒙版"后所有图层独立可编辑
