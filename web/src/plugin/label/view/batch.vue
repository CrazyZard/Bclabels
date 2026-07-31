<template>
  <div class="batch-label-editor">
    <div class="batch-toolbar">
      <div style="display:flex;align-items:center;gap:12px">
        <span style="font-size:16px;font-weight:600">生产模板</span>
        <el-select v-model="selectedTemplateName" placeholder="选择已发布的模板" size="default" style="width:220px" @change="onTemplateChange" filterable>
          <el-option v-for="tpl in publishedTemplates" :key="tpl.name" :label="tpl.name" :value="tpl.name" />
        </el-select>
        <el-button type="primary" :disabled="!selectedTemplateName" @click="handleDownloadTemplate">
          <el-icon><Download /></el-icon> 下载批量模板
        </el-button>
      </div>
      <div style="display:flex;gap:8px;align-items:center">
        <el-upload ref="uploadRef" :auto-upload="false" :show-file-list="false" accept=".xlsx,.xls" :on-change="handleFileChange">
          <el-button type="success" :disabled="!selectedTemplateName">
            <el-icon><Upload /></el-icon> 上传批量Excel
          </el-button>
        </el-upload>
        <el-button type="warning" :disabled="!dataRows.length" @click="handleExportAllPDF" :loading="exporting">
          <el-icon><Printer /></el-icon> 导出全部PDF
        </el-button>
      </div>
    </div>

    <div v-if="!selectedTemplateName" class="batch-empty">
      <el-empty description="请先选择一个已发布的模板" />
    </div>

    <div v-else class="batch-body">
      <div class="batch-canvas-area">
        <div v-if="!currentRenderEls.length" style="text-align:center;color:#bbb;padding:60px 0">
          <el-icon :size="48"><Picture /></el-icon>
          <div style="margin-top:12px">上传Excel后可预览标签效果</div>
        </div>
        <div v-else style="display:flex;flex-direction:column;align-items:center;gap:8px">
          <div style="display:flex;align-items:center;gap:12px;font-size:13px;color:#606266">
            <el-button size="small" :disabled="currentRowIndex <= 0" @click="currentRowIndex--"><el-icon><ArrowLeft /></el-icon></el-button>
            <span>第 {{ currentRowIndex + 1 }} / {{ dataRows.length }} 条</span>
            <el-button size="small" :disabled="currentRowIndex >= dataRows.length - 1" @click="currentRowIndex++"><el-icon><ArrowRight /></el-icon></el-button>
          </div>
          <div style="display:flex;flex-wrap:nowrap;gap:20px;justify-content:center;align-items:flex-start">
            <div style="display:flex;flex-direction:column;align-items:center;flex-shrink:0">
              <div style="text-align:center;font-size:11px;color:#409eff;margin-bottom:4px;font-weight:600">正面 · 预览</div>
              <div>
                <div class="seam-zone" :style="{width:labelWidth*zoom+'px',height:headSeam*zoom+'px'}"><span style="color:#999;font-size:10px;margin-right:4px">顶缝 {{headSeam}}mm</span></div>
                <div class="canvas-wrap" :style="{width:labelWidth*zoom+'px',minHeight:(labelHeight-headSeam)*zoom+'px',paddingLeft:marginLR*zoom+'px',paddingRight:marginLR*zoom+'px'}">
                  <div v-for="el in currentRenderEls" :key="el.id" class="batch-canvas-row" :style="{height:(el.height||10)*zoom+'px'}">
                    <div class="batch-element-content" :style="batchElementStyle(el)">
                      <template v-if="el.type==='text'">
                        <span class="rich-preview" v-html="el.renderedHtml || el.text || ''" />
                      </template>
                      <template v-else-if="el.type==='image'">
                        <img v-if="el.renderedSrc" :src="resolveUrl(el.renderedSrc)" style="max-width:100%;max-height:100%;display:block" />
                        <div v-else class="img-placeholder">图</div>
                      </template>
                      <template v-else-if="el.type==='table'">
                        <table class="mini-table" :style="{fontSize:(el.fontSize||5)+'pt',fontFamily:fontCss[el.fontFamily||'FZLTXIHJW--GB1-0'],borderCollapse:el.showBorder!==false?'collapse':'separate'}">
                          <tr v-for="(row,ri) in (el.cells||[])" :key="ri" :style="{height:(el.rowHeight||2.2)*zoom+'px'}">
                            <td v-for="(cell,ci) in row" :key="ci" :style="{width:(el.colWidth||4.4)*zoom+'px',border:el.showBorder!==false?'2px solid #666':'none',textAlign:cell.textAlign||el.alignment||'center'}">
                              <template v-if="cell.value"><div v-for="(line,li) in cell.value.split('\n')" :key="li">{{line||'\u00A0'}}</div></template><template v-else>&nbsp;</template>
                            </td>
                          </tr>
                        </table>
                      </template>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="needsTranslation && translateLangs.length && currentRenderBackEls.length" style="display:flex;flex-direction:column;align-items:center;flex-shrink:0">
              <div style="text-align:center;font-size:11px;color:#e6a23c;margin-bottom:4px;font-weight:600">反面 · {{translateLangs.map(l=>langLabel(l)).join('/')}}</div>
              <div>
                <div class="seam-zone" :style="{width:labelWidth*zoom+'px',height:headSeam*zoom+'px'}"><span style="color:#999;font-size:10px;margin-right:4px">顶缝 {{headSeam}}mm</span></div>
                <div class="canvas-wrap" :style="{width:labelWidth*zoom+'px',minHeight:(labelHeight-headSeam)*zoom+'px',paddingLeft:marginLR*zoom+'px',paddingRight:marginLR*zoom+'px'}">
                  <div v-for="el in currentRenderBackEls" :key="el.id" class="batch-canvas-row" :style="{height:(el.height||10)*zoom+'px'}">
                    <div class="batch-element-content" :style="batchElementStyle(el)">
                      <template v-if="el.type==='text' && el.translations">
                        <span v-for="(tr, ti) in el.translations" :key="tr.lang"
                          :style="{fontSize:el.fontSize+'pt',fontFamily:fontCss[el.fontFamily||(tr.isArabic?'ArialMT':'CenturyGothic')],fontWeight:el.bold?'bold':'normal',letterSpacing:(el.letterSpacing||0)+'pt',lineHeight:el.lineHeight||1.5,whiteSpace:'pre-wrap',wordBreak:'break-all',direction:'ltr',textAlign:el.alignment||'left',display:'block',marginTop:ti>0?'6px':'0'}"
                          v-html="tr.isArabic ? formatArabicDisplayHtml(tr.text) : escapeHtml(tr.text)"
                        ></span>
                      </template>
                      <template v-else-if="el.type==='image'">
                        <img v-if="el.src" :src="resolveUrl(el.src)" style="max-width:100%;max-height:100%;display:block" />
                        <div v-else class="img-placeholder">图</div>
                      </template>
                      <template v-else-if="el.type==='table'">
                        <table class="mini-table" :style="{fontSize:(el.fontSize||5)+'pt',fontFamily:fontCss[el.fontFamily||'FZLTXIHJW--GB1-0'],borderCollapse:el.showBorder!==false?'collapse':'separate'}">
                          <tr v-for="(row,ri) in (el.cells||[])" :key="ri" :style="{height:(el.rowHeight||2.2)*zoom+'px'}">
                            <td v-for="(cell,ci) in row" :key="ci" :style="{width:(el.colWidth||4.4)*zoom+'px',border:el.showBorder!==false?'2px solid #666':'none',textAlign:cell.textAlign||el.alignment||'center'}">
                              <template v-if="cell.value"><div v-for="(line,li) in cell.value.split('\n')" :key="li">{{line||'\u00A0'}}</div></template><template v-else>&nbsp;</template>
                            </td>
                          </tr>
                        </table>
                      </template>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="batch-data-panel">
        <div v-if="!dataRows.length" style="color:#999;text-align:center;padding:40px 0">
          <el-icon :size="32"><Document /></el-icon>
          <div style="margin-top:8px;font-size:13px">上传Excel后解析数据显示在此处</div>
        </div>
        <template v-else>
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
            <span style="font-size:14px;font-weight:600">数据预览</span>
            <el-tag type="info" size="small">共 {{ dataRows.length }} 条</el-tag>
          </div>
          <div style="overflow:auto;flex:1">
            <table class="data-preview-table">
              <thead>
                <tr>
                  <th style="width:40px">#</th>
                  <th v-for="col in columns" :key="col" style="min-width:80px">{{ col }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, ri) in dataRows" :key="ri" :class="{ 'row-selected': ri === currentRowIndex }" @click="currentRowIndex = ri" style="cursor:pointer">
                  <td style="text-align:center;color:#909399">{{ ri + 1 }}</td>
                  <td v-for="col in columns" :key="col">{{ row[col] || '' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Upload, Printer, Picture, ArrowLeft, ArrowRight, Document } from '@element-plus/icons-vue'
import * as XLSX from 'xlsx'
import { listPublishedTemplate, loadTemplate } from '@/plugin/label/api/template'
import { downloadBatchTemplate } from '@/plugin/label/api/template'
import { exportLabelPDF } from '@/plugin/label/utils/pdfExport'
import { translateMany, getCachedTranslation, makeTranslateLookup, clearDictionaryCache, formatArabicDisplayHtml, stripBidiMarks } from '@/plugin/label/utils/dictionary'

defineOptions({ name: 'BatchLabelEditor' })

const publishedTemplates = ref([])
const selectedTemplateName = ref('')
const zoom = ref(8)
const currentRowIndex = ref(0)
const exporting = ref(false)

const labelWidth = ref(80)
const labelHeight = ref(120)
const headSeam = ref(8)
const marginLR = ref(2)
const needsTranslation = ref(false)
const dictName = ref('')
const translateLangs = ref([])
const elements = ref([])
const translatedElements = ref([])

const columns = ref([])
const dataRows = ref([])
const currentRenderEls = ref([])
const currentRenderBackEls = ref([])

const fontCss = {
  'FZLTXIHJW--GB1-0': "'FZLTXIHJW--GB1-0', SimSun, sans-serif",
  CenturyGothic: "CenturyGothic, 'Century Gothic', sans-serif",
  ArialMT: 'ArialMT, Arial, sans-serif',
  'MiSans-Regular': 'MiSans-Regular, sans-serif'
}

function resolveUrl(s) {
  if (!s) return ''
  if (/^https?:\/\//.test(s) || /^data:/.test(s)) return s
  return '/' + s.replace(/^\/+/, '')
}

onMounted(async () => {
  await loadPublishedList()
})

async function loadPublishedList() {
  try {
    const res = await listPublishedTemplate()
    if (res.code === 0) {
      publishedTemplates.value = res.data || []
    }
  } catch (e) {
    console.error(e)
  }
}

async function onTemplateChange(name) {
  if (!name) return
  try {
    const res = await loadTemplate({ name })
    if (res.code === 0) {
      const t = res.data
      labelWidth.value = t.labelWidth || 80
      labelHeight.value = t.labelHeight || 120
      headSeam.value = t.headSeam || 8
      marginLR.value = t.marginLR ?? 2
      needsTranslation.value = t.needsTranslation || false
      dictName.value = t.dictName || ''
      try { translateLangs.value = JSON.parse(t.translateLangs || '[]') } catch { translateLangs.value = [] }
      try { elements.value = JSON.parse(t.elements || '[]') } catch { elements.value = [] }
      try { translatedElements.value = JSON.parse(t.translatedElements || '[]') } catch { translatedElements.value = [] }
      clearDictionaryCache()
      columns.value = []
      dataRows.value = []
      currentRowIndex.value = 0
      currentRenderEls.value = []
      currentRenderBackEls.value = []
    }
  } catch (e) {
    console.error(e)
  }
}

async function handleDownloadTemplate() {
  try {
    const res = await downloadBatchTemplate({ name: selectedTemplateName.value })
    const blob = res.data instanceof Blob ? res.data : res
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedTemplateName.value}_批量导入模板.xlsx`
    a.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('模板下载成功')
  } catch (e) {
    console.error(e)
    ElMessage.error('模板下载失败')
  }
}

function handleFileChange(file) {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const data = new Uint8Array(e.target.result)
      const workbook = XLSX.read(data, { type: 'array' })
      const sheetName = workbook.SheetNames[0]
      const sheet = workbook.Sheets[sheetName]
      const jsonData = XLSX.utils.sheet_to_json(sheet, { defval: '' })
      if (!jsonData.length) {
        ElMessage.warning('Excel文件中没有数据')
        return
      }
      columns.value = Object.keys(jsonData[0])
      dataRows.value = jsonData.map(row => {
        const obj = {}
        columns.value.forEach(col => { obj[col] = String(row[col] || '') })
        return obj
      })
      currentRowIndex.value = 0
      renderCurrentRow()
      ElMessage.success(`成功解析 ${dataRows.value.length} 条数据`)
    } catch (err) {
      console.error(err)
      ElMessage.error('Excel解析失败，请检查文件格式')
    }
  }
  reader.readAsArrayBuffer(file.raw)
}

async function renderCurrentRow() {
  if (!dataRows.value.length) {
    currentRenderEls.value = []
    currentRenderBackEls.value = []
    return
  }
  const rowData = dataRows.value[currentRowIndex.value]
  currentRenderEls.value = elements.value.map(el => {
    const copy = JSON.parse(JSON.stringify(el))
    if (el.type === 'text') {
      const key = el.key
      if (key && rowData[key] !== undefined) {
        copy.text = rowData[key]
        copy.html = '' // 清除富文本，用纯文本方式渲染
        copy.renderedHtml = elementRenderHtml({ ...el, text: rowData[key], html: '' })
      } else {
        copy.renderedHtml = elementRenderHtml(el)
      }
    } else if (el.type === 'image') {
      if (el.useMapping && el.imageMap && el.imageMap.length) {
        const key = el.key
        const value = key ? rowData[key] : ''
        const matched = el.imageMap.find(m => m.value === value)
        copy.renderedSrc = matched ? matched.src : (el.src || '')
      } else {
        copy.renderedSrc = el.src || ''
      }
    }
    return copy
  })

  // 反面翻译元素渲染（先批量请求后端）
  if (needsTranslation.value && translateLangs.value.length && translatedElements.value.length && dictName.value) {
    const items = []
    for (const el of translatedElements.value) {
      if (el.type !== 'text' || !el.key || !el.langKeys?.length) continue
      const fe = elements.value.find(e => e.key === el.key)
      const srcText = (fe && rowData[fe.key] !== undefined) ? rowData[fe.key] : (fe?.text || '')
      if (srcText && String(srcText).trim()) items.push({ text: srcText, langs: [...el.langKeys] })
    }
    try {
      await translateMany(dictName.value, items)
    } catch (e) {
      console.error(e)
    }
    currentRenderBackEls.value = translatedElements.value.map(el => {
      const copy = JSON.parse(JSON.stringify(el))
      if (el.type === 'text' && el.key) {
        const fe = elements.value.find(e => e.key === el.key)
        const srcText = (fe && rowData[fe.key] !== undefined) ? rowData[fe.key] : (fe?.text || '')
        if (el.langKeys && el.langKeys.length && srcText) {
          copy.translations = el.langKeys.map(lang => ({
            lang,
            label: langLabel(lang),
            text: stripBidiMarks(getCachedTranslation(dictName.value, srcText, lang) || '[?]'),
            isArabic: lang === 'arabic'
          }))
        } else {
          copy.translations = []
        }
      }
      return copy
    })
  } else {
    currentRenderBackEls.value = []
  }
}

function elementRenderHtml(el) {
  if (!el) return ''
  const ff = el.fontFamily || 'FZLTXIHJW--GB1-0'
  const fs = el.fontSize || 5
  const ls = el.letterSpacing ? `letter-spacing:${el.letterSpacing}pt;` : ''
  const lh = el.lineHeight ? `line-height:${el.lineHeight};` : ''
  if (el.html) return `<span style="font-size:${fs}pt;font-family:${cssFont(ff)};${ls}${lh}">${(el.html || '')}</span>`
  return `<span style="font-size:${fs}pt;font-family:${cssFont(ff)};${ls}${lh}">${escapeHtml(el.text || '')}</span>`
}

function cssFont(psName) {
  return fontCss[psName] || `'${psName}'`
}

function escapeHtml(s) { return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;') }

function langLabel(lang) {
  return { english: '英文', russian: '俄文', arabic: '阿语', indonesian: '印尼' }[lang] || lang
}

function batchElementStyle(el) {
  const v = { top: 'flex-start', middle: 'center', bottom: 'flex-end' }
  const h = { left: 'flex-start', center: 'center', right: 'flex-end' }
  return {
    maxWidth: (labelWidth.value - marginLR.value * 2) * zoom.value + 'px',
    width: (el.width || (labelWidth.value - marginLR.value * 2)) * zoom.value + 'px',
    textAlign: el.alignment || 'left',
    alignItems: h[el.alignment || 'left'],
    justifyContent: v[el.valign || 'middle']
  }
}

watch(currentRowIndex, () => {
  renderCurrentRow()
})

async function handleExportAllPDF() {
  if (!dataRows.value.length) {
    ElMessage.warning('没有可导出的数据')
    return
  }
  exporting.value = true
  try {
    if (needsTranslation.value && translateLangs.value.length && dictName.value) {
      const items = []
      for (const rowData of dataRows.value) {
        for (const el of translatedElements.value) {
          if (el.type !== 'text' || !el.key || !el.langKeys?.length) continue
          const fe = elements.value.find(e => e.key === el.key)
          const srcText = (fe && rowData[fe.key] !== undefined) ? rowData[fe.key] : (fe?.text || '')
          if (srcText && String(srcText).trim()) items.push({ text: srcText, langs: [...el.langKeys] })
        }
      }
      await translateMany(dictName.value, items)
    }

    const lookup = needsTranslation.value ? makeTranslateLookup(dictName.value) : null
    let doc = null
    for (let i = 0; i < dataRows.value.length; i++) {
      const rowData = dataRows.value[i]
      const frontEls = elements.value.map(el => {
        const copy = JSON.parse(JSON.stringify(el))
        if (el.type === 'text') {
          const key = el.key
          if (key && rowData[key] !== undefined) {
            copy.text = rowData[key]
            copy.html = ''
          }
        } else if (el.type === 'image' && el.useMapping && el.imageMap?.length) {
          const key = el.key
          const value = key ? rowData[key] : ''
          const matched = el.imageMap.find(m => m.value === value)
          if (matched) copy.src = matched.src
        }
        return copy
      })

      let backEls = null
      if (needsTranslation.value && translateLangs.value.length && translatedElements.value.length) {
        backEls = translatedElements.value.map(el => {
          const copy = JSON.parse(JSON.stringify(el))
          if (el.type === 'text' && el.key) {
            copy.enableTranslation = true
          }
          return copy
        })
      }

      doc = await exportLabelPDF({
        frontElements: frontEls,
        backElements: backEls,
        config: {
          labelWidth: labelWidth.value,
          labelHeight: labelHeight.value,
          headSeam: headSeam.value,
          marginLR: marginLR.value
        },
        translateInfo: needsTranslation.value ? {
          lookup,
          translateLangs: translateLangs.value
        } : null,
        existingDoc: doc
      })
    }

    if (doc) {
      doc.save(`${selectedTemplateName.value || '标签'}_批量_${dataRows.value.length}条.pdf`)
    }
    ElMessage.success(`成功导出 ${dataRows.value.length} 条标签`)
  } catch (e) {
    console.error(e)
    ElMessageBox.alert(e.message || '导出失败', '导出失败', { confirmButtonText: '知道了', type: 'error' })
  } finally {
    exporting.value = false
  }
}
</script>

<style scoped>
.batch-label-editor {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 90px);
  background: #f5f7fa;
}

.batch-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  flex-shrink: 0;
}

.batch-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.batch-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.batch-canvas-area {
  flex: 1;
  overflow: auto;
  padding: 24px;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  background: #f0f0f0;
}

.batch-data-panel {
  width: 340px;
  flex-shrink: 0;
  border-left: 1px solid #e8e8e8;
  background: #fff;
  display: flex;
  flex-direction: column;
  padding: 12px;
  overflow: hidden;
}

.canvas-wrap {
  background: #fff;
  border-radius: 0 0 4px 4px;
  min-height: 100px;
  box-sizing: border-box;
}

.seam-zone {
  background: repeating-linear-gradient(-45deg, #f5f5f5, #f5f5f5 4px, #ececec 4px, #ececec 8px);
  border-radius: 4px 4px 0 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 0 8px;
  box-sizing: border-box;
}

.batch-canvas-row {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e0e0e0;
  margin-bottom: 2px;
  overflow: hidden;
  background: #fff;
}

.batch-element-content {
  flex-shrink: 0;
  overflow: hidden;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 2px 4px;
  height: 100%;
}

.data-preview-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.data-preview-table th,
.data-preview-table td {
  border: 1px solid #e8e8e8;
  padding: 6px 8px;
  white-space: nowrap;
  text-align: left;
}

.data-preview-table th {
  background: #f5f7fa;
  font-weight: 600;
  color: #303133;
}

.data-preview-table tr:hover {
  background: #f5f7fa;
}

.data-preview-table .row-selected {
  background: #ecf5ff;
}

.img-placeholder {
  width: 100%;
  height: 100%;
  background: #eee;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 10px;
}

.mini-table {
  width: 100%;
  border-collapse: collapse;
}

.mini-table td {
  border: 2px solid #666;
  padding: 1px 2px;
  text-align: center;
}
</style>
