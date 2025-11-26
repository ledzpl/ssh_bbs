# SSH 텍스트 BBS - 구현 명세서

## 개요

이것은 SSH를 통해 접근 가능한 Go 기반 텍스트 게시판 시스템(BBS)입니다. 사용자는 SSH로 접속하여 게시판을 탐색하고, 게시글을 작성하고, 댓글을 추가하고, 콘텐츠를 관리할 수 있습니다. 이 시스템은 터미널을 사용자 인터페이스로 사용하며, 대화형 TUI(Terminal User Interface)를 위해 Bubbletea 프레임워크를 활용합니다.

## 기술 스택

### 핵심 의존성
- **언어**: Go 1.25.1
- **SSH 서버**: `github.com/charmbracelet/ssh` 및 `github.com/charmbracelet/wish`
- **TUI 프레임워크**: `github.com/charmbracelet/bubbletea` (터미널 UI를 위한 Elm 아키텍처)
- **UI 컴포넌트**: 
  - `github.com/charmbracelet/bubbles` (미리 만들어진 컴포넌트: textarea, textinput, viewport)
  - `github.com/charmbracelet/lipgloss` (스타일링)
  - `github.com/charmbracelet/glamour` (마크다운 렌더링)
- **암호화**: 표준 라이브러리 `crypto/aes`, `crypto/cipher`

### 프로젝트 구조

```
ag/
├── cmd/
│   └── bbs/
│       └── main.go              # 애플리케이션 진입점
├── internal/
│   ├── auth/
│   │   ├── auth.go              # 인증 로직
│   │   └── auth_test.go
│   ├── bbs/
│   │   ├── bbs.go               # 핵심 BBS 데이터 구조 및 로직
│   │   ├── comments.go          # 댓글 관리
│   │   ├── persist.go           # 게시판 목록 영속성
│   │   ├── persist_posts.go     # 게시글 영속성 (암호화 포함)
│   │   └── *_test.go            # 테스트
│   ├── server/
│   │   └── server.go            # SSH 서버 설정
│   └── ui/
│       ├── model.go             # Bubbletea 모델 정의
│       ├── update.go            # 업데이트 로직 (이벤트 처리)
│       ├── update_comments.go   # 댓글 뷰 업데이트 로직
│       ├── view.go              # 뷰 렌더링 로직
│       └── styles.go            # Lipgloss 스타일
├── data/
│   ├── boards.json              # 영구 게시판 목록
│   └── posts/
│       └── *.json               # 게시판별 게시글 파일 (버전 관리됨)
└── .ssh/
    └── term_info_ed25519        # SSH 호스트 키 (자동 생성)
```

## 아키텍처

### 1. 진입점 (`cmd/bbs/main.go`)

**책임**:
- 커맨드라인 플래그 파싱:
  - `-addr`: SSH 리슨 주소 (기본값: `:2323`)
  - `-boards`: 게시판 목록 JSON 경로 (기본값: `data/boards.json`)
  - `-posts`: 게시글 저장 디렉토리 (기본값: `data/posts`)
  - `-auth`: 인증 설정 JSON 경로 (선택사항)
- 인증 설정이 제공된 경우 로드
- 게시판 및 게시글 저장소 초기화
- 환경 변수 `BBS_ENCRYPTION_KEY`에서 암호화 키 읽기
- SSH 서버 생성 및 시작
- SIGINT/SIGTERM 시 graceful shutdown 처리

**주요 구현 세부사항**:
- CLI 파싱에 `flag` 패키지 사용
- 암호화 키는 AES-256을 위해 32바이트(64자 16진수)여야 함
- 암호화는 역호환 가능 (평문 파일 읽기 가능)
- 게시판 파일이 없는 경우 기본 게시판: `["general", "tech"]`

### 2. BBS 핵심 (`internal/bbs/`)

#### 데이터 구조

**Post** (`bbs.go`):
```go
type Post struct {
    ID        int
    Title     string
    Content   string
    Author    string
    CreatedAt time.Time
    Comments  []Comment
}
```

**Comment** (`bbs.go`):
```go
type Comment struct {
    ID        int
    PostID    int
    ParentID  int       // 최상위 댓글의 경우 0
    Author    string
    Content   string
    CreatedAt time.Time
}
```

**Board** (`bbs.go`):
```go
type Board struct {
    Name   string
    mu     sync.RWMutex
    posts  []Post
    nextID int
}
```

**BBS** (`bbs.go`):
```go
type BBS struct {
    mu     sync.RWMutex
    boards map[string]*Board
    order  []string              // 게시판 표시 순서
    now    func() time.Time     // 테스트를 위한 주입 가능한 시간
    store  BoardListStore
    posts  PostStore
}
```

