<template>
  <div class="issue-list">
    <h1>이슈 목록</h1>
    <div class="toolbar">
      <select v-model="selectedStatus" @change="fetchIssues">
        <option value="">전체</option>
        <option v-for="s in statuses" :key="s" :value="s">{{ statusLabel(s) }}</option>
      </select>
      <button @click="goNew">+ 새 이슈</button>
    </div>
    <table>
      <thead>
        <tr>
          <th>제목</th>
          <th>상태</th>
          <th>담당자</th>
          <th>생성일</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="issue in issues"
          :key="issue.id"
          @click="goDetail(issue.id)"
          style="cursor: pointer"
        >
          <td>{{ issue.title }}</td>
          <td>{{ statusLabel(issue.status) }}</td>
          <td>{{ issue.user ? issue.user.name : '-' }}</td>
          <td>{{ formatDate(issue.createdAt) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const router = useRouter()
const issues = ref([])
const selectedStatus = ref('')
const statuses = ['PENDING', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED']

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
function formatDate(dt) {
  return dt ? dt.split('T')[0] : ''
}
function goDetail(id) {
  router.push(`/issues/${id}`)
}
function goNew() {
  router.push('/issues/new')
}
async function fetchIssues() {
  let url = '/issues'
  if (selectedStatus.value) url += `?status=${selectedStatus.value}`
  const { data } = await api.get(url)
  issues.value = data.issues
}
onMounted(fetchIssues)
</script>

<style scoped>
.issue-list {
  max-width: 800px;
  margin: 2rem auto;
}
.toolbar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th,
td {
  border: 1px solid #ddd;
  padding: 8px;
}
th {
  background: #f5f5f5;
}
tr:hover {
  background: #f0f8ff;
}
</style>
