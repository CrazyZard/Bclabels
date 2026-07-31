<template>
  <div class="label-editor">
    <div v-if="viewMode === 'list'" style="display:flex;flex-direction:column;padding:20px">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
        <h3 style="margin:0">标签模板</h3>
        <div style="display:flex;gap:10px">
          <el-input v-model="searchKeyword" placeholder="搜索模板名称" clearable size="default" style="width:220px">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showSetupDialog = true"><el-icon><Plus /></el-icon> 新建模板</el-button>
        </div>
      </div>
      <div style="flex:1;overflow-y:auto">
        <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:16px">
          <div v-for="tpl in filteredTemplates" :key="tpl.name" class="tpl-card" @click="openTemplate(tpl.name)">
            <div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:8px">
              <div style="display:flex;align-items:center;gap:8px">
                <span style="font-size:16px;font-weight:600;color:#303133">{{ tpl.name }}</span>
                <el-tag v-if="tpl.isPublished" type="success" size="small">已发布</el-tag>
              </div>
              <el-button link type="danger" size="small" @click.stop="handleDeleteTpl(tpl.name)"><el-icon><Delete /></el-icon></el-button>
            </div>
            <div style="font-size:13px;color:#909399;margin-bottom:4px">{{ tpl.labelWidth }}×{{ tpl.labelHeight }}mm</div>
            <div style="font-size:12px;color:#c0c4cc;margin-bottom:8px">{{ tpl.elementCount || 0 }} 个元素</div>
            <div style="display:flex;justify-content:flex-end">
              <el-button v-if="!tpl.isPublished" type="success" size="small" plain @click.stop="handlePublish(tpl.name)">发布</el-button>
              <el-button v-else type="warning" size="small" plain @click.stop="handleUnpublish(tpl.name)">取消发布</el-button>
            </div>
          </div>
          <div v-if="filteredTemplates.length === 0" style="grid-column:1/-1;text-align:center;padding:60px 0;color:#909399">
            <el-empty description="暂无模板，点击上方【新建模板】开始创建" />
          </div>
        </div>
      </div>
    </div>

    <div v-if="viewMode === 'editor'" class="editor-wrap">
      <div class="editor-toolbar">
        <div style="display:flex;align-items:center;gap:12px">
          <el-button link @click="backToList"><el-icon><ArrowLeft /></el-icon> 返回列表</el-button>
          <span style="font-size:15px;font-weight:600">{{ templateName }}</span>
          <span style="font-size:12px;color:#909399">{{ labelWidth }}×{{ labelHeight }}mm · 顶缝{{ headSeam }}mm · 间距{{ marginLR }}mm</span>
        </div>
        <div style="display:flex;gap:8px;align-items:center">
          <template v-if="needsTranslation">
            <el-checkbox-group v-model="translateLangs" size="small" style="display:flex;gap:4px">
              <el-checkbox-button value="english" style="font-size:11px">英文</el-checkbox-button>
              <el-checkbox-button value="russian" style="font-size:11px">俄文</el-checkbox-button>
              <el-checkbox-button value="arabic" style="font-size:11px">阿语</el-checkbox-button>
              <el-checkbox-button value="indonesian" style="font-size:11px">印尼</el-checkbox-button>
            </el-checkbox-group>
          </template>
          <span style="font-size:12px;color:#909399;white-space:nowrap">缩放</span>
          <el-input-number v-model="zoomLevel" :min="0.5" :max="10" :step="0.5" size="small" controls-position="right" style="width:80px" />
          <span style="font-size:11px;color:#c0c4cc">×</span>
          <el-button type="success" @click="handleExportPDF" :loading="exporting">导出PDF</el-button>
          <el-button @click="showSetupDialog = true; setupEdit = true">修改配置</el-button>
          <el-button type="primary" @click="handleSave">保存</el-button>
        </div>
      </div>
      <div class="editor-body">
        <div class="left-panel">
          <h4>基础组件</h4>
          <div v-for="item in paletteItems" :key="item.type" class="palette-item" draggable="true" @dragstart="onDragStart($event, item.type)">
            <el-icon :size="18"><component :is="item.icon" /></el-icon><span>{{ item.label }}</span>
          </div>
          <h4 style="margin-top:16px">正面元素 ({{ elements.length }})</h4>
          <div v-for="el in elements" :key="el.id" :class="['element-item', { active: selectedId === el.id && selectedSide === 'front' }]" @click="selectElement(el.id, 'front')">
            <span>{{ typeLabel(el.type) }} - {{ el.text || '未命名' }}</span>
            <el-button link type="danger" size="small" @click.stop="removeElement(el.id, 'front')"><el-icon><Delete /></el-icon></el-button>
          </div>
          <template v-if="needsTranslation && translateLangs.length">
            <h4 style="margin-top:16px;display:flex;justify-content:space-between;align-items:center">反面元素 ({{ translatedElements.length }})<el-button size="small" text @click="addBackElement"><el-icon><Plus /></el-icon></el-button></h4>
            <div v-for="el in translatedElements" :key="el.id" :class="['element-item', { active: selectedId === el.id && selectedSide === 'back' }]" @click="selectElement(el.id, 'back')">
              <span>{{ typeLabel(el.type) }} - {{ el.text || '未命名' }}</span>
              <el-button link type="danger" size="small" @click.stop="removeElement(el.id, 'back')"><el-icon><Delete /></el-icon></el-button>
            </div>
          </template>
        </div>
        <div class="canvas-area" @dragover.prevent @dragenter.prevent @drop="onCanvasDrop">
          <div v-if="elements.length === 0" class="canvas-empty">从左侧拖入组件到此处</div>
          <div v-else style="display:flex;flex-wrap:nowrap;gap:20px;justify-content:center;align-items:flex-start">
            <div style="display:flex;flex-direction:column;align-items:center;flex-shrink:0">
              <div style="text-align:center;font-size:11px;color:#409eff;margin-bottom:4px;font-weight:600">正面 · 中文</div>
              <div>
              <div class="seam-zone" :style="{width:labelWidth*scale+'px',height:headSeam*scale+'px'}"><span style="color:#999;font-size:10px;margin-right:4px">顶缝 {{ headSeam }}mm</span></div>
              <div class="canvas-wrap" :style="{width:labelWidth*scale+'px',minHeight:canvasHeight*scale+'px',paddingLeft:marginLR*scale+'px',paddingRight:marginLR*scale+'px'}">
              <draggable v-model="elements" item-key="id" handle=".drag-handle" ghost-class="ghost" animation="200">
                <template #item="{ element }">
                  <div :class="['canvas-row',{selected:selectedId===element.id&&selectedSide==='front'}]" :style="{height:rowHeightPx(element)+'px'}" @click.stop="selectElement(element.id,'front')">
                    <div class="drag-handle"><el-icon><Menu/></el-icon></div>
                    <div class="element-content" :style="elementContentStyle(element)">
                      <template v-if="element.type==='text'"><span class="rich-preview" v-html="elementHtml(element)" /></template>
                      <template v-else-if="element.type==='image'">
                        <template v-if="element.useMapping&&element.imageMap?.length"><div class="img-placeholder"><span>动态图</span><span style="color:#409eff">{{element.imageMap.length}}条映射</span></div></template>
                        <img v-else-if="element.src" :src="resolveImageUrl(element.src)" style="max-width:100%;max-height:100%;display:block" /><div v-else class="img-placeholder">图</div>
                      </template>
                      <template v-else-if="element.type==='table'">
                        <table class="mini-table" :style="{fontSize:(element.fontSize||5)+'pt',fontFamily:fontMap[element.fontFamily||'FZLTXIHJW--GB1-0'],borderCollapse:element.showBorder!==false?'collapse':'separate'}">
                          <tr v-for="(row,ri) in (element.cells||buildDefaultCells(element))" :key="ri" :style="{height:(element.rowHeight||2.2)*scale+'px'}">
                            <td v-for="(cell,ci) in row" :key="ci" :style="{width:(element.colWidth||4.4)*scale+'px',border:element.showBorder!==false?'2px solid #666':'none',textAlign:cell.textAlign||element.alignment||'center'}">
                              <template v-if="cell.value"><div v-for="(line,li) in cell.value.split('\n')" :key="li">{{line||'\u00A0'}}</div></template><template v-else>&nbsp;</template>
                            </td>
                          </tr>
                        </table>
                      </template>
                    </div>
                    <div v-if="selectedId===element.id&&selectedSide==='front'" class="drag-mask"></div>
                    <template v-if="selectedId===element.id&&selectedSide==='front'"><div class="rh w" @pointerdown.stop="onResizeStart($event,element.id,'w','front')" @mousedown.stop.prevent @touchstart.stop.prevent/><div class="rh e" @pointerdown.stop="onResizeStart($event,element.id,'e','front')" @mousedown.stop.prevent @touchstart.stop.prevent/><div class="rh s" @pointerdown.stop="onResizeStart($event,element.id,'s','front')" @mousedown.stop.prevent @touchstart.stop.prevent/></template>
                  </div>
                </template>
              </draggable>
              </div>
              </div>
            </div>
            <div v-if="needsTranslation&&translateLangs.length" style="display:flex;flex-direction:column;align-items:center;flex-shrink:0">
              <div style="text-align:center;font-size:11px;color:#e6a23c;margin-bottom:4px;font-weight:600">反面 · {{translateLangs.map(l=>langLabel(l)).join('/')}}</div>
              <div>
              <div class="seam-zone" :style="{width:labelWidth*scale+'px',height:headSeam*scale+'px'}"><span style="color:#999;font-size:10px;margin-right:4px">顶缝 {{headSeam}}mm</span></div>
              <div class="canvas-wrap" :style="{width:labelWidth*scale+'px',minHeight:canvasHeight*scale+'px',paddingLeft:marginLR*scale+'px',paddingRight:marginLR*scale+'px'}">
              <draggable v-model="translatedElements" item-key="id" handle=".drag-handle" ghost-class="ghost" animation="200">
                <template #item="{ element }">
                  <div :class="['canvas-row',{selected:selectedId===element.id&&selectedSide==='back'}]" :style="{height:rowHeightPx(element)+'px'}" @click.stop="selectElement(element.id,'back')">
                    <div class="drag-handle"><el-icon><Menu/></el-icon></div>
                    <div class="element-content" :style="elementContentStyle(element)">
                      <template v-if="element.type==='text'">
                        <template v-if="element.langKeys&&element.langKeys.length&&dictName">
                          <span v-for="(lang,li) in element.langKeys" :key="lang" :style="{fontSize:element.fontSize+'pt',fontFamily:fontMap[element.fontFamily||(lang==='arabic'?'ArialMT':'CenturyGothic')],fontWeight:element.bold?'bold':'normal',letterSpacing:(element.letterSpacing||0)+'pt',lineHeight:element.lineHeight||1.5,whiteSpace:'pre-wrap',wordBreak:'break-all',direction:'ltr',textAlign:element.alignment||'left',display:'block',marginTop:li>0?'6px':'0'}" v-html="formatLangHtml(element,lang)" />
                        </template>
                        <span v-else class="rich-preview" v-html="elementHtml(element)" /></template>
                      <template v-else-if="element.type==='image'">
                        <template v-if="element.useMapping&&element.imageMap?.length"><div class="img-placeholder"><span>动态图</span><span style="color:#409eff">{{element.imageMap.length}}条映射</span></div></template>
                        <img v-else-if="element.src" :src="resolveImageUrl(element.src)" style="max-width:100%;max-height:100%;display:block" /><div v-else class="img-placeholder">图</div>
                      </template>
                      <template v-else-if="element.type==='table'">
                        <table class="mini-table" :style="{fontSize:(element.fontSize||5)+'pt',fontFamily:fontMap[element.fontFamily||'FZLTXIHJW--GB1-0'],borderCollapse:element.showBorder!==false?'collapse':'separate'}">
                          <tr v-for="(row,ri) in (element.cells||buildDefaultCells(element))" :key="ri" :style="{height:(element.rowHeight||2.2)*scale+'px'}">
                            <td v-for="(cell,ci) in row" :key="ci" :style="{width:(element.colWidth||4.4)*scale+'px',border:element.showBorder!==false?'2px solid #666':'none',textAlign:cell.textAlign||element.alignment||'center'}">
                              <template v-if="cell.value"><div v-for="(line,li) in cell.value.split('\n')" :key="li">{{line||'\u00A0'}}</div></template><template v-else>&nbsp;</template>
                            </td></tr></table></template></div>
                    <div v-if="selectedId===element.id&&selectedSide==='back'" class="drag-mask"></div>
                    <template v-if="selectedId===element.id&&selectedSide==='back'"><div class="rh w" @pointerdown.stop="onResizeStart($event,element.id,'w','back')" @mousedown.stop.prevent @touchstart.stop.prevent/><div class="rh e" @pointerdown.stop="onResizeStart($event,element.id,'e','back')" @mousedown.stop.prevent @touchstart.stop.prevent/><div class="rh s" @pointerdown.stop="onResizeStart($event,element.id,'s','back')" @mousedown.stop.prevent @touchstart.stop.prevent/></template>
                  </div></template></draggable></div></div></div></div></div>

        <div class="right-panel">
          <template v-if="selectedEl">
            <h4>{{typeLabel(selectedEl.type)}} 属性</h4>
            <div class="prop-row" style="display:flex;gap:10px">
              <span style="flex:1"><label>宽度 (mm)</label><el-input-number v-model="selectedEl.width" :min="5" :max="contentWidth" :step="0.5" size="small" controls-position="right" style="width:100%"/></span>
              <span style="flex:1"><label>高度 (mm)</label><el-input-number v-model="selectedEl.height" :min="2" :max="canvasHeight" :step="0.5" size="small" controls-position="right" style="width:100%"/></span>
            </div>
            <template v-if="selectedEl.type==='text'">
              <!-- 正面：富文本编辑 -->
              <template v-if="selectedSide==='front'">
                <div class="prop-row">
                  <label>内容</label>
                  <div class="rich-toolbar">
                    <span class="tool-group">
                      <el-select v-model="toolFontSize" size="small" style="width:60px" @change="applyFontSize"><el-option v-for="s in [2,3,4,5,6,7,8,10,12,14,16,20,24,28,32,36,48]" :key="s" :label="s" :value="s"/></el-select>
                      <el-select v-model="toolFontFamily" size="small" style="width:130px" @change="applyFontFamily"><el-option label="FZLTXIHJW--GB1-0" value="FZLTXIHJW--GB1-0"/><el-option label="FZLTTHPRO--GB1-4" value="FZLTTHPRO--GB1-4"/><el-option label="FZLTZHUNHPRO--GB1-4" value="FZLTZHUNHPRO--GB1-4"/><el-option label="FZLTCHPRO--GB1-4" value="FZLTCHPRO--GB1-4"/><el-option label="CenturyGothic" value="CenturyGothic"/><el-option label="Gotham-Book" value="Gotham-Book"/><el-option label="ArialMT" value="ArialMT"/><el-option label="MiSans-Regular" value="MiSans-Regular"/></el-select>
                      <el-button size="small" @click="applyBold" :type="toolBold ? 'primary' : 'default'"><strong>B</strong></el-button>
                    </span>
                  </div>
                  <div class="prop-row" style="display:flex;gap:6px;margin-bottom:6px">
                    <span style="flex:1"><label>字间距</label>
                      <el-select v-model="toolLetterSpacing" size="small" style="width:100%" @change="applyLetterSpacing">
                        <el-option label="0" :value="0"/><el-option label="0.5 pt" :value="0.5"/><el-option label="1 pt" :value="1"/><el-option label="1.5 pt" :value="1.5"/><el-option label="2 pt" :value="2"/><el-option label="3 pt" :value="3"/><el-option label="5 pt" :value="5"/>
                      </el-select>
                    </span>
                    <span style="flex:1"><label>行间距</label>
                      <el-select v-model="toolLineHeight" size="small" style="width:100%" @change="applyLineHeight">
                        <el-option label="1.0" :value="1.0"/><el-option label="1.2" :value="1.2"/><el-option label="1.5" :value="1.5"/><el-option label="1.8" :value="1.8"/><el-option label="2.0" :value="2.0"/><el-option label="2.5" :value="2.5"/><el-option label="3.0" :value="3.0"/>
                      </el-select>
                    </span>
                  </div>
                  <div ref="richEditor" class="rich-editor" contenteditable="true" :style="{lineHeight: toolLineHeight}" @input="onRichInput" @mouseup="onRichMouseUp" @keyup="onRichMouseUp" @keydown.enter.prevent="onRichEnter" @paste="onRichPaste" v-html="richHtml"></div>
                </div>
              </template>
              <!-- 反面元素：直接属性控件 -->
              <template v-if="selectedSide==='back'">
                <div class="prop-row" style="display:flex;gap:6px">
                  <span style="flex:1"><label>字号</label><el-input-number v-model="selectedEl.fontSize" :min="2" :max="48" :step="0.5" size="small" style="width:100%" controls-position="right"/></span>
                  <span style="flex:1"><label>字体</label><el-select v-model="selectedEl.fontFamily" size="small" style="width:100%"><el-option label="FZLTXIHJW--GB1-0" value="FZLTXIHJW--GB1-0"/><el-option label="FZLTTHPRO--GB1-4" value="FZLTTHPRO--GB1-4"/><el-option label="FZLTZHUNHPRO--GB1-4" value="FZLTZHUNHPRO--GB1-4"/><el-option label="FZLTCHPRO--GB1-4" value="FZLTCHPRO--GB1-4"/><el-option label="CenturyGothic" value="CenturyGothic"/><el-option label="Gotham-Book" value="Gotham-Book"/><el-option label="ArialMT" value="ArialMT"/><el-option label="MiSans-Regular" value="MiSans-Regular"/></el-select></span>
                </div>
                <div class="prop-row"><el-checkbox v-model="selectedEl.bold">粗体</el-checkbox></div>
                <div class="prop-row" style="display:flex;gap:6px">
                  <span style="flex:1"><label>字间距</label>
                    <el-select v-model="selectedEl.letterSpacing" size="small" style="width:100%"><el-option label="0" :value="0"/><el-option label="0.5 pt" :value="0.5"/><el-option label="1 pt" :value="1"/><el-option label="1.5 pt" :value="1.5"/><el-option label="2 pt" :value="2"/><el-option label="3 pt" :value="3"/><el-option label="5 pt" :value="5"/></el-select>
                  </span>
                  <span style="flex:1"><label>行间距</label>
                    <el-select v-model="selectedEl.lineHeight" size="small" style="width:100%"><el-option label="1.0" :value="1.0"/><el-option label="1.2" :value="1.2"/><el-option label="1.5" :value="1.5"/><el-option label="1.8" :value="1.8"/><el-option label="2.0" :value="2.0"/><el-option label="2.5" :value="2.5"/><el-option label="3.0" :value="3.0"/></el-select>
                  </span>
                </div>
              </template>
              <div class="prop-row"><label>水平对齐</label><el-radio-group v-model="selectedEl.alignment" size="small"><el-radio-button value="left">左</el-radio-button><el-radio-button value="center">中</el-radio-button><el-radio-button value="right">右</el-radio-button></el-radio-group></div>
              <div class="prop-row"><label>垂直对齐</label><el-radio-group v-model="selectedEl.valign" size="small"><el-radio-button value="top">顶</el-radio-button><el-radio-button value="middle">中</el-radio-button><el-radio-button value="bottom">底</el-radio-button></el-radio-group></div>
              <div class="prop-row"><label>Excel 表头</label><el-input v-model="selectedEl.key" size="small"/></div>
              <template v-if="needsTranslation && selectedSide==='back'">
                <div class="prop-row"><label>翻译语言</label><el-select v-model="selectedLangKeys" multiple size="small" style="width:100%"><el-option v-for="l in translateLangs" :key="l" :label="langLabel(l)" :value="l"/></el-select></div>
              </template>
            </template>
            <template v-if="selectedEl.type==='table'">
              <div class="prop-row"><label>列宽 (mm)</label><el-input-number v-model="selectedEl.colWidth" :min="2" :max="contentWidth" :step="0.5" size="small" controls-position="right" class="prop-input"/></div>
              <div class="prop-row"><label>行高 (mm)</label><el-input-number v-model="selectedEl.rowHeight" :min="2" :max="canvasHeight" :step="0.5" size="small" controls-position="right" class="prop-input"/></div>
              <div style="font-size:12px;font-weight:600;margin:8px 0 4px;color:#666">单元格预览</div>
              <table v-if="selectedEl.cells&&selectedEl.cells.length" class="cell-editor-grid" style="width:100%">
                <tr v-for="(row,ri) in selectedEl.cells" :key="ri" :style="{height:(selectedEl.rowHeight||2.2)*2+'px'}">
                  <td v-for="(cell,ci) in row" :key="ci" :style="{width:(selectedEl.colWidth||4.4)*2+'px'}">
                    <textarea :value="cell.value" :style="{fontSize:(selectedEl.fontSize||5)+'pt',fontFamily:fontMap[selectedEl.fontFamily||'FZLTXIHJW--GB1-0'],textAlign:cell.textAlign||selectedEl.alignment||'center'}" :rows="Math.max(1,(cell.value||'').split('\n').length)" class="cell-textarea" @input="onCellEdit(selectedEl.id,ri,ci,$event.target.value)"/>
                  </td></tr></table>
              <div v-else style="color:#999;font-size:11px;text-align:center;padding:8px">请在【表格内容】中输入数据</div>
              <div class="prop-row"><span style="display:flex;gap:8px;align-items:flex-end"><span style="flex:1"><label>行</label><el-input-number v-model="tableEdit.rows" :min="1" :max="10" size="small" controls-position="right" style="width:100%"/></span><span style="flex:1"><label>列</label><el-input-number v-model="tableEdit.cols" :min="1" :max="10" size="small" controls-position="right" style="width:100%"/></span><el-button size="small" @click="applyTableSize(selectedEl.id)">应用</el-button></span></div>
              <div class="prop-row"><label>字号</label><el-input-number v-model="selectedEl.fontSize" :min="2" :max="48" :step="0.5" size="small" controls-position="right" class="prop-input"/></div>
              <div class="prop-row"><label>font</label><el-select v-model="selectedEl.fontFamily" size="small" class="prop-input"><el-option label="FZLTXIHJW--GB1-0" value="FZLTXIHJW--GB1-0"/><el-option label="FZLTTHPRO--GB1-4" value="FZLTTHPRO--GB1-4"/><el-option label="FZLTZHUNHPRO--GB1-4" value="FZLTZHUNHPRO--GB1-4"/><el-option label="FZLTCHPRO--GB1-4" value="FZLTCHPRO--GB1-4"/><el-option label="CenturyGothic" value="CenturyGothic"/><el-option label="Gotham-Book" value="Gotham-Book"/><el-option label="ArialMT" value="ArialMT"/><el-option label="MiSans-Regular" value="MiSans-Regular"/></el-select></div>
              <div class="prop-row"><label>显示边框</label><el-switch v-model="selectedEl.showBorder" size="small"/></div>
              <div class="prop-row"><label>水平对齐</label><el-radio-group v-model="selectedEl.alignment" size="small"><el-radio-button value="left">左</el-radio-button><el-radio-button value="center">中</el-radio-button><el-radio-button value="right">右</el-radio-button></el-radio-group></div>
              <div class="prop-row"><label>垂直对齐</label><el-radio-group v-model="selectedEl.valign" size="small"><el-radio-button value="top">顶</el-radio-button><el-radio-button value="middle">中</el-radio-button><el-radio-button value="bottom">底</el-radio-button></el-radio-group></div>
              <div class="prop-row"><label>Excel表头</label><el-input v-model="selectedEl.key" size="small" placeholder="对应Excel列名"/></div>
            </template>
            <template v-if="selectedEl.type==='image'">
              <div class="prop-row"><label>选择图片</label><el-select v-model="selectedEl.src" size="small" class="prop-input" filterable clearable><el-option v-for="img in imageList" :key="img.ID" :label="img.name" :value="img.url"><div style="display:flex;align-items:center;gap:8px"><img :src="resolveImageUrl(img.url)" style="width:28px;height:28px;object-fit:contain"/><span>{{img.name}}</span></div></el-option></el-select></div>
              <div class="prop-row"><label>水平对齐</label><el-radio-group v-model="selectedEl.alignment" size="small"><el-radio-button value="left">左</el-radio-button><el-radio-button value="center">中</el-radio-button><el-radio-button value="right">右</el-radio-button></el-radio-group></div>
              <div class="prop-row"><label>垂直对齐</label><el-radio-group v-model="selectedEl.valign" size="small"><el-radio-button value="top">顶</el-radio-button><el-radio-button value="middle">中</el-radio-button><el-radio-button value="bottom">底</el-radio-button></el-radio-group></div>
              <div class="prop-row"><label>Excel表头</label><el-input v-model="selectedEl.key" size="small"/></div>
              <div class="prop-row"><label>根据值匹配图片</label><el-switch v-model="selectedEl.useMapping" size="small"/></div>
              <template v-if="selectedEl.useMapping"><div v-for="(m,mi) in (selectedEl.imageMap||[])" :key="mi" class="prop-row" style="display:flex;gap:4px"><span style="flex:1"><el-input v-model="m.value" size="small" placeholder="值"/></span><span style="flex:2"><el-select v-model="m.src" size="small" filterable clearable><el-option v-for="img in imageList" :key="img.ID" :label="img.name" :value="img.url"/></el-select></span><el-button link type="danger" size="small" @click="selectedEl.imageMap.splice(mi,1)"><el-icon><Delete/></el-icon></el-button></div><div class="prop-row"><el-button size="small" @click="addImageMapping(selectedEl)">+ 添加映射</el-button></div></template>
            </template>
            <el-button type="danger" size="small" class="prop-input" style="margin-top:16px" @click="removeElement(selectedEl.id)">删除元素</el-button>
          </template>
          <div v-else class="no-select">请选择一个元素</div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showSetupDialog" :title="setupEdit?'修改模板配置':'新建标签模板'" width="480px" destroy-on-close>
      <el-form ref="setupFormRef" :model="setupForm" :rules="setupRules" label-position="top">
        <el-form-item label="模板名称" prop="name"><el-input v-model="setupForm.name"/></el-form-item>
        <el-row :gutter="12">
          <el-col :span="6"><el-form-item label="标签宽度"><el-input-number v-model="setupForm.labelWidth" :min="10" :max="200" :step="0.5" style="width:100%"/></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="标签高度"><el-input-number v-model="setupForm.labelHeight" :min="20" :max="300" :step="0.5" style="width:100%"/></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="顶缝"><el-input-number v-model="setupForm.headSeam" :min="0" :max="30" :step="0.5" style="width:100%"/></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="左右间距"><el-input-number v-model="setupForm.marginLR" :min="0" :max="20" :step="0.5" style="width:100%"/></el-form-item></el-col>
        </el-row>
        <el-form-item label="默认字体"><el-select v-model="setupForm.defaultFontFamily" style="width:100%"><el-option label="FZLTXIHJW--GB1-0" value="FZLTXIHJW--GB1-0"/><el-option label="FZLTTHPRO--GB1-4" value="FZLTTHPRO--GB1-4"/><el-option label="FZLTZHUNHPRO--GB1-4" value="FZLTZHUNHPRO--GB1-4"/><el-option label="FZLTCHPRO--GB1-4" value="FZLTCHPRO--GB1-4"/><el-option label="CenturyGothic" value="CenturyGothic"/><el-option label="Gotham-Book" value="Gotham-Book"/><el-option label="ArialMT" value="ArialMT"/><el-option label="MiSans-Regular" value="MiSans-Regular"/></el-select></el-form-item>
        <el-form-item label="默认字号"><el-input-number v-model="setupForm.defaultFontSizePt" :min="2" :max="20" :step="0.5" style="width:100%"/></el-form-item>
        <el-form-item label="需要翻译面"><el-switch v-model="setupForm.needsTranslation"/></el-form-item>
        <el-form-item label="翻译字典" v-if="setupForm.needsTranslation"><el-select v-model="setupForm.dictName" style="width:100%"><el-option label="巴拉" value="巴拉"/><el-option label="森马" value="森马"/></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="showSetupDialog=false">取消</el-button><el-button type="primary" @click="handleSetupConfirm">{{setupEdit?'保存配置':'进入编辑'}}</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, markRaw, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { EditPen, Picture, Grid, Delete, Search, Plus, ArrowLeft, Menu } from '@element-plus/icons-vue'