#### 핵심 작업

**게시판 관리**:
- `NewWithBoards(now func() time.Time, names []string, store BoardListStore, posts PostStore) *BBS`
- `ListBoards() []BoardSummary` - 게시글 수와 함께 게시판 반환
- `ensureBoard(name string) (*Board, error)` - 존재하지 않으면 게시판 생성
- 게시판은 첫 번째 게시글이 추가될 때 암시적으로 생성됨
- 게시판 이름은 정규화됨 (공백 제거, 중복 제거)

**게시글 작업**:
- `AddPost(boardName, author, title, content string) (Post, error)`
  - 제목은 필수 (비어있으면 `ErrEmptyTitle` 반환)
  - 작성자가 비어있으면 "anonymous"로 기본 설정
  - 게시판별로 게시글 ID 자동 증가
  - 성공적으로 추가된 후 디스크에 저장
- `ListPosts(boardName string) ([]Post, error)` - 게시판의 모든 게시글 반환
- `GetPost(boardName string, id int) (Post, error)` - 단일 게시글 가져오기
- `DeletePost(boardName string, postID int, author string) error`
  - 작성자만 자신의 게시글 삭제 가능
  - 작성자가 일치하지 않으면 에러 반환
  - 삭제 후 디스크 저장소 업데이트

**댓글 작업** (`comments.go`):
- `AddComment(boardName string, postID int, author, content string, parentID int) (*Comment, error)`
  - ParentID = 0은 최상위 댓글
  - 게시글별로 댓글 ID 자동 증가
  - 댓글은 Post 구조체 내에 저장됨
  - 댓글 추가 후 전체 게시글 목록 저장
- `ListComments(boardName string, postID int) ([]Comment, error)`

**동시성**:
- BBS 레벨과 Board 레벨에서 `sync.RWMutex` 사용
- 수정 시 쓰기 락 (추가, 삭제)
- 조회 시 읽기 락 (목록, 가져오기)

**에러 처리**:
- 미리 정의된 에러: `ErrBoardNotFound`, `ErrPostNotFound`, `ErrEmptyTitle`
- 컨텍스트와 함께 영속성 에러 래핑

### 3. 영속성 레이어 (`internal/bbs/persist*.go`)

#### 게시판 목록 저장소 (`persist.go`)

**BoardFile** 구현:
```go
type BoardFile struct {
    Path string
}
```

**파일 형식** (`data/boards.json`):
```json
{
  "boards": ["general", "tech"]
}
```

**작업**:
- `Load() ([]string, error)` - JSON에서 게시판 이름 로드
- `Save(names []string) error` - 임시 파일 + 이름 변경을 사용한 원자적 쓰기
- 부모 디렉토리 자동 생성
- 파일이 없으면 빈 목록 반환 (에러 아님)

#### 게시글 저장소 (`persist_posts.go`)

**PostFile** 구현:
```go
type PostFile struct {
    Dir           string
    EncryptionKey []byte  // AES-256을 위한 32바이트
}
```

**파일 형식** (`data/posts/<board>.json`):
```json
{
  "version": 1,
  "board": "general",
  "posts": [
    {
      "ID": 1,
      "Title": "환영합니다",
      "Content": "안녕하세요",
      "Author": "admin",
      "CreatedAt": "2025-11-26T04:00:00Z",
      "comments": []
    }
  ]
}
```

**버전 관리 래퍼**:
- `version` 필드로 향후 스키마 진화 가능
- 현재 버전: 1
- 게시글은 래퍼 내 배열로 저장

**암호화** (선택사항):
- `EncryptionKey`가 제공되면 AES-256-GCM 사용
- Nonce가 암호문 앞에 추가됨
- `encrypt(plaintext []byte) ([]byte, error)`:
  - 키로 AES 암호 생성
  - GCM 모드 생성
  - 랜덤 nonce 생성
  - 반환: nonce || 암호문
- `decrypt(ciphertext []byte) ([]byte, error)`:
  - 첫 바이트에서 nonce 추출
  - 나머지 바이트 복호화
  - 복호화 실패 시 평문으로 폴백 (마이그레이션 지원)

**작업**:
- `Load(board string) ([]Post, error)` - 게시판의 게시글 로드
- `Save(board string, posts []Post) error` - 원자적 쓰기
- 게시글 디렉토리 자동 생성
- 원자적 쓰기를 위해 임시 파일 + 이름 변경 사용

### 4. 인증 (`internal/auth/`)

**설정 형식** (`data/auth.json`):
```json
{
  "users": [
    {
      "username": "alice",
      "password": "secret"
    }
  ]
}
```

