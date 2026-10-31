<script setup lang="ts">
useSeoMeta({
  title: 'Multi-Region K8s, Canary Rollout & DR Drill — Super Admin'
})

const canaryTrafficPct = ref(25)
const maintenanceMode = ref(false)
const drDrillStatus = ref<'IDLE' | 'RUNNING' | 'VERIFIED'>('VERIFIED')
const lastDrRtoSeconds = ref(412) // 6m 52s (< 15m target)

const regions = ref([
  {
    id: 'ap-northeast-2',
    name: 'Seoul Primary (EKS ap-northeast-2)',
    role: 'PRIMARY_ACTIVE',
    pods: 64,
    cpuPct: 42,
    memPct: 58,
    cacheHitRatio: '98.9%',
    p99LatencyMs: 6.4,
    status: 'HEALTHY'
  },
  {
    id: 'ap-northeast-1',
    name: 'Tokyo Edge Failover (EKS ap-northeast-1)',
    role: 'HOT_STANDBY',
    pods: 32,
    cpuPct: 28,
    memPct: 41,
    cacheHitRatio: '98.4%',
    p99LatencyMs: 11.2,
    status: 'HEALTHY'
  },
  {
    id: 'us-west-2',
    name: 'Oregon Global POP (EKS us-west-2)',
    role: 'ACTIVE_REPLICA',
    pods: 48,
    cpuPct: 37,
    memPct: 52,
    cacheHitRatio: '97.8%',
    p99LatencyMs: 19.5,
    status: 'HEALTHY'
  },
  {
    id: 'eu-central-1',
    name: 'Frankfurt EU GDPR Node (EKS eu-central-1)',
    role: 'ACTIVE_REPLICA',
    pods: 40,
    cpuPct: 35,
    memPct: 49,
    cacheHitRatio: '98.1%',
    p99LatencyMs: 15.8,
    status: 'HEALTHY'
  }
])

function triggerDrDrill() {
  drDrillStatus.value = 'RUNNING'
  setTimeout(() => {
    lastDrRtoSeconds.value = 385
    drDrillStatus.value = 'VERIFIED'
  }, 1200)
}

function rollbackCanary() {
  canaryTrafficPct.value = 0
}
</script>

