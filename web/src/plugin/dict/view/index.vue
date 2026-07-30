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
        <el-form-item label="中文">
          <el-input
            v-model="searchInfo.chinese"
            placeholder="请输入中文关键词"
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
        <el-button type="primary" icon="plus" @click="openDialog">
          新增
        </el-button>
        <el-button
          icon="delete"
          style="margin-left: 10px"
          :disabled="!multipleSelection.length"
          @click="onDelete"
        >
          删除
        </el-button>
        <el-button
          type="success"
          icon="upload"
          style="margin-left: 10px"
          @click="openImportDialog"
        >
          导入Excel
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
        <el-table-column align="left" label="字典" prop="dictName" width="80" />
        <el-table-column align="left" label="中文" prop="chinese" min-width="200" show-overflow-tooltip />
        <el-table-column align="left" label="英文" prop="english" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="俄文" prop="russian" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="阿语译文" prop="arabic" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="印尼语" prop="indonesian" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="日期" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column
          align="left"
          label="操作"
          fixed="right"
          width="160"
        >
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="edit"
              @click="updateEntryFunc(scope.row)"
            >
              编辑
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

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogFormVisible"
      :title="type === 'create' ? '新增翻译' : '编辑翻译'"
      destroy-on-close
      width="600px"
      :before-close="closeDialog"
    >
      <el-form
        ref="elFormRef"
        :model="formData"
        label-position="top"
        :rules="rule"
      >
        <el-form-item label="字典名称" prop="dictName">
          <el-select
            v-model="formData.dictName"
            placeholder="请选择字典"
            style="width: 100%"
          >
            <el-option label="巴拉" value="巴拉" />
            <el-option label="森马" value="森马" />
          </el-select>
        </el-form-item>
        <el-form-item label="中文" prop="chinese">
          <el-input
            v-model="formData.chinese"
            :clearable="true"
            placeholder="请输入中文"
          />
        </el-form-item>
        <el-form-item label="英文">
          <el-input
            v-model="formData.english"
            :clearable="true"
            placeholder="请输入英文翻译"
          />
        </el-form-item>
        <el-form-item label="俄文">
          <el-input
            v-model="formData.russian"
            :clearable="true"
            placeholder="请输入俄文翻译"
          />
        </el-form-item>
        <el-form-item label="阿语译文">
          <el-input
            v-model="formData.arabic"
            :clearable="true"
            placeholder="请输入阿语译文"
          />
        </el-form-item>
        <el-form-item label="印尼语">
          <el-input
            v-model="formData.indonesian"
            :clearable="true"
            placeholder="请输入印尼语翻译"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeDialog">取消</el-button>
          <el-button type="primary" @click="enterDialog">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 导入Excel弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入Excel"
      destroy-on-close
      width="500px"
      :before-close="closeImportDialog"
    >
      <el-form
        ref="importFormRef"
        :model="importForm"
        label-position="top"
        :rules="importRule"
      >
        <el-form-item label="目标字典" prop="dictName">
          <el-select
            v-model="importForm.dictName"
            placeholder="请选择要导入的目标字典"
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
                支持 .xlsx 和 .xls 格式的Excel文件
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeImportDialog">取消</el-button>
          <el-button type="primary" :loading="importLoading" @click="submitImport">
            开始导入
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {
  createEntry,
  deleteEntry,
  deleteEntryByIds,
  updateEntry,
  findEntry,
  getEntryList,
  importExcel
} from '@/plugin/dict/api/entry'

import { formatDate } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'

defineOptions({
  name: 'DictEntry'
})

const formData = ref({
  dictName: '巴拉',
  chinese: '',
  english: '',
  russian: '',
  arabic: '',
  indonesian: ''
})

const rule = reactive({
  dictName: [{ required: true, message: '请选择字典', trigger: 'change' }],
  chinese: [{ required: true, message: '请输入中文', trigger: 'blur' }]
})

