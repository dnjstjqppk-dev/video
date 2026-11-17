# StreamVerse / Global Video Platform — 구현 현황 및 개발 백로그 (Source of Truth)

- **작성일:** 2026-11-16 (전수 코드 대조 기준)
- **브랜드:** StreamVerse (법인: Booking Network Co., Ltd. / Master Directive `BN-GVP-YOUTUBE-CLASS-PRODUCTION-001`)
- **판정:** 상용 불가 (COMMERCIAL NO-GO) — 아키텍처 테스트베드 + 프로토타입 단계. `docs/audits/COMMERCIAL_REMEDIATION_AUDIT.md`의 NO-GO 판정을 유지한다.
- **분류 기준:** `IMPLEMENTED / PARTIAL / MOCK / BROKEN / MISSING / UNVERIFIED` (Master Directive ARTICLE 1)
- **2026-11-17 remediated:** P0-3 결제 fail-closed, P0-8 광고 인증·테넌트 격리, P0-4 재생 서명·가시성 (조회수 중복 제거·지역/연령 검사는 잔여). 상세 §4.4/§4.5/§4.6의 `FIXED 2026-11-17` 표기 및 §6 참조.

---

## 1. 이 사이트의 정체 (1줄 정의)

**StreamVerse는 YouTube급 독립 브랜드를 목표로 하는 멀티서페이스 동영상 플랫폼 모노레포다.**
4개 Nuxt 4 웹 콘솔(사용자 포털·Creator Studio·파트너/광고주·관리자) + Flutter iOS/Android 앱 + Go 모듈러 모놀리스 API + 6개 Go 워커 + PostgreSQL 마이그레이션 75테이블로 구성되며, 현재 프론트 4종은 Cloudflare Workers에 배포되고 API는 운영자 PC에서 Cloudflare Tunnel로 노출되는 하이브리드 형태로 기동한다.
결제 어댑터·S3 멀티파트·광고 결정 엔진 같은 뼈대는 실제로 동작하지만, **영구 저장·실제 트랜스코딩·실제 광고 소재 송출·실제 라이브 인제스트·NATS/Valkey/OpenSearch/ClickHouse 연동은 아직 코드에 연결되지 않았다.**

## 2. 표면(서페이스) 인벤토리

| 표면 | 경로 | 기술 | 현재 상태 |
| :--- | :--- | :--- | :--- |
| 사용자 웹 포털 | `apps/web` | Nuxt 4, Shaka/HLS 플레이어 | 배포 중. API 우선 + 시드 폴백 혼재 |
| Creator Studio | `apps/studio` | Nuxt 4, 멀티파트 업로더 | 배포 중. 업로드 실배선, 분석은 폴백 혼재 |
| 파트너/광고주 포털 | `apps/partners` | Nuxt 4 | 배포 중. 캠페인/정산/권리 화면 실배선(메모리 기반) |
| 관리자 콘솔 | `apps/admin` | Nuxt 4 | 배포 중. Ad Delivery Inspector 실배선(메모리 기반) |
| 모바일 앱 | `apps/mobile` | Flutter, Riverpod, GoRouter, Dio, media_kit | 빌드됨. 화면별 실데이터 여부 미확인 |
| API 프록시 워커 | `apps/api-proxy` | Cloudflare Worker | 존재. 터널 URL을 콘솔과 분리하는 안정 주소 |
| Go API | `services/api` | Go 1.27, chi, PostgreSQL 메모리 병행 | PC 호스팅 `:18092`. 메모리 스토어가 실서빙 주체 |
| 워커 6종 | `services/{media,analytics,recommendation,live-controller,realtime-chat,scheduler}-worker` | Go stdlib only | 실행됨. 전부 인메모리 시뮬레이터 (외부 의존성 0) |

## 3. 요청 흐름 (실제 가동 토폴로지)

```text
browser -> console workers (web/studio/partners/admin, /api)
       -> api-proxy worker (API_UPSTREAM secret)
       -> cloudflared quick tunnel (*.trycloudflare.com, 재시작 시 변경)
       -> gvp-api.exe :18092 + gvp-media-worker.exe :8081 (운영자 PC)
```