**Authenticator** 구조:
```go
type Authenticator struct {
    users map[string]string  // username -> password
}
```

**주요 기능**:
- 선택사항: 인증 파일이 지정되지 않으면 인증 비활성화
- `Enabled() bool` - 사용자가 설정되어 있으면 true 반환
- `Authenticate(r *bufio.Reader, w io.Writer, initialUser string, maxAttempts int) (string, bool)`
  - 대화형 사용자명/비밀번호 프롬프트
  - SSH 사용자명을 초기값으로 사용
  - 에코 없는 비밀번호 입력
  - 최대 시도 횟수 설정 가능
  - 사용자명과 성공 여부 반환

**참고**: 현재 구현은 SSH 미들웨어 레벨에서 인증을 통합하지 않음; 연결은 허용되고 사용자명은 SSH 세션에서 가져옴.

### 5. SSH 서버 (`internal/server/`)

**설정**:
```go
func New(addr string, hostKeyPath string, board *bbs.BBS) (*ssh.Server, error)
```

**미들웨어 스택** (Wish 프레임워크):
1. **Bubbletea 미들웨어** - TUI 애플리케이션 실행
2. **Activeterm 미들웨어** - 활성 터미널 확인
3. **Logging 미들웨어** - 요청 로깅

**세션 핸들러**:
- 활성 PTY(가상 터미널) 확인
- SSH 세션에서 사용자명 추출 (기본값 "guest")
- 세션별로 새로운 UI 모델 생성
- alt 스크린과 마우스 지원으로 Bubbletea 설정

### 6. 사용자 인터페이스 (`internal/ui/`)

#### 아키텍처: Elm/Bubbletea 패턴

UI는 Model-View-Update (MVU) 패턴을 따릅니다:
- **Model**: 애플리케이션 상태
- **View**: 상태를 문자열로 렌더링
- **Update**: 이벤트를 처리하고 상태 업데이트

#### 세션 상태

```go
type sessionState int

const (
    viewBoards    // 게시판 목록 뷰
    viewPosts     // 게시판의 게시글 목록
    viewPost      // 단일 게시글 상세 뷰
    viewCompose   // 작성/편집 뷰
    viewComments  // 댓글 뷰
)
```

#### 모델 구조 (`model.go`)

```go
type Model struct {
    // 핵심
    board    *bbs.BBS
    username string
    width, height int
    
    // 상태
    state       sessionState
    boards      []bbs.BoardSummary
    posts       []bbs.Post
    activeBoard string
    activePost  bbs.Post
    
    // 네비게이션
    boardIdx, postIdx int
    page, postsPerPage int
    
    // 컴포넌트
    viewport  viewport.Model    // 스크롤 가능한 콘텐츠용
    textInput textinput.Model   // 제목 입력용
    textarea  textarea.Model    // 내용 입력용
    searchInput textinput.Model // 검색용
    
    // 플래그
    composing   bool
    searchMode  bool
    commentMode bool  // 댓글 작성 vs 게시글 작성
    
    // 검색
    searchQuery string
    
    // 댓글
    comments   []bbs.Comment
    commentIdx int
    
    err error
}
```

#### 업데이트 로직 (`update.go`, `update_comments.go`)

**전역 키바인딩** (작성 중이 아닐 때):
- `q` - 애플리케이션 종료
- `b` 또는 `←` - 이전 뷰로 돌아가기
- `w` - 새 게시글 작성 (게시판 또는 게시글 뷰에서)

**게시판 뷰** (`viewBoards`):
- `↑`/`k` - 이전 게시판
- `↓`/`j` - 다음 게시판
- `→`/`Enter` - 게시판 열기 (게시글 뷰로 이동)

**게시글 뷰** (`viewPosts`):
- `↑`/`k` - 이전 게시글
- `↓`/`j` - 다음 게시글
- `→`/`l`/`Enter` - 게시글 상세 보기
- `w` - 새 게시글 작성
- `/` - 검색 모드 진입
- `n` - 다음 페이지
- `p` - 이전 페이지

**게시글 상세 뷰** (`viewPost`):
- `Esc`/`b`/`←` - 게시글 목록으로 돌아가기
- `r` - 답글 (댓글 추가)
- `c` - 댓글 보기
- `d` - 게시글 삭제 (작성자만)
- 마우스 또는 화살표 키로 스크롤 (viewport)

**작성 뷰** (`viewCompose`):
- `Tab` - 제목과 내용 필드 전환
- `Ctrl+S` - 게시글/댓글 제출
- `Esc` - 취소하고 돌아가기

