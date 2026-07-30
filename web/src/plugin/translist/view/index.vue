<template>
  <div>
    <div class="gva-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        class="demo-form-inline"
        @keyup.enter="onSubmit"
      >
        <el-form-item label="字典">
          <el-radio-group v-model="searchInfo.dictName" @change="onSubmit">
            <el-radio-button label="">全部</el-radio-button>
            <el-radio-button label="巴拉">巴拉</el-radio-button>
            <el-radio-button label="森马">森马</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchInfo.status"
            clearable
            placeholder="全部状态"
            style="width: 140px"
            @change="onSubmit"
          >
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="处理中" value="processing" />
          </el-select>
        </el-form-item>
        <el-form-item label="文件名">
          <el-input
            v-model="searchInfo.fileName"
            placeholder="搜索文件名"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">
            查询
          </el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="upload" @click="openUploadDialog">
          上传并翻译
        </el-button>
        <el-button
          icon="delete"
          style="margin-left: 10px"
          :disabled="!multipleSelection.length"
          @click="onDelete"
        >
          删除
        </el-button>
      </div>

      <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="ID"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column
          align="left"
          label="文件名"
          prop="fileName"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column align="left" label="字典" prop="dictName" width="80" />
        <el-table-column align="left" label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 'success'" type="success">成功</el-tag>
            <el-tag v-else-if="scope.row.status === 'failed'" type="danger">失败</el-tag>
            <el-tag v-else-if="scope.row.status === 'processing'" type="warning">处理中</el-tag>
            <el-tag v-else type="info">{{ scope.row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="行数" prop="totalRows" width="80" />
        <el-table-column align="left" label="已翻译格" prop="translatedCells" width="100" />
        <el-table-column align="left" label="未命中" prop="missingCount" width="100">
          <template #default="scope">
            <span
              v-if="scope.row.missingCount > 0"
              class="missing-count-link"
              @click="showMissingWords(scope.row)"
            >
              {{ scope.row.missingCount }}
            </span>
            <span v-else>0</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="错误信息" prop="errorMsg" min-width="160" show-overflow-tooltip />
        <el-table-column align="left" label="时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" fixed="right" width="220">
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="download"
              :disabled="scope.row.status !== 'success'"
              @click="handleExport(scope.row)"
            >
              导出
            </el-button>
            <el-button
              type="primary"
              link
              icon="refresh"
              :loading="retranslateId === scope.row.ID"
              @click="handleRetranslate(scope.row)"
            >
              重译
            </el-button>
            <el-button
              type="primary"
              link
              icon="delete"
              @click="deleteRow(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-dialog
      v-model="uploadDialogVisible"
      title="上传Excel并翻译"
      destroy-on-close
      width="520px"
      :before-close="closeUploadDialog"
    >
      <el-form
        ref="uploadFormRef"
        :model="uploadForm"
        label-position="top"
        :rules="uploadRule"
      >
        <el-form-item label="翻译字典" prop="dictName">
          <el-select
            v-model="uploadForm.dictName"
            placeholder="请选择字典"
            style="width: 100%"
          >
            <el-option label="巴拉" value="巴拉" />
            <el-option label="森马" value="森马" />
          </el-select>
        </el-form-item>
        <el-form-item label="Excel文件" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".xlsx,.xls"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            drag
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                表头需包含中文列及对应「xxx英文 / xxx俄文 / xxx阿语」列，例如洗唛成分→洗唛成分英文
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeUploadDialog">取消</el-button>
          <el-button type="primary" :loading="uploadLoading" @click="submitUpload">
            开始翻译
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="missingDialogVisible"
      :title="`未命中词条（${missingWordList.length}）`"
      destroy-on-close
      width="560px"
    >
      <div v-if="!missingWordList.length" style="color:#909399;text-align:center;padding:24px 0">
        暂无未命中词条详情，请点击「重译」后重新查看
      </div>
      <div v-else class="missing-words-wrap">
        <el-tag
          v-for="(word, idx) in missingWordList"
          :key="idx"
          type="warning"
          effect="plain"
          class="missing-word-tag"
        >
          {{ word }}
        </el-tag>
      </div>
      <template #footer>
        <el-button @click="missingDialogVisible = false">关闭</el-button>
        <el-button type="primary" :disabled="!missingWordList.length" @click="copyMissingWords">
          复制全部
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {
  uploadAndTranslate,
  deleteJob,
  deleteJobByIds,
  getJobList,
  retranslate,
  exportExcel
} from '@/plugin/translist/api/job'
import { formatDate } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'

defineOptions({
  name: 'Translist'
})

const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({ dictName: '巴拉', status: '', fileName: '' })
const multipleSelection = ref([])
const retranslateId = ref(null)
const missingDialogVisible = ref(false)
const missingWordList = ref([])

const parseMissingWords = (raw) => {
  if (!raw) return []
  if (Array.isArray(raw)) return raw
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

const showMissingWords = (row) => {
  missingWordList.value = parseMissingWords(row.missingWords)
  missingDialogVisible.value = true
}

const copyMissingWords = async () => {
  const text = missingWordList.value.join('\n')
  try {
    await navigator.clipboard.writeText(text)
    ElMessage({ type: 'success', message: '已复制到剪贴板' })
  } catch {
    ElMessage({ type: 'error', message: '复制失败' })
  }
}

const onReset = () => {
  searchInfo.value = { dictName: '', status: '', fileName: '' }
  getTableData()
}

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

const getTableData = async () => {
  const table = await getJobList({
    page: page.value,
    pageSize: pageSize.value,
    ...searchInfo.value
  })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

getTableData()

const handleSelectionChange = (val) => {
  multipleSelection.value = val
}

const deleteRow = (row) => {
  ElMessageBox.confirm('确定要删除该翻译任务吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteJobFunc(row)
  })
}

const onDelete = async () => {
  ElMessageBox.confirm('确定要删除选中的任务吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    const IDs = multipleSelection.value.map((item) => item.ID)
    if (!IDs.length) {
      ElMessage({ type: 'warning', message: '请选择要删除的数据' })
      return
    }
    const res = await deleteJobByIds({ IDs })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: '删除成功' })
      if (tableData.value.length === IDs.length && page.value > 1) {
        page.value--
      }
      getTableData()
    }
  })
}

