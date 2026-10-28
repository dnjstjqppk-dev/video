<USER_REQUEST>
# [BOOKING NETWORK CO., LTD]
# GLOBAL VIDEO PLATFORM — PRODUCTION COMMERCIALIZATION MASTER DIRECTIVE

**Command ID:** `BN-GVP-YOUTUBE-CLASS-PRODUCTION-001`

**Target:** 첨부된 `video-master.zip`의 Global Video Platform 전체 Monorepo

**Mission:** YouTube와 경쟁할 수 있는 독립 브랜드 동영상 플랫폼을 실제 사용자, 크리에이터, 광고주, 파트너가 사용할 수 있는 상용 서비스로 완성한다.

**Execution:** FULL SOURCE AUDIT → CRITICAL REMEDIATION → PRODUCT COMPLETION → DB INTEGRATION → REAL AD DELIVERY → E2E TESTS → RELEASE VERIFICATION

**Ownership:** Booking Network Co., Ltd.

---

## ARTICLE 0. 최상위 개발 원칙

본 명령은 단순 UI 개발이나 데모 프로젝트 제작이 아니다.

기존 `video-master.zip`을 기준선으로 사용하고, 전체 저장소의 실제 구현 상태를 확인한 후 상용 운영이 가능한 시스템으로 전환한다.

다음 조건을 절대 준수한다.

1. 기존 디자인의 강점을 보존하고 사용자 경험을 향상한다.
2. 기존 기능이나 코드를 무분별하게 삭제하지 않는다.
3. 모든 핵심 기능은 실제 API와 영구 저장소를 통해 동작한다.
4. Mock, Dummy, Fake Success, Synthetic Revenue, Sample Campaign 등을 운영 데이터로 표시하지 않는다.
5. 데이터가 비어 있으면 정상적인 Empty State를 표시한다.
6. API 오류는 실제 오류로 처리하고 성공한 것처럼 위장하지 않는다.
7. 미구현 기능에 대하여 Production Ready, 100% Complete라고 선언하지 않는다.
8. 사용자·크리에이터·파트너·광고주·관리자 간 데이터와 권한을 철저히 격리한다.
9. 광고 등록부터 사용자 노출, 성과 측정, 과금까지 실제 데이터를 연결한다.
10. 개인정보, 저작권, 결제, 광고 심사, 미성년자 보호와 각 국가 법규를 준수한다.
11. YouTube는 기능 벤치마크로 사용하되 자체 브랜드와 독립 디자인을 유지한다.
12. GitLab을 기준으로 CI/CD를 구성한다. 별도 지시가 없으면 GitHub 사용을 전제하지 않는다.

핵심 정의:

**CODE EXISTS ≠ FUNCTION WORKS ≠ PRODUCTION VERIFIED**

각 단계는 반드시 분리해 평가한다.

---

## ARTICLE 1. 전체 파일 무손실 전수 감사

전체 ZIP, 모든 앱, Go 서비스, SQL, Protobuf, OpenAPI, Flutter, 인프라, CI/CD 및 테스트를 전수 검사한다.

현재 확인된 기준 구성:

- `apps/web` — 사용자 Nuxt 웹
- `apps/studio` — Creator Studio
- `apps/partners` — 파트너 및 광고주 웹
- `apps/admin` — 관리자 웹
- `apps/mobile` — Flutter iOS/Android
- `services/api` — Go API
- `services/*-worker` — 미디어, 분석, 추천, 라이브, 실시간 채팅, 스케줄러
- `db/migrations` — 8개 SQL 마이그레이션
- `db/schema` — 스키마 및 인벤토리
- `packages/*` — 공통 UI, 타입, API Client
- `infra/*` — Docker, Kubernetes, Helm, Terraform, Monitoring
- `tests/*` — 계약, 보안, 통합, E2E, 성능 테스트

필수 감사 항목:

- 전체 파일/기능/라우트/API/DB 테이블 추적
- 사용하지 않는 코드와 중복 로직
- 샘플 데이터와 하드코딩
- 실제 UI 기능과 백엔드 호출의 불일치
- 인증 및 권한 우회
- API 계약 불일치
- 모바일과 웹 기능 차이
- 반응형 UI 및 접근성
- 결제·광고·정산 무결성
- 데이터 경쟁 조건
- 환경변수·비밀정보 관리
- 미디어 파이프라인 실동작
- 장애 복구 및 관측 가능성
- SEO, 성능, 운영비 구조

모든 기능은 `IMPLEMENTED`, `PARTIAL`, `MOCK`, `BROKEN`, `MISSING`, `UNVERIFIED` 중 하나로 분류한다.

감사표에는 최소한 파일 경로, 기능, 현재 상태, 결함, 심각도, 수정 방식, 의존 기능, 담당 모듈, 테스트 결과, 증거를 기재한다.

발견된 결함은 누락하지 말고 실제로 수정한다.

---

## ARTICLE 2. P0 — 운영 차단 결함 우선 정상화

다음 항목부터 처리한다.

### 2.1 인증 체계

현재 `services/api/internal/identity/middleware.go`의 개발용 토큰 허용과 미확인 토큰에 대한 기본 사용자 부여 방식을 운영 코드에서 제거한다.

실제 OIDC/JWT 검증을 연결한다.