- `8080`은 타 프로세스 점유로 사용 금지, API는 `:18092`로 기동한다 (`infra/local/README.md`).
- 상세 절차는 [infra/local/README.md](../infra/local/README.md)를 따른다.

## 4. 모듈별 구현 상태 (증거 포함)

### 4.1 인증·인가 — PARTIAL (배포 설정은 P0)

- OIDC Authorization Code + PKCE 로그인/콜백/로그아웃 라우트: 코드상 IMPLEMENTED (`apps/*/server/routes/auth/*.ts`, `services/api/internal/security/*`, `services/api/internal/identity/middleware.go`).
- JWKS RS256 검증·issuer/audience/exp/nbf 검증: 코드상 IMPLEMENTED, 실테넌트 연동은 UNVERIFIED (프로덕션 ZITADEL 테넌트 없음).
- 개발 토큰(`admin/creator/viewer/advertiser-token`)의 운영 차단 게이트: 코드상 IMPLEMENTED (production + 미설정 시 불가).
- **P0:** 현 배포는 `GVP_ENV=test` + `GVP_ALLOW_DEV_TOKENS=true`로 기동 중이라 데모 로그인이 열려 있다. 실사용자 전환 전 OIDC 필수 (`docs/operations/OIDC_SETUP.md`).
- RBAC 역할 게이트 + BOLA 소유권 확인(채널/영상/업로드/댓글): IMPLEMENTED (핸들러 내 검사 + `ownership_test.go`).
- 관리자 IP 화이트리스트: IMPLEMENTED (`security.AdminGuard`, `/v1/admin` 그룹 적용).
- 사용자 TOTP MFA 발급/검증/해제 엔드포인트: IMPLEMENTED (`/v1/me/mfa/*`, `users/handler.go`).
- 관리자 MFA 강제·민감 작업 재인증: MISSING (역할 게이트에 MFA 클레임 검사 없음).
- 조직 단위 ABAC(광고주·파트너 격리): PARTIAL (광고 엔진은 캠페인 단위 검사, 조직 스코프 강제 미확인).

### 4.2 영속 저장 — PARTIAL (이중 스키마 해소됨, 코어 4종 영속화)

- `db/migrations/00001..00010`: **75개 테이블 DDL IMPLEMENTED** (계정 10, 콘텐츠 14, 소셜 9, 라이브 5, 광고 6, 결제 10, 저작권 8, 운영 5, 광고엔진 8 + 00010 누락 역할 9종·결제 컬럼 2종 추가).
- **FIXED 2026-11-17:** 간이 DDL `InitSchema` 삭제 → `VerifySchema`(생성 없이 검증; `DATABASE_URL` 설정 + 스키마 불일치 시 부팅 거부)로 교체.
  `SaveUser/SaveChannel/SaveVideo/SaveTransaction`이 실스키마(users+user_profiles·UUID·tags·원장 계정 매핑) 기준으로 재작성됐고,
  부팅 시 `Hydrate`로 DB 행을 메모리에 적재한다. 사용자·채널·영상·결제+원장은 재시작 내구성 실증 완료(실서버 재시작 증명).
- 잔여: 댓글·광고·라이브·소셜 등 나머지 도메인은 메모리 전용 (동일 패턴으로 확장 예정). 비-UUID 구형 fixture ID는 경고 후 스킵.
- sqlc 쿼리 5종(`db/queries`) PARTIAL, 시드(`db/seeds/seeds.sql`) 존재.
- `db/schema/schema.sql`, `db/schema/inventory.csv`, `docs/architecture/erd/*`는 실제 마이그레이션과 **다른 88테이블 스키마를 기술** → STALE (본 문서 §7 참조).

### 4.3 업로드·미디어 파이프라인 — PARTIAL (인제스트 실동작, 트랜스코딩 MOCK)