<template>
  <div>
    <!-- Infrastructure Simulation Banner (GV-07) -->
    <div class="mb-6 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-xs text-amber-300">
      <div class="flex items-center gap-2 font-bold">
        <span>⚠️ 인프라 시뮬레이션 안내 (Simulation Demonstration Mode)</span>
        <span class="rounded bg-amber-500/20 px-2 py-0.5 font-mono text-[10px] text-amber-200">DEMO ONLY</span>
      </div>
      <p class="mt-1 text-[11px] text-neutral-300">
        본 화면에 표시된 멀티 리전 EKS 클러스터 상태(4개 리전), 파드 수, 카나리 트래픽 시프팅, WAL 아카이빙 지표는 아키텍처 검증을 위한 목업 시뮬레이션 데이터입니다.
        실제 EKS 클러스터와의 실시간 텔레메트리 연동 상태가 아니며, DR 훈련 버튼은 가상 타이머 시뮬레이션만을 동작시킵니다.
      </p>
    </div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">멀티 리전 K8s 클러스터, 카나리 배포 및 재해복구(DR) 관제 (시뮬레이션)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          EKS Multi-Region Topology · Canary Progressive Traffic Shifting · PITR WAL Archiving & DR Failover Drill (RTO &lt; 15m, RPO &lt; 1m)
        </p>
      </div>
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="rounded-xl border px-4 py-2 text-xs font-bold transition"
          :class="maintenanceMode ? 'border-amber-400 bg-amber-500/20 text-amber-300' : 'border-white/15 bg-[#131722] text-neutral-300'"
          @click="maintenanceMode = !maintenanceMode"
        >
          {{ maintenanceMode ? '⚠️ Maintenance Mode ACTIVE' : '정기 점검 모드 전환 (Maintenance Mode)' }}
        </button>
        <button
          type="button"
          class="rounded-xl bg-red-600 px-4 py-2 text-xs font-bold text-white shadow hover:bg-red-500"
          @click="triggerDrDrill"
        >
          {{ drDrillStatus === 'RUNNING' ? 'Executing Failover...' : 'DR 페일오버 시뮬레이션 시연 (Simulation Only)' }}
        </button>
      </div>
    </div>

    <!-- Canary Rollout & PITR Backup Bar -->
    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white">카나리 점진적 배포 제어 (Canary Deployment v2026.10.30)</h2>
            <p class="text-[11px] text-neutral-400">Cosign OIDC Verified · Trivy CVE Critical: 0 · SonarQube Coverage: 91.4%</p>
          </div>
          <span class="rounded bg-[#3EA6FF]/20 px-2.5 py-1 font-mono text-xs font-bold text-[#3EA6FF]">
            Traffic: {{ canaryTrafficPct }}%
          </span>
        </div>

        <div class="mt-4 flex items-center gap-2">
          <button
            v-for="pct in [0, 5, 25, 50, 100]"
            :key="pct"
            type="button"
            class="flex-1 rounded-xl border py-2 text-xs font-bold transition"
            :class="canaryTrafficPct === pct ? 'border-[#3EA6FF] bg-[#3EA6FF] text-black' : 'border-white/10 bg-[#090B10] text-neutral-300'"
            @click="canaryTrafficPct = pct"
          >
            {{ pct }}%
          </button>
          <button
            type="button"
            class="rounded-xl border border-red-500/40 bg-red-950/40 px-3 py-2 text-xs font-bold text-red-300 hover:bg-red-900/50"
            @click="rollbackCanary"
          >
            Instant Rollback
          </button>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white">PostgreSQL 18 PITR 백업 & 재해 복구(DR) 지표</h2>
            <p class="text-[11px] text-neutral-400">Hourly WAL Archiving to Cloudflare R2 · Cross-Region Synchronous Replication</p>
          </div>
          <span class="rounded bg-emerald-500/20 px-2.5 py-1 font-mono text-xs font-bold text-emerald-400">
            {{ drDrillStatus }}
          </span>
        </div>
        <div class="mt-4 grid grid-cols-3 gap-3 text-center text-xs">
          <div class="rounded-xl bg-[#090B10] p-3">
            <p class="text-neutral-400">시뮬레이션 RTO (Target < 15m)</p>
            <p class="mt-1 font-mono text-base font-extrabold text-emerald-400">{{ Math.floor(lastDrRtoSeconds / 60) }}m {{ lastDrRtoSeconds % 60 }}s</p>
          </div>
          <div class="rounded-xl bg-[#090B10] p-3">
            <p class="text-neutral-400">시뮬레이션 RPO (Target < 1m)</p>
            <p class="mt-1 font-mono text-base font-extrabold text-[#3EA6FF]">0.8 sec</p>
          </div>
          <div class="rounded-xl bg-[#090B10] p-3">
            <p class="text-neutral-400">Global CDN Hit Ratio</p>
            <p class="mt-1 font-mono text-base font-extrabold text-white">98.7%</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Multi-Region Kubernetes Clusters -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5">
      <h2 class="text-sm font-bold text-white">글로벌 멀티 리전 쿠버네티스 클러스터 토폴로지 (Multi-Region K8s Fleet)</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
        <div
          v-for="reg in regions"
          :key="reg.id"
          class="rounded-xl border border-white/10 bg-[#090B10] p-4 text-xs"
        >
          <div class="flex items-center justify-between">
            <div>
              <p class="font-bold text-white">{{ reg.name }}</p>
              <p class="font-mono text-[11px] text-neutral-400">{{ reg.id }} · {{ reg.role }}</p>
            </div>
            <span class="rounded bg-emerald-500/20 px-2 py-0.5 font-mono text-[10px] font-bold text-emerald-400">
              {{ reg.status }}
            </span>
          </div>
          <div class="mt-3 grid grid-cols-4 gap-2 border-t border-white/10 pt-3 text-center">
            <div>
              <p class="text-[10px] text-neutral-400">Active Pods</p>
              <p class="font-mono font-bold text-white">{{ reg.pods }}</p>
            </div>
            <div>
              <p class="text-[10px] text-neutral-400">CPU / Mem</p>
              <p class="font-mono font-bold text-neutral-200">{{ reg.cpuPct }}% / {{ reg.memPct }}%</p>
            </div>
            <div>
              <p class="text-[10px] text-neutral-400">Edge Cache Hit</p>
              <p class="font-mono font-bold text-emerald-400">{{ reg.cacheHitRatio }}</p>
            </div>
            <div>
              <p class="text-[10px] text-neutral-400">P99 Latency</p>
              <p class="font-mono font-bold text-[#3EA6FF]">{{ reg.p99LatencyMs }}ms</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