- 유효 서명 검증
- issuer / audience / expiration / not-before 검증
- JWKS 키 회전
- 비활성 사용자 차단
- 세션 폐기·기기 관리
- 관리자 MFA
- 관리자 민감 작업 재인증
- RBAC + 조직 단위 ABAC
- 모든 객체별 소유권 확인
- API Rate Limiting
- 인증 실패는 401, 권한 실패는 403
- 비밀키는 Secret Manager 사용

데모 인증은 테스트 전용 환경에만 격리한다.

### 2.2 영구 데이터베이스

`store.NewMemoryStore()`를 운영 데이터 저장 수단으로 사용하는 구조를 제거한다.

PostgreSQL 기반 Repository / Service / Transaction 계층을 구현하고 모든 운영 도메인을 연결한다.

기존 스키마를 전수 비교하고 누락된 테이블만 전진 마이그레이션으로 추가한다.

기존 데이터를 파괴하거나 임의로 DB를 초기화하지 않는다.

### 2.3 진짜 영상 업로드

현재 업로드 완료 시 테스트 HLS URL을 붙이고 READY 상태를 부여하는 방식을 제거한다.

실제 multipart upload, 객체 저장소, 인코딩 워커, manifest 생성, 썸네일 생성, 메타데이터 검증, CDN 배포, 재생 확인을 연결한다.

### 2.4 진짜 결제

Stripe/Toss의 임시 URL 생성과 단순 문자열 서명 검증을 제거한다.

공식 결제 API, 실제 webhook 서명, 중복 방지, 환불, 원장 반영, 정산 재조정 기능을 구현한다.

### 2.5 실제 Worker 실행

메모리에서 함수를 호출하고 종료되는 Worker를 지속 실행 가능한 소비자·처리기로 전환한다.

Queue Consumer, DB Persistence, Retry, Backoff, Dead Letter Queue, Health Check, Graceful Shutdown을 구현한다.

---

## ARTICLE 3. 사용자 서비스 — YouTube급 기능 범위

PC, 모바일 웹, iOS, Android에서 다음 기능을 구현한다.

### 3.1 계정과 개인화

- 회원가입·로그인·SSO
- 프로필·프로필 이미지
- 사용자명·핸들
- 다중 채널 생성
- 구독 관리
- 알림 설정
- 시청 기록
- 검색 기록
- 이어보기
- 나중에 볼 동영상
- 좋아요한 동영상
- 비공개 재생목록
- 계정 데이터 다운로드
- 개인정보 삭제 요청
- 차단 채널·사용자
- 비추천 콘텐츠 관리
- 언어·국가·시간대·테마
- 비로그인 사용자 제한 정책
- 연령 확인과 아동 보호 기능

### 3.2 홈 및 탐색

- 개인화 추천 홈
- 최신 동영상
- 인기·급상승
- 구독 피드
- 카테고리
- 탐색 메뉴
- Shorts 피드
- 라이브 피드
- 채널 추천
- 검색 자동완성
- 검색어 오타 교정
- 인기 검색어
- 검색 필터
- 정렬
- 무한스크롤
- 중복 영상 억제
- 신고·차단 콘텐츠 제외
- 추천 이유·사용자 피드백

추천 시스템은 처음부터 과도한 AI 비용을 유발하지 않도록 규칙 기반, 행동 기반, 후보 생성 및 랭킹 단계로 확장한다.

### 3.3 동영상 시청

- HLS 적응형 화질
- 지원 가능한 코덱에 따른 DASH
- 240p~4K 이상 선택
- 자동 화질
- 배속·일시정지·탐색
- 5초/10초 건너뛰기
- 자막
- 다국어 자막
- 챕터
- 전체화면
- Picture-in-Picture
- 미니플레이어
- 자동재생
- 다음 영상
- 재생목록 연속 재생
- 시청 위치 동기화
- 반복 재생
- 키보드 단축키
- 공유·임베드
- 클립 생성
- 저작권·연령·지역 제한 안내
- 실제 재생 실패 대응
- QoE 측정: 시작 시간·버퍼링·오류·시청 지속시간

### 3.4 소셜 기능

- 좋아요·취소
- 구독·해지
- 알림 종
- 댓글 및 답글
- 댓글 좋아요
- 댓글 고정
- 댓글 정렬
- 댓글 신고
- 댓글 차단어
- 댓글 작성 제한
- 커뮤니티 게시물
- 투표
- 멘션
- 공유
- 크리에이터와 팬의 상호작용

### 3.5 Shorts

- 세로형 무한 피드
- 스와이프 전환
- 자동재생 및 정지
- 좋아요·댓글·공유·구독
- 음악·오디오 출처
- 해시태그·챌린지
- 관련 긴 영상 연결
- 광고 삽입
- 시청 이력
- 추천 최적화
- 모바일 촬영·업로드
- 권리가 허용된 음악에 대한 Remix 기능

Shorts 데이터와 일반 영상 데이터는 통합 검색 가능하되 형식별 최적화한다.

---

## ARTICLE 4. Creator Studio 전면 상용화

크리에이터 전용 서비스에 다음 기능을 제공한다.

### 콘텐츠 관리
- 영상 업로드·일괄 업로드
- 업로드 중단 및 이어올리기
- 제목·설명·태그·카테고리
- 썸네일 업로드 및 수정
- 공개·비공개·일부공개·회원전용
- 예약 공개
- 플레이리스트 관리
- 영상 교체 정책
- 영상 삭제·복원 정책
- 연령·지역 제한
- 댓글 정책
- 라이선스·저작권 상태