const elFormRef = ref()

// =========== 表格控制 ===========
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({ dictName: '巴拉' })

const onReset = () => {
  searchInfo.value = { dictName: '' }
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
  const table = await getEntryList({
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

// =========== 多选 ===========
const multipleSelection = ref([])
const handleSelectionChange = (val) => {
  multipleSelection.value = val
}

// =========== 删除 ===========
const deleteRow = (row) => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteEntryFunc(row)
  })
}

const onDelete = async () => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    const IDs = []
    if (multipleSelection.value.length === 0) {
      ElMessage({ type: 'warning', message: '请选择要删除的数据' })
      return
    }
    multipleSelection.value.map((item) => IDs.push(item.ID))
    const res = await deleteEntryByIds({ IDs })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: '删除成功' })
      if (tableData.value.length === IDs.length && page.value > 1) {
        page.value--
      }
      getTableData()
    }
  })
}

const deleteEntryFunc = async (row) => {
  const res = await deleteEntry({ ID: row.ID })
  if (res.code === 0) {
    ElMessage({ type: 'success', message: '删除成功' })
    if (tableData.value.length === 1 && page.value > 1) {
      page.value--
    }
    getTableData()
  }
}

// =========== 新增/编辑弹窗 ===========
const type = ref('')
const dialogFormVisible = ref(false)

const updateEntryFunc = async (row) => {
  const res = await findEntry({ ID: row.ID })
  type.value = 'update'
  if (res.code === 0) {
    formData.value = res.data
    dialogFormVisible.value = true
  }
}

const openDialog = () => {
  type.value = 'create'
  formData.value = {
    dictName: searchInfo.value.dictName || '巴拉',
    chinese: '',
    english: '',
    russian: '',
    arabic: '',
    indonesian: ''
  }
  dialogFormVisible.value = true
}

const closeDialog = () => {
  dialogFormVisible.value = false
}

const enterDialog = async () => {
  elFormRef.value?.validate(async (valid) => {
    if (!valid) return
    let res
    switch (type.value) {
      case 'create':
        res = await createEntry(formData.value)
        break
      case 'update':
        res = await updateEntry(formData.value)
        break
      default:
        res = await createEntry(formData.value)
        break
    }
    if (res.code === 0) {
      ElMessage({ type: 'success', message: type.value === 'create' ? '创建成功' : '更新成功' })
      closeDialog()
      getTableData()
    }
  })
}

// =========== 导入Excel弹窗 ===========
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importForm = reactive({
  dictName: '巴拉',
  file: null
})
const importRule = reactive({
  dictName: [{ required: true, message: '请选择目标字典', trigger: 'change' }]
})
const importFormRef = ref()
const uploadRef = ref()

const openImportDialog = () => {
  importForm.dictName = searchInfo.value.dictName || '巴拉'
  importForm.file = null
  importDialogVisible.value = true
}

const closeImportDialog = () => {
  importDialogVisible.value = false
  importForm.file = null
}

const handleFileChange = (uploadFile) => {
  importForm.file = uploadFile.raw
}

const handleFileRemove = () => {
  importForm.file = null
}

const submitImport = async () => {
  if (!importForm.file) {
    ElMessage({ type: 'warning', message: '请选择要上传的Excel文件' })
    return
  }
  importFormRef.value?.validate(async (valid) => {
    if (!valid) return
    importLoading.value = true
    try {
      const formData = new FormData()
      formData.append('dictName', importForm.dictName)
      formData.append('file', importForm.file)

      const res = await importExcel(formData)
      if (res.code === 0) {
        const result = res.data
        ElMessage({
          type: 'success',
          message: `导入完成! 成功 ${result.success || 0} 条, 失败 ${result.fail || 0} 条`
        })
        closeImportDialog()
        getTableData()
      }
    } finally {
      importLoading.value = false
    }
  })
}
</script>

<style scoped>
</style>
