<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑翻译' : '新增翻译'"
    destroy-on-close
    width="600px"
    :before-close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      label-position="top"
      :rules="rules"
    >
      <el-form-item label="字典名称" prop="dictName">
        <el-select v-model="form.dictName" placeholder="请选择字典" style="width: 100%" :disabled="isEdit">
          <el-option label="巴拉" value="巴拉" />
          <el-option label="森马" value="森马" />
        </el-select>
      </el-form-item>
      <el-form-item label="中文" prop="chinese">
        <el-input v-model="form.chinese" placeholder="请输入中文" clearable />
      </el-form-item>
      <el-form-item label="英文">
        <el-input v-model="form.english" placeholder="请输入英文翻译" clearable />
      </el-form-item>
      <el-form-item label="俄文">
        <el-input v-model="form.russian" placeholder="请输入俄文翻译" clearable />
      </el-form-item>
      <el-form-item label="阿语译文">
        <el-input v-model="form.arabic" placeholder="请输入阿语译文" clearable />
      </el-form-item>
      <el-form-item label="印尼语">
        <el-input v-model="form.indonesian" placeholder="请输入印尼语翻译" clearable />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch, reactive } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  data: {
    type: Object,
    default: () => ({})
  },
  isEdit: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'submit'])

const visible = ref(false)
const formRef = ref()

const form = reactive({
  dictName: '巴拉',
  chinese: '',
  english: '',
  russian: '',
  arabic: '',
  indonesian: ''
})

const rules = reactive({
  dictName: [{ required: true, message: '请选择字典', trigger: 'change' }],
  chinese: [{ required: true, message: '请输入中文', trigger: 'blur' }]
})

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val && props.data) {
    Object.assign(form, {
      dictName: props.data.dictName || '巴拉',
      chinese: props.data.chinese || '',
      english: props.data.english || '',
      russian: props.data.russian || '',
      arabic: props.data.arabic || '',
      indonesian: props.data.indonesian || ''
    })
  }
}, { immediate: true })

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const handleClose = () => {
  visible.value = false
}

const handleSubmit = () => {
  formRef.value?.validate((valid) => {
    if (!valid) return
    emit('submit', { ...form })
    visible.value = false
  })
}
</script>