### 영상 편집
- 구간 자르기
- 블러
- 자막 편집
- 챕터 편집
- 엔드스크린
- 카드
- 음원 라이브러리
- 쇼츠 변환
- 클립 생성
- AI 기반 자동 자막 초안
- 번역·다국어 오디오 확장
- 썸네일 후보 및 A/B 테스트

### 분석
- 조회수·유효 조회수
- 시청 시간
- 평균 시청 지속시간
- 시청 유지율 그래프
- CTR
- 노출수
- 신규·재방문 시청자
- 구독 전환율
- 유입 경로
- 검색 유입
- 국가·기기·OS
- 영상별 성과
- Shorts 성과
- 라이브 성과
- 광고 수익
- 구독·후원 수익
- 지급 예정액
- 기간 비교 및 CSV 다운로드

### 채널 관리
- 채널 배너·로고·설명
- 채널 홈 구성
- 채널 관리자 초대
- 권한별 업무 분리
- 정책 위반 내역
- 신고·이의 신청
- 브랜드 협업
- MCN 계약 연결
- 수익화 신청
- 지급 계좌 및 세무 정보 관리

실제 발생하지 않은 수익이나 조회수를 생성해 표시하면 안 된다.

---

## ARTICLE 5. LIVE Streaming 플랫폼

라이브는 별도 서비스 수준으로 완성한다.

- 방송 생성
- 예약 라이브
- Stream Key 발급·회전
- RTMPS / 지원 가능한 SRT ingest
- OBS 연결
- 방송 상태 전환
- ABR 라이브 인코딩
- 저지연 HLS
- 라이브 상태 감시
- 동시 시청자
- 라이브 채팅
- 채팅 금칙어·도배 방지
- 관리자 채팅 모더레이션
- 느린 채팅 모드
- 슈퍼챗형 후원
- 멤버 전용 라이브
- 방송 중 광고
- 방송 종료 후 VOD 변환
- DVR 및 되감기
- 방송 분석
- 방송 장애 경고
- 스트림 키 유출 대응

외부 미디어 사업자를 사용하는 경우 실제 지원 프로토콜과 API를 먼저 검증한다. 지원되지 않는 기능을 명목상 구현하지 않는다.

---

## ARTICLE 6. Advertiser Portal 독립 구현

기존 Partner Portal은 유지하되 광고주를 위한 별도 화면과 권한 체계를 확립한다.

권장 경로는 `/advertiser`이며 독립 배포 여부는 운영 요구에 따라 결정한다.

### 광고주 등록

- 광고주 계정
- 회사·사업자 정보
- 광고주 조직
- 조직 멤버 초대
- 광고 계정별 권한
- 광고 정책 동의
- 결제 수단
- 예산 충전 또는 청구 계약
- 캠페인 승인 현황
- 광고 계정 제재 및 이의 신청

### 광고 제작

- 이미지 광고
- 영상 광고
- 홈 피드 광고
- Shorts 광고
- 검색 광고
- 영상 시작 전 광고
- 영상 중간 광고
- 영상 종료 광고
- 채널 페이지 광고
- 라이브 광고
- 브랜드 배너
- 자체 프로모션 광고

### 캠페인 설정

- 캠페인명
- 광고 목적
- 광고 형식
- 시작·종료 일시
- 광고 소재
- 클릭 목적지
- CTA
- 국가·언어·지역
- 콘텐츠 카테고리
- 기기·플랫폼
- 허용된 조건의 타기팅
- 제외 카테고리
- 일 예산·총예산
- CPM / CPC / CPV
- 빈도 제한
- 광고 노출 순위
- 목표 노출수
- 예약 인벤토리
- 광고 배치
- 광고 소재 심사

### 성과 관리

- 광고 요청
- Fill Rate
- Served Ads
- Verified Impressions
- Views / Quartile Completion
- Clicks
- CTR
- Spend
- CPM·CPC·CPV
- Conversions
- 예산 소진율
- 국가별 성과
- 기기별 성과
- 캠페인별 수익성
- 청구서
- 환불·크레딧
- 데이터 내보내기

광고주가 자신의 광고만 조회·수정할 수 있도록 조직별 접근 통제를 적용한다.

---

## ARTICLE 7. 가장 중요한 요구 — 광고 등록과 실제 노출의 완전한 연결

**이는 프로젝트의 최상위 P0 비즈니스 요구사항이다.**

광고주 또는 관리자가 광고를 등록했을 때, 사용자의 광고 슬롯에서 운영 DB에 저장된 해당 광고를 실제로 가져와 렌더링해야 한다.

단순히 광고주 화면에 등록 목록을 추가하는 것은 구현 완료가 아니다.

### 7.1 광고 상태 머신

`DRAFT → PENDING_REVIEW → APPROVED → SCHEDULED → ACTIVE → PAUSED / COMPLETED / EXHAUSTED`

별도로 `REJECTED`, `SUSPENDED`, `EXPIRED` 상태를 처리한다.

등록 즉시 승인 정책이 허용된 자체 광고와 심사가 필요한 외부 유료 광고를 구분한다.

### 7.2 광고 노출 흐름