import draggable from 'vuedraggable'
import { saveTemplate, loadTemplate, listTemplate, deleteTemplate, publishTemplate, unpublishTemplate } from '@/plugin/label/api/template'
import { getImageList } from '@/plugin/image/api/image'
import { translateMany, getCachedTranslation, makeTranslateLookup, clearDictionaryCache, formatArabicDisplayHtml, stripBidiMarks } from '@/plugin/label/utils/dictionary'
import { exportLabelPDF } from '@/plugin/label/utils/pdfExport'

defineOptions({ name: 'LabelEditor' })

const resolveImageUrl = (url) => { if (!url) return ''; if (/^https?:\/\//.test(url) || /^data:/.test(url)) return url; return '/' + url.replace(/^\/+/, '') }
const imageList = ref([])
async function fetchImageList() { try { const res = await getImageList({ page: 1, pageSize: 999 }); if (res.code === 0) imageList.value = res.data?.list || [] } catch { imageList.value = [] } }

const paletteItems = [ { type: 'text', label: '文本', icon: markRaw(EditPen) }, { type: 'image', label: '图片', icon: markRaw(Picture) }, { type: 'table', label: '表格', icon: markRaw(Grid) } ]
function onDragStart(e, type) { e.dataTransfer.effectAllowed = 'copy'; e.dataTransfer.setData('text/plain', type) }

const exporting = ref(false), viewMode = ref('list'), searchKeyword = ref(''), allTemplates = ref([])
const templateName = ref(''), labelWidth = ref(80), labelHeight = ref(120), headSeam = ref(8), marginLR = ref(2)
const defaultFont = ref('FZLTXIHJW--GB1-0'), defaultFontSize = ref(5), elements = ref([]), selectedId = ref(null), selectedSide = ref('front')
const zoomLevel = ref(8), scale = computed(() => zoomLevel.value), canvasHeight = computed(() => Math.max(5, labelHeight.value - headSeam.value))
const needsTranslation = ref(false), translateLangs = ref([]), dictName = ref('巴拉'), translatedElements = ref([])
const translationTick = ref(0)

const tableEdit = ref({ rows: 2, cols: 6, text: '130\t140\t150\t160\t170\t175\n127\t118\t143\t159\t177\t183' })
function syncTableEdit(el) { if (!el || el.type !== 'table') return; tableEdit.value.rows = el.cells?.length || el.rows || 2; tableEdit.value.cols = (el.cells && el.cells[0]?.length) || el.cols || 6; tableEdit.value.text = cellsToText(el.cells || buildDefaultCells(el)) }
function cellsToText(cells) { if (!cells || !cells.length) return ''; return cells.map(r => r.map(c => c.value || '').join('\t')).join('\n') }
function buildDefaultCells(el) { const rows = el.rows || 2, cols = el.cols || 6; return Array.from({ length: rows }, () => Array.from({ length: cols }, () => ({ value: '', textAlign: 'center' }))) }
function applyTableSize(elId) { const el = activeElements.value.find(e => e.id === elId); if (!el || el.type !== 'table') return; const nr = tableEdit.value.rows, nc = tableEdit.value.cols; const old = el.cells || buildDefaultCells(el); const cells = []; for (let r = 0; r < nr; r++) { const row = []; for (let c = 0; c < nc; c++) row.push((old[r] && old[r][c]) ? { ...old[r][c] } : { value: '', textAlign: el.alignment || 'center' }); cells.push(row) } el.cells = cells; el.rows = nr; el.cols = nc; syncTableEdit(el) }
function onCellEdit(elId, ri, ci, value) { const el = activeElements.value.find(e => e.id === elId); if (!el || el.type !== 'table') return; if (!el.cells) el.cells = buildDefaultCells(el); while (el.cells.length <= ri) el.cells.push([]); while (el.cells[ri].length <= ci) el.cells[ri].push({ value: '', textAlign: el.alignment || 'center' }); el.cells[ri][ci].value = value; tableEdit.value.text = cellsToText(el.cells) }

const contentWidth = computed(() => Math.max(0, labelWidth.value - marginLR.value * 2))
const filteredTemplates = computed(() => { if (!searchKeyword.value) return allTemplates.value; const kw = searchKeyword.value.toLowerCase(); return allTemplates.value.filter(t => t.name.toLowerCase().includes(kw)) })

const showSetupDialog = ref(false), setupEdit = ref(false), setupFormRef = ref()
const setupForm = ref({ name: '', labelWidth: 80, labelHeight: 120, headSeam: 8, marginLR: 2, defaultFontFamily: 'FZLTXIHJW--GB1-0', defaultFontSizePt: 5, needsTranslation: false, dictName: '巴拉' })
const setupRules = { name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }] }