- 멀티파트 세션 생성·presigned 파트 URL(SigV4, R2/MinIO)·완료 검증: IMPLEMENTED (`uploads/s3.go`, `uploads/handler.go`).
- 로컬 릴레이 모드(실바이트 수신·조립·디스크 저장): IMPLEMENTED.
- 완료 → 미디어 워커 HTTP 디스패치(`/jobs/transcode`): IMPLEMENTED (`uploads/dispatch.go`).
- **P0 MOCK:** 디스패치는 probe 메타데이터 없이 전송하고, 워커는 ffprobe/ffmpeg 실행·R2 I/O 없이 **가정 프로파일(1080p/60s)로 READY를 반환**한다
  (`services/media-worker/cmd/worker/main.go` — `exec.Command` 전무, `ProcessJob`은 인메모리 전이 + 아티팩트명 나열).
  즉 업로드 완료→READY는 아직 **가짜 완료**이며 Master Directive ARTICLE 2.3 미달이다.
- 워커 도달 불가 시 재시도·큐·DLQ 복구 없음 (QUEUED 영구 정체): MISSING.
- 썸네일·스토리보드·WebP·자막·챕터 생성: MISSING (인자 빌더 함수만 존재).
- READY 전환 시 영상을 강제 `public`으로 뒤집는다 (`dispatch.go`) — 비공개 초안 정책과 충돌하므로 P1 수정 대상.

### 4.4 재생·시청 — PARTIAL

- `GET /v1/videos/{id}/playback`: **FIXED 2026-11-17** — HMAC-SHA256 서명 토큰(`v1.<b64id>.<exp>.<sig>`, 1h, 영상 바인딩) 발급 + private(비소유 404)·members_only(미자격 403) 검사 (`videos/playback_token.go`, `videos/visibility.go`). 지역/연령 검사는 정책 데이터 부재로 잔여.
- 조회수 증가가 재생 API와 분석 이벤트(`valid_view`)에서 **중복 가산**되고 중복 제거 없음 → P1 (P0-4 잔여, P1-14에서 처리).
- HLS 서명·AES-128 키 발급·Widevine/FairPlay: MISSING (문서만 존재).

### 4.5 광고 (최상위 P0 비즈니스 요구) — PARTIAL

- 결정 API(`/v1/ads/decision`): 상태·예산·일정·카테고리·포맷 검사 후 후보 선정 — 인메모리 ENGINE IMPLEMENTED (`ads/handler.go`).
- 이벤트 수집·미리보기·캠페인 CRUD·소재 승인/거절·광고주 리포트·Admin 송출 진단: IMPLEMENTED (인메모리).
- 홈 피드 배너 슬롯: 실배선 (`apps/web/app/pages/index.vue`의 `getAdDecision` + 노출/클릭 전송).
- Admin Ad Delivery Inspector: 실배선 (`apps/admin/app/pages/ads.vue`).
- **MISSING:** 시청 페이지 프리롤/미드롤 연동 (watch 페이지에 광고 호출 전무), 소재 파일의 실제 스토리지 업로드·검수 파이프라인,
  예약형 강제 배정 검증 모드, 노출 조건 정의(플레이어 수신≠노출) 강제, 빈도 제한 집행 확인, E2E(등록→심사→송출→집계) 실증.
- 노출 이벤트가 마운트 시점에 무조건 발사되어 유효 노출 정의 미달 소지: P1.
- **P0 인증 구멍: FIXED 2026-11-17** — `/v1/ads/*`·`/v1/advertiser/*` 캠페인/리포트/미리보기는 광고주 역할 게이트 + 광고주 스코프 격리, `/v1/ads/admin/*`는 플랫폼 관리자 게이트.
  생성 시 `draft` + 소재 `pending` (ARTICLE 7.1 상태머신 강제), 승인/거절 감사 로그 기록, `/advertiser` 전용 라우터로 중복 마운트 해소(143→137라우트).
  상세: [ROUTE_CENSUS.md](api/ROUTE_CENSUS.md) ads/advertiser 그룹.

### 4.6 결제·정산 — PARTIAL + P0 FAIL-OPEN 1건

- Stripe Checkout Sessions 실 API + Toss 승인 실 API + 웹훅 HMAC 검증 + 멱등키 + 복식부기 원장: IMPLEMENTED (`payments/*`).
- **P0 FAIL-OPEN: FIXED 2026-11-17** — `ConfirmPayment`가 미구성 프로바이더(503)·미지정 프로바이더(400)·금액 불일치(409)를 거부하고,
  Toss(DONE + 금액 일치 확인)와 Stripe(Checkout Session `paid` + 금액/통화/트랜잭션 바인딩 확인)를 서버에서 실검증한다 (`payments/handler.go`, `payments/stripe.go`).