1. 광고주가 캠페인과 실제 소재를 등록한다.
2. 광고 데이터가 PostgreSQL에 저장된다.
3. 소재 파일이 실제 스토리지에 업로드된다.
4. 포맷·목적지 URL·정책·바이러스·저작권 검사를 수행한다.
5. 관리자 승인 결과를 기록한다.
6. 적격 캠페인을 Active 상태로 만든다.
7. 광고 배치별 Ad Inventory에 등록한다.
8. 사용자 웹 또는 앱이 광고 슬롯을 요청한다.
9. Ad Decision API가 자격 있는 캠페인을 선택한다.
10. 실제 소재 URL과 광고 메타데이터를 반환한다.
11. 사용자 화면 또는 플레이어가 소재를 렌더링한다.
12. 클라이언트가 렌더링·노출·영상 재생 이벤트를 전송한다.
13. 서버가 유효 이벤트를 검증·중복 제거한다.
14. 실적·예산·광고주 대시보드를 갱신한다.
15. 실패 시 운영 관리자에게 구체적인 원인을 제공한다.

### 7.3 필수 규칙

- 승인된 광고만 일반 사용자에게 송출한다.
- 캠페인 상태·일정·예산·타기팅을 검증한다.
- 해당 광고 슬롯의 형식과 실제 소재가 호환되어야 한다.
- 광고주 자신의 캠페인이 아닌 타인의 데이터를 수정할 수 없다.
- 광고 소재가 없으면 가짜 광고를 만들어 내지 않는다.
- 광고를 플레이어가 받기만 한 것은 유효 노출이 아니다.
- 실제 정의된 노출 조건이 충족되어야 노출을 집계한다.
- 새로고침·재전송·중복 이벤트가 과금 중복으로 이어지면 안 된다.
- 광고비보다 많은 금액이 집행되어서는 안 된다.
- 광고 차단기·네트워크 장애·Premium·아동 보호·사용자 동의 상태를 존중한다.
- 미승인 또는 부적합 광고를 억지로 노출해서는 안 된다.

**광고 노출 보장의 의미를 명확히 구분한다.**

A. 관리자 자체 홍보 또는 예약형 광고: 지정 인벤토리와 적격 테스트 사용자 조건에서 반드시 광고가 표시되도록 할당·검증한다.

B. 일반 유료 캠페인: 승인, 기간, 예산, 타기팅, 재고와 정책이 충족되면 Ad Decision 대상에 포함하고 실제 집행·성과를 추적한다.

C. 특정 노출량 보장 계약: 판매 가능한 인벤토리 기반으로 예약하고 미달 예상 시 사전 경고·보상·조정 절차를 제공한다.

무조건 모든 방문자에게 모든 광고를 보여주는 방식은 요구하지 않는다. 그 대신 **적격 광고가 있는데도 사용자 슬롯에 나오지 않는 시스템 오류를 방치하지 않는 것**을 필수 조건으로 한다.

### 7.4 광고 미노출 진단 기능

관리자에 `Ad Delivery Inspector` 구현.

각 광고에 대해 다음을 실시간 조회할 수 있어야 한다.

- 캠페인 상태
- 심사 상태
- 소재 처리 상태
- 유효 기간
- 광고 슬롯 연결
- 타기팅 적격성
- 예산 잔액
- 빈도 제한
- 최근 광고 요청
- 최근 광고 선택
- 실제 노출 이벤트
- 노출되지 않은 이유
- 최근 실패 코드
- 관련 Request ID
- 수동 테스트 결과

`광고 미노출 사유를 알 수 없음`은 정상적인 운영 상태가 아니다.

### 7.5 광고 미리보기

광고주는 사용자 페이지와 같은 렌더러로 광고를 미리 볼 수 있어야 한다.

관리자에는 **검증용 강제 배정 모드**를 제공한다.

이 모드는 권한 있는 내부 테스트 세션과 프리뷰 환경에만 적용하고, 실제 과금 및 운영 통계와 분리한다.

테스트 결과에는 실제 소재 ID, 캠페인 ID, 슬롯 ID, 렌더링 여부, 이벤트 수신 여부를 남긴다.

### 7.6 광고 자동 검증

광고 시스템의 배포 및 운영 환경에 다음 E2E 시나리오를 만든다.

`광고 등록 → 소재 업로드 → 심사 승인 → 예산 활성화 → 적격 사용자 접속 → 광고 결정 → 실제 화면 렌더링 → 유효 노출 기록 → 광고주 대시보드 집계`

이 전 과정이 테스트에서 입증되지 않으면 광고 기능을 완료 처리하지 않는다.

---

## ARTICLE 8. Partner / MCN / Rights Management

Partner 서비스는 광고주 서비스와 역할을 구분한다.

필수 기능:

- MCN 회사 등록
- 소속 크리에이터 초대
- 계약 생성·전자서명 연계
- 채널 소유권 확인
- 계약 기간
- 수익 배분율
- 채널별 성과 분석
- 공동 수익 정산
- 월간 지급 명세
- 파트너 관리자 권한
- 콘텐츠 라이선스
- 권리 소유자 등록
- 참조 콘텐츠 등록
- 영상 지문 생성·매칭
- 저작권 침해 의심 탐지
- 수동 저작권 클레임
- 이의 신청 및 재심
- 국가별 사용 권리
- 라이선스 기간
- 권리 침해 처리 이력
- 법률 검토 기록

권리 침해 판정은 자동 탐지만으로 확정하지 않으며, 적절한 심사와 이의 신청 절차를 둔다.

---