async function refreshList() { try { const res = await listTemplate(); if (res.code === 0) { allTemplates.value = (res.data || []).map(t => ({ name: t.name, labelWidth: t.labelWidth || 80, labelHeight: t.labelHeight || 120, isPublished: t.isPublished || false, elementCount: (() => { try { return JSON.parse(t.elements||'[]').length } catch { return 0 } })() })) } } catch (e) { console.error(e) } }
refreshList()

async function handleDeleteTpl(name) { try { await ElMessageBox.confirm(`确定删除模板"${name}"？`, '提示', { type: 'warning' }); const res = await deleteTemplate({ name }); if (res.code === 0) { ElMessage.success('删除成功'); refreshList() } } catch {} }

async function handlePublish(name) { try { const res = await publishTemplate({ name }); if (res.code === 0) { ElMessage.success('发布成功'); refreshList() } } catch { ElMessage.error('发布失败') } }

async function handleUnpublish(name) { try { const res = await unpublishTemplate({ name }); if (res.code === 0) { ElMessage.success('已取消发布'); refreshList() } } catch { ElMessage.error('取消发布失败') } }

async function openTemplate(name) {
  fetchImageList()
  const res = await loadTemplate({ name })
  if (res.code === 0) {
    const t = res.data
    templateName.value = t.name
    labelWidth.value = t.labelWidth || 80
    labelHeight.value = t.labelHeight || 120
    headSeam.value = t.headSeam || 8
    marginLR.value = t.marginLR ?? 2
    needsTranslation.value = t.needsTranslation || false
    dictName.value = t.dictName || '巴拉'
    try { translateLangs.value = JSON.parse(t.translateLangs || '[]') } catch { translateLangs.value = [] }
    try {
      elements.value = JSON.parse(t.elements || '[]')
      const maxId = elements.value.reduce((max, e) => Math.max(max, parseInt(e.id?.replace('el_', '') || '0')), 0)
      idCounter = maxId + 1
    } catch { elements.value = [] }
    try { translatedElements.value = JSON.parse(t.translatedElements || '[]') } catch { translatedElements.value = [] }
    selectedId.value = null
    selectedSide.value = 'front'
    viewMode.value = 'editor'
    clearDictionaryCache()
    await refreshBackTranslations()
  }
}
function backToList() { viewMode.value = 'list'; refreshList() }

