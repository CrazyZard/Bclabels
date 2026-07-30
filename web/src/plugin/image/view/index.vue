<template>
  <div>
    <div class="gva-search-box">
      <el-form ref="searchFormRef" :inline="true" :model="searchInfo" @keyup.enter="onSubmit">
        <el-form-item label="图片类型">
          <el-select v-model="searchInfo.type" placeholder="全部" clearable style="width:160px" @change="onSubmit">
            <el-option label="logo 图片" value="logo" />
            <el-option label="水洗标标志" value="washLabel" />
          </el-select>
        </el-form-item>
        <el-form-item label="图片名称">
          <el-input v-model="searchInfo.name" placeholder="搜索名称" clearable style="width:180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog('create')">
          <el-icon><Plus /></el-icon> 新增图片
        </el-button>
        <el-button @click="onDeleteBatch" :disabled="multipleSelection.length === 0">
          <el-icon><Delete /></el-icon> 批量删除
        </el-button>
      </div>

      <el-table
        ref="tableRef"
        :data="tableData"
        style="width:100%"
        @selection-change="handleSelectionChange"
        row-key="ID"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column label="预览" width="100" align="center">
          <template #default="{ row }">
            <img v-if="row.url" :src="resolveImageUrl(row.url)" style="width:60px;height:60px;object-fit:contain;border:1px solid #eee;border-radius:4px" />
            <span v-else style="color:#ccc">无图</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column label="类型" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'logo' ? 'primary' : 'success'" size="small">
              {{ row.type === 'logo' ? 'logo 图片' : '水洗标标志' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="图片地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="UpdatedAt" label="更新时间" width="160" align="center">
          <template #default="{ row }">{{ formatDate(row.UpdatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDialog('update', row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          @size-change="getTableData"
          @current-change="getTableData"
        />
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogType === 'create' ? '新增图片' : '编辑图片'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="rules" label-width="90px">
        <el-form-item label="图片名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入图片名称" />
        </el-form-item>
        <el-form-item label="图片类型" prop="type">
          <el-select v-model="formData.type" style="width:100%">
            <el-option label="logo 图片" value="logo" />
            <el-option label="水洗标标志" value="washLabel" />
          </el-select>
        </el-form-item>
        <el-form-item label="图片地址" prop="url">
          <el-input v-model="formData.url" placeholder="输入图片URL或上传" />
        </el-form-item>
        <el-form-item label="上传图片">
          <el-upload
            class="image-uploader"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :show-file-list="false"
            :on-success="handleUploadSuccess"
            :before-upload="beforeUpload"
            accept="image/svg+xml"
          >
            <el-button type="primary" plain>选择SVG图片</el-button>
            <template #tip>
              <div class="el-upload__tip">仅支持 .svg 格式图片</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="formData.sort" :min="0" :step="1" style="width:100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">{{ dialogType === 'create' ? '新增' : '保存' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { createImage, deleteImage, deleteImageByIds, updateImage, getImageList } from '@/plugin/image/api/image'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'ImageManager' })

// ====== 工具：将相对路径的图片URL转为完整地址（修复dev环境下跨端口问题） ======
const backendBase = import.meta.env.VITE_BASE_PATH + ':' + import.meta.env.VITE_SERVER_PORT
const resolveImageUrl = (url) => {
  if (!url) return ''
  if (/^https?:\/\//.test(url)) return url
  return backendBase + '/' + url.replace(/^\/+/, '')
}

// ====== 搜索 ======
const searchInfo = reactive({ type: '', name: '' })
const searchFormRef = ref()

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const onReset = () => {
  searchInfo.type = ''
  searchInfo.name = ''
  page.value = 1
  getTableData()
}

// ====== 表格 ======
const tableRef = ref()
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const multipleSelection = ref([])

const handleSelectionChange = (val) => {
  multipleSelection.value = val
}

const getTableData = async () => {
  const res = await getImageList({
    type: searchInfo.type,
    name: searchInfo.name,
    page: page.value,
    pageSize: pageSize.value
  })
  if (res.code === 0) {
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  }
}
getTableData()

// ====== 上传 ======
const uploadUrl = import.meta.env.VITE_BASE_API + '/fileUploadAndDownload/upload'
const uploadHeaders = { 'x-token': localStorage.getItem('token') || '' }

const handleUploadSuccess = (res, file) => {
  if (res.code === 0 && res.data?.file?.url) {
    formData.value.url = res.data.file.url
    // 上传成功后，如果图片名称为空，自动填入文件名（去掉扩展名）
    if (!formData.value.name) {
      formData.value.name = file.name.replace(/\.svg$/i, '')
    }
    ElMessage.success('上传成功')
  }
}

const beforeUpload = (file) => {
  const isSVG = file.type === 'image/svg+xml' || file.name.toLowerCase().endsWith('.svg')
  if (!isSVG) {
    ElMessage.error('只支持 SVG 格式图片')
    return false
  }
  return true
}

// ====== 弹窗 ======
const dialogVisible = ref(false)
const dialogType = ref('create')
const formRef = ref()
const formData = ref({ name: '', type: 'logo', url: '', sort: 0 })
const rules = {
  name: [{ required: true, message: '请输入图片名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择图片类型', trigger: 'change' }]
}

const openDialog = (type, row) => {
  dialogType.value = type
  if (type === 'update' && row) {
    formData.value = {
      ID: row.ID,
      name: row.name,
      type: row.type,
      url: row.url,
      sort: row.sort
    }
  } else {
    formData.value = { name: '', type: 'logo', url: '', sort: 0 }
  }
  dialogVisible.value = true
}

const submitForm = () => {
  formRef.value?.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'create' ? createImage : updateImage
    const res = await api(formData.value)
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'create' ? '新增成功' : '更新成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

// ====== 删除 ======
const deleteRow = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该图片？', '提示', { type: 'warning' })
    const res = await deleteImage({ ID: row.ID })
    if (res.code === 0) { ElMessage.success('删除成功'); getTableData() }
  } catch { /* cancelled */ }
}

const onDeleteBatch = async () => {
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${multipleSelection.value.length} 张图片？`, '提示', { type: 'warning' })
    const ids = multipleSelection.value.map(v => v.ID)
    const res = await deleteImageByIds({ IDs: ids })
    if (res.code === 0) { ElMessage.success('批量删除成功'); getTableData() }
  } catch { /* cancelled */ }
}
</script>

<style scoped>
.image-uploader .el-upload {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