## ARTICLE 9. Admin Console 완전한 운영 시스템

기존 관리자 디자인을 유지하되 전 영역을 실제 DB·API에 연결한다.

관리자 메뉴를 다음 도메인으로 확장한다.

### 운영 관제
- DAU/MAU
- 동시 사용자
- 재생 성공률
- 업로드 성공률
- 라이브 상태
- 광고 송출 상태
- API 및 DB 상태
- 미디어 Worker 상태
- 장애·경고
- CDN 트래픽·비용
- 총 매출·정산 부채

### 회원·크리에이터
- 회원 검색
- 계정 상태
- 채널 상태
- 제재
- 제재 해제
- 신고
- 수익화 신청
- 본인·사업자 확인
- 조직별 접근 관리

### 콘텐츠 운영
- 영상 검색
- 공개 상태
- 콘텐츠 심사
- 신고 큐
- 저작권
- 연령 제한
- 아동 보호
- 지역 차단
- 영상 재처리
- 삭제·복구

### 광고 통합 운영
- 전체 광고주
- 전체 캠페인
- 광고 소재 심사
- 광고 슬롯 관리
- 배치 우선순위
- 예약 광고
- 집행 상태
- 실제 노출 로그
- 미노출 진단
- 광고비 및 환불
- 사기성 광고 탐지
- 광고 성과 분석
- 자체 프로모션
- 광고 운영 킬스위치

### 수익·결제·정산
- 결제 내역
- 환불
- 원장
- 광고 매출
- Premium 매출
- 후원 매출
- 크리에이터 지급
- 파트너 지급
- 수수료
- 세금
- 미지급금
- 월 결산

### 시스템 관리
- 권한 관리
- 기능 플래그
- 국가별 정책
- 언어
- 카테고리
- SEO 설정
- 추천 설정
- 광고 설정
- API 키 관리
- Webhook 관리
- 스토리지 관리
- 보안 설정
- 감사 로그
- 운영 변경 이력
- 배포 버전·롤백

중대한 변경에는 승인 워크플로와 감사 로그가 필요하다.

관리자 화면에서 역할을 선택했다고 실제 권한이 변경되면 안 된다. 서버가 발급하고 검증한 권한만 사용한다.

---

## ARTICLE 10. PostgreSQL 중심 DB 아키텍처

기존 67개 테이블 정의를 버리지 않는다.

먼저 기존 테이블, 키, FK, 제약조건, 인덱스, 참조 관계를 비교하고 누락된 데이터 도메인을 확장한다.

### A. Identity / Organization

기존:
`users`, `user_profiles`, `user_preferences`, `sessions`, `devices`, `roles`, `permissions`, `role_permissions`, `organizations`, `organization_members`

추가 검토:
`account_security_events`, `user_consents`, `parental_control_profiles`, `user_blocks`, `login_audit_events`, `creator_verifications`, `tax_profiles`

### B. Content / Video

기존:
`channels`, `channel_members`, `videos`, `video_assets`, `video_variants`, `upload_sessions`, `upload_parts`, `media_jobs`, `thumbnails`, `captions`, `chapters`, `playlists`, `playlist_items`, `visibility_rules`

추가:
`video_edits`, `video_translations`, `video_audio_tracks`, `video_end_screens`, `video_cards`, `video_clips`, `video_processing_attempts`, `media_asset_checksums`, `content_availability_rules`

### C. Engagement

기존:
`subscriptions`, `video_reactions`, `comments`, `comment_reactions`, `watch_history`, `watch_progress`, `notifications`, `community_posts`, `community_votes`

추가:
`user_follows`, `channel_blocks`, `user_feedback`, `notification_preferences`, `watch_sessions`, `shares`, `comment_moderation_actions`, `search_feedback`

### D. Live

기존:
`live_streams`, `live_inputs`, `live_chat_messages`, `live_recordings`, `live_moderators`

추가:
`live_health_events`, `live_viewer_sessions`, `live_chat_sanctions`, `live_schedule_events`, `live_ad_breaks`

### E. Advertising — 중점 확장

기존:
`advertisers`, `ad_campaigns`, `ad_creatives`, `ad_placements`, `ad_events`, `ad_budget_allocations`

추가 권장:
`ad_accounts`, `ad_account_members`, `ad_slots`, `ad_flights`, `ad_targeting_rules`, `ad_creative_assets`, `ad_creative_reviews`, `ad_delivery_decisions`, `ad_delivery_events`, `ad_budget_reservations`, `ad_frequency_caps`, `ad_inventory_forecasts`, `ad_reporting_daily`, `ad_conversions`, `ad_adjustments`, `ad_contracts`

중요 관계:

- Advertiser 1:N Ad Accounts
- Ad Account 1:N Campaigns
- Campaign 1:N Creatives
- Campaign 1:N Flights
- Campaign 1:N Targeting Rules
- Ad Slot 1:N Delivery Decisions
- Delivery Decision 1:N Delivery Events
- Campaign 1:N Budget Reservations
- Campaign 1:N Daily Reports

### F. Finance

기존:
`monetization_accounts`, `memberships`, `subscriptions_billing`, `payment_transactions`, `refunds`, `ledger_accounts`, `ledger_entries`, `creator_earnings`, `payouts`, `invoices`

추가 검토:
`payment_webhook_receipts`, `payout_methods`, `payout_batches`, `revenue_share_contracts`, `settlement_adjustments`, `tax_documents`, `billing_disputes`