function handleSetupConfirm() {
  setupFormRef.value?.validate(async (valid) => {
    if (!valid) return
    const f = setupForm.value
    templateName.value = f.name
    labelWidth.value = f.labelWidth
    labelHeight.value = f.labelHeight
    headSeam.value = f.headSeam || 0
    marginLR.value = f.marginLR ?? 2
    defaultFont.value = f.defaultFontFamily
    defaultFontSize.value = f.defaultFontSizePt
    needsTranslation.value = f.needsTranslation
    dictName.value = f.dictName || '巴拉'
    if (setupEdit.value) {
      ElMessage.success('配置已更新')
      clearDictionaryCache()
      await refreshBackTranslations()
    } else {
      elements.value = []
      translatedElements.value = []
      idCounter = 1
      selectedId.value = null
      selectedSide.value = 'front'
      viewMode.value = 'editor'
      fetchImageList()
      clearDictionaryCache()
    }
    showSetupDialog.value = false
    setupEdit.value = false
  })
}

async function handleSave() { const data = { name: templateName.value, labelWidth: labelWidth.value, labelHeight: labelHeight.value, headSeam: headSeam.value, marginLR: marginLR.value, needsTranslation: needsTranslation.value, dictName: dictName.value, translateLangs: JSON.stringify(translateLangs.value), elements: JSON.stringify(elements.value), translatedElements: JSON.stringify(translatedElements.value) }; const res = await saveTemplate(data); if (res.code === 0) ElMessage.success('保存成功') }