- 환불·차지백·분쟁·정산 배치 실행·실지급(Bank/PG payout): MISSING 또는 UNVERIFIED.
- KRW `amount_cents` 혼용 여부: UNVERIFIED (통화 최소단위 정책 점검 필요).

### 4.7 라이브 — PARTIAL (메타·채팅만, 인제스트 MISSING)

- 방송 CRUD·상태 전이·스트림키 발급/회전· ll-HLS URL 필드·WebSocket 채팅 허브·슬로모드·슈퍼챗·투표: 인메모리 IMPLEMENTED (`live/*`).
- **MISSING:** 실제 RTMP(S)/SRT 인제스트 서버, Cloudflare Stream 연동 호출, ABR 라이브 인코딩, DVR/되감기, 종료 후 VOD 아카이브,
  시청자 수 실집계, 채팅 금칙어/도배 방지 실집행.

### 4.8 Studio·파트너·관리자·소셜 — PARTIAL (메모리 CRUD)

- Studio 대시보드/콘텐츠/업로드/댓글/분석/수익 화면과 API: 실배선이나 집계는 인메모리·시드 혼재.
- 파트너 캠페인/MCN/권리/정산, 관리자 사용자/모더레이션/감사로그/정책 화면과 API: 실배선(메모리 기반).
- 저작권 지문 생성·매칭: MISSING (클레임/분쟁/DMCA 폼만 존재).
- 댓글/좋아요/구독/재생목록/시청기록: 인메모리 IMPLEMENTED.

### 4.9 이벤트·분석·추천·검색 — PARTIAL (수집구만, 싱크 MISSING)

- `POST /v1/events/analytics`: 수락 + 인메모리 조회수++ 만 수행. `event_id` 중복 제거·NATS 발행·ClickHouse 적재 없음 (`events/handler.go`).
- Go API와 워커 6종에 NATS/Valkey/ClickHouse/OpenSearch 클라이언트 의존성이 **전무** (`go.mod` 전수 확인) — compose의 해당 서비스는 미접속 장식물이다.
- 추천/검색: 인메모리 필터·페이지네이션만. 랭킹·OpenSearch 색인·Valkey 캐시·자동완성 실랭킹 없음.
- **P0 개인정보 유출: FIXED 2026-11-17** — 검색·자동완성이 private/unlisted을 제외하고 members_only는 유자격자에게만 반환한다 (`recommendations/handler.go` + `videos.SearchVisible`).
- 피드백 API는 저장 없이 `"recorded"` 가짜 성공, 검색어 제안은 빈 결과에 하드코딩 제안 반환: P1.
- 관리자 대시보드 API에 하드코딩 지표(`activeUsersToday: 1240`, `totalAdRevenueKrw: 18450000`) 혼재: P1.
- `comments`·`copyright`·`moderation`·`settlements` 패키지의 `Routes()`는 마운트되지 않은 데드 코드 (main.go가 메서드 직접 배선): P2 정리 대상.
- Outbox: API가 `outbox_events`에 쓰지 않고 스케줄러는 자체 인메모리 큐만 보유 → 트랜잭셔널 아웃박스 MISSING.

### 4.10 프론트 공통 결함

- **BROKEN:** `apps/studio/app/pages/upload.vue`, `apps/admin/app/pages/ads.vue`, `apps/web/app/pages/watch/[id].vue`에 **깨진 UTF-8 바이트**(한글 모지바케) 존재. P1로 정상 UTF-8 재작성 필요.
- **STALE:** `apps/*/nuxt.config.ts`의 콘솔 상호 링크가 구 서브도메인(`dnjstjqppk`)을 가리키고 실배포(`dnjstjq0824`)와 불일치.
- 시드 폴백: 홈/시청/분석/Shorts·라이브 선반이 API 실패 시 가짜 수치와 함께 렌더링된다. Empty State 정책(지시 ARTICLE 0.5) 위반 소지 → P1.
- `cloudflare-deploy.yml`은 4개 콘솔만 배포하고 `apps/api-proxy`를 배포하지 않는다 → MISSING (수동 배포 절차 필요).
- 메모리상의 `/api 404 error 1042` 이슈는 본 세션에서 미재현 → UNVERIFIED (재검증 필요).