const deleteJobFunc = async (row) => {
  const res = await deleteJob({ ID: row.ID })
  if (res.code === 0) {
    ElMessage({ type: 'success', message: '删除成功' })
    if (tableData.value.length === 1 && page.value > 1) {
      page.value--
    }
    getTableData()
  }
}

const handleExport = async (row) => {
  try {
    const res = await exportExcel({ ID: row.ID })
    const blob = res instanceof Blob ? res : res?.data instanceof Blob ? res.data : null
    if (!blob) {
      ElMessage({ type: 'error', message: '导出失败' })
      return
    }
    const contentType = String(res?.headers?.['content-type'] || blob.type || '').toLowerCase()
    if (contentType.includes('application/json')) {
      const text = await blob.text()
      try {
        const json = JSON.parse(text)
        ElMessage({ type: 'error', message: json.msg || '导出失败' })
      } catch {
        ElMessage({ type: 'error', message: '导出失败' })
      }
      return
    }
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const base = (row.fileName || 'translated').replace(/\.(xlsx|xls)$/i, '')
    a.download = `${base}_已翻译.xlsx`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch (e) {
    ElMessage({ type: 'error', message: '导出失败' })
  }
}

const handleRetranslate = async (row) => {
  retranslateId.value = row.ID
  try {
    const res = await retranslate({ ID: row.ID })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: `重译完成：${res.data?.translatedCells || 0} 格，未命中 ${res.data?.missingCount || 0}`
      })
      getTableData()
    }
  } finally {
    retranslateId.value = null
  }
}

const uploadDialogVisible = ref(false)
const uploadLoading = ref(false)
const uploadForm = reactive({
  dictName: '巴拉',
  file: null
})
const uploadRule = reactive({
  dictName: [{ required: true, message: '请选择字典', trigger: 'change' }]
})
const uploadFormRef = ref()
const uploadRef = ref()

const openUploadDialog = () => {
  uploadForm.dictName = searchInfo.value.dictName || '巴拉'
  uploadForm.file = null
  uploadDialogVisible.value = true
}

const closeUploadDialog = () => {
  uploadDialogVisible.value = false
  uploadForm.file = null
}

const handleFileChange = (uploadFile) => {
  uploadForm.file = uploadFile.raw
}

const handleFileRemove = () => {
  uploadForm.file = null
}

const submitUpload = async () => {
  if (!uploadForm.file) {
    ElMessage({ type: 'warning', message: '请选择要上传的Excel文件' })
    return
  }
  uploadFormRef.value?.validate(async (valid) => {
    if (!valid) return
    uploadLoading.value = true
    try {
      const formData = new FormData()
      formData.append('dictName', uploadForm.dictName)
      formData.append('file', uploadForm.file)

      const res = await uploadAndTranslate(formData)
      if (res.code === 0) {
        const job = res.data || {}
        ElMessage({
          type: job.status === 'success' ? 'success' : 'warning',
          message:
            job.status === 'success'
              ? `翻译完成：${job.translatedCells || 0} 格，未命中 ${job.missingCount || 0}`
              : res.msg || '翻译失败'
        })
        closeUploadDialog()
        getTableData()
      }
    } finally {
      uploadLoading.value = false
    }
  })
}
</script>

<style scoped>
.missing-count-link {
  color: #e6a23c;
  cursor: pointer;
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.missing-count-link:hover {
  color: #cf9236;
}
.missing-words-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-height: 360px;
  overflow: auto;
}
.missing-word-tag {
  margin: 0;
  max-width: 100%;
  white-space: normal;
  height: auto;
  line-height: 1.4;
  padding: 4px 8px;
}
</style>