async function handleExportPDF() {
  if (!elements.value.length) { ElMessage.warning('请先添加元素'); return }
  exporting.value = true
  try {
    const r = (s) => { if (!s) return ''; if (/^https?:\/\//.test(s) || /^data:/.test(s)) return s; return '/' + s.replace(/^\/+/, '') }
    const re = arr => arr.map(el => el.type === 'image' && el.src ? { ...el, src: r(el.src) } : el)
    if (needsTranslation.value && translateLangs.value.length && dictName.value) {
      await refreshBackTranslations()
    }
    const doc = await exportLabelPDF({
      frontElements: re(elements.value),
      backElements: needsTranslation.value && translateLangs.value.length ? re(translatedElements.value) : null,
      config: { labelWidth: labelWidth.value, labelHeight: labelHeight.value, headSeam: headSeam.value, marginLR: marginLR.value },
      translateInfo: needsTranslation.value ? { lookup: makeTranslateLookup(dictName.value), translateLangs: translateLangs.value } : null,
      fileName: `${templateName.value || '标签'}.pdf`
    })
    if (doc) doc.save(`${templateName.value || '标签'}.pdf`)
    ElMessage.success('PDF 导出成功')
  } catch (e) {
    console.error(e)
    ElMessageBox.alert(e.message || '导出失败', '导出失败', { confirmButtonText: '知道了', type: 'error', dangerouslyUseHTMLString: true })
  } finally {
    exporting.value = false
  }
}
function langLabel(lang) { return { english: '英文', russian: '俄文', arabic: '阿语', indonesian: '印尼' }[lang] || lang }

const activeElements = computed(() => selectedSide.value === 'front' ? elements.value : translatedElements.value)
const selectedEl = computed(() => activeElements.value.find(e => e.id === selectedId.value) || null)
const fontMap = { 'FZLTXIHJW--GB1-0': "'FZLTXIHJW--GB1-0', SimSun, sans-serif", CenturyGothic: "CenturyGothic, 'Century Gothic', sans-serif", ArialMT: 'ArialMT, Arial, sans-serif', 'MiSans-Regular': 'MiSans-Regular, sans-serif', 'FZLTTHPRO--GB1-4': "'FZLTTHPRO--GB1-4', 'FZLTXIHJW--GB1-0', SimSun, sans-serif", 'FZLTZHUNHPRO--GB1-4': "'FZLTZHUNHPRO--GB1-4', 'FZLTXIHJW--GB1-0', SimSun, sans-serif", 'Gotham-Book': "'Gotham-Book', CenturyGothic, 'Century Gothic', sans-serif", 'FZLTCHPRO--GB1-4': "'FZLTCHPRO--GB1-4', 'FZLTXIHJW--GB1-0', SimSun, sans-serif", FZ: "'FZLTXIHJW--GB1-0', SimSun, sans-serif", GO: "CenturyGothic, 'Century Gothic', sans-serif", ARIAL: 'ArialMT, Arial, sans-serif', MiSans: 'MiSans-Regular, sans-serif' }

let idCounter = 1
function typeLabel(t) { return { text:'text', image:'image', table:'table' }[t] || t }
function selectElement(id, side) { selectedId.value = id; selectedSide.value = side || 'front'; if (id) { const el = activeElements.value.find(e => e.id === id); if (el?.type === 'text') loadRichHtml(el) } }

function onCanvasDrop(e) { const type = e.dataTransfer.getData('text/plain'); if (!['text','image','table'].includes(type)) return; const base = { id: 'el_' + (idCounter++), type, width: contentWidth.value, height: type === 'table' ? 10 : 10, fontSize: defaultFontSize.value, fontFamily: defaultFont.value, alignment: type === 'table' ? 'center' : 'left', valign: 'middle', key: '' }; const el = type === 'text' ? { ...base, text: 'text', bold: false, letterSpacing: 0, lineHeight: 1.5, enableTranslation: false, langKeys: [] } : type === 'image' ? { ...base, src: '', fit: 'contain' } : { ...base, rows: 2, cols: 6, colWidth: 4.4, rowHeight: 2.2, showBorder: true, cells: [[{ value:'130',textAlign:'center'},{ value:'140',textAlign:'center'},{ value:'150',textAlign:'center'},{ value:'160',textAlign:'center'},{ value:'170',textAlign:'center'},{ value:'175',textAlign:'center'}],[{ value:'127',textAlign:'center'},{ value:'118',textAlign:'center'},{ value:'143',textAlign:'center'},{ value:'159',textAlign:'center'},{ value:'177',textAlign:'center'},{ value:'183',textAlign:'center'}]], width: contentWidth.value, height: 10 }; elements.value.push(el); if (needsTranslation.value) { const copy = JSON.parse(JSON.stringify(el)); copy.text = ''; copy.langKeys = []; translatedElements.value.push(copy) } selectedId.value = el.id; selectedSide.value = 'front' }

function elementContentStyle(el) { const v = { top:'flex-start', middle:'center', bottom:'flex-end' }; const h = { left:'flex-start', center:'center', right:'flex-end' }; return { maxWidth: contentWidth.value * scale.value + 'px', width: (el.width || contentWidth.value) * scale.value + 'px', textAlign: el.alignment || 'left', alignItems: h[el.alignment||'left'], justifyContent: v[el.valign||'middle'] } }
function rowHeightPx(el) { return (el.height || 10) * scale.value }
function addImageMapping(el) { if (!el.imageMap) el.imageMap = []; el.imageMap.push({ value: '', src: '' }) }
function removeElement(id, side) { const s = side || selectedSide.value; if (s === 'front') elements.value = elements.value.filter(e => e.id !== id); else translatedElements.value = translatedElements.value.filter(e => e.id !== id); if (selectedId.value === id) selectedId.value = null }
/** 新增一个反面文本元素 */
function addBackElement() {
  const be = {
    id: 'el_' + (idCounter++), type: 'text', width: contentWidth.value, height: 10,
    fontSize: defaultFontSize.value, fontFamily: defaultFont.value,
    alignment: 'left', valign: 'middle', key: '',
    text: '', bold: false, letterSpacing: 0, lineHeight: 1.5,
    enableTranslation: false, langKeys: []
  }
  translatedElements.value.push(be)
  selectedId.value = be.id; selectedSide.value = 'back'
}

// 反面元素选语言 → 自动标记翻译状态 + 触发翻译
const selectedLangKeys = ref([])
watch(() => JSON.stringify({ id: selectedId.value, side: selectedSide.value }), () => {
  const el = activeElements.value.find(e => e.id === selectedId.value)
  selectedLangKeys.value = el?.langKeys ? [...el.langKeys] : []
}, { immediate: true })
watch(selectedLangKeys, (v) => {
  const el = activeElements.value.find(e => e.id === selectedId.value)
  if (!el || selectedSide.value !== 'back') return
  el.langKeys = [...v]
  if (v.length) { el.enableTranslation = true; translateOne(el) }
  else el.enableTranslation = false
}, { deep: true })

function findFrontByKey(key) { return key ? elements.value.find(e => e.key === key) : null }

/** 收集反面所需的翻译任务并请求后端 */
async function refreshBackTranslations() {
  if (!needsTranslation.value || !dictName.value) return
  const items = []
  for (const backEl of translatedElements.value) {
    if (backEl.type !== 'text' || !backEl.langKeys?.length) continue
    const fe = findFrontByKey(backEl.key)
    const text = fe?.text
    if (!text || !String(text).trim() || text === '文本' || text === 'text') continue
    items.push({ text, langs: [...backEl.langKeys] })
  }
  if (!items.length) {
    translationTick.value++
    return
  }
  try {
    await translateMany(dictName.value, items)
  } catch (e) {
    console.error(e)
    ElMessage.warning(e.message || '翻译失败')
  }
  translationTick.value++
}

/** 反面元素 → 获取指定语言的翻译文本（从正面按 key 取源文本） */
function getLangTranslation(backEl, lang) {
  void translationTick.value
  if (!lang || !dictName.value) return '...'
  const fe = findFrontByKey(backEl.key)
  if (!fe || !fe.text || fe.text === '文本' || fe.text === 'text') return '文本'
  const cached = getCachedTranslation(dictName.value, fe.text, lang)
  return cached !== null ? stripBidiMarks(cached) : '...'
}

function formatLangHtml(backEl, lang) {
  const text = getLangTranslation(backEl, lang)
  if (lang === 'arabic' && text && text !== '...' && text !== '文本') {
    return formatArabicDisplayHtml(text)
  }
  return escapeHtml(text)
}

/** 执行翻译：仅更新当前元素的 enableTranslation + langKeys，不创建/合并其他元素 */
async function translateOne(el) {
  const srcKey = el.key
  const srcEl = srcKey ? findFrontByKey(srcKey) : el
  const t = srcEl?.text
  if (!t || !String(t).trim() || t === '文本' || t === 'text') { ElMessage.warning('请先在正面元素中输入文本内容'); return }
  const langs = el.langKeys
  if (!langs || !langs.length) { ElMessage.warning('请选择至少一种翻译语言'); return }
  if (!dictName.value) { ElMessage.warning('请先选择翻译字典'); return }
  try {
    await translateMany(dictName.value, [{ text: t, langs: [...langs] }])
    translationTick.value++
  } catch (e) {
    ElMessage.warning(e.message || '翻译失败')
    return
  }
  el.langKeys = [...langs]
  el.key = srcKey
  // 不覆盖已有字体/大小/宽高等属性
}

let dragState = null
function mm(v) { return Math.round(v / scale.value * 10) / 10 }
function onResizeStart(e, id, handle, side) { e.stopPropagation(); e.preventDefault(); const el = (side === 'back' ? translatedElements.value : elements.value).find(e => e.id === id) || (side === 'back' ? elements.value : translatedElements.value).find(e => e.id === id); if (!el) return; dragState = { id, handle, side, sx: e.clientX, sy: e.clientY, sw: el.width, sh: el.height }; e.target.setPointerCapture(e.pointerId) }
function onPointerMove(e) { if (!dragState) return; const dx = mm(e.clientX - dragState.sx), dy = mm(e.clientY - dragState.sy); const el = (dragState.side === 'back' ? translatedElements.value : elements.value).find(e => e.id === dragState.id) || (dragState.side === 'back' ? elements.value : translatedElements.value).find(e => e.id === dragState.id); if (!el) return; const h = dragState.handle; if (h.includes('w')) el.width = Math.max(5, Math.min(contentWidth.value, dragState.sw - dx)); if (h.includes('e')) el.width = Math.max(5, Math.min(contentWidth.value, dragState.sw + dx)); if (h.includes('s')) el.height = Math.max(5, Math.min(canvasHeight.value, dragState.sh + dy)) }
function onPointerUp() { dragState = null }

// rich text editor state
const richEditor = ref(null)
const richHtml = ref('')
const toolFontSize = ref(5)
const toolFontFamily = ref('FZLTXIHJW--GB1-0')
const toolBold = ref(false)
const toolLetterSpacing = ref(0)
const toolLineHeight = ref(1.5)
let richSaveTimer = null

function elementHtml(el) {
  if (!el) return ''
  const ff = el.fontFamily || 'FZLTXIHJW--GB1-0'
  const fs = el.fontSize || 5
  const ls = el.letterSpacing ? `letter-spacing:${el.letterSpacing}pt;` : ''
  const lh = el.lineHeight ? `line-height:${el.lineHeight};` : ''
  if (el.html) return `<span style="font-size:${fs}pt;font-family:${cssFont(ff)};${ls}${lh}">${(el.html||'')}</span>`
  return `<span style="font-size:${fs}pt;font-family:${cssFont(ff)};${ls}${lh}">${escapeHtml(el.text||'text')}</span>`
}

function buildDefaultHtml(el) {
  if (el.html) return el.html
  const fs = el.fontSize || 5
  const ff = el.fontFamily || 'FZLTXIHJW--GB1-0'
  const bw = el.bold ? 'bold' : 'normal'
  const ls = el.letterSpacing ? `letter-spacing:${el.letterSpacing}pt;` : ''
  const lh = el.lineHeight ? `line-height:${el.lineHeight};` : ''
  const text = el.text || 'text'
  return `<span style="font-size:${fs}pt;font-family:${cssFont(ff)};font-weight:${bw};${ls}${lh}white-space:pre-wrap">${escapeHtml(text)}</span>`
}

function loadRichHtml(el) {
  richHtml.value = buildDefaultHtml(el)
  toolFontSize.value = el.fontSize || 5
  toolFontFamily.value = el.fontFamily || 'FZLTXIHJW--GB1-0'
  toolBold.value = !!el.bold
  toolLetterSpacing.value = el.letterSpacing || 0
  toolLineHeight.value = el.lineHeight || 1.5
}

function onRichInput() {
  if (!selectedEl.value || selectedEl.value.type !== 'text') return
  const el = selectedEl.value
  const div = richEditor.value
  if (!div) return
  el.html = div.innerHTML
  el.text = div.textContent || ''
  // 同步工具栏状态：检测当前光标处的格式
  syncToolbarFromSelection()
}

function syncToolbarFromSelection() {
  const div = richEditor.value
  if (!div) return
  const sel = window.getSelection()
  if (!sel.rangeCount) return
  let node = sel.rangeCount ? sel.getRangeAt(0).commonAncestorContainer : null
  if (!node) return
  if (node.nodeType === 3) node = node.parentNode
  // 向上查找最近的格式化 span 或 b 标签
  let el = node
  while (el && el !== div) {
    if (el.nodeType !== 1) { el = el.parentNode; continue }
    const st = el.style || {}
    if (st.fontSize) {
      const m = st.fontSize.match(/^([\d.]+)pt/)
      if (m) toolFontSize.value = parseFloat(m[1])
    }
    if (st.fontFamily) {
      const ff = st.fontFamily.replace(/['"]/g, '').split(',')[0].trim()
      // 反向映射 CSS font-family → PostScript name
      for (const [ps, css] of Object.entries(fontMap)) {
        if (css.startsWith("'" + ff + "'") || css.startsWith(ff)) {
          toolFontFamily.value = ps
          break
        }
      }
    }
    if (st.fontWeight === 'bold' || st.fontWeight === '700') toolBold.value = true
    else if (st.fontWeight === 'normal' || st.fontWeight === '400') toolBold.value = false
    if (el.tagName === 'B' || el.tagName === 'STRONG') toolBold.value = true
    if (st.letterSpacing) {
      const m = st.letterSpacing.match(/^([\d.]+)pt/)
      if (m) toolLetterSpacing.value = parseFloat(m[1])
    }
    el = el.parentNode
  }
}

function onRichEnter() {
  const div = richEditor.value
  if (!div) return
  const sel = window.getSelection()
  if (!sel.rangeCount) return
  const range = sel.getRangeAt(0)
  const br = document.createElement('br')
  range.deleteContents()
  range.insertNode(br)
  range.setStartAfter(br)
  range.setEndAfter(br)
  sel.removeAllRanges()
  sel.addRange(range)
  onRichInput()
}

function onRichPaste(e) {
  e.preventDefault()
  const text = (e.clipboardData || window.clipboardData).getData('text/plain')
  document.execCommand('insertText', false, text)
  onRichInput()
}

function onRichMouseUp() {
  // 鼠标点击/键盘移动光标时同步工具栏状态
  syncToolbarFromSelection()
}

function applyFontSize() {
  restoreSelection()
  const div = richEditor.value
  if (!div) return
  const sel = window.getSelection()
  if (!sel.rangeCount) return
  const range = sel.getRangeAt(0)
  if (range.collapsed) return

  // 直接在选中文字外包裹带 font-size 的 span
  wrapSelectionWithStyle(div, range, 'fontSize', toolFontSize.value + 'pt', 'font-size')
  div.normalize()
  saveSelection()
  onRichInput()
}

function applyFontFamily() {
  restoreSelection()
  const div = richEditor.value
  if (!div) return
  const sel = window.getSelection()
  if (!sel.rangeCount) return
  const range = sel.getRangeAt(0)
  if (range.collapsed) return

  wrapSelectionWithStyle(div, range, 'fontFamily', cssFont(toolFontFamily.value), 'font-family')
  div.normalize()
  saveSelection()
  onRichInput()
}

/** 辅助：在选中文字外包裹 span 并设置样式；若已在同属性 span 内则更新 */
function wrapSelectionWithStyle(div, range, attrKey, attrValue, cssProp) {
  // 检查是否已在同属性 span 内
  let parent = range.commonAncestorContainer
  if (parent.nodeType === 3) parent = parent.parentNode
  let targetSpan = null
  while (parent && parent !== div) {
    if (parent.nodeType === 1 && parent.style && parent.style[attrKey]) {
      targetSpan = parent
      break
    }
    parent = parent.parentNode
  }

  if (targetSpan) {
    // 已在 span 内，更新样式
    targetSpan.style[attrKey] = attrValue
    return
  }

  const span = document.createElement('span')
  span.style[attrKey] = attrValue

  try {
    range.surroundContents(span)
  } catch (e) {
    const fragment = range.extractContents()
    span.appendChild(fragment)
    range.insertNode(span)
  }
}

function applyBold() {
  restoreSelection()
  const div = richEditor.value
  if (!div) return
  const sel = window.getSelection()
  if (!sel.rangeCount) return
  const range = sel.getRangeAt(0)
  if (range.collapsed) return

  toolBold.value = !toolBold.value

  // 检查选中内容是否已经在粗体 span 或 b 标签内
  let parent = range.commonAncestorContainer
  if (parent.nodeType === 3) parent = parent.parentNode
  let boldSpan = null
  while (parent && parent !== div) {
    if (parent.nodeType === 1) {
      const st = parent.style || {}
      if (st.fontWeight === 'bold' || st.fontWeight === '700' || parent.tagName === 'B' || parent.tagName === 'STRONG') {
        boldSpan = parent
        break
      }
    }
    parent = parent.parentNode
  }

  if (boldSpan) {
    // 取消粗体：unwrap
    const p = boldSpan.parentNode
    while (boldSpan.firstChild) {
      p.insertBefore(boldSpan.firstChild, boldSpan)
    }
    p.removeChild(boldSpan)
    p.normalize()
  } else if (toolBold.value) {
    // 添加粗体
    const span = document.createElement('span')
    span.style.fontWeight = 'bold'
    try {
      range.surroundContents(span)
    } catch (e) {
      const fragment = range.extractContents()
      span.appendChild(fragment)
      range.insertNode(span)
    }
  }

  div.normalize()
  saveSelection()
  onRichInput()
}

function applyLetterSpacing() {
  restoreSelection()
  const div = richEditor.value
  if (!div) return

  const sel = window.getSelection()
  if (!sel.rangeCount) return

  const range = sel.getRangeAt(0)
  const ls = toolLetterSpacing.value

  // 检查是否在已有的 letter-spacing span 内
  let parent = range.commonAncestorContainer
  if (parent.nodeType === 3) parent = parent.parentNode
  while (parent && parent !== div) {
    if (parent.nodeType === 1 && parent.style && parent.style.letterSpacing) {
      if (ls === 0) {
        // 移除字间距：unwrap span
        const p = parent.parentNode
        while (parent.firstChild) {
          p.insertBefore(parent.firstChild, parent)
        }
        p.removeChild(parent)
        p.normalize()
      } else {
        parent.style.letterSpacing = ls + 'pt'
      }
      saveSelection()
      onRichInput()
      return
    }
    parent = parent.parentNode
  }

  if (ls === 0) return // 无需添加

  // 选中了文字，包裹在带 letter-spacing 的 span 中
  if (range.collapsed) return // 无选中内容

  const span = document.createElement('span')
  span.style.letterSpacing = ls + 'pt'

  try {
    range.surroundContents(span)
  } catch (e) {
    // 跨元素边界时 extract + wrap
    const fragment = range.extractContents()
    span.appendChild(fragment)
    range.insertNode(span)
  }

  div.normalize()
  saveSelection()
  onRichInput()
}

function applyLineHeight() {
  if (!selectedEl.value || selectedEl.value.type !== 'text') return
  selectedEl.value.lineHeight = toolLineHeight.value
  const div = richEditor.value
  if (div) div.style.lineHeight = toolLineHeight.value
}

let savedRange = null
function saveSelection() {
  const sel = window.getSelection()
  if (sel.rangeCount) savedRange = sel.getRangeAt(0).cloneRange()
}
function restoreSelection() {
  if (!savedRange) return
  const div = richEditor.value
  if (!div) return
  const sel = window.getSelection()
  sel.removeAllRanges()
  try { sel.addRange(savedRange) } catch { div.focus() }
}

watch(selectedId, () => { const el = activeElements.value.find(e => e.id === selectedId.value); if (el?.type === 'text') loadRichHtml(el); if (el?.type === 'table') syncTableEdit(el) })

// 正面文本变更后，防抖请求后端刷新反面译文（保留换行缩进）
let refreshTimer = null
watch(
  () => elements.value.filter(e => e.type === 'text').map(e => `${e.key}\0${e.text}`).join('\x1e'),
  () => {
    if (!needsTranslation.value) return
    clearTimeout(refreshTimer)
    refreshTimer = setTimeout(() => { refreshBackTranslations() }, 400)
  }
)

onMounted(() => { document.addEventListener('pointermove', onPointerMove); document.addEventListener('pointerup', onPointerUp) })
onUnmounted(() => {
  document.removeEventListener('pointermove', onPointerMove)
  document.removeEventListener('pointerup', onPointerUp)
  clearTimeout(refreshTimer)
})

function escapeHtml(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;') }

/** Convert PostScript font name to CSS font-family string (uses @font-face rules) */
function cssFont(psName) {
  return fontMap[psName] || `'${psName}'`
}
</script>
<style>
@font-face { font-family:'FZLTXIHJW--GB1-0'; src:url('/fonts/FZ.TTF') format('truetype') }
@font-face { font-family:'CenturyGothic'; src:url('/fonts/GO.TTF') format('truetype') }
@font-face { font-family:'ArialMT'; src:url('/fonts/ARIAL.TTF') format('truetype') }
@font-face { font-family:'MiSans-Regular'; src:url('/fonts/MiSans-Regular.ttf') format('truetype') }
@font-face { font-family:'FZLTTHPRO--GB1-4'; src:url('/fonts/FZLTTHPRO--GB1-4_0.OTF') format('opentype') }
@font-face { font-family:'FZLTZHUNHPRO--GB1-4'; src:url('/fonts/FZLTZHUNHPRO--GB1-4_0.OTF') format('opentype') }
@font-face { font-family:'Gotham-Book'; src:url('/fonts/GOTHAM-BOOK_0.OTF') format('opentype') }
@font-face { font-family:'FZLTCHPRO--GB1-4'; src:url('/fonts/FZLTCHPRO--GB1-4_0.OTF') format('opentype') }

.label-editor { min-height: 100%; }
.tpl-card { border: 1px solid #e8e8e8; border-radius: 8px; padding: 16px; cursor: pointer; transition: all .2s; background: #fff; }
.tpl-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.1); }

.editor-wrap { display: flex; flex-direction: column; height: calc(100vh - 90px); }
.editor-toolbar { display: flex; align-items: center; justify-content: space-between; padding: 8px 16px; border-bottom: 1px solid #e8e8e8; background: #fff; flex-shrink: 0; overflow-x: auto; }
.editor-body { flex: 1; display: flex; overflow: hidden; }

.left-panel { width: 200px; padding: 12px; overflow-y: auto; flex-shrink: 0; border-right: 1px solid #e8e8e8; background: #fafafa; }
.left-panel h4 { margin: 0 0 8px; font-size: 14px; }

.palette-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; border: 1px dashed #c0c4cc; border-radius: 6px; cursor: grab; background: #fff; font-size: 13px; transition: all .15s; margin-bottom: 6px; }
.palette-item:hover { border-color: #409eff; color: #409eff; background: #ecf5ff; }

.element-item { display: flex; align-items: center; justify-content: space-between; padding: 6px 8px; border-radius: 4px; cursor: pointer; margin-bottom: 4px; border: 1px solid #e8e8e8; font-size: 12px; background: #fff; transition: all .15s; }
.element-item.active { background: #ecf5ff; border-color: #409eff; }
.element-item:hover { border-color: #409eff; }
.element-item span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }

.canvas-area { flex: 1; background: #f0f0f0; overflow: auto; padding: 24px; display: flex; align-items: flex-start; justify-content: center; }
.canvas-empty { color: #bbb; font-size: 15px; text-align: center; padding: 80px 0; border: 2px dashed #e0e0e0; border-radius: 8px; width: 100%; max-width: 400px; }
.canvas-wrap { background: #fff; border-radius: 0 0 4px 4px; min-height: 100px; box-sizing: border-box; }
.seam-zone { background: repeating-linear-gradient(-45deg, #f5f5f5, #f5f5f5 4px, #ececec 4px, #ececec 8px); border-radius: 4px 4px 0 0; display: flex; align-items: center; justify-content: flex-end; padding: 0 8px; box-sizing: border-box; }

.canvas-row { position: relative; display: flex; align-items: center; justify-content: center; border: 1px solid #e0e0e0; margin-bottom: 2px; transition: border-color .2s, background .2s; overflow: hidden; background: #fff; }
.canvas-row:hover { background: #fafafa; }
.canvas-row.selected { border: 1px dashed #409eff; background: #ecf5ff; }

.drag-mask { position: absolute; top: 0; left: 0; right: 0; bottom: 0; border: 2px dashed #409eff; pointer-events: none; z-index: 5; background: rgba(64,158,255,0.04); border-radius: 2px; }

.drag-handle { position: absolute; left: 0; top: 0; bottom: 0; width: 24px; display: flex; align-items: center; justify-content: center; cursor: grab; color: #c0c4cc; z-index: 10; opacity: 0; transition: opacity .15s; border-radius: 3px 0 0 3px; background: rgba(0,0,0,0.03); }
.canvas-row:hover .drag-handle, .canvas-row.selected .drag-handle { opacity: 1; }
.canvas-row.selected .drag-handle { color: #409eff; }

.element-content { flex-shrink: 0; overflow: hidden; box-sizing: border-box; display: flex; flex-direction: column; justify-content: center; padding: 2px 4px; height: 100%; }

.ghost { opacity: 0.4; background: #d9ecff !important; }

.rh { position: absolute; width: 8px; height: 8px; background: #409eff; border: 1px solid #fff; z-index: 20; }
.rh.w { left: -4px; top: 50%; transform: translateY(-50%); cursor: w-resize; }
.rh.e { right: -4px; top: 50%; transform: translateY(-50%); cursor: e-resize; }
.rh.s { bottom: -4px; left: 50%; transform: translateX(-50%); cursor: s-resize; }

.img-placeholder { width: 100%; height: 100%; background: #eee; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #999; font-size: 10px; }
.mini-table { width: 100%; border-collapse: collapse; }
.mini-table td { border: 2px solid #666; padding: 1px 2px; text-align: center; }

.right-panel { width: 240px; padding: 12px; overflow-y: auto; flex-shrink: 0; border-left: 1px solid #e8e8e8; background: #fafafa; }
.right-panel h4 { margin: 0 0 10px; font-size: 14px; }
.prop-row { margin-bottom: 12px; }
.prop-row label { font-size: 12px; color: #666; display: block; margin-bottom: 4px; }
.prop-input { width: 100%; }
.no-select { color: #999; text-align: center; padding-top: 60px; font-size: 13px; }

.cell-editor-grid { border-collapse: collapse; }
.cell-editor-grid td { padding: 1px; vertical-align: top; }
.cell-textarea { width: 100%; height: 100%; border: 1px solid #dcdfe6; border-radius: 2px; resize: none; overflow: hidden; box-sizing: border-box; padding: 2px 3px; font-family: inherit; line-height: 1.3; background: #fff; }
.cell-textarea:focus { border-color: #409eff; outline: none; }

.rich-toolbar { display: flex; align-items: center; gap: 4px; margin-bottom: 4px; }
.rich-toolbar .tool-group { display: flex; align-items: center; gap: 4px; }

.rich-editor { min-height: 60px; max-height: 200px; overflow-y: auto; border: 1px solid #dcdfe6; border-radius: 4px; padding: 6px 8px; font-size: 12px; line-height: 1.6; background: #fff; outline: none; white-space: pre-wrap; word-break: break-all; }
.rich-editor:focus { border-color: #409eff; }
.rich-editor br + br { display: none; }

.rich-preview { white-space: pre-wrap; word-break: break-all; }
.rich-preview span { white-space: pre-wrap; }
</style>