**댓글 뷰** (`viewComments`):
- `Esc`/`b` - 게시글 뷰로 돌아가기
- `↑`/`k` - 이전 댓글
- `↓`/`j` - 다음 댓글
- `r` - 답글 (댓글 추가)

**검색**:
- 제목 또는 내용으로 게시글 필터링 (대소문자 무시)
- 사용자가 입력할 때 실시간 필터링
- `Esc`로 검색 모드 종료
- `Enter`로 검색 확인

#### 뷰 렌더링 (`view.go`)

**스타일링** (`styles.go`):
- 색상, 테두리, 패딩을 위해 Lipgloss 사용
- 스타일: 제목, 부제목, 선택된 항목, 도움말 텍스트, 에러
- 터미널 크기에 따른 적응형 스타일링

**뷰 컴포넌트**:
1. **헤더** - 현재 위치와 사용자명 표시
2. **콘텐츠 영역** - 메인 콘텐츠 (게시판 목록, 게시글, 게시글 상세)
3. **푸터** - 사용 가능한 키바인딩 도움말
4. **에러 표시** - 에러가 있을 경우 표시

**게시판 목록**:
- 게시판 이름과 게시글 수 표시
- 선택된 게시판 강조
- 네비게이션 힌트 표시

**게시글 목록**:
- 페이지네이션 (페이지당 10개 게시글)
- 표시: 게시글 ID, 제목, 작성자, 타임스탬프
- 선택된 게시글 강조
- 검색 모드일 때 검색 바
- 페이지 표시기

**게시글 상세**:
- 제목과 메타데이터 (작성자, 타임스탬프)
- Glamour로 렌더링된 콘텐츠 (마크다운)
- 댓글 수
- 긴 콘텐츠를 위한 스크롤 가능한 viewport

**작성 뷰**:
- 제목 입력 필드 (게시글용)
- 내용 textarea
- Tab으로 전환, Ctrl+S로 제출을 보여주는 도움말 텍스트
- 댓글 vs 게시글에 따라 다른 프롬프트

**댓글 뷰**:
- 작성자 및 타임스탬프와 함께 댓글 목록
- 중첩된 댓글을 위한 들여쓰기 (parentID > 0인 경우)
- 선택된 댓글 강조

## 기능 요약

### 핵심 기능
1. ✅ 다중 게시판 (설정 가능, 자동 생성)
2. ✅ 게시글 생성, 조회, 삭제
3. ✅ 게시글에 댓글 달기 (중첩 구조 지원)
4. ✅ 게시글 검색 (제목 및 내용)
5. ✅ 사용자 귀속 (작성자 추적)
6. ✅ 게시글 페이지네이션
7. ✅ 게시글 내용 마크다운 렌더링
8. ✅ 영구 저장소 (JSON 파일)
9. ✅ 선택적 암호화 (AES-256-GCM)
10. ✅ 선택적 인증 (사용자명/비밀번호)

### 주목할 만한 구현 세부사항
- **동시성 안전**: 모든 데이터 접근에 mutex 사용
- **원자적 쓰기**: 데이터 무결성을 위한 임시 파일 + 이름 변경 패턴
- **버전 관리 영속성**: 스키마 진화 지원
- **역호환 가능한 암호화**: 복호화 실패 시 평문 읽기
- **반응형 TUI**: 터미널 크기에 적응
- **Vim 스타일 네비게이션**: hjkl 키 지원
- **작성자 전용 삭제**: 사용자는 자신의 게시글만 삭제 가능
- **기본 익명 게시**: 작성자가 없는 게시글은 "anonymous"로 귀속

## 환경 변수

- `BBS_ENCRYPTION_KEY`: AES-256 암호화를 위한 64자 16진수 문자열 (32바이트)

## 커맨드라인 사용법

```bash
# 기본 사용법
go run ./cmd/bbs -addr :2323

# 인증 및 암호화 사용
export BBS_ENCRYPTION_KEY=$(openssl rand -hex 32)
go run ./cmd/bbs -addr :2323 -auth data/auth.json

# 사용자 정의 데이터 경로
go run ./cmd/bbs -boards data/boards.json -posts data/posts
```

## 접속

```bash
ssh guest@localhost -p 2323
```

## 테스트 전략

### 단위 테스트
- `internal/auth/auth_test.go` - 인증 로직
- `internal/bbs/bbs_test.go` - 핵심 BBS 작업
- `internal/bbs/persist_test.go` - 게시판 영속성
- `internal/bbs/persist_posts_test.go` - 게시글 영속성 및 암호화

### 통합 테스트
- `internal/bbs/bbs_integration_test.go` - 종단간 BBS 워크플로우

