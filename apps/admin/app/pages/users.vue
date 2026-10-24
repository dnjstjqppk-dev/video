<script setup lang="ts">
import { UserRole, SubscriptionTier } from '@gvp/shared-types'

const api = useAdminApi()

useSeoMeta({
  title: 'User & Channel Governance — RBAC & Sanctions'
})

const searchFilter = ref('')
const sanctionReason = ref('Repeated Copyright Policy Violation / Spam Upload')

interface ManagedUserRow {
  id: string
  email: string
  username: string
  displayName: string
  role: UserRole
  subscriptionTier: SubscriptionTier
  accountState: 'ACTIVE' | 'WARNED' | 'UPLOAD_FROZEN' | 'DEMONETIZED' | 'SUSPENDED'
}

const users = ref<ManagedUserRow[]>([
  {
    id: 'usr_1',
    email: 'creator@gvptech.media',
    username: 'gvptech',
    displayName: 'GVP Tech Engineering',
    role: UserRole.CREATOR,
    subscriptionTier: SubscriptionTier.PREMIUM,
    accountState: 'ACTIVE'
  },
  {
    id: 'usr_adv_1',
    email: 'ads@samsung-global.com',
    username: 'samsung_media',
    displayName: 'Samsung Electronics Media',
    role: UserRole.ADVERTISER,
    subscriptionTier: SubscriptionTier.FREE,
    accountState: 'ACTIVE'
  },
  {
    id: 'usr_suspect_88',
    email: 'reuploader88@example.com',
    username: 'movie_reupload_88',
    displayName: 'Fast Movie Clips 88',
    role: UserRole.CREATOR,
    subscriptionTier: SubscriptionTier.FREE,
    accountState: 'WARNED'
  }
])

const auditHistory = ref<Array<{ id: string; target: string; action: string; reason: string; timestamp: string }>>([
  {
    id: 'san_1',
    target: '@movie_reupload_88',
    action: 'WARNED',
    reason: 'Automated Content ID threshold exceeded',
    timestamp: '2026-10-23T19:00:00Z'
  }
])

const filteredUsers = computed(() =>
  users.value.filter(
    u =>
      u.username.toLowerCase().includes(searchFilter.value.toLowerCase()) ||
      u.email.toLowerCase().includes(searchFilter.value.toLowerCase())
  )
)

async function applySanction(
  user: ManagedUserRow,
  sanction: 'WARNED' | 'UPLOAD_FROZEN' | 'DEMONETIZED' | 'SUSPENDED' | 'ACTIVE'
) {
  user.accountState = sanction
  auditHistory.value.unshift({
    id: `san_${Date.now()}`,
    target: `@${user.username}`,
    action: sanction,
    reason: sanctionReason.value,
    timestamp: new Date().toISOString()
  })
  try {
    await api.updateAdminUserStatus(
      user.id,
      sanction === 'SUSPENDED' ? 'suspended' : 'active',
      sanctionReason.value
    )
  } catch {
    // Non-blocking fallback
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">사용자 및 채널 거버넌스 (User & RBAC Governance)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Search users & channels · Assign 15-role RBAC permissions · Execute progressive sanctions (Warn, Upload Freeze, Demonetize, Suspend)
    </p>

    <!-- Search & Sanction Reason Bar -->
    <div class="mt-6 grid grid-cols-1 gap-4 rounded-2xl border border-white/10 bg-[#131722] p-4 md:grid-cols-2">
      <div>
        <label class="text-xs text-neutral-400">Search by Username or Email</label>
        <input
          v-model="searchFilter"
          type="search"
          placeholder="Search @username or email..."
          class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3.5 py-2 text-xs text-white outline-none"
        />
      </div>
      <div>
        <label class="text-xs text-neutral-400">Mandatory Audit Reason for Sanction / Role Change</label>
        <input
          v-model="sanctionReason"
          type="text"
          class="mt-1 w-full rounded-xl border border-white/15 bg-[#090B10] px-3.5 py-2 text-xs text-white outline-none"
        />
      </div>
    </div>

    <!-- Users Governance Table -->
    <div class="mt-6 overflow-x-auto rounded-2xl border border-white/10 bg-[#131722]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th class="p-3.5">User / Channel</th>
            <th class="p-3.5">RBAC Role</th>
            <th class="p-3.5">Account State</th>
            <th class="p-3.5">Progressive Sanctions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="u in filteredUsers" :key="u.id" class="hover:bg-white/5">
            <td class="p-3.5">
              <p class="font-bold text-white">{{ u.displayName }} (@{{ u.username }})</p>
              <p class="text-[11px] text-neutral-400">{{ u.email }}</p>
            </td>
            <td class="p-3.5">
              <select
                v-model="u.role"
                class="rounded-lg border border-white/15 bg-[#090B10] px-2.5 py-1 font-mono text-xs text-[#3EA6FF]"
              >
                <option v-for="role in Object.values(UserRole)" :key="role" :value="role">
                  {{ role }}
                </option>
              </select>
            </td>
            <td class="p-3.5">
              <span
                class="rounded px-2.5 py-0.5 text-[10px] font-bold uppercase"
                :class="
                  u.accountState === 'ACTIVE'
                    ? 'bg-emerald-500/20 text-emerald-400'
                    : 'bg-red-500/20 text-red-300'
                "
              >
                {{ u.accountState }}
              </span>
            </td>
            <td class="p-3.5">
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded bg-amber-500/20 px-2.5 py-1 text-[11px] font-bold text-amber-300 hover:bg-amber-500/30"
                  @click="applySanction(u, 'WARNED')"
                >
                  Warn
                </button>
                <button
                  type="button"
                  class="rounded bg-orange-500/20 px-2.5 py-1 text-[11px] font-bold text-orange-300 hover:bg-orange-500/30"
                  @click="applySanction(u, 'UPLOAD_FROZEN')"
                >
                  Freeze Uploads
                </button>
                <button
                  type="button"
                  class="rounded bg-purple-500/20 px-2.5 py-1 text-[11px] font-bold text-purple-300 hover:bg-purple-500/30"
                  @click="applySanction(u, 'DEMONETIZED')"
                >
                  Demonetize
                </button>
                <button
                  type="button"
                  class="rounded bg-red-600/30 px-2.5 py-1 text-[11px] font-bold text-red-300 hover:bg-red-600/40"
                  @click="applySanction(u, 'SUSPENDED')"
                >
                  Suspend
                </button>
                <button
                  v-if="u.accountState !== 'ACTIVE'"
                  type="button"
                  class="rounded bg-emerald-500/20 px-2.5 py-1 text-[11px] font-bold text-emerald-300"
                  @click="applySanction(u, 'ACTIVE')"
                >
                  Restore
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Governance Audit Log -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5">
      <h2 class="text-sm font-bold text-white">최근 제재 및 권한 변경 감사 기록 (Sanction Audit Trail)</h2>
      <div class="mt-3 space-y-2 text-xs">
        <div
          v-for="entry in auditHistory"
          :key="entry.id"
          class="flex items-center justify-between rounded-xl bg-[#090B10] px-4 py-2.5"
        >
          <span>
            <strong class="text-red-400">[{{ entry.action }}]</strong> {{ entry.target }} — {{ entry.reason }}
          </span>
          <span class="font-mono text-neutral-500">{{ entry.timestamp }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