### 4.11 모바일 — UNVERIFIED (표면적 PARTIAL)

- Dio 에러 envelope 처리·OIDC 서비스·7개 화면·라우터·오프라인/네이티브 브릿지 파일 존재. 화면별 실데이터/목업 여부는 본 세션 미판독 → 전수 확인 필요.

### 4.12 테스트·CI/CD — PARTIAL

- Go 단위/통합, 계약 6종, OWASP 4종, API 레벨 E2E 4종, k6 스펙, Flutter 테스트: 존재. **브라우저 실렌더링 E2E·실결제·실송출 검증은 없음.**
- OpenAPI 스펙은 chi.Walk 실측 143라우트 기준으로 동기화됐다 (121 경로, 상세 23 + 인벤토리 98,
  `docs/generate_route_census.py`, 계약 게이트 테스트 포함). 엔드포인트별 상세 스키마는 P1 잔여.
- ERD 9종도 실측 마이그레이션 기준으로 재생성됐다 (`docs/generate_erd.py`). P1-8/P1-9의 문서 동기화분은 본 회차 완료.
- GitHub CI/CD 그린 이력 있음(11-07, d1ce1f6). 최신 `5d2c347`의 파이프라인 결과는 UNVERIFIED.
- 원격은 GitHub인데 지시는 GitLab 우선(ARTICLE 0.12/19) → GitLab 파이프라인 실가동 UNVERIFIED.

## 5. 실제 개발 필요 항목 (P0 → P2)

### P0 — 상용 차단 (순서대로)

1. **이중 스키마 해소 + 영속 저장 완성** (ARTICLE 2.2): ~~`InitSchema` 간이 DDL 삭제, 마이그레이션 기준 저장 계층, 부팅 시 읽기, 재시작 내구성 검증~~ **DONE 2026-11-17 (코어 4종)** —
   사용자·채널·영상·결제+원장 영속화 + 실서버 재시작 증명 + PG 라운드트립 3종. 댓글·광고·라이브·소셜 전 도메인 확장은 후속 과제로 잔여.
2. **진짜 트랜스코딩** (ARTICLE 2.3): ffprobe 실측 → 디스패치에 메타 전달 → 워커에서 ffmpeg 실실행(HLS ladder·썸네일·자막·챕터) → R2 업로드 → 실패 시 FAILED + 재시도/DLQ. `REQUIRE_PROBED_SOURCE=true`를 기본으로.
3. **결제 fail-open 제거** (ARTICLE 2.4): ~~Toss/Stripe 시크릿 없으면 `ConfirmPayment` 거부~~ **DONE 2026-11-17** (목 테스트 8건). 환불·웹훅 재조정·원장 정합 테스트는 P1-17 잔여.
4. **재생·검색 권한 검사** (ARTICLE 16): ~~HMAC 서명 playback 토큰(1h, 경로 바인딩) + 가시성/회원권 검사 + 검색 가시성 필터~~ **DONE 2026-11-17** (목 테스트 9건). 지역/연령 검사·유효 조회수 정의·중계복 제거는 잔여 (P1-14).
5. **광고 E2E 실증** (ARTICLE 7): 소재 실업로드 → 심사 → 승인 → 인벤토리 → 결정 → **시청 프리롤 렌더링** → 유효 노출 집계 → 광고주 대시보드 전 구간 테스트 통과.
6. **운영 인증 전환**: ZITADEL 실테넌트 발급 + 4콘솔/모바일 클라이언트 등록 + `GVP_ENV=production`, 개발 토큰 완전 차단 검증.
7. **관리자 MFA + 민감 작업 재인증 + 감사 무결성** (ARTICLE 2.1/9).
8. **광고 서피스 인증·테넌트 격리** (ARTICLE 2.1/7): ~~`/v1/ads`·`/v1/advertiser` 마운트에 인증/역할/광고주 스코프 강제,
   생성 시 자동 승인 제거(DRAFT→심사 상태머신), 중복 마운트 해소, 승인/거절 감사 로그~~ **DONE 2026-11-17** (목 테스트 6건 + E2E 상태머신).