KRW, USD 및 다국가 통화에서는 통화별 최소 단위 또는 명시적 Decimal 정책을 일관되게 사용한다.

KRW를 무조건 cents로 가정하지 않는다.

### G. Copyright / Moderation

기존:
`rights_owners`, `rights_assets`, `rights_territories`, `copyright_claims`, `copyright_disputes`, `dispute_evidence`, `reports`, `moderation_cases`

추가:
`fingerprint_matches`, `content_policy_versions`, `moderation_decisions`, `appeals`, `repeat_infringement_events`, `legal_hold_records`

### H. Operations

기존:
`admin_actions`, `audit_logs`, `feature_flags`, `outbox_events`, `idempotency_keys`

추가:
`system_config_versions`, `background_job_runs`, `incident_records`, `outbox_delivery_attempts`, `webhook_endpoints`, `webhook_deliveries`, `service_usage_costs`

### DB 공통 원칙

- PK / FK / UNIQUE / CHECK 완비
- 적절한 `NOT NULL`
- UTC 시간 저장
- 명확한 통화 단위
- 조직 단위 데이터 격리
- 중요한 수정의 감사 이력
- PII 접근 제한
- 개인정보 보관·삭제 정책
- 외래키 인덱스 검토
- 대용량 이벤트 파티셔닝
- 결제 멱등성
- 트랜잭션 원자성
- Zero Downtime Migration
- PITR Backup 및 복구 훈련
- 읽기 복제본·캐시·검색 인덱스 확장

기존 운영 데이터에 영향이 있는 파괴적 SQL은 사용자 승인 없이 실행하지 않는다.

---

## ARTICLE 11. API 계약의 전면 정리

`openapi/gvp-v1.yaml`을 실제 구현된 전체 API와 동기화한다.

라우트마다 다음을 명세한다.

- HTTP Method
- 인증 요구
- 권한
- Path / Query / Body
- Request Schema
- Response Schema
- Error Schema
- Pagination
- Rate Limit
- Idempotency
- Versioning

추가 또는 개선해야 할 광고 API 예시:

- `POST /v1/advertiser/campaigns`
- `GET /v1/advertiser/campaigns`
- `PATCH /v1/advertiser/campaigns/{id}`
- `POST /v1/advertiser/campaigns/{id}/submit`
- `POST /v1/admin/ads/creatives/{id}/approve`
- `POST /v1/admin/ads/creatives/{id}/reject`
- `GET /v1/admin/ads/delivery-diagnostics`
- `POST /v1/ads/decision`
- `POST /v1/ads/events`
- `POST /v1/ads/preview`
- `GET /v1/advertiser/reports`
- `GET /v1/admin/ads/inventory`
- `GET /v1/admin/ads/attribution`

모든 관리자 API와 광고주 API에서 실제 서버 측 권한 및 소유권 검사를 수행한다.

DTO의 camelCase/snake_case 변환과 API Client를 자동화하거나 명확히 표준화한다.

---

## ARTICLE 12. 미디어 처리 및 플랫폼 인프라

권장 흐름:

`Client Upload → Signed Multipart Upload → Object Storage → Media Job Queue → Validator/Transcoder → HLS/DASH Asset → CDN → Playback API → Player`

- 파일 무결성 확인
- MIME 실검사
- 업로드 크기 제한
- 악성 파일 검사
- 메타데이터 추출
- 지원 포맷 검증
- 영상·오디오 변환
- 다양한 해상도
- 자막
- 썸네일
- 에러 재시도
- 큐 장애 복구
- CDN 접근제어
- 권한별 Signed URL
- 저작권 삭제 반영
- 원본·변환물 보관 정책

PostgreSQL은 운영 트랜잭션의 기준 데이터베이스로 사용한다.

Valkey는 캐시와 단기 데이터, NATS는 이벤트 메시지, OpenSearch는 검색, ClickHouse는 대용량 분석 이벤트에 사용한다.

모든 제품에 모든 인프라를 무조건 도입하지 말고 실제 규모에 따른 비용·운영 복잡성을 검토한다.

서버리스 및 관리형 서비스와 자체 인코딩 워커의 단가를 비교하고 영상 분당 처리 비용, 재생 GB당 비용, 저장 GB당 비용을 측정한다.

---

## ARTICLE 13. UI/UX 디자인 보존 및 고도화

기존 디자인을 대규모 재설계하지 않는다.

다음 요소를 유지한다.

- 다크 모드 중심 정체성
- 영상 중심 카드 레이아웃
- StreamVerse 브랜드 콘셉트
- 메인 내비게이션
- Shorts 전용 경험
- Creator Studio의 프로페셔널 대시보드
- Partner 및 Admin의 구분된 콘솔

다만 다음을 전면 개선한다.

- 색상·타이포그래피 토큰 통합
- 본문·표·버튼 글씨 가독성
- 카드 밀도
- 모바일 터치 타깃
- 반응형 내비게이션
- 영상 카드 썸네일 비율
- 광고 콘텐츠 구분
- 검색·자동완성 UX
- 로딩·빈 화면·에러 화면
- 다이얼로그·토스트 피드백
- 키보드 접근성
- 스크린리더
- 레이아웃 이동 최소화
- 동영상 플레이어 조작
- 관리자 테이블·필터·일괄 작업

검사 해상도:

360 / 390 / 430 / 768 / 1024 / 1280 / 1440 / 1920px

모바일 앱은 실제 Android/iOS 기기 또는 에뮬레이터에서 검증한다.

WCAG 2.2 AA 수준을 목표로 접근성을 검증한다.

실제 렌더링 스크린샷과 수정 전후 비교를 제출한다.

---

## ARTICLE 14. 검색·추천·발견성·SEO

다음 기능을 추가한다.

- 동영상/채널/Shorts/Live 통합 검색
- 자막 기반 검색
- 카테고리별 색인
- 다국어 색인
- 자동완성
- 오타 교정
- 최신순·인기순
- 지역별 트렌드
- 중복 콘텐츠 억제
- 시청자 피드백 반영
- 개인화 추천
- 탐색 다양성
- 신생 채널 노출
- 콘텐츠 안전 필터
- 광고와 자연 검색 결과 분리 표시
- Sitemap
- VideoObject 구조화 데이터
- 동영상별 OG 메타데이터
- Canonical
- 공개 콘텐츠만 검색엔진 색인
- 비공개·삭제 콘텐츠 색인 제거

추천·광고 랭킹은 별도 모델과 목적함수로 관리한다.

광고를 일반 콘텐츠로 위장하지 않는다.

---

## ARTICLE 15. 수익화 및 글로벌 사업 확장

지원할 수익화 모델:

- 영상 광고
- Shorts 광고
- 라이브 광고
- Premium 구독
- 채널 멤버십
- 크리에이터 후원
- 라이브 유료 메시지
- 브랜드 캠페인
- MCN 계약 정산
- 합법적인 제휴 쇼핑 링크
- 라이선스 콘텐츠
- 광고주 대상 예약 광고 상품

향후 확장:

- 다국어 AI 자막·더빙
- AI 클립 추천
- 자동 하이라이트
- AI 기반 정책 심사 보조
- 크리에이터 브랜드 마켓플레이스
- TV 및 CTV 앱
- 미디어 권리 거래
- 외부 파트너용 API
- 영상 기반 커머스
- 라이브 쇼핑
- 크리에이터 공동 제작
- 글로벌 광고 성과 API

모든 유료 기능은 제공 국가별 결제·세금·환불·디지털 콘텐츠 정책을 확인한 뒤 출시한다.

---

## ARTICLE 16. 보안·개인정보·저작권·법규

필수 점검:

- OWASP Web/API 보안
- JWT/OIDC 검증
- 조직별 권한 격리
- CSRF/XSS/SSRF 예방
- 파일 업로드 보호
- Webhook 서명
- 결제 중복 방지
- 개인정보 최소 수집
- 저장 및 전송 암호화
- 민감 데이터 접근 감사
- 관리자 MFA
- 민감 작업 재인증
- 봇·스팸·광고 사기 대응
- 저작권 신고·삭제·이의 신청
- 불법 콘텐츠 신고
- 미성년자 보호
- 광고 표기
- 개인정보 및 맞춤형 광고 동의
- 국가별 개인정보·소비자 보호 법규
- 데이터 삭제 및 보관 기한

외부 사업자의 저작권 보호 기능을 근거 없이 구현 완료라고 선언해서는 안 된다.

---

## ARTICLE 17. Release Quality Gates

최소 다음 게이트를 만족해야 한다.

**Gate 1 — 정적 검사**

- Lint / Typecheck / Build 통과
- 마이그레이션 검증
- API 계약 검사
- 의존성 보안 검사
- Secret Scan
- 코드 감사

**Gate 2 — 기능 검사**

- 실제 회원 생성·로그인·로그아웃
- 실제 영상 업로드
- 실제 인코딩
- 실제 영상 재생
- 조회 및 시청 이력 영구 저장
- 댓글·구독 영구 저장
- 라이브 ingest 및 재생
- 광고 등록·심사·송출
- 광고 노출·클릭 집계
- 결제 성공·실패·환불 테스트
- 정산 원장 검증
- 관리자 권한 격리

**Gate 3 — PC/Mobile E2E**

- Chromium / Firefox / WebKit
- 여러 화면 크기
- Android / iOS
- 플레이어 동작
- 검색·자동완성
- Shorts 스와이프
- 로그인 상태 전환
- 광고 슬롯 렌더링
- 접근성
- 오류 처리

**Gate 4 — 보안**

- 미검증 토큰은 401
- 타 사용자 데이터 접근은 차단
- 조직 간 캠페인 격리
- 미승인 광고 송출 금지
- 위조 결제 Webhook 차단
- 재전송 이벤트 이중 과금 금지

**Gate 5 — 운영**

- 장애 복구
- 백업 복구
- 실제 부하 테스트
- 서비스 SLO
- CDN/미디어 비용 예측
- 에러 추적
- 광고 송출 알림
- Rollback 검증

성능·보안·결제·미디어·광고 테스트는 실제 통합 시스템 또는 격리된 Staging 환경에서 수행한다.

테스트 코드 안에서 결과를 흉내 낸 경우 Production 검증으로 인정하지 않는다.

---

## ARTICLE 18. 단계별 실행 우선순위

### PHASE 0 — 기준선 및 전수 감사
모든 파일·라우트·권한·데이터·서비스·테스트 인벤토리 작성.

### PHASE 1 — P0 복구
인증, 영구 DB, 업로드, 미디어 처리, 결제, 광고 연결, 보안 취약점 정상화.

