<script setup lang="ts">
useSeoMeta({
  title: '개인정보처리방침',
  description: 'StreamVerse 개인정보처리방침: 수집 항목, 보관 기간, 파기 절차, 이용자 권리와 보호 조치를 안내합니다.'
})
</script>

<template>
  <article class="mx-auto max-w-3xl bg-white px-4 py-12 text-slate-900 sm:px-6">
    <p class="text-xs font-bold uppercase tracking-wider text-blue-700">Privacy</p>
    <h1 class="mt-2 text-3xl font-extrabold text-slate-900">개인정보처리방침</h1>
    <p class="mt-2 text-xs text-slate-500 tabular-nums">시행일: 2026년 1월 1일 · 최종 개정: 2026년 11월 1일 · 버전 1.4</p>

    <div class="mt-8 space-y-6 text-sm leading-relaxed text-slate-600 tabular-nums">
      <section class="rounded-lg border border-slate-200 bg-slate-50 p-5">
        <h2 class="text-base font-bold text-slate-900">1. 수집하는 개인정보</h2>
        <ul class="mt-2 list-disc space-y-1.5 pl-5">
          <li><strong class="text-slate-900">필수:</strong> OIDC 식별자(이메일, 표시 이름), 국가·언어 설정, 결제 영수증 ID</li>
          <li><strong class="text-slate-900">서비스 이용 기록:</strong> 시청 기록·좋아요·구독·댓글, 기기 정보, IP·접속 로그(보안 목적 90일 보관)</li>
          <li><strong class="text-slate-900">선택:</strong> 프로필 이미지, 2단계 인증(TOTP) 시크릿, 마케팅 수신 동의</li>
          <li>신용카드 번호 등 결제 수단은 TossPayments·Stripe가 직접 처리하며 회사는 보관하지 않습니다.</li>
        </ul>
      </section>
      <section class="rounded-lg border border-slate-200 bg-slate-50 p-5">
        <h2 class="text-base font-bold text-slate-900">2. 이용 목적</h2>
        <p class="mt-2">수집 정보는 본인 인증, 맞춤 추천·검색, 광고 집행·정산, 부정 이용 방지, 고객 지원에만 사용되며 목적 외 이용 시 별도 동의를 받습니다.</p>
      </section>
      <section class="rounded-lg border border-slate-200 bg-slate-50 p-5">
        <h2 class="text-base font-bold text-slate-900">3. 보관과 파기</h2>
        <ul class="mt-2 list-disc space-y-1.5 pl-5">
          <li>계정 삭제 요청 시 30일 유예 후 영상·댓글·시청 기록을 복구 불가 방식으로 파기합니다.</li>
          <li>전자상거래법상 계약·결제 기록은 5년, 접속 로그는 3개월 보관 후 파기합니다.</li>
          <li>감사 로그의 IP·이메일은 마스킹 처리되어 최소 권한 관리자만 열람할 수 있습니다.</li>
        </ul>
      </section>
      <section class="rounded-lg border border-slate-200 bg-slate-50 p-5">
        <h2 class="text-base font-bold text-slate-900">4. 제3자 제공과 국외 이전</h2>
        <p class="mt-2">원칙적으로 제3자에게 제공하지 않습니다. 결제(TossPayments·Stripe), 인증(OIDC 제공자), CDN·미디어 저장(Cloudflare·R2) 위탁 범위에서만 필요한 정보가 처리되며, 국외 이전 시 항목·국가·보유 기간을 사전 고지합니다.</p>
      </section>
      <section class="rounded-lg border border-slate-200 bg-slate-50 p-5">
        <h2 class="text-base font-bold text-slate-900">5. 이용자의 권리</h2>
        <p class="mt-2">이용자는 열람·정정·삭제·처리 정지·동의 철회를 언제든 요구할 수 있으며, 만 14세 미만 아동의 정보는 법정대리인 동의를 확인한 뒤에만 처리됩니다(COPPA 준수). 요청은 <span class="rounded-md border border-slate-200 bg-white px-1.5 py-0.5 font-mono text-xs text-slate-900">privacy@streamverse.example</span>로 접수되며 10일 이내에 처리 결과를 통지합니다.</p>
      </section>
      <section class="rounded-lg border border-slate-200 bg-slate-50 p-5">
        <h2 class="text-base font-bold text-slate-900">6. 안전 조치</h2>
        <p class="mt-2">전송 구간 TLS 암호화, 비밀번호 Argon2id 해시(해당 시), 관리자 IP 허용 목록·15분 세션·2단계 인증, 분당 요청 제한과 이상 접속 탐지를 적용합니다. 침해 사고 발생 시 72시간 이내에 이용자와 관계 기관에 통지합니다.</p>
      </section>
    </div>
  </article>
</template>