### P1 — 상용 핵심

8. OpenAPI 전 라우트 동기화 + 계약 테스트 확대 (ARTICLE 11).
9. ERD·inventory·schema.sql을 실제 마이그레이션 기준으로 재생성.
10. 모지바케 3파일 UTF-8 복구 + 전수 바이트 검사.
11. 시드 폴백의 Empty State 전환 (데이터 없음은 빈 화면, 오류는 오류로).
12. 크로스 콘솔 URL 현행화 + api-proxy 배포 자동화 + `/api 1042` 재검증.
13. NATS JetStream 실연동 (미디어잡·분석 이벤트·아웃박스 릴레이) 또는 범위 축소 결정.
14. 조회수/분석 중복 제거 + ClickHouse 싱크 또는 PG 롤업 확정.
15. 라이브 실인제스트 범위 확정 (Cloudflare Stream 실연동 또는 RTMP 자체 구현).
16. Shorts 촬영·업로드·음악 권리, 댓글 정책(차단어/제한), 알림·구독 피드 완성.
17. 환불·정산 배치·실지급·세금계산서·광고 인보이스·미달 경고/보상 절차.
18. RBAC 15롤·조직 ABAC 전수 강제 + 침투/권한 격리 테스트.
19. 반응형 8해상도 + WCAG 2.2 AA + 플레이어 단축키/자막/챕터/PiP.
20. 부하(k6 1k CCU)·복구(PITR/DLQ)·SLO·비용 실측.

### P2 — 고도화

21. 추천 랭킹 파이프라인 (규칙→행동→랭킹 단계적).
22. OpenSearch 통합 검색·자동완성·오타 교정·SEO 구조화 데이터.
23. DRM(Widevine/FairPlay), DASH, 4K+ ladder, 다국어 더빙.
24. Content ID 지문, MCN 전자서명·수익 분배, 브랜드 마켓.
25. TV/CTV, 커머스/라이브쇼핑, 외부 파트너 API.

## 6. 문서 정합성 메모 (본 회차 수정분)

- 본 파일이 현황의 단일 진실 공급원이다. 아래 문서들은 전수 대조 후 최신화했다.
- `README.md`: 디렉터리 수치를 실측으로 정정 + 본 문서 링크.
- `docs/architecture/SYSTEM_ARCHITECTURE.md`: 실가동 토폴로지(§0) 추가.
- `docs/architecture/erd/`: 9종 다이어그램을 실측 마이그레이션 기준으로 재생성 + `docs/generate_erd.py` 커밋. 구 10종 삭제.
- `docs/api/API_INVENTORY.md`: 목표 계약 문서로 유지, 커버리지 경고문 추가.
- `docs/api/ROUTE_CENSUS.md` + `docs/api/routes.json` (신규): chi.Walk 실측 143라우트 + 인증 가드 표.
- `openapi/gvp-v1.yaml`: 23 상세 경로 + 98 인벤토리 경로로 동기화 (121 경로). 생성기 `docs/generate_route_census.py`.
- `tests/contract/openapi_contract.test.ts`: 센서스-스펙 정합 게이트 테스트 추가.
- `docs/operations/PRODUCTION_CHECKLIST.md` (신규): 상용 전환 게이트·시크릿·순서서.
- `docs/security/SECURITY_AND_COMPLIANCE.md`: 미구현 통제 목록(§5) 추가.
- `docs/operations/RUNBOOK_AND_DR.md`: 현행 호스팅 주의문 추가.
- `docs/operations/OIDC_SETUP.md`: 현 배포의 개발 토큰 상태 주의문 추가.
- 감사 보고서 2종은 역사 기록이므로 수정하지 않았다.
- **2026-11-17 P0-3/P0-8/P0-4 수정분:** `docs/api/ROUTE_CENSUS.md` + `routes.json` + `openapi/gvp-v1.yaml`을 chi 실측 137라우트로 재생성(HOLE 0건).
  증거: Go 25pkg + 워커 6 + 통합 1 그린, 계약/OWASP 9/9, Playwright 4/4(로컬 18086), 신규 목 테스트 23건(결제 8·광고 6·재생/검색 9).
