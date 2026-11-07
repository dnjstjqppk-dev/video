<script setup lang="ts">
import { SubscriptionTier, type User } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'

const api = useAdminApi()

useSeoMeta({
  title: 'User & Channel Governance — RBAC & Sanctions'
})

const searchFilter = ref('')
const sanctionReason = ref('')
const page = ref(1)
const limit = 20
const total = ref(0)
const actionError = ref<string | null>(null)
const actingIds = ref<Set<string>>(new Set())
const states = ref<Record<string, string>>({})

const { data, pending, error, refresh } = await useAsyncData(
  () => `admin-users-${page.value}-${searchFilter.value}`,
  async () => {
    const res = await api.getAdminUsers(page.value, limit)
    total.value = res.total
    return res.items
  },
  { watch: [page] }
)

const { data: auditData, pending: auditPending, refresh: refreshAudit } = await useAsyncData(
  'admin-users-audit',
  async () => {
    const res = await api.getAdminAuditLogs(1, 10)
    return res.items
  }
)

const users = computed(() => data.value ?? [])
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const filteredUsers = computed(() => {
  const q = searchFilter.value.trim().toLowerCase()
  if (!q) return users.value
  return users.value.filter(
    u => u.username.toLowerCase().includes(q) || u.email.toLowerCase().includes(q)
  )
})

function stateOf(u: User): string {
  return states.value[u.id] || 'ACTIVE'
}

async function applySanction(user: User, sanction: 'suspended' | 'banned' | 'active') {
  if (!sanctionReason.value.trim()) {
    actionError.value = '제재 사유를 입력해야 합니다. 감사 기록에 필수로 남습니다.'
    return
  }
  actionError.value = null
  actingIds.value.add(user.id)
  const prev = stateOf(user)
  states.value[user.id] = sanction.toUpperCase()
  try {
    await api.updateAdminUserStatus(user.id, sanction, sanctionReason.value.trim())
    await refreshAudit()
  } catch (e: unknown) {
    states.value[user.id] = prev
    actionError.value = e instanceof Error ? e.message : '상태 변경에 실패했습니다.'
  } finally {
    actingIds.value.delete(user.id)
  }
}