**테스트 패턴**:
- 여러 시나리오를 위한 테이블 주도 테스트
- 결정론적 타임스탬프를 위한 주입 가능한 시간 함수
- 인메모리 및 파일 기반 테스트 케이스
- 스레드 안전성을 위한 동시 접근 테스트

**테스트 실행**:
```bash
go test ./...
```

## 보안 고려사항

1. **암호화**:
   - AES-256-GCM은 인증된 암호화 제공
   - Nonce는 암호화마다 무작위로 생성
   - 키는 안전하게 저장 및 전송되어야 함

2. **인증**:
   - 비밀번호가 설정 파일에 평문으로 저장됨 (⚠️ 프로덕션 준비 안됨)
   - 로그인 시도에 대한 속도 제한 없음 (TODO)
   - SSH 레벨에서 인증 강제되지 않음 (애플리케이션 로직에 의존)

3. **입력 검증**:
   - 게시글 제목 필수 (빈 값 확인)
   - 게시판 이름 정규화
   - 삭제 시 작성자 일치 확인

4. **접근 제어**:
   - 사용자는 자신의 게시글만 삭제 가능
   - 관리자/모더레이터 역할 구현 안됨

## 향후 개선사항 (구현되지 않음)

- 비밀번호 해싱 (bcrypt/argon2)
- SSH 키 기반 인증
- 속도 제한
- 관리자 역할 및 모더레이션
- 다이렉트 메시지 시스템
- 게시판 권한
- 파일 첨부
- 사용자 프로필
- 읽음/읽지 않음 추적
- 이메일 알림

## 개발 가이드라인

다음은 `AGENTS.md`를 참조하세요:
- 코드 스타일 규칙
- 테스트 요구사항  
- 커밋 가이드라인
- 빌드 명령어

## 다른 LLM을 위한 구현 체크리스트

이 시스템을 처음부터 구현하려면:

1. **Go 프로젝트 설정**
   - Go 모듈 초기화: `go mod init ag`
   - `go.mod`에서 의존성 추가

2. **데이터 구조 구현** (`internal/bbs/bbs.go`)
   - Post, Comment, Board, BBS 구조체 정의
   - 핵심 메서드 구현: AddPost, ListPosts, DeletePost, AddComment

3. **영속성 구현** 
   - `internal/bbs/persist.go`: 게시판 목록을 위한 BoardFile
   - `internal/bbs/persist_posts.go`: 암호화를 포함한 PostFile

4. **인증 구현** (`internal/auth/`)
   - 설정 로딩
   - 비밀번호 확인 기능이 있는 Authenticator
   - 에코 없는 비밀번호 입력

5. **SSH 서버 구현** (`internal/server/`)
   - Wish 서버 설정
   - Bubbletea 미들웨어 통합
   - 세션 핸들러

6. **UI 구현** (`internal/ui/`)
   - Model, sessionState 정의
   - Init, Update, View 메서드 구현
   - 각 상태에 대한 뷰 함수 생성
   - 네비게이션을 위한 키보드 입력 처리

7. **main 구현** (`cmd/bbs/main.go`)
   - CLI 플래그 파싱
   - 컴포넌트 초기화
   - 서버 시작
   - 종료 처리

8. **테스트 추가**
   - 각 패키지에 대한 단위 테스트
   - 전체 워크플로우에 대한 통합 테스트

9. **데이터 디렉토리 구조 생성**
   - `data/boards.json`
   - `data/posts/`

## 이해해야 할 핵심 추상화

1. **Bubbletea MVU 패턴**: 전체 UI는 상태의 순수 함수
2. **Wish 미들웨어**: 조합 가능한 SSH 서버 미들웨어
3. **원자적 파일 쓰기**: 항상 임시 파일에 쓴 다음 이름 변경
4. **인터페이스 기반 저장소**: BoardListStore, PostStore가 테스트 허용
5. **세션 상태 머신**: viewBoards → viewPosts → viewPost → viewCompose

## 주의사항 및 중요 노트

- 터미널 앱에는 PTY가 필요함 (`s.Pty()` 확인)
- Lipgloss 렌더링은 비용이 클 수 있음; 가능하면 캐시
- Mutex 순서: 항상 교착 상태를 방지하기 위해 Board mu 전에 BBS mu
- Glamour 렌더링은 실패할 수 있음; 에러를 우아하게 처리
- 게시글 ID는 게시판별로 있으며, 전역이 아님
- 댓글은 Post에 내장되어 있으며, 별도 저장소가 아님
- 암호화 키는 정확히 32바이트여야 함
- 빈 게시판 이름은 "general"로 기본 설정됨
