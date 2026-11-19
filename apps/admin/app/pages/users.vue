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
const sanctionTotp = ref('')
const page = ref(1)
const limit = 20
const actionError = ref<string | null>(null)
const actingIds = ref<Set<string>>(new Set())
const states = ref<Record<string, string>>({})

const { data, pending, error, refresh } = await useAsyncData(
  () => `admin-users-${page.value}-${searchFilter.value}`,
  async () => {
    return await api.getAdminUsers(page.value, limit)
  },
  { watch: [page] }
)

const total = computed(() => data.value?.total ?? 0)

const { data: auditData, pending: auditPending, refresh: refreshAudit } = await useAsyncData(
  'admin-users-audit',
  async () => {
    const res = await api.getAdminAuditLogs(1, 10)
    return res.items
  }
)

const users = computed(() => data.value?.items ?? [])
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
  if (sanctionTotp.value.trim().length !== 6) {
    actionError.value = '6자리 TOTP 코드를 입력하세요.'
    return
  }
  actionError.value = null
  actingIds.value.add(user.id)
  const prev = stateOf(user)
  states.value[user.id] = sanction.toUpperCase()
  try {
    await api.updateAdminUserStatus(user.id, sanction, sanctionReason.value.trim(), sanctionTotp.value.trim())
    sanctionTotp.value = ''
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
  <div class="space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-[var(--gvp-text)]">
        사용자 및 채널 거버넌스 (User &amp; RBAC Governance)
      </h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Search users · Review RBAC roles · Execute account sanctions (Suspend, Ban, Restore) with mandatory audit reasons
      </p>
    </div>

    <!-- Search & Sanction Reason Bar -->
    <div class="grid grid-cols-1 gap-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4 md:grid-cols-3">
      <div>
        <label for="admin-user-search" class="block text-xs font-medium text-[var(--gvp-text-secondary)]">Search by Username or Email</label>
        <input
          id="admin-user-search"
          v-model="searchFilter"
          type="search"
          placeholder="Search @username or email..."
          class="mt-1.5 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3.5 py-2 text-sm text-[var(--gvp-text)] placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
        />
      </div>
      <div>
        <label for="admin-sanction-reason" class="block text-xs font-medium text-[var(--gvp-text-secondary)]">Mandatory Audit Reason for Sanction</label>
        <input
          id="admin-sanction-reason"
          v-model="sanctionReason"
          type="text"
          placeholder="예: 반복 저작권 위반 신고 3회 누적"
          class="mt-1.5 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3.5 py-2 text-sm text-[var(--gvp-text)] placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
        />
      </div>
      <div>
        <label for="admin-sanction-totp" class="block text-xs font-medium text-[var(--gvp-text-secondary)]">Step-Up TOTP (6 digits)</label>
        <input
          id="admin-sanction-totp"
          v-model="sanctionTotp"
          type="text"
          inputmode="numeric"
          maxlength="6"
          placeholder="123456"
          class="mt-1.5 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3.5 py-2 text-sm text-[var(--gvp-text)] placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
        />
      </div>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <!-- Loading -->
    <div v-if="pending" class="space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 5" :key="n" class="h-14 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" />
    </div>

    <!-- Error -->
    <GvpBanner v-else-if="error" tone="danger">
      사용자 목록을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <!-- Empty -->
    <GvpEmptyState
      v-else-if="filteredUsers.length === 0"
      title="조건에 맞는 사용자가 없습니다"
      description="검색어를 바꾸거나 필터를 초기화하세요."
    />

    <!-- Users Governance Table -->
    <div v-else class="overflow-x-auto rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr>
            <th scope="col" class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">User / Channel</th>
            <th scope="col" class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">RBAC Role</th>
            <th scope="col" class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Tier</th>
            <th scope="col" class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Account State</th>
            <th scope="col" class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider py-2.5 px-4">Sanctions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr v-for="u in filteredUsers" :key="u.id" class="hover:bg-[var(--gvp-canvas)] transition-colors">
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <p class="font-semibold text-[var(--gvp-text)]">{{ u.displayName }} <span class="font-normal text-[var(--gvp-text-secondary)]">(@{{ u.username }})</span></p>
              <p class="text-xs text-[var(--gvp-text-secondary)]">{{ u.email }}</p>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <span class="inline-flex items-center rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] px-2.5 py-0.5 font-mono text-xs font-medium text-[var(--gvp-text)]">
                {{ u.role }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums">
              {{ tierLabel(u.subscriptionTier) }}
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <span
                class="inline-flex items-center rounded-md px-2.5 py-0.5 text-xs font-semibold uppercase"
                :class="stateOf(u) === 'ACTIVE' ? 'border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]' : 'border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)]'"
              >
                {{ stateOf(u) }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <div class="flex flex-wrap gap-2">
                <button
                  type="button"
                  :disabled="actingIds.has(u.id)"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
                  @click="applySanction(u, 'suspended')"
                >
                  Suspend
                </button>
                <button
                  type="button"
                  :disabled="actingIds.has(u.id)"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors disabled:opacity-40"
                  @click="applySanction(u, 'banned')"
                >
                  Ban
                </button>
                <button
                  v-if="stateOf(u) !== 'ACTIVE'"
                  type="button"
                  :disabled="actingIds.has(u.id)"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
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
    <div v-if="!pending && !error && totalPages > 1" class="flex items-center justify-center gap-3 text-sm">
      <button
        type="button"
        :disabled="page <= 1"
        class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
        @click="page -= 1"
      >
        이전
      </button>
      <span class="text-[var(--gvp-text-secondary)] tabular-nums" role="status">{{ page }} / {{ totalPages }} (총 {{ total }}명)</span>
      <button
        type="button"
        :disabled="page >= totalPages"
        class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
        @click="page += 1"
      >
        다음
      </button>
    </div>

    <!-- Governance Audit Log -->
    <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <div class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3.5">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">
          최근 제재 및 권한 변경 감사 기록 (Sanction Audit Trail)
        </h2>
      </div>
      <div v-if="auditPending" class="p-4 space-y-2" role="status" aria-label="감사 기록 불러오는 중">
        <div v-for="n in 3" :key="n" class="h-10 animate-pulse rounded-md bg-[var(--gvp-elevated)]" />
      </div>
      <div v-else-if="!auditData || auditData.length === 0" class="p-5 text-xs text-[var(--gvp-text-secondary)]">
        기록된 감사 로그가 없습니다.
      </div>
      <div v-else class="divide-y divide-[var(--gvp-border)] text-xs">
        <div
          v-for="entry in auditData"
          :key="entry.id"
          class="flex flex-wrap items-center justify-between gap-2 px-5 py-3 hover:bg-[var(--gvp-canvas)]"
        >
          <span>
            <strong class="font-mono font-semibold text-[var(--gvp-live)]">[{{ entry.action }}]</strong>
            <span class="font-medium text-[var(--gvp-text)]"> {{ entry.target }}</span>
            <span class="text-[var(--gvp-text-secondary)] tabular-nums"> · {{ entry.actorId }} · {{ entry.ip }}</span>
          </span>
          <span class="font-mono text-[var(--gvp-text-secondary)] tabular-nums">{{ entry.timestamp }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