function tierLabel(t: SubscriptionTier): string {
  return String(t)
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">사용자 및 채널 거버넌스 (User & RBAC Governance)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Search users · Review RBAC roles · Execute account sanctions (Suspend, Ban, Restore) with mandatory audit reasons
    </p>

    <!-- Search & Sanction Reason Bar -->
    <div class="mt-6 grid grid-cols-1 gap-4 rounded-2xl border border-white/10 bg-[#131722] p-4 md:grid-cols-2">
      <div>
        <label for="admin-user-search" class="text-xs text-neutral-400">Search by Username or Email</label>
        <input
          id="admin-user-search"
          v-model="searchFilter"
          type="search"
          placeholder="Search @username or email..."
          class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3.5 py-2 text-xs text-white outline-none"
        />
      </div>
      <div>
        <label for="admin-sanction-reason" class="text-xs text-neutral-400">Mandatory Audit Reason for Sanction</label>
        <input
          id="admin-sanction-reason"
          v-model="sanctionReason"
          type="text"
          placeholder="예: 반복 저작권 위반 신고 3회 누적"
          class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3.5 py-2 text-xs text-white outline-none"
        />
      </div>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible class="mt-4" @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <!-- Loading -->
    <div v-if="pending" class="mt-6 space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 5" :key="n" class="h-14 animate-pulse rounded-xl bg-white/5" />
    </div>

    <!-- Error -->
    <GvpBanner v-else-if="error" tone="danger" class="mt-6">
      사용자 목록을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <!-- Empty -->
    <GvpEmptyState
      v-else-if="filteredUsers.length === 0"
      class="mt-6"
      title="조건에 맞는 사용자가 없습니다"
      description="검색어를 바꾸거나 필터를 초기화하세요."
    />

    <!-- Users Governance Table -->
    <div v-else class="mt-6 overflow-x-auto rounded-2xl border border-white/10 bg-[#131722]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th scope="col" class="p-3.5">User / Channel</th>
            <th scope="col" class="p-3.5">RBAC Role</th>
            <th scope="col" class="p-3.5">Tier</th>
            <th scope="col" class="p-3.5">Account State</th>
            <th scope="col" class="p-3.5">Sanctions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="u in filteredUsers" :key="u.id" class="hover:bg-white/5">
            <td class="p-3.5">
              <p class="font-bold text-white">{{ u.displayName }} (@{{ u.username }})</p>
              <p class="text-[11px] text-neutral-400">{{ u.email }}</p>
            </td>
            <td class="p-3.5">
              <span class="rounded-lg bg-white/5 px-2.5 py-1 font-mono text-xs text-[#3EA6FF]">{{ u.role }}</span>
            </td>
            <td class="p-3.5 text-neutral-300">{{ tierLabel(u.subscriptionTier) }}</td>
            <td class="p-3.5">
              <span
                class="rounded px-2.5 py-0.5 text-[10px] font-bold uppercase"
                :class="stateOf(u) === 'ACTIVE' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-300'"
              >
                {{ stateOf(u) }}
              </span>
            </td>
            <td class="p-3.5">
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  :disabled="actingIds.has(u.id)"
                  class="rounded bg-red-600/30 px-2.5 py-1 text-[11px] font-bold text-red-300 hover:bg-red-600/40 disabled:opacity-40"
                  @click="applySanction(u, 'suspended')"
                >
                  Suspend
                </button>
                <button
                  type="button"
                  :disabled="actingIds.has(u.id)"
                  class="rounded bg-red-900/50 px-2.5 py-1 text-[11px] font-bold text-red-200 hover:bg-red-900/70 disabled:opacity-40"
                  @click="applySanction(u, 'banned')"
                >
                  Ban
                </button>
                <button
                  v-if="stateOf(u) !== 'ACTIVE'"
                  type="button"
                  :disabled="actingIds.has(u.id)"
                  class="rounded bg-emerald-500/20 px-2.5 py-1 text-[11px] font-bold text-emerald-300 disabled:opacity-40"
                  @click="applySanction(u, 'active')"
                >
                  Restore
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="!pending && !error && totalPages > 1" class="mt-4 flex items-center justify-center gap-2 text-xs">
      <button
        type="button"
        :disabled="page <= 1"
        class="rounded-lg border border-white/15 px-3 py-1.5 text-neutral-200 disabled:opacity-40"
        @click="page -= 1"
      >
        이전
      </button>
      <span class="text-neutral-400" role="status">{{ page }} / {{ totalPages }} (총 {{ total }}명)</span>
      <button
        type="button"
        :disabled="page >= totalPages"
        class="rounded-lg border border-white/15 px-3 py-1.5 text-neutral-200 disabled:opacity-40"
        @click="page += 1"
      >
        다음
      </button>
    </div>

    <!-- Governance Audit Log -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5">
      <h2 class="text-sm font-bold text-white">최근 제재 및 권한 변경 감사 기록 (Sanction Audit Trail)</h2>
      <div v-if="auditPending" class="mt-3 space-y-2" role="status" aria-label="감사 기록 불러오는 중">
        <div v-for="n in 3" :key="n" class="h-10 animate-pulse rounded-xl bg-white/5" />
      </div>
      <div v-else-if="!auditData || auditData.length === 0" class="mt-3 text-xs text-neutral-400">
        기록된 감사 로그가 없습니다.
      </div>
      <div v-else class="mt-3 space-y-2 text-xs">
        <div
          v-for="entry in auditData"
          :key="entry.id"
          class="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-[#090B10] px-4 py-2.5"
        >
          <span>
            <strong class="text-red-400">[{{ entry.action }}]</strong>
            <span class="text-neutral-300"> {{ entry.target }}</span>
            <span class="text-neutral-500"> · {{ entry.actorId }} · {{ entry.ip }}</span>
          </span>
          <span class="font-mono text-neutral-500">{{ entry.timestamp }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
