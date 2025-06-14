<template>
  <div class="issue-form">
    <h1>{{ isEdit ? '이슈 수정' : '이슈 생성' }}</h1>
    <form @submit.prevent="onSubmit">
      <div>
        <label>제목</label>
        <input v-model="form.title" :disabled="isCompletedOrCancelled" required />
      </div>
      <div>
        <label>설명</label>
        <textarea v-model="form.description" :disabled="isCompletedOrCancelled" />
      </div>
      <div>
        <label>상태</label>
        <select v-model="form.status" :disabled="statusSelectDisabled">
          <option v-for="s in statuses" :key="s" :value="s" :disabled="statusOptionDisabled(s)">
            {{ statusLabel(s) }}
          </option>
        </select>
      </div>
      <div>
        <label>담당자</label>
        <select v-model="userIdString" :disabled="assigneeSelectDisabled">
          <option value="">(없음)</option>
          <option v-for="u in users" :key="u.id" :value="String(u.id)">{{ u.name }}</option>
        </select>
      </div>
      <div class="actions">
        <button type="submit" :disabled="isCompletedOrCancelled">저장</button>
        <button type="button" @click="goList">목록</button>
      </div>
    </form>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { users as userList } from '../data/mockData'

const route = useRoute()
const router = useRouter()
const isEdit = route.path !== '/issues/new'
const users = userList
const statuses = ['PENDING', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED']
const form = reactive({ title: '', description: '', status: 'PENDING', userId: null })
const error = ref('')
const issueId = isEdit ? route.params.id : null
const isCompletedOrCancelled = ref(false)
const userIdString = ref('')

function statusLabel(s) {
  switch (s) {
    case 'PENDING':
      return '대기'
    case 'IN_PROGRESS':
      return '진행중'
    case 'COMPLETED':
      return '완료'
    case 'CANCELLED':
      return '취소'
    default:
      return s
  }
}
function statusOptionDisabled(s) {
  if (form.userId == null && s !== 'PENDING') return true
  return false
}
const statusSelectDisabled = computed(() => form.userId == null || isCompletedOrCancelled.value)
const assigneeSelectDisabled = computed(() => isCompletedOrCancelled.value)

function goList() {
  router.push('/issues')
}
async function fetchIssue() {
  if (!isEdit) return
  try {
    const { data } = await api.get(`/issue/${issueId}`)
    form.title = data.title
    form.description = data.description
    form.status = data.status
    form.userId = data.user ? data.user.id : null
    isCompletedOrCancelled.value = data.status === 'COMPLETED' || data.status === 'CANCELLED'
  } catch (e) {
    error.value = e.response?.data?.error || '이슈를 불러올 수 없습니다.'
  }
}
onMounted(fetchIssue)

watch(
  () => form.status,
  (val) => {
    isCompletedOrCancelled.value = val === 'COMPLETED' || val === 'CANCELLED'
  },
)

watch(
  () => form.userId,
  (val) => {
    userIdString.value = val == null ? '' : String(val)
  },
  { immediate: true },
)
watch(userIdString, (val) => {
  form.userId = val === '' ? null : Number(val)
})

async function onSubmit() {
  error.value = ''
  try {
    const payload = {
      title: form.title,
      description: form.description,
      status: form.status,
    }
    if (form.userId !== null) payload.userId = form.userId
    if (!isEdit) {
      await api.post('/issue', payload)
    } else {
      await api.patch(`/issue/${issueId}`, payload)
    }
    goList()
  } catch (e) {
    error.value = e.response?.data?.error || '저장 실패'
  }
}
</script>

<style scoped>
.issue-form {
  max-width: 600px;
  margin: 2rem auto;
}
form > div {
  margin-bottom: 1rem;
}
label {
  display: block;
  margin-bottom: 0.3rem;
}
input,
textarea,
select {
  width: 100%;
  padding: 0.5rem;
}
.actions {
  display: flex;
  gap: 1rem;
}
.error {
  color: red;
  margin-top: 1rem;
}
</style>