### PHASE 2 — 상용 핵심 기능
사용자 웹, 모바일, Creator Studio, Advertiser, Partner, Admin 주요 기능 완성.

### PHASE 3 — 광고 수익화
Ad Server, 인벤토리, 타기팅, 예산 제어, 유효 노출, 광고주 통계, 청구·정산.

### PHASE 4 — 서비스 고도화
라이브, Shorts, 추천, 다국어, 저작권, 수익화, SEO, 접근성.

### PHASE 5 — 검증
기능·보안·실사용자 E2E·성능·복원력·광고 송출 검사.

### PHASE 6 — 배포 준비
운영 DB 백업, 마이그레이션 검증, GitLab CI/CD, Staging 배포, Canary/Rollback 계획.

### PHASE 7 — 정식 배포
운영 접근 권한과 PO 배포 승인이 있는 경우에 한해 Production에 배포한다.

배포 이후 실제 URL 및 운영 환경에서 재검증한다.

각 Phase별 완료 결과는 코드·로그·API 응답·테스트 결과로 증명한다.

---

## ARTICLE 19. GitLab / Commit / Deployment

- 기존 Git 커밋 이력을 보존한다.
- 기준선 SHA를 기록한다.
- 원격 GitLab 저장소와 작업 대상 브랜치를 검증한다.
- 변경 파일을 전수 확인한다.
- 비밀키·고객 데이터·결제 정보는 Git에 포함하지 않는다.
- 테스트 실패 상태에서 완료 선언을 금지한다.
- 최종 검증 후 관련 변경을 하나의 명확한 최종 통합 커밋으로 정리한다.
- GitLab CI/CD 성공 여부를 확인한다.
- DB 백업 및 롤백 지점을 보존한다.
- 배포 권한과 승인이 있는 경우에만 실서버 배포한다.
- 배포 성공 후 실제 운영 서비스에 변경이 반영되었는지 재확인한다.

단순히 Git Push나 Deploy 성공 메시지를 확인하는 것으로 완료 처리하지 않는다.

---

## ARTICLE 20. 최종 인수인계 산출물

다음을 반드시 제출한다.

1. 전체 코드베이스 감사 보고서
2. P0/P1/P2 결함 목록과 수정 결과
3. 화면별 PC/Mobile 디자인 감사
4. 사용자·Studio·Advertiser·Partner·Admin 기능 매트릭스
5. API Inventory 및 OpenAPI 스펙
6. 실제 DB ERD 및 마이그레이션
7. 광고 등록 → 사용자 노출 → 통계·과금 검증 보고서
8. 영상 업로드 → 인코딩 → 실제 재생 증거
9. 결제·정산 검증 기록
10. 보안 및 권한 격리 테스트
11. 성능·부하·장애 복구 결과
12. 미구현·보류·외부 의존 기능 목록
13. 운영 비용 추정과 인프라 용량 계획
14. GitLab Commit SHA
15. 배포 환경과 실제 배포 상태
16. 롤백 및 장애 대응 매뉴얼

**최종 완료 선언 조건:**

모든 기능은 `코드 존재`, `빌드 성공`, `기능 테스트 통과`, `실제 통합 검증`, `운영 검증`을 별도로 표시한다.

미완료 항목은 숨기지 않는다.

확인되지 않은 기능은 `UNVERIFIED`로 표시한다.

가짜 광고 노출, 가짜 조회수, 가짜 결제, 가짜 업로드 완료, 가짜 관리자 지표를 근거로 상용화 완료라고 선언하면 안 된다.

**최종 목표는 YouTube를 흉내 낸 데모가 아니라 실제 콘텐츠가 업로드되고, 실제 시청자가 재생하며, 실제 광고주가 광고를 집행하고, 실제 크리에이터에게 수익이 정산되는 독립 상용 동영상 플랫폼이다.**

`END OF MASTER DIRECTIVE — BN-GVP-YOUTUBE-CLASS-PRODUCTION-001`
</USER_REQUEST>
<ADDITIONAL_METADATA>
The current local time is: 2026-10-26T16:53:45+09:00.

The user's current state is as follows:
Active Document: c:\Users\dnjst\Desktop\project\video\.github\workflows\cloudflare-deploy.yml (LANGUAGE_UNSPECIFIED)
Cursor is on line: 1
Other open documents:
- c:\Users\dnjst\Desktop\project\video\.github\workflows\cd.yml (LANGUAGE_UNSPECIFIED)
- c:\Users\dnjst\Desktop\project\video\.github\workflows\ci.yml (LANGUAGE_UNSPECIFIED)
- c:\Users\dnjst\Desktop\project\ott\clients\web-admin\tsconfig.json (LANGUAGE_UNSPECIFIED)
- c:\Users\dnjst\Desktop\project\ott\clients\web-user\tsconfig.json (LANGUAGE_UNSPECIFIED)
- c:\Users\dnjst\Desktop\project\video\infra\docker\Dockerfile.nuxt (LANGUAGE_UNSPECIFIED)
</ADDITIONAL_METADATA>
<USER_SETTINGS_CHANGE>
The user changed setting `Model Selection` from Gemini 3.8 Flash (High) to Claude Sonnet 5.5 (High). No need to comment on this change if the user doesn't ask about it. If reporting what model you are, please use a human readable name instead of the exact string.
</USER_SETTINGS_CHANGE>