<script setup lang="ts">
useSeoMeta({
  title: '이용약관',
  description: 'StreamVerse Global Video Platform 이용약관: 계정, 콘텐츠, 유료 결제, 저작권, 분쟁 해결 규정을 안내합니다.'
})
</script>

<template>
  <article class="mx-auto max-w-3xl px-4 py-12 sm:px-6">
    <p class="text-xs font-bold uppercase tracking-wider text-[#FF0033]">Legal</p>
    <h1 class="mt-2 text-3xl font-extrabold">이용약관</h1>
    <p class="mt-2 text-xs text-neutral-400">시행일: 2026년 1월 1일 · 최종 개정: 2026년 11월 1일 · 버전 1.2</p>

    <div class="mt-8 space-y-8 text-sm leading-relaxed text-neutral-300">
      <section>
        <h2 class="text-base font-bold text-white">제1조 (목적)</h2>
        <p class="mt-2">이 약관은 StreamVerse Global Video Platform(이하 “서비스”)이 제공하는 동영상 스트리밍, 라이브 방송, 크리에이터 스튜디오, 파트너 광고 및 정산 서비스의 이용 조건과 회사·이용자의 권리·의무를 정합니다.</p>
      </section>
      <section>
        <h2 class="text-base font-bold text-white">제2조 (계정과 인증)</h2>
        <ul class="mt-2 list-disc space-y-1.5 pl-5">
          <li>회원가입과 로그인은 OIDC(OpenID Connect) 기반 외부 인증 제공자를 통해 수행되며, 회사는 비밀번호를 직접 보관하지 않습니다.</li>
          <li>이용자는 14세 이상이어야 하며, 14세 미만은 법정대리인 동의 없이 가입할 수 없습니다. 연령 제한 콘텐츠는 별도 인증이 필요합니다.</li>
          <li>계정 도용이 의심되면 회사는 추가 인증(2단계 인증, TOTP)을 요구하거나 세션을 강제 종료할 수 있습니다.</li>
        </ul>
      </section>
      <section>
        <h2 class="text-base font-bold text-white">제3조 (콘텐츠와 라이선스)</h2>
        <ul class="mt-2 list-disc space-y-1.5 pl-5">
          <li>이용자가 업로드한 영상의 저작권은 이용자에게 있으며, 이용자는 서비스 운영·송출·홍보에 필요한 범위에서 회사에 비독점적 이용 권한을 부여합니다.</li>
          <li>타인의 저작권을 침해하는 업로드는 삭제·수익화 중단·경고 누적(스트라이크) 조치 대상이며, 반복 시 계정이 정지될 수 있습니다.</li>
          <li>저작권자는 권리 보호 센터를 통해 침해 신고·이의 제기를 할 수 있으며, 회사는 접수일로부터 10영업일 이내에 1차 판정을 통지합니다.</li>
        </ul>
      </section>
      <section>
        <h2 class="text-base font-bold text-white">제4조 (유료 결제와 Premium)</h2>
        <ul class="mt-2 list-disc space-y-1.5 pl-5">
          <li>Premium 멤버십 결제는 TossPayments와 Stripe를 통해 처리되며, 모든 거래는 복식부기 원장에 기록되고 영수증 ID가 발급됩니다.</li>
          <li>결제는 서버 승인(Fail-Closed) 방식으로만 확정됩니다. 승인 없이 등급이 변경되거나 과금되는 일은 없습니다.</li>
          <li>청약 철회는 결제 후 7일 이내, 시청·다운로드 등 이용 기록이 없는 경우 전액 환불됩니다. 부분 환불과 정기결제 해지는 마이페이지에서 언제든 신청할 수 있습니다.</li>
        </ul>
      </section>
      <section>
        <h2 class="text-base font-bold text-white">제5조 (금지 행위)</h2>
        <p class="mt-2">불법 촬영물·아동 성착취물·혐오 조장·스팸·조회수 조작·계정 거래·리버스 엔지니어링·서비스 장애 유발 행위를 금지합니다. 위반 시 콘텐츠 삭제, 수익 회수, 계정 정지 및 수사기관 신고 조치를 할 수 있습니다.</p>
      </section>
      <section>
        <h2 class="text-base font-bold text-white">제6조 (계정 삭제와 데이터 파기)</h2>
        <p class="mt-2">이용자는 언제든 계정 삭제를 요청할 수 있으며, 요청 후 30일의 유예 기간 동안 철회할 수 있습니다. 유예 종료 후 영상·댓글·시청 기록 등 개인 데이터는 복구 불가 방식으로 파기되고, 법령상 보관 의무가 있는 결제 기록만 별도 보관됩니다.</p>
      </section>
      <section>
        <h2 class="text-base font-bold text-white">제7조 (책임과 분쟁 해결)</h2>
        <p class="mt-2">회사는 천재지변·통신 장애 등 불가항력으로 인한 손해에 책임을 지지 않습니다. 분쟁은 대한민국 법을 준거법으로 하며, 서울중앙지방법원을 전속 관할로 합니다. 문의: <span class="font-mono text-xs">legal@streamverse.example</span></p>
      </section>
    </div>
  </article>
</template>
